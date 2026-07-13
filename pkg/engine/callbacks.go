package engine

import (
	"github.com/krigsherre/aerocast/pkg/fanout"
	"github.com/krigsherre/aerocast/pkg/spatial"
	"go.uber.org/zap"
)

func (e *Engine) OnSubscribe(subID uint64, lat, lng, radiusM float64) {
	e.fanout.Subscribe(fanout.SubscriberID(subID), lat, lng, radiusM)
	if e.cfg.Logger != nil {
		e.cfg.Logger.Debug("subscriber connected",
			zap.Uint64("subID", subID),
			zap.Float64("lat", lat),
			zap.Float64("lng", lng),
			zap.Float64("radius", radiusM))
	}
}

func (e *Engine) OnUnsubscribe(subID uint64) {
	e.fanout.Unsubscribe(fanout.SubscriberID(subID))
}

func (e *Engine) OnFollow(subID uint64, entityID uint32) {
	e.fanout.Follow(fanout.SubscriberID(subID), spatial.EntityID(entityID))
}

func (e *Engine) OnUnfollow(subID uint64, entityID uint32) {
	e.fanout.Unfollow(fanout.SubscriberID(subID), spatial.EntityID(entityID))
}

func (e *Engine) OnDisconnect(subID uint64) {
	e.fanout.Unsubscribe(fanout.SubscriberID(subID))
}

func (e *Engine) Channels() *fanout.ChannelManager {
	return e.fanout.Channels()
}
