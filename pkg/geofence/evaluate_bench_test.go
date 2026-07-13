package geofence

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/krigsherre/aerocast/pkg/binary"
)

func BenchmarkEvaluateMove(b *testing.B) {
	for _, fenceCount := range []int{0, 10, 100, 1000} {
		fenceCount := fenceCount
		name := fmt.Sprintf("Fences%d", fenceCount)

		b.Run(name, func(b *testing.B) {
			store := NewStore()
			for i := 0; i < fenceCount; i++ {
				store.Register(&Circle{
					NameStr:   fmt.Sprintf("fence-%d", i),
					CenterLat: -90 + rand.Float64()*180,
					CenterLng: -180 + rand.Float64()*360,
					RadiusM:   500 + rand.Float64()*5000,
				})
			}

			eval := NewEvaluator(store)
			from := &binary.CoordPacket{Lat: 37.7749, Lng: -122.4194}
			to := binary.CoordPacket{Lat: 37.7750, Lng: -122.4195}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				to.Lat = 37.7749 + math.Sin(float64(i))*0.001
				to.Lng = -122.4194 + math.Cos(float64(i))*0.001
				eval.EvaluateMove(uint32(i), from, to)
			}
		})
	}
}

func BenchmarkEvaluateMoveNoFencesNearby(b *testing.B) {
	store := NewStore()
	store.Register(&Circle{
		NameStr:   "remote-fence",
		CenterLat: -33.8688,
		CenterLng: 151.2093,
		RadiusM:   1000,
	})

	eval := NewEvaluator(store)
	from := &binary.CoordPacket{Lat: 37.7749, Lng: -122.4194}
	to := binary.CoordPacket{Lat: 37.7750, Lng: -122.4195}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		eval.EvaluateMove(42, from, to)
	}
}
