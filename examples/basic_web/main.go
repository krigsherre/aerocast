package main

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krigsherre/aerocast"
	"go.uber.org/zap"
)

func main() {
	cfg := aerocast.DefaultConfig()
	cfg.DisableUDP = true

	logger, _ := zap.NewDevelopment()
	cfg.Logger = logger

	cfg.OnGeofenceEnter = func(entityID uint32, fenceName string) {
		logger.Info("Geofence ENTER", zap.Uint32("entity", entityID), zap.String("fence", fenceName))
	}
	cfg.OnGeofenceExit = func(entityID uint32, fenceName string) {
		logger.Info("Geofence EXIT", zap.Uint32("entity", entityID), zap.String("fence", fenceName))
	}

	engine, err := aerocast.New(cfg)
	if err != nil {
		logger.Error("Failed to start engine", zap.Error(err))
		os.Exit(1)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	go func() {
		logger.Info("Frontend HTTP server running on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			logger.Error("HTTP Server failed", zap.Error(err))
		}
	}()

	go simulateEntities(engine)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	engine.Run(ctx)
}

func simulateEntities(engine *aerocast.Engine) {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	type Entity struct {
		ID  uint32
		Lat float64
		Lng float64
	}
	entities := make([]*Entity, 100)
	for i := 0; i < 100; i++ {
		entities[i] = &Entity{
			ID:  uint32(i),
			Lat: 37.7749 + (rand.Float64()*0.02 - 0.01),
			Lng: -122.4194 + (rand.Float64()*0.02 - 0.01),
		}
	}

	for range ticker.C {
		for _, e := range entities {
			e.Lat += (rand.Float64() - 0.5) * 0.0001
			e.Lng += (rand.Float64() - 0.5) * 0.0001

			engine.Publish(e.ID, e.Lat, e.Lng)
		}
	}
}
