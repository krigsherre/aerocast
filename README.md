<div align="center">
  <br>
  <img src="assets/aerocast.png" width="350" alt="Aerocast Logo">
  <h1>🚀 Aerocast</h1>
  <p><b>Hyper-Scale Spatial Multiplexing & Geofencing for Real-Time State</b></p>
  <br>
  
  [![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
  [![License](https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge)](LICENSE)
  [![Zero Allocs](https://img.shields.io/badge/Zero_Allocations-True-success?style=for-the-badge)](#performance)
  [![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen?style=for-the-badge)](#)
</div>

---

**Aerocast** is an ultra-low latency, zero-allocation spatial routing engine built in Go. It is designed to ingest massive streams of coordinate data (e.g. from IoT devices, ride-sharing apps, or MMO games), evaluate complex Geofences instantly, and fan-out state to millions of WebSocket clients in real-time.

Whether you're building the next Uber, a real-time multiplayer game, or tracking a fleet of 100,000 drones, Aerocast handles the spatial math so you don't have to.

## 🔥 Why Aerocast?

Aerocast is built for raw throughput. By utilizing a 16x16 fixed spatial grid, advanced probabilistic data structures, and raw binary serialization, it achieves **~5,000,000 packets per second** on a single CPU core.

| Feature | Description |
|---|---|
| **🏎️ Zero Allocations** | The core routing loop allocates exactly `0 B/op`. The Garbage Collector sleeps while your data flies. |
| **🛡️ Probabilistic Defenses** | Built-in **Tick-Level Bloom Filters** instantly drop duplicate spam, and **Count-Min Sketches** automatically rate-limit abusive clients. |
| **📊 HLL Analytics** | Uses **HyperLogLog** to estimate Unique Daily Active Entities passing through every spatial shard with only 64 bytes of memory per region. |
| **⚡ Binary Codecs** | Coordinate data is packed into highly-optimized 12-byte little-endian structs for maximum network and CPU cache efficiency. |
| **🌍 Actionable Geofences** | Native callbacks (`OnGeofenceEnter` / `OnGeofenceExit`) let you trigger custom business logic the millisecond a boundary is crossed. |

---

## 📈 Benchmarks

Tested on an **Apple M1 Max (10 cores)**. The engine routes coordinate packets through the pipeline, evaluates geofences, and stages egress buffers for WebSocket clients concurrently.

```text
goos: darwin
goarch: arm64
pkg: github.com/krigsherre/aerocast/pkg/engine
cpu: Apple M1 Max

// Standard pipeline
BenchmarkEnginePipeline-10              5555604       216.4 ns/op     4621264 pkts/sec       0 B/op       0 allocs/op

// Simulating 99% spam from malicious clients (Bloom filter interception)
BenchmarkEnginePipeline_WithSpam-10     8382646       141.0 ns/op     7093905 pkts/sec       0 B/op       0 allocs/op

// Realistic load (100 geofences, 1,000 subscribers, 100,000 active entities)
BenchmarkEnginePipeline_Realistic-10    5326860       199.0 ns/op     5025355 pkts/sec       0 B/op       0 allocs/op
```

> **Takeaway:** Aerocast can ingest, geofence, and route roughly **4.7 Million** location updates per second, without allocating a single byte on the heap.

---

## 🏗️ Architecture

Aerocast is built as an asynchronous, lock-free (where possible) pipeline:

```mermaid
graph LR
    subgraph Ingress
        UDP[UDP Devices] --> Listener
        API[Your Go App] --> Engine
    end
    
    subgraph Aerocast Pipeline
        Listener -->|Binary| Engine
        Engine -->|Bloom Filter Dedupe| Engine
        Engine -->|Count-Min Sketch Throttle| Engine
        Engine --> SpatialGrid[(16x16 Grid)]
        Engine --> Geofence[Geofence Evaluator]
        Geofence --"Trigger"--> Callbacks[OnEnter/OnExit]
        Engine --> Fanout[Fanout Channel]
    end
    
    subgraph Egress
        Fanout --> WS1[WebSocket Client]
        Fanout --> WS2[WebSocket Client]
    end
```

---

## 🛠️ Integration Guide

Aerocast is primarily designed as an **Embeddable Go Library** so you can wire it directly into your application's business logic. However, it also ships with a Standalone Daemon (`aerocastd`) if you just want a pre-compiled router.

### 1. As a Library (Recommended)

Embedding Aerocast into your existing Go monolith is extremely easy.

```go
package main

import (
	"context"
	"github.com/krigsherre/aerocast"
	"go.uber.org/zap"
)

func main() {
	// 1. Get default configuration
	cfg := aerocast.DefaultConfig()
	
	// Optional: Use Zap for blazing fast structured logging
	logger, _ := zap.NewProduction()
	cfg.Logger = logger

	// 2. Define your business logic!
	cfg.OnGeofenceEnter = func(entityID uint32, fenceName string) {
		logger.Info("Entity Entered Zone!", zap.Uint32("id", entityID), zap.String("zone", fenceName))
		// e.g., Send a Push Notification, unlock a scooter, alert police, etc.
	}

	// 3. Initialize the Engine
	engine, err := aerocast.New(cfg)
	if err != nil {
		logger.Fatal("failed to start aerocast", zap.Error(err))
	}

	// 4. Feed data programmatically (if bypassing UDP)
	go func() {
		// Pushing a location update for Entity ID 42 in San Francisco
		engine.Publish(42, 37.7749, -122.4194)
	}()

	// 5. Run the engine (blocks until context is cancelled)
	ctx := context.Background()
	engine.Run(ctx)
}
```

### 2. Full Example (Web Frontend)

To see Aerocast in action with a real HTML/Canvas visualizer connecting via WebSocket:

```bash
cd examples/basic_web
go run main.go
```
Then navigate to `http://localhost:8080` in your browser. You will see 100 entities being actively routed through the Aerocast engine in real-time.

---

## 🐳 Running as a Standalone Daemon

If you don't want to write Go code, you can run Aerocast as a standalone daemon that ingests raw UDP packets and exposes a WebSocket endpoint.

```bash
# Build the daemon
make build

# Run it
./bin/aerocastd -config configs/default.yaml
```

**Prometheus Metrics:** The daemon automatically exposes an endpoint at `http://localhost:9090/metrics` containing advanced telemetry like:
- `aerocast_packets_routed_total`
- `aerocast_unique_entities_estimated` (Powered by HyperLogLog)
- `aerocast_geofence_events_total`

---

## 🤝 Contributing
PRs are welcome! When contributing to the core engine, please ensure that you do not introduce heap allocations on the hot path. Verify with `go test -bench=. -benchmem`.
