package spatial

import (
	"sync"

	"github.com/krigsherre/aerocast/pkg/binary"
)

const GridStripes = 16

type Shard struct {
	mu      [GridStripes]sync.Mutex
	entries [GridStripes]map[EntityID]EntityState
}

type SpatialGrid struct {
	shards [ShardCount]Shard
}

func NewSpatialGrid() *SpatialGrid {
	g := &SpatialGrid{}
	for i := range g.shards {
		for s := 0; s < GridStripes; s++ {
			g.shards[i].entries[s] = make(map[EntityID]EntityState, 8)
		}
	}
	return g
}

func (g *SpatialGrid) Route(id EntityID, coord binary.CoordPacket) {
	key := ShardKey(coord.Lat, coord.Lng)
	s := &g.shards[key]
	stripe := id % GridStripes

	s.mu[stripe].Lock()
	s.entries[stripe][id] = EntityState{Coord: coord}
	s.mu[stripe].Unlock()
}

func (g *SpatialGrid) RouteWithPrevious(id EntityID, coord binary.CoordPacket, prevCoord *binary.CoordPacket) bool {
	newKey := ShardKey(coord.Lat, coord.Lng)
	stripe := id % GridStripes

	if prevCoord != nil {
		oldKey := ShardKey(prevCoord.Lat, prevCoord.Lng)
		if oldKey != newKey {
			oldShard := &g.shards[oldKey]
			oldShard.mu[stripe].Lock()
			delete(oldShard.entries[stripe], id)
			oldShard.mu[stripe].Unlock()
		}
	}

	s := &g.shards[newKey]
	s.mu[stripe].Lock()
	s.entries[stripe][id] = EntityState{Coord: coord}
	s.mu[stripe].Unlock()

	if prevCoord == nil {
		return true
	}
	return ShardKey(prevCoord.Lat, prevCoord.Lng) != newKey
}

func (g *SpatialGrid) Remove(id EntityID, lat, lng float64) {
	key := ShardKey(lat, lng)
	s := &g.shards[key]
	stripe := id % GridStripes
	s.mu[stripe].Lock()
	delete(s.entries[stripe], id)
	s.mu[stripe].Unlock()
}

func (g *SpatialGrid) Get(id EntityID) (EntityState, bool) {
	stripe := id % GridStripes
	for i := range g.shards {
		s := &g.shards[i]
		s.mu[stripe].Lock()
		state, ok := s.entries[stripe][id]
		s.mu[stripe].Unlock()
		if ok {
			return state, true
		}
	}
	return EntityState{}, false
}

func (g *SpatialGrid) QueryCircle(lat, lng, radiusM float64) []EntityID {
	shardKeys := ShardsForRadius(lat, lng, radiusM)
	var result []EntityID

	for _, key := range shardKeys {
		s := &g.shards[key]
		for stripe := 0; stripe < GridStripes; stripe++ {
			s.mu[stripe].Lock()
			for id, state := range s.entries[stripe] {
				dist := HaversineM(lat, lng, state.Coord.Lat, state.Coord.Lng)
				if dist <= radiusM {
					result = append(result, id)
				}
			}
			s.mu[stripe].Unlock()
		}
	}

	return result
}

func (g *SpatialGrid) QueryCircleWithCoords(lat, lng, radiusM float64) []EntityState {
	shardKeys := ShardsForRadius(lat, lng, radiusM)
	var result []EntityState

	for _, key := range shardKeys {
		s := &g.shards[key]
		for stripe := 0; stripe < GridStripes; stripe++ {
			s.mu[stripe].Lock()
			for _, state := range s.entries[stripe] {
				dist := HaversineM(lat, lng, state.Coord.Lat, state.Coord.Lng)
				if dist <= radiusM {
					result = append(result, state)
				}
			}
			s.mu[stripe].Unlock()
		}
	}

	return result
}

func (g *SpatialGrid) ShardStats() [ShardCount]int {
	var counts [ShardCount]int
	for i := range g.shards {
		s := &g.shards[i]
		total := 0
		for stripe := 0; stripe < GridStripes; stripe++ {
			s.mu[stripe].Lock()
			total += len(s.entries[stripe])
			s.mu[stripe].Unlock()
		}
		counts[i] = total
	}
	return counts
}

func (g *SpatialGrid) Count() int {
	total := 0
	for i := range g.shards {
		s := &g.shards[i]
		for stripe := 0; stripe < GridStripes; stripe++ {
			s.mu[stripe].Lock()
			total += len(s.entries[stripe])
			s.mu[stripe].Unlock()
		}
	}
	return total
}

func (g *SpatialGrid) EgressRecords(shardIndex uint8) []binary.EgressRecord {
	s := &g.shards[shardIndex]
	var totalLen int
	for stripe := 0; stripe < GridStripes; stripe++ {
		s.mu[stripe].Lock()
		totalLen += len(s.entries[stripe])
		s.mu[stripe].Unlock()
	}

	if totalLen == 0 {
		return nil
	}
	
	records := make([]binary.EgressRecord, 0, totalLen)
	for stripe := 0; stripe < GridStripes; stripe++ {
		s.mu[stripe].Lock()
		for id, state := range s.entries[stripe] {
			records = append(records, binary.EgressRecord{
				EntityID: uint32(id),
				Lat:      state.Coord.Lat,
				Lng:      state.Coord.Lng,
			})
		}
		s.mu[stripe].Unlock()
	}
	
	return records
}
