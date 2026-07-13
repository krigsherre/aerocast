package geofence

import (
	"sync"
	"time"

	"github.com/krigsherre/aerocast/pkg/binary"
)

type Evaluator struct {
	store          *Store
	callbacks      map[string][]Callback
	GlobalCallback func(FenceEvent)
	mu             sync.RWMutex
}

func NewEvaluator(store *Store) *Evaluator {
	return &Evaluator{
		store:     store,
		callbacks: make(map[string][]Callback),
	}
}

func (e *Evaluator) OnEvent(fenceName string, fn Callback) {
	e.mu.Lock()
	e.callbacks[fenceName] = append(e.callbacks[fenceName], fn)
	e.mu.Unlock()
}

func (e *Evaluator) EvaluateMove(
	entityID uint32,
	from *binary.CoordPacket,
	to binary.CoordPacket,
) []FenceEvent {
	nearTo := e.store.FencesNear(to.Lat, to.Lng)

	if len(nearTo) == 0 && from == nil {
		return nil
	}

	var events []FenceEvent
	now := time.Now()

	for _, f := range nearTo {
		nowInside := f.Contains(to.Lat, to.Lng)

		if from == nil {
			if nowInside {
				events = append(events, FenceEvent{
					EntityID:  entityID,
					FenceName: f.Name(),
					Type:      EventEnter,
					Lat:       to.Lat,
					Lng:       to.Lng,
					Timestamp: now,
				})
			}
			continue
		}

		wasInside := f.Contains(from.Lat, from.Lng)

		if !wasInside && nowInside {
			events = append(events, FenceEvent{
				EntityID:  entityID,
				FenceName: f.Name(),
				Type:      EventEnter,
				Lat:       to.Lat,
				Lng:       to.Lng,
				Timestamp: now,
			})
		} else if wasInside && !nowInside {
			events = append(events, FenceEvent{
				EntityID:  entityID,
				FenceName: f.Name(),
				Type:      EventExit,
				Lat:       to.Lat,
				Lng:       to.Lng,
				Timestamp: now,
			})
		}
	}

	if len(events) > 0 {
		e.fireCallbacks(events)
	}

	return events
}

func (e *Evaluator) fireCallbacks(events []FenceEvent) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, ev := range events {
		if e.GlobalCallback != nil {
			e.GlobalCallback(ev)
		}
		if fns, ok := e.callbacks[ev.FenceName]; ok {
			for _, fn := range fns {
				fn(ev)
			}
		}
	}
}
