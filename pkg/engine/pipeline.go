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

	if e.bloom.checkAndAdd(id) {
		e.packetsDropped.Add(1)
		return
	}

	e.prevMu.RLock()
	prev := e.prevPos[spatial.EntityID(id)]
	e.prevMu.RUnlock()

	e.grid.Route(spatial.EntityID(id), coord)

	events := e.evaluator.EvaluateMove(spatial.EntityID(id), prev, coord)
	e.eventsFired.Add(uint64(len(events)))

	e.prevMu.Lock()
	if prev == nil {
		e.prevPos[spatial.EntityID(id)] = &binary.CoordPacket{}
		prev = e.prevPos[spatial.EntityID(id)]
	}
	*prev = coord
	e.prevMu.Unlock()

	shard := spatial.ShardKey(coord.Lat, coord.Lng)
	e.shardHLL[shard].Add(id)

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
			e.bloom.reset()
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
