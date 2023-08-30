package geomath

import (
	"math"
)

const EarthCircumference = 40_075_016.686 // meters at equator

type TileMap struct {
	Size uint16
}

func (t *TileMap) TileCoordinate(coord GeodesicCoordinate, zoom uint8) Point {
	n := float64(int(1) << zoom)
	latRad := coord.Latitude.Radian()
	lat := (1.0 - math.Asinh(math.Tan(latRad))/math.Pi) / 2.0 * n
	lon := n / 360 * float64(coord.Longitude+180)
	return Point{
		Latitude:  math.Floor(lat),
		Longitude: math.Floor(lon),
	}
}

func (t *TileMap) Tile(p Point, zoom uint8) Tile {
	return Tile{
		Center: t.tileCenter(p, zoom),
		Zoom:   zoom,
		Map:    t,
	}
}

func (t *TileMap) tileCenter(p Point, zoom uint8) GeodesicCoordinate {
	return t.tileNW(Point{
		Latitude:  p.Latitude + 0.5,
		Longitude: p.Longitude + 0.5,
	}, zoom)
}

func (t *TileMap) tileNW(p Point, zoom uint8) GeodesicCoordinate {
	n := float64(int(1) << zoom)

	latRad := math.Atan(math.Sinh(math.Pi * (1 - 2*p.Latitude/n)))
	return GeodesicCoordinate{
		Latitude:  RadianToDegree(latRad),
		Longitude: Degree(p.Longitude/n*360 - 180),
	}
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

func (t Tile) Position(coord GeodesicCoordinate) (Point, bool) {
	box := t.Boundaries()
	if !box.Inside(coord) {
		return Point{}, false
	}

	mapSize := float64(t.Map.Size)
	stepLat := mapSize / float64(box.North-box.South)
	stepLon := mapSize / float64(box.East-box.West)

	relLat := float64(coord.Latitude - box.South)
	relLon := float64(coord.Longitude - box.West)

	return Point{
		Latitude:  math.Round(mapSize - stepLat*relLat),
		Longitude: math.Round(stepLon * relLon),
	}, true
}

func (t Tile) Boundaries() GeodesicBox {
	p := t.Map.TileCoordinate(t.Center, t.Zoom)

	nw := t.Map.tileNW(p, t.Zoom)
	se := t.Map.tileNW(Point{
		Latitude:  p.Latitude + 1,
		Longitude: p.Longitude + 1,
	}, t.Zoom)

	return NewBox(nw, se)
}

func (t Tile) TilePosition() Point {
	return t.Map.TileCoordinate(t.Center, t.Zoom)
}
