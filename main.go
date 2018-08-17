package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"

	colorful "github.com/lucasb-eyer/go-colorful"
)

const (
	maxMotion = 15
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

	motionCorrection := make([]Motion, len(loadedImages))
	motionCorrection[1] = estimateMotion(loadedImages[0], loadedImages[1])

	bounds := loadedImages[0].Bounds()

	output := image.NewRGBA(bounds)

	//currentColor := make([]colorful.Color, len(images))
	var currentColor []colorful.Color

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			currentColor = make([]colorful.Color, 0)

			for i := range images {
				currX := x + motionCorrection[i].x
				currY := y + motionCorrection[i].y
				if currX < 0 || currX > bounds.Max.X ||
					currY < 0 || currY > bounds.Max.Y {
					continue
				}

				//currentColor[i] = rgbaToColorful(loadedImages[i].At(x, y))
				currentColor = append(currentColor, rgbaToColorful(loadedImages[i].At(currX, currY)))
			}
			output.Set(x, y, averageColor(currentColor))
		}
	}

	f, _ := os.Create("output.png")
	defer f.Close()
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

func estimateMotion(reference, candidate image.Image) Motion {
	var bestXMotion, bestYMotion int
	scale := 10
	bounds := reference.Bounds()

	var bestDist = math.MaxFloat64
	var currentDist float64

	for xMotion := -maxMotion; xMotion <= maxMotion; xMotion++ {
		for yMotion := -maxMotion; yMotion <= maxMotion; yMotion++ {
			currentDist = 0
			numberOfPixelsCompared := 0

			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					if x+xMotion < bounds.Min.X || x+xMotion > bounds.Max.X ||
						y+yMotion < bounds.Min.Y || y+yMotion > bounds.Max.Y {
						continue
					}

					if x%scale != 0 || y%scale != 0 {
						continue
					}

					referencePoint := reference.At(x, y)
					candidatePoint := candidate.At(x+xMotion, y+yMotion)

					d := distance(rgbaToColorful(referencePoint), rgbaToColorful(candidatePoint))
					currentDist += d * d
					numberOfPixelsCompared++
				}
			}

			currentDist = currentDist / float64(numberOfPixelsCompared)

			if currentDist < bestDist {
				//fmt.Printf("Best dist so far\n")
				bestXMotion = xMotion
				bestYMotion = yMotion
				bestDist = currentDist
			}
		}
		fmt.Printf("xMotion=%d dist=%f bestDist=%f\n", xMotion, currentDist, bestDist)

	}

	return Motion{x: bestXMotion, y: bestYMotion}
}

type Motion struct {
	x int
	y int
}
