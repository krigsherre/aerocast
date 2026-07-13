package udp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/krigsherre/aerocast/pkg/binary"
)

type Config struct {
	ListenAddr  string
	BatchSize   int
	ReadBuffer  int
	ChannelSize int
	PacketSize  int
}

func DefaultConfig() Config {
	return Config{
		ListenAddr:  ":9101",
		BatchSize:   32,
		ReadBuffer:  4 * 1024 * 1024,
		ChannelSize: 65536,
		PacketSize:  binary.CoordPacketSize,
	}
}

type Listener struct {
	cfg            Config
	conn           *net.UDPConn
	out            chan IngressPacket
	wg             sync.WaitGroup
	done           chan struct{}
	droppedPackets atomic.Uint64
}

func NewListener(cfg Config) (*Listener, error) {
	addr, err := net.ResolveUDPAddr("udp", cfg.ListenAddr)
	if err != nil {
		return nil, fmt.Errorf("udp: resolve addr: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("udp: listen: %w", err)
	}

	if err := conn.SetReadBuffer(cfg.ReadBuffer); err != nil {
		conn.Close()
		return nil, fmt.Errorf("udp: set read buffer: %w", err)
	}

	return &Listener{
		cfg:  cfg,
		conn: conn,
		out:  make(chan IngressPacket, cfg.ChannelSize),
		done: make(chan struct{}),
	}, nil
}

func (l *Listener) Packets() <-chan IngressPacket {
	return l.out
}

func (l *Listener) Run(ctx context.Context) error {
	l.wg.Add(1)
	defer l.wg.Done()

	defer close(l.out)
	defer l.conn.Close()

	bufSize := l.cfg.BatchSize * (l.cfg.PacketSize + 28)
	buf := make([]byte, bufSize)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := l.readAndProcessBatch(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			select {
			case <-l.done:
				return nil
			default:
				return fmt.Errorf("udp: read batch: %w", err)
			}
		}
	}
}

func (l *Listener) processPacket(data []byte) error {
	if len(data) < l.cfg.PacketSize {
		return fmt.Errorf("udp: packet too short: %d < %d", len(data), l.cfg.PacketSize)
	}

	var pkt binary.CoordPacket
	if err := binary.DecodeCoordInPlace(data, &pkt); err != nil {
		return err
	}

	var entityID uint32
	if len(data) >= 20 {
		entityID = binary.LittleEndian.Uint32(data[0:4])
		if err := binary.DecodeCoordInPlace(data[4:20], &pkt); err != nil {
			return err
		}
	}

	ip := IngressPacket{
		EntityID:  entityID,
		Coord:     pkt,
		Timestamp: time.Now(),
	}

	select {
	case l.out <- ip:
	default:
		l.droppedPackets.Add(1)
	}

	return nil
}

func (l *Listener) Close() error {
	close(l.done)
	l.conn.SetReadDeadline(time.Now())
	l.wg.Wait()
	return nil
}

func (l *Listener) DroppedPackets() uint64 {
	return l.droppedPackets.Load()
}

func (l *Listener) LocalAddr() string {
	return l.conn.LocalAddr().String()
}
