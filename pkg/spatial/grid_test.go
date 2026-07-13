package spatial

import (
	"testing"

	"github.com/krigsherre/aerocast/pkg/binary"
)

func TestGridRouteAndQuery(t *testing.T) {
	grid := NewSpatialGrid()

	entities := []struct {
		id  EntityID
		lat float64
		lng float64
	}{
		{1, 37.7749, -122.4194},
		{2, 37.7750, -122.4195},
		{3, 37.7800, -122.4100},
		{4, 40.7128, -74.0060},
	}

	for _, e := range entities {
		grid.Route(e.id, binary.CoordPacket{Lat: e.lat, Lng: e.lng})
	}

	results := grid.QueryCircle(37.7749, -122.4194, 500.0)

	found := make(map[EntityID]bool)
	for _, id := range results {
		found[id] = true
	}

	if !found[1] {
		t.Error("expected entity 1 (exact match)")
	}
	if !found[2] {
		t.Error("expected entity 2 (11m away)")
	}
	if found[3] {
		t.Error("entity 3 should NOT be in 500m radius (700m away)")
	}
	if found[4] {
		t.Error("entity 4 (NYC) should NOT be in SF radius")
	}
}

func TestGridRouteUpdate(t *testing.T) {
	grid := NewSpatialGrid()

	grid.Route(1, binary.CoordPacket{Lat: 37.7749, Lng: -122.4194})
	grid.Route(1, binary.CoordPacket{Lat: 37.7800, Lng: -122.4100})

	results := grid.QueryCircle(37.7800, -122.4100, 200.0)
	if len(results) != 1 || results[0] != 1 {
		t.Errorf("expected entity 1 at new position, got %v", results)
	}

	oldResults := grid.QueryCircle(37.7749, -122.4194, 50.0)
	if len(oldResults) != 0 {
		t.Errorf("expected 0 entities at old position, got %d", len(oldResults))
	}
}
