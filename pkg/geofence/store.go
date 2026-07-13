package geofence

import (
	"sync"

	"github.com/krigsherre/aerocast/pkg/spatial"
)

type Store struct {
	mu     sync.RWMutex
	fences []Fence
	index  [256][]int
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Register(f Fence) {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := len(s.fences)
	s.fences = append(s.fences, f)

	clat, clng := f.Center()
	radius := f.BoundingRadius()
	shardKeys := spatial.ShardsForRadius(clat, clng, radius)
	for _, key := range shardKeys {
		s.index[key] = append(s.index[key], idx)
	}
}

func (s *Store) FencesNear(lat, lng float64) []Fence {
	key := spatial.ShardKey(lat, lng)

	s.mu.RLock()
	indices := s.index[key]
	if len(indices) == 0 {
		s.mu.RUnlock()
		return nil
	}

	result := make([]Fence, len(indices))
	for i, idx := range indices {
		result[i] = s.fences[idx]
	}
	s.mu.RUnlock()

	return result
}

func (s *Store) All() []Fence {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Fence, len(s.fences))
	copy(out, s.fences)
	return out
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.fences)
}
