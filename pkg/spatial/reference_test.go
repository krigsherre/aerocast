package spatial

import (
	"sync"

	"github.com/krigsherre/aerocast/pkg/binary"
)

type GlobalMapRouter struct {
	mu       sync.RWMutex
	entities map[EntityID]binary.CoordPacket
}

func NewGlobalMapRouter() *GlobalMapRouter {
	return &GlobalMapRouter{
		entities: make(map[EntityID]binary.CoordPacket, 1024),
	}
}

func (r *GlobalMapRouter) Route(id EntityID, coord binary.CoordPacket) {
	r.mu.Lock()
	r.entities[id] = coord
	r.mu.Unlock()
}
