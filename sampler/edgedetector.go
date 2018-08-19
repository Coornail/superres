package sampler

import (
	"fmt"
	"image"
	"image/color"

	"github.com/disintegration/imaging"
	colorful "github.com/lucasb-eyer/go-colorful"
)

type EdgeDetector struct {
	Reference image.Image
	bounds    image.Rectangle
	Edges     int

	Treshold uint32

	x int
	y int
}

func (ed EdgeDetector) HasMore() bool {
	return ed.y < ed.bounds.Max.Y
}

func (ed *EdgeDetector) Next() (x, y int) {
	for ed.y < ed.bounds.Max.Y {
		ed.x++
		if ed.x > ed.bounds.Max.X {
			ed.y++
			ed.x = 0
		}

		if getBrightness(ed.Reference.At(ed.x, ed.y)) > ed.Treshold {
			ed.Edges++
			return ed.x, ed.y
		}
	}

	return ed.bounds.Max.X, ed.bounds.Max.Y
}

func getBrightness(p color.Color) uint32 {
	r, g, b, _ := p.RGBA()
	// https://ieeexplore.ieee.org/document/5329404/
	return 2*r + 3*g + 4*b

}

func (ed *EdgeDetector) Reset() {
	ed.x = 0
	ed.y = 0
}

func (ed EdgeDetector) NumberOfEdges() int {
	edges := 0
	for y := ed.bounds.Min.Y; y < ed.bounds.Max.Y; y++ {
		for x := ed.bounds.Min.X; x < ed.bounds.Max.X; x++ {
			p := ed.Reference.At(x, y)
			if getBrightness(p) > ed.Treshold {
				edges++
			}
		}
	}

	return edges
}

func NewEdgeDetector(img image.Image, samples int) *EdgeDetector {
	// Denoise using a gauss operator.
	// https://en.wikipedia.org/wiki/Canny_edge_detector
	denoised := imaging.Convolve5x5(
		img,
		[25]float64{
			2.0 / 159.0, 4.0 / 159.0, 5.0 / 159.0, 4.0 / 159.0, 2.0 / 159.0,
			4.0 / 159.0, 9.0 / 159.0, 12.0 / 159.0, 9.0 / 159.0, 4.0 / 159.0,
			5.0 / 159.0, 12.0 / 159.0, 15.0 / 159.0, 12.0 / 159.0, 5.0 / 159.0,
			4.0 / 159.0, 9.0 / 159.0, 12.0 / 159.0, 9.0 / 159.0, 4.0 / 159.0,
			2.0 / 159.0, 4.0 / 159.0, 5.0 / 159.0, 4.0 / 159.0, 2.0 / 159.0,
		},
		nil,
	)

	// Use two directional sobal operator.
	// https://en.wikipedia.org/wiki/Sobel_operator
	horizontal := imaging.Convolve3x3(
		denoised,
		[9]float64{
			1.0, 0, -1.0,
			2.0, 0, -2.0,
			1.0, 0, -1.0,
		},
		nil,
	)

	vertical := imaging.Convolve3x3(
		denoised,
		[9]float64{
			1.0, 2.0, 1.0,
			0.0, 0.0, 0.0,
			-1.0, -2.0, -1.0,
		},
		nil,
	)

	bounds := horizontal.Bounds()
	res := image.NewNRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			h := horizontal.At(x, y)
			v := vertical.At(x, y)

			res.Set(x, y, sumColors([]colorful.Color{rgbaToColorful(h), rgbaToColorful(v)}))
		}
	}

	// Line thinning.
	// @TODO improve
	black := colorful.LinearRgb(0.0, 0.0, 0.0)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if getBrightness(res.At(x, y)) > 128 {
				res.Set(x-1, y-1, black)
				res.Set(x-1, y, black)
				res.Set(x-1, y+1, black)
				res.Set(x, y-1, black)
				res.Set(x, y+1, black)
				res.Set(x+1, y-1, black)
				res.Set(x+1, y, black)
				res.Set(x+1, y+1, black)
			}
		}
	}

	ed := EdgeDetector{
		Reference: res,
		bounds:    img.Bounds(),
	}

	// Calculate the treshold dynamically to be one step smaller than the samples.
	ed.Treshold = 1
	for ed.NumberOfEdges() > samples {
		ed.Treshold *= 2
	}

	// We went too far, let's take a step back.
	if ed.NumberOfEdges() == 0 {
		ed.Treshold /= 2
	}

	return &ed
}

// @TODO move color to its own package.
func sumColors(colors []colorful.Color) colorful.Color {
	var l, a, b float64

	for i := range colors {
		currL, currA, currB := colors[i].Lab()
		l += currL
		a += currA
		b += currB
	}

	return colorful.Lab(l, a, b).Clamped()
}

func rgbaToColorful(c color.Color) colorful.Color {
	r, g, b, _ := c.RGBA()
	res := colorful.Color{
		R: float64(r) / 65535.0,
		G: float64(g) / 65535.0,
		B: float64(b) / 65535.0,
	}

	if res.R > 1.0 {
		res.R = 1.0
	}

	if res.G > 1.0 {
		res.G = 1.0
	}

	if res.B > 1.0 {
		res.B = 1.0
	}

	if !res.IsValid() {
		fmt.Printf("%#v\n", res)
		panic("invalid color")
	}

	return res
}
