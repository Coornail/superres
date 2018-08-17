package main

import (
	"fmt"
	"image"
	"math"
)

type Motion struct {
	x int
	y int
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
