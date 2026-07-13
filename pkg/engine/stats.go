package engine

type EngineStats struct {
	PacketsRouted  uint64
	PacketsDropped uint64
	EventsFired    uint64
	UniqueEntities uint64
	Connections    int
	Subscribers    int
	Entities       int
	Fences         int
	ConnectionsTotal uint64
	BroadcastBytes   uint64
}

func (e *Engine) Stats() EngineStats {
	conns := 0
	var totalConns uint64
	var broadcastBytes uint64
	if e.ws != nil {
		conns = e.ws.ConnCount()
		totalConns = e.ws.TotalConnections()
		broadcastBytes = e.ws.BroadcastBytes()
	}

	var dropped uint64 = e.packetsDropped.Load()
	if e.udp != nil {
		dropped += e.udp.DroppedPackets()
	}
	dropped += e.fanout.DroppedFrames()

	var totalUnique uint64
	for i := 0; i < 256; i++ {
		totalUnique += e.shardHLL[i].Estimate()
	}

	return EngineStats{
		PacketsRouted:  e.packetsRouted.Load(),
		PacketsDropped: dropped,
		EventsFired:    e.eventsFired.Load(),
		UniqueEntities: totalUnique,
		Connections:    conns,
		Subscribers:    e.fanout.SubscriberCount(),
		Entities:       e.grid.Count(),
		Fences:         e.fences.Count(),
		ConnectionsTotal: totalConns,
		BroadcastBytes:   broadcastBytes,
	}
}
