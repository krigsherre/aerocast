package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krigsherre/aerocast/pkg/config"
	"github.com/krigsherre/aerocast/pkg/engine"
	"github.com/krigsherre/aerocast/pkg/metrics"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	version   = "dev"
	commitSHA = "unknown"
)

func main() {
	var (
		configPath = flag.String("config", "configs/default.yaml", "path to config file")
		showVer    = flag.Bool("version", false, "print version and exit")
	)
	flag.Parse()

	if *showVer {
		fmt.Printf("aerocast %s (%s)\n", version, commitSHA)
		os.Exit(0)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	var logLevel zapcore.Level
	switch cfg.Logging.Level {
	case "debug":
		logLevel = zap.DebugLevel
	case "warn":
		logLevel = zap.WarnLevel
	case "error":
		logLevel = zap.ErrorLevel
	default:
		logLevel = zap.InfoLevel
	}

	var logger *zap.Logger
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(logLevel)

	if cfg.Logging.Format != "json" {
		zapConfig.Encoding = "console"
	}

	logger, _ = zapConfig.Build()
	defer logger.Sync()

	logger.Info("starting aerocast",
		zap.String("version", version),
		zap.String("commit", commitSHA),
		zap.String("ws", cfg.Server.WSListen),
		zap.String("udp", cfg.Server.UDPListen),
	)

	engineCfg := engine.Config{
		WS:              cfg.WebSocket.ToTransportConfig(),
		UDP:             cfg.UDP.ToTransportConfig(),
		GridShards:      cfg.Spatial.GridShards,
		CellSizeM:       cfg.Spatial.CellSizeM,
		TickRate:        cfg.Engine.TickRate,
		PipelineWorkers: cfg.Engine.PipelineWorkers,
		MaxEntities:     cfg.Engine.MaxEntities,
		Logger:          logger,
	}
	engineCfg.WS.ListenAddr = cfg.Server.WSListen
	engineCfg.UDP.ListenAddr = cfg.Server.UDPListen

	eng, err := engine.NewEngine(engineCfg)
	if err != nil {
		logger.Error("failed to create engine", zap.Error(err))
		os.Exit(1)
	}

	m := metrics.New()
	mux := http.NewServeMux()
	mux.Handle("/metrics", m.Handler())
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		stats := eng.Stats()
		if stats.Connections >= 0 {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"status":"ok","connections":%d,"entities":%d}`,
				stats.Connections, stats.Entities)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	})

	metricsServer := &http.Server{
		Addr:    cfg.Server.MetricsListen,
		Handler: mux,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := metricsServer.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error("metrics server error", zap.Error(err))
		}
	}()

	go func(ctx context.Context) {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		var lastStats engine.EngineStats

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stats := eng.Stats()
				m.ConnectionsActive.Set(float64(stats.Connections))
				m.SubscribersActive.Set(float64(stats.Subscribers))
				m.EntitiesActive.Set(float64(stats.Entities))
				m.UniqueEntities.Set(float64(stats.UniqueEntities))

				if stats.PacketsRouted > lastStats.PacketsRouted {
					m.PacketsRouted.Add(float64(stats.PacketsRouted - lastStats.PacketsRouted))
				}
				if stats.PacketsDropped > lastStats.PacketsDropped {
					m.PacketsDropped.Add(float64(stats.PacketsDropped - lastStats.PacketsDropped))
				}
				if stats.EventsFired > lastStats.EventsFired {
					m.GeofenceEvents.Add(float64(stats.EventsFired - lastStats.EventsFired))
				}
				if stats.ConnectionsTotal > lastStats.ConnectionsTotal {
					m.ConnectionsTotal.Add(float64(stats.ConnectionsTotal - lastStats.ConnectionsTotal))
				}
				if stats.BroadcastBytes > lastStats.BroadcastBytes {
					m.BroadcastBytes.Add(float64(stats.BroadcastBytes - lastStats.BroadcastBytes))
				}

				lastStats = stats
			}
		}
	}(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-sigCh
		logger.Info("received signal, shutting down", zap.String("signal", sig.String()))
		cancel()
		metricsServer.Shutdown(context.Background())
	}()

	if err := eng.Run(ctx); err != nil {
		logger.Error("engine stopped with error", zap.Error(err))
		os.Exit(1)
	}

	stats := eng.Stats()
	logger.Info("shutdown complete",
		zap.Uint64("packets_routed", stats.PacketsRouted),
		zap.Uint64("events_fired", stats.EventsFired),
	)
}
