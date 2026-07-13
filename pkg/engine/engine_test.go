package engine

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/krigsherre/aerocast/pkg/binary"
	"github.com/krigsherre/aerocast/pkg/spatial"
)

func TestEnginePacketRouting(t *testing.T) {
	cfg := DefaultConfig()
	cfg.WS.ListenAddr = "127.0.0.1:0"
	cfg.UDP.ListenAddr = "127.0.0.1:0"

	engine, err := NewEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go engine.Run(ctx)
	time.Sleep(100 * time.Millisecond)

	udpAddr := engine.udp.LocalAddr()
	conn, err := net.Dial("udp", udpAddr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	data := make([]byte, 20)
	binary.LittleEndian.PutUint32(data[0:4], 42)
	coord := binary.CoordPacket{Lat: 37.7749, Lng: -122.4194}
	binary.EncodeCoord(data[4:20], &coord)

	_, err = conn.Write(data)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(200 * time.Millisecond)

	state, found := engine.grid.Get(spatial.EntityID(42))
	if !found {
		t.Fatal("entity 42 not found in grid")
	}
	if state.Coord.Lat != 37.7749 {
		t.Errorf("entity 42 lat = %f, want 37.7749", state.Coord.Lat)
	}

	stats := engine.Stats()
	if stats.PacketsRouted == 0 {
		t.Error("expected PacketsRouted > 0")
	}
	if stats.Entities == 0 {
		t.Error("expected Entities > 0")
	}
}
