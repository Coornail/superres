package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"math"
	"os"

	colorful "github.com/lucasb-eyer/go-colorful"
)

type Motion struct {
	X int
	Y int
}

const (
	// We make  (imageX*maxMotion)*(imageY*maxMotion)*(imageX*ImageSampleX)*(imageY*ImageSampleY)
	// comparison on the potentially supersampled image.
	maxMotion = 0.01

	ImageSamples = 1024
)

func estimateMotion(reference, candidate image.Image) Motion {
	var bestXMotion, bestYMotion int
	bounds := reference.Bounds()

	ref := NewImageCache(reference)

	var bestDist = math.MaxFloat64
	var currentDist float64
	var numberOfPixelsCompared int

	maxXMotion := int(math.Round(float64(bounds.Max.X) * maxMotion))
	maxYMotion := int(math.Round(float64(bounds.Max.Y) * maxMotion))

	for xMotion := -maxXMotion; xMotion <= maxXMotion; xMotion++ {
		for yMotion := -maxYMotion; yMotion <= maxYMotion; yMotion++ {
			currentDist = 0
			numberOfPixelsCompared = 0

			us := NewUniformSampler(reference, ImageSamples)
			for us.HasMore() {
				x, y := us.Next()
				if x+xMotion < bounds.Min.X || x+xMotion > bounds.Max.X ||
					y+yMotion < bounds.Min.Y || y+yMotion > bounds.Max.Y {
					//fmt.Printf("Out of bounds: %d %d", x, y)
					// @todo why does it go out of bounds?
					continue
				}

				referencePoint := ref.At(x, y)
				candidatePoint := candidate.At(x+xMotion, y+yMotion)

				d := distance(referencePoint, rgbaToColorful(candidatePoint))
				currentDist += d * d
				numberOfPixelsCompared++
			}

			currentDist = currentDist / float64(numberOfPixelsCompared)

			if currentDist < bestDist {
				bestXMotion = xMotion
				bestYMotion = yMotion
				bestDist = currentDist
			}
		}
	}

	if bestXMotion == maxXMotion || bestXMotion == -maxXMotion ||
		bestYMotion == maxYMotion || bestYMotion == -maxYMotion {
		fmt.Printf("Warning: Hit motion limit, consider raising maxMotion! bestXmotion: %d maxXMotion: %d | bestYMotion: %d maxYMotion: %d\n", bestXMotion, maxXMotion, bestYMotion, maxYMotion)
	}

	return Motion{X: bestXMotion, Y: bestYMotion}
}

type ImageCache struct {
	Img       image.Image
	cache     map[int]map[int]colorful.Color
	CacheHit  int
	CacheMiss int
}

func (ic *ImageCache) At(x, y int) (res colorful.Color) {
	if cacheX, foundX := ic.cache[x]; foundX {
		if res, foundY := cacheX[y]; foundY {
			ic.CacheHit++
			return res
		}
	} else {
		ic.cache[x] = make(map[int]colorful.Color)
	}

	ic.CacheMiss++
	color := rgbaToColorful(ic.Img.At(x, y))
	ic.cache[x][y] = color

	return color
}

func NewImageCache(img image.Image) *ImageCache {
	var ic ImageCache

	ic.Img = img
	ic.cache = make(map[int]map[int]colorful.Color, 0)

	return &ic
}

type MotionCache map[string]Motion

func (ms MotionCache) WriteToFile(filename string) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(ms); err != nil {
		return err
	}

	return ioutil.WriteFile(filename, buf.Bytes(), os.FileMode(0666|os.O_CREATE|os.O_TRUNC))
}

func (ms *MotionCache) ReadFromFile(filename string) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, &ms)
}
