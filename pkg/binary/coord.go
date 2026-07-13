package binary

import (
	"sync"
	"unsafe"
)

type CoordPacket struct {
	Lng float64
	Lat float64
}

var _ [16]byte = [unsafe.Sizeof(CoordPacket{})]byte{}

type ExtendedPacket struct {
	Version  uint8
	_        [7]byte
	Lng      float64
	Lat      float64
	Bearing  float32
	Speed    float32
	EntityID uint32
	_        [3]byte
}

var (
	coordPool = sync.Pool{
		New: func() any { return new(CoordPacket) },
	}
	extendedPool = sync.Pool{
		New: func() any { return new(ExtendedPacket) },
	}
)

func GetCoord() *CoordPacket {
	return coordPool.Get().(*CoordPacket)
}

func PutCoord(p *CoordPacket) {
	coordPool.Put(p)
}

func GetExtended() *ExtendedPacket {
	return extendedPool.Get().(*ExtendedPacket)
}

func PutExtended(p *ExtendedPacket) {
	extendedPool.Put(p)
}
