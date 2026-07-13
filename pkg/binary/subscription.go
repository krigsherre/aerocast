package binary

import (
	"errors"
	"math"
)

var (
	ErrSubFrameTooShort = errors.New("binary: subscription frame too short")
	ErrUnknownOpcode    = errors.New("binary: unknown opcode")
)

type SubFrame struct {
	Opcode    uint8
	CenterLat float64
	CenterLng float64
	RadiusM   float64
	HasRadius bool
	EntityID  uint32
}

func DecodeSubFrame(data []byte) (SubFrame, error) {
	if len(data) < 1 {
		return SubFrame{}, ErrSubFrameTooShort
	}

	f := SubFrame{
		Opcode: data[0],
	}

	switch f.Opcode {
	case OpSubscribe, OpUnsubscribe:
		if len(data) < SubFrameSize {
			return SubFrame{}, ErrSubFrameTooShort
		}
		f.CenterLng = math.Float64frombits(LittleEndian.Uint64(data[1:9]))
		f.CenterLat = math.Float64frombits(LittleEndian.Uint64(data[9:17]))
		if len(data) >= 25 {
			f.RadiusM = math.Float64frombits(LittleEndian.Uint64(data[17:25]))
			f.HasRadius = true
		}

	case OpFollow, OpUnfollow:
		if len(data) < 5 {
			return SubFrame{}, ErrSubFrameTooShort
		}
		f.EntityID = LittleEndian.Uint32(data[1:5])

	case OpPing:
	default:
		return SubFrame{}, ErrUnknownOpcode
	}

	return f, nil
}

func EncodeSubscribe(lat, lng, radiusM float64) []byte {
	buf := make([]byte, 25)
	buf[0] = OpSubscribe
	LittleEndian.PutUint64(buf[1:9], math.Float64bits(lng))
	LittleEndian.PutUint64(buf[9:17], math.Float64bits(lat))
	LittleEndian.PutUint64(buf[17:25], math.Float64bits(radiusM))
	return buf
}
