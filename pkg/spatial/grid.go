package spatial

import (
	"sync"

	"github.com/krigsherre/aerocast/pkg/binary"
)

type Shard struct {
	mu      sync.Mutex
	entries map[EntityID]EntityState
	_pad    [40]byte
}

type SpatialGrid struct {
	shards [ShardCount]Shard
}

func NewSpatialGrid() *SpatialGrid {
	g := &SpatialGrid{}
	for i := range g.shards {
		g.shards[i].entries = make(map[EntityID]EntityState, 64)
	}
	return g
}

func (g *SpatialGrid) Route(id EntityID, coord binary.CoordPacket) {
	key := ShardKey(coord.Lat, coord.Lng)
	s := &g.shards[key]

	s.mu.Lock()
	s.entries[id] = EntityState{Coord: coord}
	s.mu.Unlock()
}

func (g *SpatialGrid) RouteWithPrevious(id EntityID, coord binary.CoordPacket, prevCoord *binary.CoordPacket) bool {
	newKey := ShardKey(coord.Lat, coord.Lng)

	if prevCoord != nil {
		oldKey := ShardKey(prevCoord.Lat, prevCoord.Lng)
		if oldKey != newKey {
			oldShard := &g.shards[oldKey]
			oldShard.mu.Lock()
			delete(oldShard.entries, id)
			oldShard.mu.Unlock()
		}
	}

	s := &g.shards[newKey]
	s.mu.Lock()
	s.entries[id] = EntityState{Coord: coord}
	s.mu.Unlock()

	if prevCoord == nil {
		return true
	}
	return ShardKey(prevCoord.Lat, prevCoord.Lng) != newKey
}

func (g *SpatialGrid) Remove(id EntityID, lat, lng float64) {
	key := ShardKey(lat, lng)
	s := &g.shards[key]
	s.mu.Lock()
	delete(s.entries, id)
	s.mu.Unlock()
}

func (g *SpatialGrid) Get(id EntityID) (EntityState, bool) {
	for i := range g.shards {
		s := &g.shards[i]
		s.mu.Lock()
		state, ok := s.entries[id]
		s.mu.Unlock()
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
		s.mu.Lock()
		for id, state := range s.entries {
			dist := HaversineM(lat, lng, state.Coord.Lat, state.Coord.Lng)
			if dist <= radiusM {
				result = append(result, id)
			}
		}
		s.mu.Unlock()
	}

	return result
}

func (g *SpatialGrid) QueryCircleWithCoords(lat, lng, radiusM float64) []EntityState {
	shardKeys := ShardsForRadius(lat, lng, radiusM)
	var result []EntityState

	for _, key := range shardKeys {
		s := &g.shards[key]
		s.mu.Lock()
		for _, state := range s.entries {
			dist := HaversineM(lat, lng, state.Coord.Lat, state.Coord.Lng)
			if dist <= radiusM {
				result = append(result, state)
			}
		}
		s.mu.Unlock()
	}

	return result
}

func (g *SpatialGrid) ShardStats() [ShardCount]int {
	var counts [ShardCount]int
	for i := range g.shards {
		g.shards[i].mu.Lock()
		counts[i] = len(g.shards[i].entries)
		g.shards[i].mu.Unlock()
	}
	return counts
}

func (g *SpatialGrid) Count() int {
	total := 0
	for i := range g.shards {
		g.shards[i].mu.Lock()
		total += len(g.shards[i].entries)
		g.shards[i].mu.Unlock()
	}
	return total
}

func (g *SpatialGrid) EgressRecords(shardIndex uint8) []binary.EgressRecord {
	s := &g.shards[shardIndex]
	s.mu.Lock()
	if len(s.entries) == 0 {
		s.mu.Unlock()
		return nil
	}
	
	records := make([]binary.EgressRecord, 0, len(s.entries))
	for id, state := range s.entries {
		records = append(records, binary.EgressRecord{
			EntityID: uint32(id),
			Lat:      state.Coord.Lat,
			Lng:      state.Coord.Lng,
		})
	}
	s.mu.Unlock()
	
	return records
}
