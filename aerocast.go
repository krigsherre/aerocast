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
