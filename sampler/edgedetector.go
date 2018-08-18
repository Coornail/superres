package sampler

import (
	"image"
	"image/color"

	"github.com/disintegration/imaging"
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
	res := imaging.Convolve3x3(
		img,
		[9]float64{
			-0.25, 0, 0.25,
			0, 0, 0,
			0.25, 0, -0.25,
		},
		nil,
	)

	ed := EdgeDetector{
		Reference: res,
		bounds:    img.Bounds(),
	}

	// Calculate the treshold dynamically to be one step smaller than the samples.
	ed.Treshold = 2048
	for ed.NumberOfEdges() > samples {
		ed.Treshold *= 2
	}

	return &ed
}
