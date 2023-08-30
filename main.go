package main

import (
	"cassini/geomath"
	"fmt"
	shapefile "github.com/jonas-p/go-shp"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

var uc = map[byte]rune{
	0xe5: 'å',
	0xe4: 'ä',
	0xf6: 'ö',
	0xc5: 'Å',
	0xc4: 'Ä',
	0xd6: 'Ö',
}

func fromLatin1(input string) (string, error) {
	output := strings.Builder{}

	for _, c := range []byte(input) {
		if c <= unicode.MaxASCII {
			output.WriteByte(c)
			continue
		}

		translation, exists := uc[c]
		if !exists {
			return "", fmt.Errorf("failed to find character %x in string %s", c, input)
		}
		output.WriteRune(translation)
	}

	return output.String(), nil
}

type PointFeature struct {
	Position   geomath.GeodesicCoordinate
	Color      color.Color
	Attributes map[string]string
}

type Image struct {
	Tile  geomath.Tile
	Image *image.RGBA
}

func NewImage(t geomath.Tile) *Image {
	return &Image{
		Tile:  t,
		Image: image.NewRGBA(image.Rect(0, 0, int(t.Map.Size), int(t.Map.Size))),
	}
}

func (i *Image) DrawPoint(coord geomath.GeodesicCoordinate, col color.Color) {
	p, ok := i.Tile.Position(coord)
	if !ok {
		return
	}

	i.Image.Set(int(p.Longitude), int(p.Latitude), col)
}

type grid struct {
	Point    geomath.Point
	Features []PointFeature
}

func main() {
	shape, err := shapefile.Open("/Users/emiltullstedt/Stockholm/Sweref_99_TM/shape/tk_01_Sweref_99_TM_shape/terrang/01/bs_01.shp")
	if err != nil {
		panic(err)
	}

	fields := shape.Fields()
	boundaries := shape.BBox()

	tiles := []geomath.Tile{}
	zoomLevels := []uint8{8, 9, 10, 11}

	for _, zoom := range zoomLevels {
		m := &geomath.TileMap{Size: 256}

		nw := m.TileCoordinate(geomath.Sweref99TM.ToCoordinate(geomath.Point{
			Latitude:  boundaries.MaxY,
			Longitude: boundaries.MinX,
		}), zoom)
		se := m.TileCoordinate(geomath.Sweref99TM.ToCoordinate(geomath.Point{
			Latitude:  boundaries.MinY,
			Longitude: boundaries.MaxX,
		}), zoom)

		for lat := nw.Latitude; lat <= se.Latitude; lat++ {
			for lon := nw.Longitude; lon <= se.Longitude; lon++ {
				tiles = append(tiles, m.Tile(geomath.Point{
					Latitude:  lat,
					Longitude: lon,
				}, zoom))
			}
		}
	}

	feats := make([]PointFeature, 0)
	for shape.Next() {
		n, p := shape.Shape()
		box := p.BBox()

		coord := geomath.Sweref99TM.ToCoordinate(geomath.Point{box.MinY, box.MinX})

		feat := PointFeature{
			Position:   coord,
			Color:      nil,
			Attributes: map[string]string{},
		}

		// print attributes
		for k, f := range fields {
			val, err := fromLatin1(shape.ReadAttribute(n, k))
			if err != nil {
				panic(err)
			}
			feat.Attributes[f.String()] = val
		}

		switch feat.Attributes["KKOD"] {
		case "355":
			feat.Color = color.RGBA{R: 0xFF, A: 0xFF}
		case "351":
			feat.Color = color.RGBA{G: 0xFF, A: 0xFF}
		default:
			feat.Color = color.RGBA{B: 0xFF, A: 0xFF}
		}

		feats = append(feats, feat)
	}

	os.Mkdir("out", 0o750)
	for _, t := range tiles {
		img := NewImage(t)
		for _, feat := range feats {
			img.DrawPoint(feat.Position, feat.Color)
		}

		pos := t.TilePosition()
		os.Mkdir(filepath.Join("out", fmt.Sprintf("%d", t.Zoom)), 0o750)
		os.Mkdir(filepath.Join("out", fmt.Sprintf("%d/%d", t.Zoom, int(pos.Longitude))), 0o750)
		f, _ := os.Create(fmt.Sprintf("out/%d/%d/%d.png", t.Zoom, int(pos.Longitude), int(pos.Latitude)))
		png.Encode(f, img.Image)
	}
}
