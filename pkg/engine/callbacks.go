package engine

import (
	"github.com/krigsherre/aerocast/pkg/binary"
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

func (e *Engine) SeedPosition(entityID uint32, lat, lng float64) {
	shardIdx := entityID % prevPosShards
	shard := &e.prevPosShards[shardIdx]

	coord := binary.CoordPacket{
		Lat: lat,
		Lng: lng,
	}

	shard.mu.Lock()
	shard.m[spatial.EntityID(entityID)] = &coord
	shard.mu.Unlock()

	e.grid.Route(spatial.EntityID(entityID), coord)
}

func (e *Engine) SubscribersInCell(cell uint8) int {
	return e.fanout.SubscribersInCell(cell)
}

func (e *Engine) ShardStats() [spatial.ShardCount]int {
	return e.grid.ShardStats()
}

func (e *Engine) Follow(subID uint64, entityID uint32) {
	e.fanout.Follow(fanout.SubscriberID(subID), spatial.EntityID(entityID))
}

func (e *Engine) Unfollow(subID uint64, entityID uint32) {
	e.fanout.Unfollow(fanout.SubscriberID(subID), spatial.EntityID(entityID))
}
