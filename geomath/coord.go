package geomath

import (
	"fmt"
	"math"
)

// A GeodesicCoordinate is a coordinate on an ellipsoid given in degrees from
// the central point (where the equator meets meridian).
//
// This should always be expressed in WGS84's EPSG:4326 format.
type GeodesicCoordinate struct {
	Latitude  Degree
	Longitude Degree
}

type Degree float64

func (d Degree) Radian() float64 {
	return float64(d) * math.Pi / 180
}

func RadianToDegree(r float64) Degree {
	return Degree(r * 180 / math.Pi)
}

func (c GeodesicCoordinate) String() string {
	return fmt.Sprintf("%.6fN %.6fE", c.Latitude, c.Longitude)
}

// Point is a coordinate in an undetermined two-dimensional coordinate system.
type Point struct {
	Latitude  float64
	Longitude float64
}

func (s Point) String() string {
	return fmt.Sprintf("lat %7.1f, lon %6.1f", s.Latitude, s.Longitude)
}

type GeodesicBox struct {
	North Degree
	East  Degree
	South Degree
	West  Degree
}

func NewBox(nw GeodesicCoordinate, se GeodesicCoordinate) GeodesicBox {
	return GeodesicBox{
		North: nw.Latitude,
		East:  se.Longitude,
		South: se.Latitude,
		West:  nw.Longitude,
	}
}

func (b GeodesicBox) String() string {
	return fmt.Sprintf("%.6fN(max) %.6fE(max) %.6fN(min) %.6fE(min)", b.North, b.East, b.South, b.West)
}

func (b GeodesicBox) Inside(c GeodesicCoordinate) bool {
	return b.North >= c.Latitude && b.South <= c.Latitude && b.East >= c.Longitude && b.West <= c.Longitude
}

var (
	Sweref99TM = GaussKreger{
		axis:            6_378_137,
		flattening:      1.0 / 298.257222101,
		centralMeridian: 15,
		scale:           0.9996,
		falseNorthing:   0,
		falseEasting:    5e5,
	}
)

type GaussKreger struct {
	axis            float64
	flattening      float64
	centralMeridian Degree
	scale           float64
	falseNorthing   float64
	falseEasting    float64
}

// ToCoordinate translates a [Point] on in a coordinate system into a geodesic
// [GeodesicCoordinate]. This is essentially a magic function for anyone who isn't very
// fond of math.
//
// References:
//
//   - https://www.lantmateriet.se/sv/geodata/gps-geodesi-och-swepos/Om-geodesi/Formelsamling/
//   - https://www.trafiklab.se/sv/docs/using-trafiklab-data/combining-data/converting-sweref99-to-wgs84/
func (g GaussKreger) ToCoordinate(s Point) GeodesicCoordinate {
	n := g.flattening / (2 - g.flattening)
	aRoof := g.axis / (1 + n) * (1 + n*n/4 + n*n*n*n/64)
	xi := (s.Latitude - g.falseNorthing) / (g.scale * aRoof)
	eta := (s.Longitude - g.falseEasting) / (g.scale * aRoof)

	primeCalc := g.primeCalc(n, xi, eta)
	xiPrime := xi - primeCalc(math.Sin, math.Cosh)
	etaPrime := eta - primeCalc(math.Cos, math.Sinh)

	return GeodesicCoordinate{
		Latitude:  g.calcLat(xiPrime, etaPrime),
		Longitude: g.calcLon(xiPrime, etaPrime),
	}
}

func (g GaussKreger) primeCalc(n, xi, eta float64) func(func(float64) float64, func(float64) float64) float64 {
	delta1 := mulFactors(n, 1./2, -2./3, 37./96, -1./360)
	delta2 := mulFactors(n, 0, 1./48, 1./15, -437./1440)
	delta3 := mulFactors(n, 0, 0, 17./480, -37./840)
	delta4 := mulFactors(n, 0, 0, 0, 4397./161_280)
	return func(fst func(float64) float64, snd func(float64) float64) float64 {
		return delta1*fst(2*xi)*snd(2*eta) -
			delta2*fst(4*xi)*snd(4*eta) -
			delta3*fst(6*xi)*snd(6*eta) -
			delta4*fst(8*xi)*snd(8*eta)
	}
}

func (g GaussKreger) calcLon(xiPrime, etaPrime float64) Degree {
	lambdaZero := g.centralMeridian.Radian()
	deltaLambda := math.Atan(math.Sinh(etaPrime) / math.Cos(xiPrime))
	return RadianToDegree(lambdaZero + deltaLambda)
}

func (g GaussKreger) calcLat(xiPrime, etaPrime float64) Degree {
	e2 := g.flattening * (2 - g.flattening)

	aStar := mulFactors(e2, 1, 1, 1, 1)
	bStar := mulFactors(e2, 0, 7, 17, 30) / -6
	cStar := mulFactors(e2, 0, 0, 224, 889) / 120
	dStar := mulFactors(e2, 0, 0, 0, 4279) / -1260

	phiStar := math.Asin(math.Sin(xiPrime) / math.Cosh(etaPrime))
	sinPhiStar := math.Sin(phiStar)

	phi := phiStar + sinPhiStar*math.Cos(phiStar)*
		(aStar+
			bStar*math.Pow(sinPhiStar, 2)+
			cStar*math.Pow(sinPhiStar, 4)+
			dStar*math.Pow(sinPhiStar, 6))

	return RadianToDegree(phi)
}

func mulFactors(n, a, b, c, d float64) float64 {
	return a*n + b*n*n + c*n*n*n + d*n*n*n*n
}
