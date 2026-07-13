package engine

import (
	"math/bits"
	"sync"
)

const hllRegisters = 64

type HyperLogLog struct {
	mu        sync.Mutex
	registers [hllRegisters]uint8
}

func (h *HyperLogLog) Add(id uint32) {
	hash := uint32(id * 0xc2b2ae35)
	idx := hash >> (32 - 6)
	remaining := (hash << 6) | 0x3f
	val := uint8(bits.LeadingZeros32(remaining)) + 1

	h.mu.Lock()
	if val > h.registers[idx] {
		h.registers[idx] = val
	}
	h.mu.Unlock()
}

func (h *HyperLogLog) Estimate() uint64 {
	h.mu.Lock()
	defer h.mu.Unlock()

	var sum float64
	for _, v := range h.registers {
		sum += 1.0 / float64(uint64(1)<<v)
	}

	estimate := 0.709 * float64(hllRegisters*hllRegisters) / sum
	return uint64(estimate)
}

func (h *HyperLogLog) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i := 0; i < hllRegisters; i++ {
		h.registers[i] = 0
	}
}
