package engine

import (
	"sync"
)

const (
	cmsRows = 4
	cmsCols = 1024
)

type countMinSketch struct {
	mu    sync.Mutex
	table [cmsRows][cmsCols]uint32
}

func (c *countMinSketch) addAndEstimate(id uint32) uint32 {
	h1 := uint32(id * 0x85ebca6b)
	h2 := uint32(id * 0xc2b2ae35)

	var minVal uint32 = 0xFFFFFFFF

	c.mu.Lock()
	defer c.mu.Unlock()

	for i := uint32(0); i < cmsRows; i++ {
		hash := h1 + i*h2
		col := hash % cmsCols

		c.table[i][col]++
		val := c.table[i][col]
		if val < minVal {
			minVal = val
		}
	}

	return minVal
}

func (c *countMinSketch) reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := 0; i < cmsRows; i++ {
		for j := 0; j < cmsCols; j++ {
			c.table[i][j] = 0
		}
	}
}
