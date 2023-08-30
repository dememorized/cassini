package geomath

import (
	"fmt"
	"math"
	"testing"
)

func TestSweref99_ToWGS84(t *testing.T) {
	const epsilon = 1e-5
	tests := map[Point]GeodesicCoordinate{
		Point{6539980.38, 689323.231}: {58.957488, 18.292103},
		Point{6580479.43, 674768.606}: {59.326855, 18.071839},
		Point{6529507.61, 666048.223}: {58.873208, 17.880177},
	}

	for origin, expected := range tests {
		t.Run(fmt.Sprintf("%s = %s", origin, expected), func(t *testing.T) {
			coordinate := Sweref99TM.ToCoordinate(origin)
			deltaLat := math.Abs(float64(coordinate.Latitude - expected.Latitude))
			deltaLon := math.Abs(float64(coordinate.Longitude - expected.Longitude))

			if deltaLon > epsilon || deltaLat > epsilon {
				t.Errorf("expected %s, got %s. Diff: (%f, %f) Tolerated: %f", expected, coordinate, deltaLat, deltaLon, epsilon)
			}
		})
	}
}
