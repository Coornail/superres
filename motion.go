package main

import (
	"bytes"
	"encoding/json"
	"image"
	"io/ioutil"
	"math"
	"os"
)

type Motion struct {
	X int
	Y int
}

const (
	// We are making maxMotion^2*ImageSampleX*ImageSampleY distance comparisons per motion estimation.
	maxMotion = 50

	ImageSampleX = 64
	ImageSampleY = 64
)

func estimateMotion(reference, candidate image.Image) Motion {
	var bestXMotion, bestYMotion int
	bounds := reference.Bounds()

	var bestDist = math.MaxFloat64
	var currentDist float64
	var numberOfPixelsCompared int

	for xMotion := -maxMotion; xMotion <= maxMotion; xMotion++ {
		for yMotion := -maxMotion; yMotion <= maxMotion; yMotion++ {
			currentDist = 0
			numberOfPixelsCompared = 0

			for y := bounds.Min.Y; y < bounds.Max.Y; y += stepY {
				for x := bounds.Min.X; x < bounds.Max.X; x += stepX {
					if x+xMotion < bounds.Min.X || x+xMotion > bounds.Max.X ||
						y+yMotion < bounds.Min.Y || y+yMotion > bounds.Max.Y {
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
				bestXMotion = xMotion
				bestYMotion = yMotion
				bestDist = currentDist
			}
		}
	}

	return Motion{X: bestXMotion, Y: bestYMotion}
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
