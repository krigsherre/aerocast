package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	reg *prometheus.Registry

	ConnectionsActive prometheus.Gauge
	ConnectionsTotal  prometheus.Counter
	PacketsRouted     prometheus.Counter
	PacketsDropped    prometheus.Counter
	RoutingLatency    prometheus.Histogram
	BroadcastBytes    prometheus.Counter
	GeofenceEvents    prometheus.Counter
	ShardLockWaitNs   prometheus.Histogram
	SubscribersActive prometheus.Gauge
	EntitiesActive    prometheus.Gauge
	UniqueEntities    prometheus.Gauge
}

func New() *Metrics {
	reg := prometheus.NewRegistry()
	reg.MustRegister(prometheus.NewGoCollector())
	reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	m := &Metrics{
		reg: reg,

		ConnectionsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "aerocast_connections_active",
			Help: "Current open WebSocket connections",
		}),
		ConnectionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "aerocast_connections_total",
			Help: "Total connections accepted since startup",
		}),
		PacketsRouted: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "aerocast_packets_routed_total",
			Help: "Total UDP packets successfully routed",
		}),
		PacketsDropped: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "aerocast_packets_dropped_total",
			Help: "Total UDP packets dropped (backpressure)",
		}),
		RoutingLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "aerocast_routing_latency_ns",
			Help:    "Per-packet routing latency in nanoseconds",
			Buckets: []float64{10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000},
		}),
		BroadcastBytes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "aerocast_broadcast_bytes_total",
			Help: "Total bytes broadcast to WebSocket clients",
		}),
		GeofenceEvents: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "aerocast_geofence_events_total",
			Help: "Total geofence enter/exit events triggered",
		}),
		ShardLockWaitNs: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "aerocast_shard_lock_wait_ns",
			Help:    "Time spent waiting on shard mutex in nanoseconds",
			Buckets: []float64{10, 50, 100, 250, 500, 1000, 5000, 10000, 50000},
		}),
		SubscribersActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "aerocast_subscribers_active",
			Help: "Current active spatial subscriptions",
		}),
		EntitiesActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "aerocast_entities_active",
			Help: "Current tracked entities in spatial grid",
		}),
		UniqueEntities: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "aerocast_unique_entities_estimated",
			Help: "Estimated unique entities (HLL) over analytics window",
		}),
	}

	reg.MustRegister(
		m.ConnectionsActive,
		m.ConnectionsTotal,
		m.PacketsRouted,
		m.PacketsDropped,
		m.RoutingLatency,
		m.BroadcastBytes,
		m.GeofenceEvents,
		m.ShardLockWaitNs,
		m.SubscribersActive,
		m.EntitiesActive,
		m.UniqueEntities,
	)

	return m
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.reg, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}
