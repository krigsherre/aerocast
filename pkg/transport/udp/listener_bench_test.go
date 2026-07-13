package udp

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/krigsherre/aerocast/pkg/binary"
)

func BenchmarkProcessPacket(b *testing.B) {
	l := &Listener{
		cfg:  DefaultConfig(),
		out:  make(chan IngressPacket, 65536),
		done: make(chan struct{}),
	}

	data := make([]byte, 20)
	binary.LittleEndian.PutUint32(data[0:4], 42)
	binary.LittleEndian.PutUint64(data[4:12], binary.LittleEndian.Uint64(func() []byte {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, 4638487916524879872)
		return b
	}()))
	binary.LittleEndian.PutUint64(data[12:20], binary.LittleEndian.Uint64(func() []byte {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, 4633664324253351936)
		return b
	}()))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = l.processPacket(data)
		select {
		case <-l.out:
		default:
		}
	}
}

func BenchmarkEndToEndUDP(b *testing.B) {
	cfg := DefaultConfig()
	cfg.ListenAddr = "127.0.0.1:0"
	l, err := NewListener(cfg)
	if err != nil {
		b.Fatal(err)
	}
	defer l.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go l.Run(ctx)

	time.Sleep(10 * time.Millisecond)

	addr := l.conn.LocalAddr().String()
	conn, err := net.Dial("udp", addr)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	data := make([]byte, 20)
	binary.LittleEndian.PutUint32(data[0:4], 1)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		binary.LittleEndian.PutUint32(data[0:4], uint32(i))
		_, _ = conn.Write(data)
	}
}
