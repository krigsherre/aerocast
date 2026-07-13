package engine

import (
	"sync"
)

const (
	bloomShards   = 64
	wordsPerShard = 32
	bitsPerShard  = wordsPerShard * 64
)

type tickBloomFilter struct {
	shards [bloomShards]struct {
		sync.Mutex
		bits [wordsPerShard]uint64
	}
}

func (f *tickBloomFilter) checkAndAdd(id uint32) bool {
	h1 := uint32(id * 0x85ebca6b)
	h2 := uint32(id * 0xc2b2ae35)

	shardIdx := h1 % bloomShards
	shard := &f.shards[shardIdx]

	bit1 := h1 % bitsPerShard
	bit2 := h2 % bitsPerShard

	idx1, mask1 := bit1/64, uint64(1)<<(bit1%64)
	idx2, mask2 := bit2/64, uint64(1)<<(bit2%64)

	shard.Lock()
	present := (shard.bits[idx1]&mask1 != 0) && (shard.bits[idx2]&mask2 != 0)
	shard.bits[idx1] |= mask1
	shard.bits[idx2] |= mask2
	shard.Unlock()

	return present
}

func (f *tickBloomFilter) reset() {
	for i := 0; i < bloomShards; i++ {
		shard := &f.shards[i]
		shard.Lock()
		for j := 0; j < wordsPerShard; j++ {
			shard.bits[j] = 0
		}
		shard.Unlock()
	}
}
