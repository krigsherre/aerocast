package binary

const (
	CoordPacketSize    = 16
	ExtendedPacketSize = 32
	EgressRecordSize   = 20
	SubFrameSize       = 17

	VersionCoord    uint8 = 0x01
	VersionExtended uint8 = 0x02

	OpSubscribe   uint8 = 0x01
	OpUnsubscribe uint8 = 0x02
	OpFollow      uint8 = 0x03
	OpUnfollow    uint8 = 0x04
	OpPing        uint8 = 0xFF

	byteOrder = 0
)
