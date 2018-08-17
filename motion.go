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
	maxMotion  = 50
	ImageScale = 100
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

			for y := bounds.Min.Y; y < bounds.Max.Y; y += ImageScale {
				for x := bounds.Min.X; x < bounds.Max.X; x += ImageScale {
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
		//fmt.Printf("xMotion=%d dist=%f bestDist=%f\n", xMotion, currentDist, bestDist)

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
