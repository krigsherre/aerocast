package binary

import (
	"math"
	"sync"
)

type EgressRecord struct {
	EntityID uint32
	Lat      float64
	Lng      float64
}

type EgressBuffer struct {
	data []byte
	n    int
}

var egressBufPool = sync.Pool{
	New: func() any {
		return &EgressBuffer{
			data: make([]byte, 0, 4096),
		}
	},
}

func GetEgressBuffer() *EgressBuffer {
	buf := egressBufPool.Get().(*EgressBuffer)
	buf.n = 0
	buf.data = buf.data[:0]
	return buf
}

func PutEgressBuffer(b *EgressBuffer) {
	egressBufPool.Put(b)
}

func (b *EgressBuffer) Reset() {
	b.n = 0
	b.data = b.data[:0]
}

func (b *EgressBuffer) Append(r EgressRecord) {
	off := len(b.data)
	if off+EgressRecordSize > cap(b.data) {
		newData := make([]byte, off+EgressRecordSize, cap(b.data)*2)
		copy(newData, b.data)
		b.data = newData
	} else {
		b.data = b.data[:off+EgressRecordSize]
	}

	LittleEndian.PutUint32(b.data[off:off+4], r.EntityID)
	LittleEndian.PutUint64(b.data[off+4:off+12], math.Float64bits(r.Lat))
	LittleEndian.PutUint64(b.data[off+12:off+20], math.Float64bits(r.Lng))
	b.n++
}

func (b *EgressBuffer) Bytes() []byte {
	return b.data
}

func (b *EgressBuffer) Len() int {
	return b.n
}

func DecodeEgressFrame(data []byte) ([]EgressRecord, error) {
	if len(data)%EgressRecordSize != 0 {
		return nil, ErrPacketWrongSize
	}

	count := len(data) / EgressRecordSize
	records := make([]EgressRecord, count)

	for i := 0; i < count; i++ {
		off := i * EgressRecordSize
		records[i].EntityID = LittleEndian.Uint32(data[off : off+4])
		records[i].Lat = math.Float64frombits(LittleEndian.Uint64(data[off+4 : off+12]))
		records[i].Lng = math.Float64frombits(LittleEndian.Uint64(data[off+12 : off+20]))
	}

	return records, nil
}
