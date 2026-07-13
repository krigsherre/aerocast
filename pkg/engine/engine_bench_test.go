package engine

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/krigsherre/aerocast/pkg/binary"
	udpTransport "github.com/krigsherre/aerocast/pkg/transport/udp"
	"go.uber.org/zap"
)

func BenchmarkEnginePipeline(b *testing.B) {
	cfg := DefaultConfig()
	cfg.DisableUDP = true
	cfg.DisableWS = true
	cfg.Logger = zap.NewNop()

	eng, err := NewEngine(cfg)
	if err != nil {
		b.Fatalf("failed to create engine: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go eng.Run(ctx)
	time.Sleep(10 * time.Millisecond)

	const numPackets = 100000
	packets := make([]udpTransport.IngressPacket, numPackets)
	for i := 0; i < numPackets; i++ {
		packets[i] = udpTransport.IngressPacket{
			EntityID: uint32(i),
			Coord: binary.CoordPacket{
				Lat: rand.Float64()*180 - 90,
				Lng: rand.Float64()*360 - 180,
			},
			Timestamp: time.Now(),
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			eng.ProcessPacket(packets[i%numPackets])
			i++
		}
	})
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "pkts/sec")
}

func BenchmarkEnginePipeline_WithSpam(b *testing.B) {
	cfg := DefaultConfig()
	cfg.DisableUDP = true
	cfg.DisableWS = true
	cfg.Logger = zap.NewNop()

	eng, err := NewEngine(cfg)
	if err != nil {
		b.Fatalf("failed to create engine: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go eng.Run(ctx)
	time.Sleep(10 * time.Millisecond)

	const numPackets = 100000
	packets := make([]udpTransport.IngressPacket, numPackets)
	for i := 0; i < numPackets; i++ {
		id := uint32(i % 10)
		packets[i] = udpTransport.IngressPacket{
			EntityID: id,
			Coord: binary.CoordPacket{
				Lat: 37.0 + float64(id),
				Lng: -122.0 + float64(id),
			},
			Timestamp: time.Now(),
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			eng.ProcessPacket(packets[i%numPackets])
			i++
		}
	})
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "pkts/sec")
}

func BenchmarkEnginePipeline_Realistic(b *testing.B) {
	cfg := DefaultConfig()
	cfg.DisableUDP = true
	cfg.DisableWS = true
	cfg.Logger = zap.NewNop()

	eng, err := NewEngine(cfg)
	if err != nil {
		b.Fatalf("failed to create engine: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < 100; i++ {
		eng.OnSubscribe(uint64(i), rand.Float64()*180-90, rand.Float64()*360-180, 5000)
	}

	go eng.Run(ctx)
	time.Sleep(10 * time.Millisecond)

	const numPackets = 100000
	packets := make([]udpTransport.IngressPacket, numPackets)
	for i := 0; i < numPackets; i++ {
		packets[i] = udpTransport.IngressPacket{
			EntityID: uint32(i),
			Coord: binary.CoordPacket{
				Lat: rand.Float64()*180 - 90,
				Lng: rand.Float64()*360 - 180,
			},
			Timestamp: time.Now(),
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			eng.ProcessPacket(packets[i%numPackets])
			i++
		}
	})
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "pkts/sec")
}
