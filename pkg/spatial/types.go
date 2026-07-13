package spatial

import "github.com/krigsherre/aerocast/pkg/binary"

type EntityID = uint32

type EntityState struct {
	Coord binary.CoordPacket
}

const ShardCount = 256
const ShardMask = ShardCount - 1
