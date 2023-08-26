package geomath

import "math"

const EarthCircumference = 40_075_016.686 // meters at equator

type TileMap struct {
	Size uint16
}

func (t TileMap) TileCoordinate(coord GeodesicCoordinate, zoom uint8) Point {
	n := float64(int(1) << zoom)
	latRad := coord.Latitude.Radian()
	lat := n / 2 * (1 - (math.Asinh(math.Tan(latRad)))/math.Pi)
	lon := n / 360 * float64(coord.Longitude+180)
	return Point{
		Latitude:  math.Floor(lat),
		Longitude: math.Floor(lon),
	}
}

func (t TileMap) TileNW(p Point, zoom uint8) GeodesicCoordinate {
	n := float64(int(1) << zoom)

	latRad := math.Atan(math.Sinh(math.Pi * (1 - 2*p.Latitude/n)))
	return GeodesicCoordinate{
		Latitude:  RadianToDegree(latRad),
		Longitude: Degree(p.Longitude/n*360 - 180),
	}
}

func (t TileMap) TileBox(p Point, zoom uint8) GeodesicBox {
	nw := t.TileNW(p, zoom)
	se := t.TileNW(Point{
		Latitude:  p.Latitude + 1,
		Longitude: p.Longitude + 1,
	}, zoom)

	return NewBox(nw, se)
}

func (t TileMap) TileCenter(p Point, zoom uint8) GeodesicCoordinate {
	return t.TileNW(Point{
		Latitude:  p.Latitude + 0.5,
		Longitude: p.Longitude + 0.5,
	}, zoom)
}

type Tile struct {
	Center GeodesicCoordinate
	Zoom   uint8
	Map    *TileMap
}

func (t Tile) MetersPerTile() float64 {
	scale := 1 << t.Zoom
	return EarthCircumference * math.Cos(t.Center.Latitude.Radian()) / float64(scale)
}

func (t Tile) MetersPerPixel() float64 {
	return t.MetersPerTile() / float64(t.Map.Size)
}

func sec(n float64) float64 {
	return 1 / math.Cos(n)
}
