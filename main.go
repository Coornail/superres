package main

import (
	"fmt"
	"image"
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

	cache := make(MotionCache, len(images))
	cache.ReadFromFile("/tmp/motion.json")

	motionCorrection := make([]Motion, len(loadedImages))

	var motion Motion
	var found bool
	for i := 1; i < len(images); i++ {
		if motion, found = cache[images[i]]; found {
			motionCorrection[i] = motion
			fmt.Printf("Cached motion %s\t: %d %d\n", images[i], motion.X, motion.Y)
			continue
		}

		motion = estimateMotion(loadedImages[0], loadedImages[i])
		motionCorrection[i] = motion
		cache[images[i]] = motion
		fmt.Printf("Motion calculated %s\t: %d %d\n", images[i], motionCorrection[i].X, motionCorrection[i].Y)
	}

	cache.WriteToFile("/tmp/motion.json")

	bounds := loadedImages[0].Bounds()

	output := image.NewRGBA(bounds)

	var currentColor []colorful.Color

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			currentColor = make([]colorful.Color, 0)

			for i := range images {
				currX := x + motionCorrection[i].X
				currY := y + motionCorrection[i].Y
				if currX < bounds.Min.X || currX >= bounds.Max.X ||
					currY < bounds.Min.Y || currY >= bounds.Max.Y {
					continue
				}

				currentColor = append(currentColor, rgbaToColorful(loadedImages[i].At(currX, currY)))
			}
			output.Set(x, y, averageColor(currentColor))
		}
	}

	f, _ := os.Create("output.png")
	defer f.Close()
	png.Encode(f, output)
}
