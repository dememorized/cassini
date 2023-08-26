package geomath

import (
	"fmt"
	"math"
	"testing"
)

var (
	Map       = &TileMap{Size: 256}
	OrigoTile = Tile{
		Center: GeodesicCoordinate{0, 0},
	}
	StockholmTile = Tile{
		Center: GeodesicCoordinate{58.960351, 18.314639},
	}
)

func TestTile_MetersPerPixel(t *testing.T) {
	tests := map[Tile][]float64{
		OrigoTile: {
			// https://wiki.openstreetmap.org/wiki/Zoom_levels
			156_543, 78_272, 39_136, 19_568, 9_784,
			4_892, 2_446, 1_223, 611.496, 305.748,
			152.874, 76.437, 38.219, 19.109, 9.555,
			4.777, 2.389, 1.194, 0.597,
		},
		StockholmTile: {
			7: 630,
		},
	}

	for k, v := range tests {
		t.Run(fmt.Sprintf("%s", k.Center), func(t *testing.T) {
			for level, expected := range v {
				if expected == 0 {
					continue
				}
				margin := expected * 0.001

				tile := k
				tile.Map = Map
				tile.Zoom = uint8(level)
				m := tile.MetersPerPixel()
				delta := math.Abs(expected - m)

				if delta > margin {
					diff := delta / expected * 100
					t.Errorf("expected %f, got %f. %f%% off", expected, m, diff)
				}
			}
		})
	}
}

func TestTileCoordinate(t *testing.T) {
	type T struct {
		Coord    GeodesicCoordinate
		Zoom     uint8
		Expected Point
	}

	tests := []T{
		{StockholmTile.Center, 8, Point{75, 141}},
	}

	// Generate tests for origo
	for i := uint8(0); i < 20; i++ {
		lonlat := 1 << i
		lonlat = lonlat >> 1
		tests = append(tests, T{OrigoTile.Center, i, Point{float64(lonlat), float64(lonlat)}})
	}

	for _, v := range tests {
		t.Run(fmt.Sprintf("%s @ %d", v.Coord, v.Zoom), func(t *testing.T) {
			coord := Map.TileCoordinate(v.Coord, v.Zoom)
			if coord.Latitude != v.Expected.Latitude || coord.Longitude != v.Expected.Longitude {
				t.Errorf("expected %s, got %s", v.Expected, coord)
			}
		})
	}
}

func TestTileMap_TileCenter(t *testing.T) {
	t.Log(Map.TileBox(Point{75, 141}, 8))
}
