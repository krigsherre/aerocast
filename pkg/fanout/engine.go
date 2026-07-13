package fanout

import (
	"sync"
	"sync/atomic"

	"github.com/krigsherre/aerocast/pkg/binary"
	"github.com/krigsherre/aerocast/pkg/spatial"
)

type Engine struct {
	mu            sync.RWMutex
	subs          map[SubscriberID]*Subscription
	cellIndex     [spatial.ShardCount][]SubscriberID
	channels      *ChannelManager
	egressPool    sync.Pool
	droppedFrames atomic.Uint64
}

func NewEngine() *Engine {
	return &Engine{
		subs:     make(map[SubscriberID]*Subscription, 1024),
		channels: NewChannelManager(),
		egressPool: sync.Pool{
			New: func() any { return binary.GetEgressBuffer() },
		},
	}
}

func (e *Engine) Channels() *ChannelManager {
	return e.channels
}

func (e *Engine) Subscribe(id SubscriberID, lat, lng, radiusM float64) {
	cells := spatial.ShardsForRadius(lat, lng, radiusM)

	sub := &Subscription{
		ID:        id,
		CenterLat: lat,
		CenterLng: lng,
		RadiusM:   radiusM,
		Cells:     cells,
	}

	e.mu.Lock()
	e.subs[id] = sub
	for _, cell := range cells {
		e.cellIndex[cell] = append(e.cellIndex[cell], id)
	}
	e.mu.Unlock()

	e.channels.Create(id)
}

func (e *Engine) Unsubscribe(id SubscriberID) {
	e.mu.Lock()
	sub, ok := e.subs[id]
	if !ok {
		e.mu.Unlock()
		return
	}

	for _, cell := range sub.Cells {
		subs := e.cellIndex[cell]
		for i, sid := range subs {
			if sid == id {
				e.cellIndex[cell] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
	}

	delete(e.subs, id)
	e.mu.Unlock()

	e.channels.Remove(id)
}

func (e *Engine) Follow(subID SubscriberID, entityID spatial.EntityID) {
	e.mu.Lock()
	if sub, ok := e.subs[subID]; ok {
		if sub.Following == nil {
			sub.Following = make(map[spatial.EntityID]struct{}, 8)
		}
		sub.Following[entityID] = struct{}{}
	}
	e.mu.Unlock()
}

func (e *Engine) Unfollow(subID SubscriberID, entityID spatial.EntityID) {
	e.mu.Lock()
	if sub, ok := e.subs[subID]; ok && sub.Following != nil {
		delete(sub.Following, entityID)
	}
	e.mu.Unlock()
}

func (e *Engine) Broadcast(cell uint8, records []binary.EgressRecord) int {
	e.mu.RLock()
	subIDs := e.cellIndex[cell]
	if len(subIDs) == 0 {
		e.mu.RUnlock()
		return 0
	}

	subs := make([]*Subscription, len(subIDs))
	for i, id := range subIDs {
		subs[i] = e.subs[id]
	}
	e.mu.RUnlock()

	sent := 0
	for _, sub := range subs {
		buf := binary.GetEgressBuffer()

		for _, rec := range records {
			if sub.WantsEntity(rec.EntityID) {
				buf.Append(rec)
			}
		}

		if buf.Len() == 0 {
			binary.PutEgressBuffer(buf)
			continue
		}

		frame := make([]byte, len(buf.Bytes()))
		copy(frame, buf.Bytes())
		binary.PutEgressBuffer(buf)

		ch := e.channels.Get(sub.ID)
		if ch != nil {
			select {
			case ch <- frame:
				sent++
			default:
				e.droppedFrames.Add(1)
			}
		}
	}

	return sent
}

func (e *Engine) DroppedFrames() uint64 {
	return e.droppedFrames.Load()
}

func (e *Engine) BroadcastMulti(records []binary.EgressRecord) int {
	cellRecords := make(map[uint8][]binary.EgressRecord, 8)
	for _, rec := range records {
		cell := spatial.ShardKey(rec.Lat, rec.Lng)
		cellRecords[cell] = append(cellRecords[cell], rec)
	}

	total := 0
	for cell, recs := range cellRecords {
		total += e.Broadcast(cell, recs)
	}
	return total
}

func (e *Engine) SubscriberCount() int {
	e.mu.RLock()
	n := len(e.subs)
	e.mu.RUnlock()
	return n
}

func (e *Engine) SubscribersInCell(cell uint8) int {
	e.mu.RLock()
	n := len(e.cellIndex[cell])
	e.mu.RUnlock()
	return n
}
