package engine

import (
	"time"

	"go.uber.org/zap"

	udpTransport "github.com/krigsherre/aerocast/pkg/transport/udp"
	wsTransport "github.com/krigsherre/aerocast/pkg/transport/ws"
)

type Config struct {
	WS              wsTransport.Config
	UDP             udpTransport.Config
	DisableWS       bool
	DisableUDP      bool
	GridShards      int
	CellSizeM       float64
	TickRate        time.Duration
	PipelineWorkers int
	MaxEntities     int
	OnGeofenceEnter func(entityID uint32, fenceName string)
	OnGeofenceExit  func(entityID uint32, fenceName string)
	Logger          *zap.Logger
}

func DefaultConfig() Config {
	return Config{
		WS:              wsTransport.DefaultConfig(),
		UDP:             udpTransport.DefaultConfig(),
		GridShards:      256,
		CellSizeM:       500,
		TickRate:        50 * time.Millisecond,
		PipelineWorkers: 4,
		MaxEntities:     2_000_000,
	}
}
