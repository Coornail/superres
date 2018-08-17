package main

import (
	"image"
	_ "image/jpeg"
	"image/png"
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
