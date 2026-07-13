package engine

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/krigsherre/aerocast/pkg/binary"
	"github.com/krigsherre/aerocast/pkg/fanout"
	"github.com/krigsherre/aerocast/pkg/geofence"
	"github.com/krigsherre/aerocast/pkg/spatial"
	udpTransport "github.com/krigsherre/aerocast/pkg/transport/udp"
	wsTransport "github.com/krigsherre/aerocast/pkg/transport/ws"
	"go.uber.org/zap"
)

type Engine struct {
	cfg            Config
	grid           *spatial.SpatialGrid
	fences         *geofence.Store
	evaluator      *geofence.Evaluator
	fanout         *fanout.Engine
	udp            *udpTransport.Listener
	ws             *wsTransport.Server
	prevMu         sync.RWMutex
	prevPos        map[spatial.EntityID]*binary.CoordPacket
	bloom          *tickBloomFilter
	cms            *countMinSketch
	shardHLL       [256]*HyperLogLog
	packetsRouted  atomic.Uint64
	packetsDropped atomic.Uint64
	eventsFired    atomic.Uint64
	wg             sync.WaitGroup
	done           chan struct{}
}

func NewEngine(cfg Config) (*Engine, error) {
	grid := spatial.NewSpatialGrid()
	fenceStore := geofence.NewStore()
	evaluator := geofence.NewEvaluator(fenceStore)

	if cfg.Logger == nil {
		cfg.Logger = zap.NewNop()
	}

	evaluator.GlobalCallback = func(ev geofence.FenceEvent) {
		if ev.Type == geofence.EventEnter && cfg.OnGeofenceEnter != nil {
			cfg.OnGeofenceEnter(ev.EntityID, ev.FenceName)
		} else if ev.Type == geofence.EventExit && cfg.OnGeofenceExit != nil {
			cfg.OnGeofenceExit(ev.EntityID, ev.FenceName)
		}
	}

	fanoutEngine := fanout.NewEngine()

	var udpListener *udpTransport.Listener
	var err error
	if !cfg.DisableUDP {
		udpListener, err = udpTransport.NewListener(cfg.UDP)
		if err != nil {
			return nil, fmt.Errorf("engine: udp listener: %w", err)
		}
	}

	e := &Engine{
		cfg:       cfg,
		grid:      grid,
		fences:    fenceStore,
		evaluator: evaluator,
		fanout:    fanoutEngine,
		udp:       udpListener,
		prevPos:   make(map[spatial.EntityID]*binary.CoordPacket, 65536),
		bloom:     &tickBloomFilter{},
		cms:       &countMinSketch{},
		done:      make(chan struct{}),
	}

	for i := 0; i < 256; i++ {
		e.shardHLL[i] = &HyperLogLog{}
	}

	if !cfg.DisableWS {
		wsServer, err := wsTransport.NewServer(cfg.WS, e, cfg.Logger)
		if err != nil {
			return nil, fmt.Errorf("engine: ws server: %w", err)
		}
		e.ws = wsServer
	}

	return e, nil
}

func (e *Engine) Fences() *geofence.Store {
	return e.fences
}

func (e *Engine) Run(ctx context.Context) error {
	e.cfg.Logger.Info("engine starting...")

	if e.udp != nil {
		e.wg.Add(1)
		go func() {
			defer e.wg.Done()
			if err := e.udp.Run(ctx); err != nil {
				e.cfg.Logger.Error("udp error", zap.Error(err))
			}
		}()
	}

	for i := 0; i < e.cfg.PipelineWorkers; i++ {
		e.wg.Add(1)
		go e.pipelineWorker(ctx)
	}

	e.wg.Add(1)
	go e.tickLoop(ctx)

	if e.ws != nil {
		e.wg.Add(1)
		go func() {
			defer e.wg.Done()
			if err := e.ws.Run(ctx); err != nil {
				e.cfg.Logger.Error("ws error", zap.Error(err))
			}
		}()
	}

	e.cfg.Logger.Info("engine running",
		zap.Int("workers", e.cfg.PipelineWorkers),
		zap.Duration("tick", e.cfg.TickRate))

	<-ctx.Done()
	e.cfg.Logger.Info("engine shutting down...")

	if e.udp != nil {
		e.udp.Close()
	}

	e.wg.Wait()

	e.cfg.Logger.Info("engine stopped")
	return ctx.Err()
}
