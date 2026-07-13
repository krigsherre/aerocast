package spatial

import "math"

func ShardKey(lat, lng float64) uint8 {
	x := int32((lat + 90.0) / 180.0 * 16.0)
	if x > 15 {
		x = 15
	}
	if x < 0 {
		x = 0
	}
	y := int32((lng + 180.0) / 360.0 * 16.0)
	if y > 15 {
		y = 15
	}
	if y < 0 {
		y = 0
	}
	return uint8(x<<4) | uint8(y)
}

const earthRadiusM = 6_371_000.0

func HaversineM(lat1, lng1, lat2, lng2 float64) float64 {
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLng := (lng2 - lng1) * math.Pi / 180.0
	lat1Rad := lat1 * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0
	sinDLat := math.Sin(dLat / 2)
	sinDLng := math.Sin(dLng / 2)
	a := sinDLat*sinDLat + math.Cos(lat1Rad)*math.Cos(lat2Rad)*sinDLng*sinDLng
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusM * c
}

func HaversineMPrecomputed(lat1Rad, cosLat1, lng1, lat2, lng2 float64) float64 {
	dLat := (lat2 - lat1Rad*180.0/math.Pi) * math.Pi / 180.0
	lat2Rad := lat2 * math.Pi / 180.0
	dLng := (lng2 - lng1) * math.Pi / 180.0
	sinDLat := math.Sin(dLat / 2)
	sinDLng := math.Sin(dLng / 2)
	a := sinDLat*sinDLat + cosLat1*math.Cos(lat2Rad)*sinDLng*sinDLng
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusM * c
}

func ShardsForRadius(lat, lng, radiusM float64) []uint8 {
	latDegPerM := 1.0 / 111_320.0
	lngDegPerM := 1.0 / (111_320.0 * math.Cos(lat*math.Pi/180.0))
	dLatDeg := radiusM * latDegPerM
	dLngDeg := radiusM * lngDegPerM
	latMin := lat - dLatDeg
	latMax := lat + dLatDeg
	lngMin := lng - dLngDeg
	lngMax := lng + dLngDeg
	if latMin < -90 {
		latMin = -90
	}
	if latMax > 90 {
		latMax = 90
	}
	seen := [ShardCount]bool{}
	var result []uint8
	latStep := 180.0 / 16.0
	lngStep := 360.0 / 16.0
	latStart := int(math.Floor((latMin + 90.0) / latStep))
	latEnd := int(math.Floor((latMax + 90.0) / latStep))
	lngStart := int(math.Floor((lngMin + 180.0) / lngStep))
	lngEnd := int(math.Floor((lngMax + 180.0) / lngStep))
	for ix := latStart; ix <= latEnd; ix++ {
		for iy := lngStart; iy <= lngEnd; iy++ {
			x := ix & 0x0F
			y := iy & 0x0F
			key := uint8(x<<4) | uint8(y)
			if !seen[key] {
				seen[key] = true
				result = append(result, key)
			}
		}
	}
	return result
}
