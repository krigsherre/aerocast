package aerocast

import (
	"context"
	"time"

	"github.com/krigsherre/aerocast/pkg/binary"
	"github.com/krigsherre/aerocast/pkg/engine"
	"github.com/krigsherre/aerocast/pkg/geofence"
	udpTransport "github.com/krigsherre/aerocast/pkg/transport/udp"
)

type Engine struct {
	eng *engine.Engine
}

func New(cfg engine.Config) (*Engine, error) {
	eng, err := engine.NewEngine(cfg)
	if err != nil {
		return nil, err
	}
	return &Engine{eng: eng}, nil
}

func (e *Engine) Run(ctx context.Context) error {
	return e.eng.Run(ctx)
}

func (e *Engine) Fences() *geofence.Store {
	return e.eng.Fences()
}

func (e *Engine) Stats() engine.EngineStats {
	return e.eng.Stats()
}

func (e *Engine) Publish(entityID uint32, lat, lng float64) {
	pkt := udpTransport.IngressPacket{
		EntityID: entityID,
		Coord: binary.CoordPacket{
			Lat: lat,
			Lng: lng,
		},
		Timestamp: time.Now(),
	}
	e.eng.ProcessPacket(pkt)
}

func (e *Engine) SeedPosition(entityID uint32, lat, lng float64) {
	e.eng.SeedPosition(entityID, lat, lng)
}

func (e *Engine) SubscribersInCell(cell uint8) int {
	return e.eng.SubscribersInCell(cell)
}

func (e *Engine) ShardStats() [256]int {
	return e.eng.ShardStats()
}

func (e *Engine) Follow(subID uint64, entityID uint32) {
	e.eng.Follow(subID, entityID)
}

func (e *Engine) Unfollow(subID uint64, entityID uint32) {
	e.eng.Unfollow(subID, entityID)
}
