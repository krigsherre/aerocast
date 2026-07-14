package main

import (
	"bufio"
	"context"
	"net"
	"testing"
	"time"

	"github.com/krigsherre/aerocast"
)

func BenchmarkAerocastInternal(b *testing.B) {
	cfg := aerocast.DefaultConfig()
	cfg.DisableUDP = true
	cfg.DisableWS = true

	engine, err := aerocast.New(cfg)
	if err != nil {
		b.Fatalf("failed to create engine: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = engine.Run(ctx)
	}()
	time.Sleep(10 * time.Millisecond)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := uint32(1)
		for pb.Next() {
			engine.Publish(i, 37.7749, -122.4194)
			i++
		}
	})
}

func BenchmarkRedisGeo(b *testing.B) {
	conn, err := net.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		b.Skip("Redis is not running on 127.0.0.1:6379, skipping...")
	}
	defer conn.Close()

	go func(c net.Conn) {
		tmp := make([]byte, 4096)
		for {
			_, err := c.Read(tmp)
			if err != nil {
				return
			}
		}
	}(conn)

	writer := bufio.NewWriterSize(conn, 65536)
	buf := make([]byte, 0, 128)

	b.ResetTimer()
	b.ReportAllocs()

	i := uint32(1)
	for k := 0; k < b.N; k++ {
		_ = writeRedisCmd(writer, buf, -122.4194, 37.7749, i)
		i++
	}
	_ = writer.Flush()
}

func BenchmarkTile38SET(b *testing.B) {
	conn, err := net.Dial("tcp", "127.0.0.1:9851")
	if err != nil {
		b.Skip("Tile38 is not running on 127.0.0.1:9851, skipping...")
	}
	defer conn.Close()

	go func(c net.Conn) {
		tmp := make([]byte, 4096)
		for {
			_, err := c.Read(tmp)
			if err != nil {
				return
			}
		}
	}(conn)

	writer := bufio.NewWriterSize(conn, 65536)
	buf := make([]byte, 0, 128)

	b.ResetTimer()
	b.ReportAllocs()

	i := uint32(1)
	for k := 0; k < b.N; k++ {
		_ = writeTile38Cmd(writer, buf, i, 37.7749, -122.4194)
		i++
	}
	_ = writer.Flush()
}
