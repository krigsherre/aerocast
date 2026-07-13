package fanout

import (
	"sync"
)

const defaultChannelBuffer = 1024

type ChannelManager struct {
	mu    sync.RWMutex
	chans map[SubscriberID]chan []byte
}

func NewChannelManager() *ChannelManager {
	return &ChannelManager{
		chans: make(map[SubscriberID]chan []byte),
	}
}

func (cm *ChannelManager) Create(id SubscriberID) chan []byte {
	ch := make(chan []byte, defaultChannelBuffer)
	cm.mu.Lock()
	cm.chans[id] = ch
	cm.mu.Unlock()
	return ch
}

func (cm *ChannelManager) Get(id SubscriberID) chan []byte {
	cm.mu.RLock()
	ch := cm.chans[id]
	cm.mu.RUnlock()
	return ch
}

func (cm *ChannelManager) Remove(id SubscriberID) {
	cm.mu.Lock()
	if ch, ok := cm.chans[id]; ok {
		close(ch)
		delete(cm.chans, id)
	}
	cm.mu.Unlock()
}

func (cm *ChannelManager) Count() int {
	cm.mu.RLock()
	n := len(cm.chans)
	cm.mu.RUnlock()
	return n
}
