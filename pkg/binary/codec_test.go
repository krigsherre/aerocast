package binary

import (
	"math"
	"testing"
)

func TestDecodeCoordInPlace(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantLng float64
		wantLat float64
		wantErr error
	}{
		{
			name:    "san francisco",
			data:    encodeTestCoord(-122.4194, 37.7749),
			wantLng: -122.4194,
			wantLat: 37.7749,
		},
		{
			name:    "null island",
			data:    encodeTestCoord(0, 0),
			wantLng: 0,
			wantLat: 0,
		},
		{
			name:    "north pole",
			data:    encodeTestCoord(0, 90),
			wantLng: 0,
			wantLat: 90,
		},
		{
			name:    "south pole",
			data:    encodeTestCoord(0, -90),
			wantLng: 0,
			wantLat: -90,
		},
		{
			name:    "antimeridian east",
			data:    encodeTestCoord(179.999, 0),
			wantLng: 179.999,
			wantLat: 0,
		},
		{
			name:    "antimeridian west",
			data:    encodeTestCoord(-179.999, 0),
			wantLng: -179.999,
			wantLat: 0,
		},
		{
			name:    "packet too short",
			data:    []byte{0x01, 0x02},
			wantErr: ErrPacketTooShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pkt CoordPacket
			err := DecodeCoordInPlace(tt.data, &pkt)
			if err != tt.wantErr {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if math.Abs(pkt.Lng-tt.wantLng) > 1e-6 {
				t.Errorf("Lng = %f, want %f", pkt.Lng, tt.wantLng)
			}
			if math.Abs(pkt.Lat-tt.wantLat) > 1e-6 {
				t.Errorf("Lat = %f, want %f", pkt.Lat, tt.wantLat)
			}
		})
	}
}

func TestEncodeDecodeRoundtrip(t *testing.T) {
	orig := CoordPacket{Lng: -73.9857, Lat: 40.7484}
	dst := make([]byte, CoordPacketSize)
	EncodeCoord(dst, &orig)

	var decoded CoordPacket
	if err := DecodeCoordInPlace(dst, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded != orig {
		t.Errorf("roundtrip failed: got %+v, want %+v", decoded, orig)
	}
}

func TestEgressBufferAppend(t *testing.T) {
	buf := GetEgressBuffer()
	defer PutEgressBuffer(buf)

	buf.Append(EgressRecord{EntityID: 1, Lat: 40.0, Lng: -73.0})
	buf.Append(EgressRecord{EntityID: 2, Lat: 41.0, Lng: -74.0})

	if buf.Len() != 2 {
		t.Fatalf("Len = %d, want 2", buf.Len())
	}

	records, err := DecodeEgressFrame(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("decoded %d records, want 2", len(records))
	}
	if records[0].EntityID != 1 || records[1].EntityID != 2 {
		t.Errorf("entity IDs mismatch: %+v", records)
	}
}

func encodeTestCoord(lng, lat float64) []byte {
	pkt := CoordPacket{Lng: lng, Lat: lat}
	dst := make([]byte, CoordPacketSize)
	EncodeCoord(dst, &pkt)
	return dst
}
