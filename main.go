package main

import (
	"fmt"
	shapefile "github.com/jonas-p/go-shp"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
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

type point struct {
	X float64
	Y float64
}

func (p point) ImageCoords(origin point, scale point) image.Point {
	return image.Point{
		X: int((p.X - origin.X) / scale.X),
		Y: int((p.Y - origin.Y) / scale.Y),
	}
}

type PointFeature struct {
	Position   point
	Color      color.Color
	Attributes map[string]string
}

type Image struct {
	Min   point
	Max   point
	Steps point

	Image *image.RGBA
}

func NewImage(min point, max point, w, h int) *Image {
	dX := math.Abs(max.X - min.X)
	dY := math.Abs(max.Y - min.Y)

	return &Image{
		Min: min,
		Max: max,
		Steps: point{
			X: dX / float64(w),
			Y: dY / float64(h),
		},
		Image: image.NewRGBA(image.Rect(0, 0, w, h)),
	}
}

func (i *Image) DrawPoint(p point, col color.Color) {
	c := p.ImageCoords(i.Min, i.Steps)
	i.Image.Set(c.X, c.Y, col)
}

type grid struct {
	Min point
	Max point

	Features []PointFeature
}

func main() {
	shape, err := shapefile.Open("terrang/01/js_01.shp")
	if err != nil {
		panic(err)
	}

	fields := shape.Fields()
	boundaries := shape.BBox()

	g := grid{
		Min: point{
			X: boundaries.MinX,
			Y: boundaries.MinY,
		},
		Max: point{
			X: boundaries.MaxX,
			Y: boundaries.MaxY,
		},
	}

	for shape.Next() {
		n, p := shape.Shape()
		box := p.BBox()

		feat := PointFeature{
			Position:   point{X: box.MinX, Y: box.MinY},
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

		g.Features = append(g.Features, feat)
	}

	img := NewImage(g.Min, g.Max, 256, 256)

	for _, feat := range g.Features {
		img.DrawPoint(feat.Position, feat.Color)
	}

	f, _ := os.Create("image.png")
	png.Encode(f, img.Image)
	fmt.Printf("%#v", g)
}
