package spatial

import (
	"testing"
)

func TestShardKeyDistribution(t *testing.T) {
	coords := []struct{ lat, lng float64 }{
		{37.7749, -122.4194},
		{40.7128, -74.0060},
		{51.5074, -0.1278},
		{35.6762, 139.6503},
		{-33.8688, 151.2093},
		{55.7558, 37.6173},
		{19.4326, -99.1332},
		{-23.5505, -46.6333},
		{1.3521, 103.8198},
		{28.6139, 77.2090},
	}

	counts := make(map[uint8]int)
	for _, c := range coords {
		key := ShardKey(c.lat, c.lng)
		counts[key]++
	}

	if len(counts) != len(coords) {
		t.Logf("collision detected: %d unique shards for %d coords", len(counts), len(coords))
	}
}

func TestHaversineM(t *testing.T) {
	tests := []struct {
		name       string
		lat1       float64
		lng1       float64
		lat2       float64
		lng2       float64
		wantM      float64
		toleranceM float64
	}{
		{
			name: "same point",
			lat1: 37.7749, lng1: -122.4194,
			lat2: 37.7749, lng2: -122.4194,
			wantM:      0,
			toleranceM: 0.01,
		},
		{
			name: "sf to oakland ~13km",
			lat1: 37.7749, lng1: -122.4194,
			lat2: 37.8044, lng2: -122.2712,
			wantM:      13_200,
			toleranceM: 500,
		},
		{
			name: "new york to london ~5570km",
			lat1: 40.7128, lng1: -74.0060,
			lat2: 51.5074, lng2: -0.1278,
			wantM:      5_570_000,
			toleranceM: 50_000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HaversineM(tt.lat1, tt.lng1, tt.lat2, tt.lng2)
			diff := got - tt.wantM
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.toleranceM {
				t.Errorf("HaversineM() = %.1f m, want %.1f ± %.1f m", got, tt.wantM, tt.toleranceM)
			}
		})
	}
}
