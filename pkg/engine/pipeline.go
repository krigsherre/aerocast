package engine

import (
	"context"
	"time"

	"github.com/krigsherre/aerocast/pkg/binary"
	"github.com/krigsherre/aerocast/pkg/spatial"
	udpTransport "github.com/krigsherre/aerocast/pkg/transport/udp"
)

func (e *Engine) pipelineWorker(ctx context.Context) {
	defer e.wg.Done()
	if e.udp == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case pkt, ok := <-e.udp.Packets():
			if !ok {
				return
			}
			e.ProcessPacket(pkt)
		}
	}
}

func (e *Engine) ProcessPacket(pkt udpTransport.IngressPacket) {
	id := pkt.EntityID
	coord := pkt.Coord

	if e.cms.addAndEstimate(id) > 50 {
		e.packetsDropped.Add(1)
		return
	}

	shardIdx := id % prevPosShards
	shard := &e.prevPosShards[shardIdx]

	shard.mu.RLock()
	prev := shard.m[spatial.EntityID(id)]
	shard.mu.RUnlock()

	if prev != nil && prev.Lat == coord.Lat && prev.Lng == coord.Lng {
		e.packetsDropped.Add(1)
		return
	}

	e.grid.RouteWithPrevious(spatial.EntityID(id), coord, prev)

	events := e.evaluator.EvaluateMove(spatial.EntityID(id), prev, coord)
	e.eventsFired.Add(uint64(len(events)))

	shard.mu.Lock()
	p := shard.m[spatial.EntityID(id)]
	if p == nil {
		p = &binary.CoordPacket{}
		shard.m[spatial.EntityID(id)] = p
	}
	*p = coord
	shard.mu.Unlock()

	gridShard := spatial.ShardKey(coord.Lat, coord.Lng)
	e.shardHLL[gridShard].Add(id)

	e.packetsRouted.Add(1)
}

func (e *Engine) tickLoop(ctx context.Context) {
	defer e.wg.Done()

	ticker := time.NewTicker(e.cfg.TickRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.cms.reset()
			e.broadcastTick()
		}
	}
}

func (e *Engine) broadcastTick() {
	for cell := 0; cell < spatial.ShardCount; cell++ {
		subCount := e.fanout.SubscribersInCell(uint8(cell))
		if subCount == 0 {
			continue
		}

		records := e.grid.EgressRecords(uint8(cell))
		if len(records) > 0 {
			e.fanout.Broadcast(uint8(cell), records)
		}
	}
}
