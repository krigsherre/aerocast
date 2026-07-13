package udp

import (
	"time"

	"github.com/krigsherre/aerocast/pkg/binary"
)

type IngressPacket struct {
	EntityID  uint32
	Coord     binary.CoordPacket
	Timestamp time.Time
}
