package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"os"

	colorful "github.com/lucasb-eyer/go-colorful"
)

func main() {
	images := os.Args[1:]

	var loadedImages []image.Image
	for i := range images {
		currImg, _ := os.Open(images[i])
		defer currImg.Close()
		decoded, _, err := image.Decode(currImg)
		if err != nil {
			panic(err)
		}

		loadedImages = append(loadedImages, decoded)
	}

	bounds := loadedImages[0].Bounds()

	output := image.NewRGBA(bounds)

	currentColor := make([]colorful.Color, len(images))

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for i := range images {
				currentColor[i] = rgbaToColorful(loadedImages[i].At(x, y))
			}
			output.Set(x, y, averageColor(currentColor))
		}
	}

	f, _ := os.Create("output.png")
	png.Encode(f, output)
}

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
