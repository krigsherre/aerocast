package spatial

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/krigsherre/aerocast/pkg/binary"
)

func BenchmarkSpatialRouting(b *testing.B) {
	grid := NewSpatialGrid()
	coords := generateTestCoords(10_000)

	b.Run("ShardedGrid", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := i % len(coords)
			grid.Route(EntityID(i), coords[idx])
		}
	})

	b.Run("GlobalMutex", func(b *testing.B) {
		ref := NewGlobalMapRouter()
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			idx := i % len(coords)
			ref.Route(EntityID(i), coords[idx])
		}
	})
}

func BenchmarkConcurrentRouting(b *testing.B) {
	for _, parallelism := range []int{1, 2, 4, 8, 16, 32} {
		parallelism := parallelism
		name := fmt.Sprintf("ShardedGrid/P%d", parallelism)

		b.Run(name, func(b *testing.B) {
			grid := NewSpatialGrid()
			coords := generateTestCoords(10_000)
			b.SetParallelism(parallelism)
			b.ReportAllocs()
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					idx := i % len(coords)
					grid.Route(EntityID(i), coords[idx])
					i++
				}
			})
		})
	}

	for _, parallelism := range []int{1, 2, 4, 8, 16, 32} {
		parallelism := parallelism
		name := fmt.Sprintf("GlobalMutex/P%d", parallelism)

		b.Run(name, func(b *testing.B) {
			ref := NewGlobalMapRouter()
			coords := generateTestCoords(10_000)
			b.SetParallelism(parallelism)
			b.ReportAllocs()
			b.ResetTimer()

			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					idx := i % len(coords)
					ref.Route(EntityID(i), coords[idx])
					i++
				}
			})
		})
	}
}

func BenchmarkQueryCircle(b *testing.B) {
	grid := NewSpatialGrid()
	for i := 0; i < 100_000; i++ {
		lat := -90.0 + rand.Float64()*180.0
		lng := -180.0 + rand.Float64()*360.0
		grid.Route(EntityID(i), binary.CoordPacket{Lat: lat, Lng: lng})
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = grid.QueryCircle(37.7749, -122.4194, 5000.0)
	}
}

func generateTestCoords(n int) []binary.CoordPacket {
	coords := make([]binary.CoordPacket, n)
	for i := range coords {
		coords[i] = binary.CoordPacket{
			Lat: -90.0 + rand.Float64()*180.0,
			Lng: -180.0 + rand.Float64()*360.0,
		}
	}
	return coords
}
