package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"os"

	"github.com/disintegration/imaging"
	colorful "github.com/lucasb-eyer/go-colorful"
)

const (
	supersample = false

	motionCachePath = "/tmp/motion.json"
)

func main() {
	images := os.Args[1:]

	loadedImages, err := loadImages(images)
	if err != nil {
		panic(err)
	}

	motionCorrection := getMotionCorrection(images, loadedImages)

	if supersample {
		loadedImages = upscale(loadedImages)
	}

	bounds := loadedImages[0].Bounds()

	output := image.NewNRGBA(bounds)

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
			//output.Set(x, y, medianColor(currentColor))
		}
	}

	// Downscale if necessary.
	var o image.Image
	if supersample {
		o = downscale(output)
	} else {
		o = output
	}

	f, _ := os.Create("output.png")
	defer f.Close()
	png.Encode(f, o)
}

func getMotionCorrection(imageNames []string, imgs []image.Image) []Motion {
	motionCorrection := make([]Motion, len(imgs))

	motionCache := make(MotionCache, len(imgs))
	motionCache.ReadFromFile(motionCachePath)

	fmt.Printf("Reference %s\t 0 0\n", imageNames[0])

	var motion Motion
	var found bool
	for i := 1; i < len(imgs); i++ {
		if motion, found = motionCache[imageNames[i]]; found {
			motionCorrection[i] = motion
			fmt.Printf("Cached motion %s\t: %d %d\n", imageNames[i], motion.X, motion.Y)
			continue
		}

		motion = estimateMotion(imgs[0], imgs[i])
		motionCorrection[i] = motion
		motionCache[imageNames[i]] = motion
		fmt.Printf("Motion calculated %s\t: %d %d\n", imageNames[i], motionCorrection[i].X, motionCorrection[i].Y)
		go motionCache.WriteToFile(motionCachePath)
	}

	motionCache.WriteToFile(motionCachePath)

	return motionCorrection
}

func loadImages(images []string) ([]image.Image, error) {
	var loadedImages []image.Image
	for i := range images {
		currImg, _ := os.Open(images[i])
		defer currImg.Close()
		decoded, _, err := image.Decode(currImg)
		if err != nil {
			return loadedImages, err
		}

		loadedImages = append(loadedImages, decoded)
	}

	return loadedImages, nil
}

func upscale(images []image.Image) []image.Image {
	bounds := images[0].Bounds()
	width := bounds.Max.X * 2
	height := bounds.Max.Y * 2

	for i := range images {
		images[i] = imaging.Resize(images[i], width, height, imaging.Gaussian)
	}

	return images
}

func downscale(img *image.NRGBA) image.Image {
	bounds := img.Bounds()
	width := bounds.Max.X / 2
	height := bounds.Max.Y / 2

	img = imaging.Sharpen(img, 0.5)
	return imaging.Resize(img, width, height, imaging.Lanczos)
}
