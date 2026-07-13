package geofence

import (
	"math"

	"github.com/krigsherre/aerocast/pkg/spatial"
)

type Fence interface {
	Name() string
	Contains(lat, lng float64) bool
	BoundingRadius() float64
	Center() (lat, lng float64)
}

type Circle struct {
	NameStr   string
	CenterLat float64
	CenterLng float64
	RadiusM   float64
}

func (c *Circle) Name() string               { return c.NameStr }
func (c *Circle) BoundingRadius() float64    { return c.RadiusM }
func (c *Circle) Center() (float64, float64) { return c.CenterLat, c.CenterLng }

func (c *Circle) Contains(lat, lng float64) bool {
	return spatial.HaversineM(c.CenterLat, c.CenterLng, lat, lng) <= c.RadiusM
}

type Polygon struct {
	NameStr     string
	Vertices    []Vertex
	centerLat   float64
	centerLng   float64
	boundRadius float64
}

type Vertex struct {
	Lat float64
	Lng float64
}

func NewPolygon(name string, vertices []Vertex) *Polygon {
	p := &Polygon{
		NameStr:  name,
		Vertices: vertices,
	}

	var sumLat, sumLng float64
	for _, v := range vertices {
		sumLat += v.Lat
		sumLng += v.Lng
	}
	p.centerLat = sumLat / float64(len(vertices))
	p.centerLng = sumLng / float64(len(vertices))

	for _, v := range vertices {
		d := spatial.HaversineM(p.centerLat, p.centerLng, v.Lat, v.Lng)
		if d > p.boundRadius {
			p.boundRadius = d
		}
	}

	return p
}

func (p *Polygon) Name() string               { return p.NameStr }
func (p *Polygon) BoundingRadius() float64    { return p.boundRadius }
func (p *Polygon) Center() (float64, float64) { return p.centerLat, p.centerLng }

func (p *Polygon) Contains(lat, lng float64) bool {
	n := len(p.Vertices)
	if n < 3 {
		return false
	}

	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		vi := p.Vertices[i]
		vj := p.Vertices[j]

		if ((vi.Lat > lat) != (vj.Lat > lat)) &&
			(lng < (vj.Lng-vi.Lng)*(lat-vi.Lat)/(vj.Lat-vi.Lat)+vi.Lng) {
			inside = !inside
		}
		j = i
	}

	return inside
}

func inBoundingBox(centerLat, centerLng, radiusM, lat, lng float64) bool {
	latDeg := radiusM / 111_320.0
	lngDeg := radiusM / (111_320.0*math.Abs(math.Cos(centerLat*math.Pi/180.0)) + 1e-10)

	return lat >= centerLat-latDeg && lat <= centerLat+latDeg &&
		lng >= centerLng-lngDeg && lng <= centerLng+lngDeg
}
