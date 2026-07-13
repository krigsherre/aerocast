package binary

import (
	"encoding/json"
	"testing"
)

func BenchmarkDecodeCoordInPlace(b *testing.B) {
	data := encodeTestCoord(-122.4194, 37.7749)
	var pkt CoordPacket

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = DecodeCoordInPlace(data, &pkt)
	}
}

func BenchmarkDecodeCoordPool(b *testing.B) {
	data := encodeTestCoord(-122.4194, 37.7749)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		pkt, _ := DecodeCoord(data)
		PutCoord(pkt)
	}
}

func BenchmarkEncodeCoord(b *testing.B) {
	pkt := CoordPacket{Lng: -122.4194, Lat: 37.7749}
	dst := make([]byte, CoordPacketSize)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		EncodeCoord(dst, &pkt)
	}
}

type jsonCoord struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

func BenchmarkJSONDecode(b *testing.B) {
	data := []byte(`{"lng":-122.4194,"lat":37.7749}`)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var c jsonCoord
		_ = json.Unmarshal(data, &c)
	}
}

func BenchmarkJSONEncode(b *testing.B) {
	c := jsonCoord{Lng: -122.4194, Lat: 37.7749}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(c)
	}
}

func BenchmarkEgressBufferAppend(b *testing.B) {
	buf := GetEgressBuffer()
	defer PutEgressBuffer(buf)

	rec := EgressRecord{EntityID: 42, Lat: 37.7749, Lng: -122.4194}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Append(rec)
		if buf.Len() >= 1000 {
			buf.Reset()
		}
	}
}

func BenchmarkFullCodecPipeline(b *testing.B) {
	data := encodeTestCoord(-122.4194, 37.7749)
	var pkt CoordPacket
	egressBuf := GetEgressBuffer()
	defer PutEgressBuffer(egressBuf)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = DecodeCoordInPlace(data, &pkt)
		egressBuf.Append(EgressRecord{
			EntityID: 42,
			Lat:      pkt.Lat,
			Lng:      pkt.Lng,
		})
		if egressBuf.Len() >= 200 {
			egressBuf.Reset()
		}
	}
}
