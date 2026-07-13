package binary

import (
	"encoding/binary"
	"errors"
	"math"
)

var (
	ErrPacketTooShort  = errors.New("binary: packet too short")
	ErrPacketWrongSize = errors.New("binary: unexpected packet size")
	ErrInvalidVersion  = errors.New("binary: unknown version tag")
)

var LittleEndian = binary.LittleEndian

func DecodeCoord(data []byte) (*CoordPacket, error) {
	if len(data) < CoordPacketSize {
		return nil, ErrPacketTooShort
	}
	pkt := GetCoord()
	pkt.Lng = math.Float64frombits(LittleEndian.Uint64(data[0:8]))
	pkt.Lat = math.Float64frombits(LittleEndian.Uint64(data[8:16]))
	return pkt, nil
}

func DecodeCoordInPlace(data []byte, dst *CoordPacket) error {
	if len(data) < CoordPacketSize {
		return ErrPacketTooShort
	}
	dst.Lng = math.Float64frombits(LittleEndian.Uint64(data[0:8]))
	dst.Lat = math.Float64frombits(LittleEndian.Uint64(data[8:16]))
	return nil
}

func EncodeCoord(dst []byte, pkt *CoordPacket) {
	LittleEndian.PutUint64(dst[0:8], math.Float64bits(pkt.Lng))
	LittleEndian.PutUint64(dst[8:16], math.Float64bits(pkt.Lat))
}

func DecodeExtended(data []byte) (*ExtendedPacket, error) {
	if len(data) < ExtendedPacketSize {
		return nil, ErrPacketTooShort
	}
	pkt := GetExtended()
	pkt.Version = data[0]
	pkt.Lng = math.Float64frombits(LittleEndian.Uint64(data[8:16]))
	pkt.Lat = math.Float64frombits(LittleEndian.Uint64(data[16:24]))
	pkt.Bearing = math.Float32frombits(LittleEndian.Uint32(data[24:28]))
	pkt.Speed = math.Float32frombits(LittleEndian.Uint32(data[28:32]))
	return pkt, nil
}

func DecodePacket(data []byte) (*CoordPacket, *ExtendedPacket, error) {
	switch len(data) {
	case CoordPacketSize:
		coord, err := DecodeCoord(data)
		return coord, nil, err
	case ExtendedPacketSize:
		ext, err := DecodeExtended(data)
		return nil, ext, err
	default:
		return nil, nil, ErrPacketWrongSize
	}
}
