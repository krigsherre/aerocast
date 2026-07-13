package fanout

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/krigsherre/aerocast/pkg/binary"
	"github.com/krigsherre/aerocast/pkg/spatial"
)

func BenchmarkBroadcast(b *testing.B) {
	for _, subCount := range []int{100, 1_000, 10_000, 100_000} {
		subCount := subCount
		name := fmt.Sprintf("Subs%d", subCount)

		b.Run(name, func(b *testing.B) {
			engine := NewEngine()

			for i := 0; i < subCount; i++ {
				lat := 37.7 + rand.Float64()*0.1
				lng := -122.5 + rand.Float64()*0.1
				engine.Subscribe(SubscriberID(i), lat, lng, 5000)
			}

			cell := spatial.ShardKey(37.7749, -122.4194)
			records := []binary.EgressRecord{
				{EntityID: 1, Lat: 37.7749, Lng: -122.4194},
				{EntityID: 2, Lat: 37.7750, Lng: -122.4195},
				{EntityID: 3, Lat: 37.7751, Lng: -122.4196},
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				engine.Broadcast(cell, records)
			}
		})
	}
}

func BenchmarkBroadcastWithEntityFollow(b *testing.B) {
	engine := NewEngine()

	for i := 0; i < 10_000; i++ {
		lat := 37.7 + rand.Float64()*0.1
		lng := -122.5 + rand.Float64()*0.1
		sid := SubscriberID(i)
		engine.Subscribe(sid, lat, lng, 5000)

		for j := 0; j < 5; j++ {
			engine.Follow(sid, spatial.EntityID(j*100))
		}
	}

	cell := spatial.ShardKey(37.7749, -122.4194)
	records := []binary.EgressRecord{
		{EntityID: 0, Lat: 37.7749, Lng: -122.4194},
		{EntityID: 500, Lat: 37.7750, Lng: -122.4195},
		{EntityID: 99999, Lat: 37.7751, Lng: -122.4196},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		engine.Broadcast(cell, records)
	}
}

func BenchmarkSubscribeUnsubscribe(b *testing.B) {
	engine := NewEngine()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sid := SubscriberID(i)
		engine.Subscribe(sid, 37.7749, -122.4194, 5000)
		engine.Unsubscribe(sid)
	}
}
