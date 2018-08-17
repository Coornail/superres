package main

import (
	"fmt"
	"image/color"
	"math"

	colorful "github.com/lucasb-eyer/go-colorful"
)

func averageColor(colors []colorful.Color) color.Color {
	var l, a, b float64

	for i := range colors {
		currL, currA, currB := colors[i].Lab()
		l += currL
		a += currA
		b += currB
	}

	c := float64(len(colors))

	return colorful.Lab(l/c, a/c, b/c).Clamped()
}

func medianColor(colors []colorful.Color) colorful.Color {
	var l, a, b []float64

	c := len(colors)

	if c == 1 {
		return colors[0]
	}

	l = make([]float64, c)
	a = make([]float64, c)
	b = make([]float64, c)

	for i := range colors {
		l[i], a[i], b[i] = colors[i].Lab()
	}

	if c%2 == 0 {
		return colorful.Lab(l[c/2], a[c/2], b[c/2])
	}

	i := int(math.Floor(float64(c) / 2.0))
	return colorful.Lab((l[i]+l[i+1])/2.0, (a[i]+a[i+1])/2.0, (b[i]+b[i+1])/2.0).Clamped()
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

func distance(c1, c2 colorful.Color) float64 {
	// We are using CIE76 as CIE94 sometimes returns NaN.
	d := c1.DistanceCIE76(c2)
	if math.IsNaN(d) {
		fmt.Printf("%s + %s = %#v\n", c1.Hex(), c2.Hex(), d)
		panic("Color distance is NaN")
	}

	// @todo why is this bigger than 1.0?
	if d < -1.0 {
		return -1.0
	}

	if d > 1.0 {
		return 1.0
	}

	return d
}
