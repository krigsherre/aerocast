package fanout

import "github.com/krigsherre/aerocast/pkg/spatial"

type SubscriberID = uint64

type Subscription struct {
	ID        SubscriberID
	CenterLat float64
	CenterLng float64
	RadiusM   float64
	Cells     []uint8
	Following map[spatial.EntityID]struct{}
}

func (s *Subscription) HasEntityFollow() bool {
	return len(s.Following) > 0
}

func (s *Subscription) WantsEntity(id spatial.EntityID) bool {
	if s.Following == nil {
		return true
	}
	_, ok := s.Following[id]
	return ok
}
