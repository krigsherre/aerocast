package geofence

import "time"

type EventType uint8

const (
	EventEnter EventType = iota
	EventExit
)

func (e EventType) String() string {
	if e == EventEnter {
		return "entered"
	}
	return "exited"
}

type FenceEvent struct {
	EntityID  uint32
	FenceName string
	Type      EventType
	Lat       float64
	Lng       float64
	Timestamp time.Time
}

type Callback func(event FenceEvent)
