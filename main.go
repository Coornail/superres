package main

import (
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"

	"github.com/disintegration/imaging"
	colorful "github.com/lucasb-eyer/go-colorful"
)

const (
	motionCachePath = "/tmp/motion.json"

	sharpenSigma = 0.5
)

var (
	supersample bool
	sharpen     bool
	verbose     bool
	parallelism int
	mergeMethod string
	samplerName string
	outputFile  string
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	flag.BoolVar(&supersample, "supersample", true, "Supersample image")
	flag.BoolVar(&sharpen, "sharpen", true, "Sharpen output image")
	flag.BoolVar(&verbose, "verbose", true, "Verbose output")
	flag.IntVar(&parallelism, "parallelism", runtime.NumCPU(), "Number of threads to download the articles")
	flag.StringVar(&mergeMethod, "mergeMethod", "median", "Method to merge pixels from the input images (median, average)")
	flag.StringVar(&samplerName, "sampler", "gauss", "Sample images for motion detection (gauss, uniform, edge)")
	flag.StringVar(&outputFile, "output", "output.png", "Output file name")
	flag.Parse()

	images := flag.Args()

	loadedImages, err := loadImages(images)
	if err != nil {
		panic(err)
	}

	if supersample {
		loadedImages = upscale(loadedImages)
	}

	motionCorrection := getMotionCorrection(images, loadedImages)

	var colorMergeMethod ColorMerge = medianColor
	if mergeMethod == "average" {
		colorMergeMethod = averageColor
	}

	output := superres(loadedImages, motionCorrection, colorMergeMethod)

	if sharpen {
		output = imaging.Sharpen(output, sharpenSigma)
	}

	if supersample {
		output = downscale(output)
	}

	f, _ := os.Create("output.png")
	defer f.Close()
	png.Encode(f, output)
}

func superres(images []image.Image, motionCorrection []Motion, colorMergeMethod ColorMerge) *image.NRGBA {
	bounds := images[0].Bounds()
	output := image.NewNRGBA(bounds)

	var currentColor []colorful.Color
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			currentColor = []colorful.Color{}

			for i := range images {
				currX := x + motionCorrection[i].X
				currY := y + motionCorrection[i].Y
				if currX < bounds.Min.X || currX >= bounds.Max.X ||
					currY < bounds.Min.Y || currY >= bounds.Max.Y {
					continue
				}

				currentColor = append(currentColor, rgbaToColorful(images[i].At(currX, currY)))
			}
			output.Set(x, y, colorMergeMethod(currentColor))
		}
	}

	return output
}

func verboseOutput(format string, args ...interface{}) {
	if verbose {
		fmt.Printf(format, args...)

	}
}

func getMotionCorrection(imageNames []string, imgs []image.Image) []Motion {
	motionCorrection := make([]Motion, len(imgs))

	motionCache := make(MotionCache, len(imgs))
	motionCache.ReadFromFile(motionCachePath)

	fmt.Printf("Reference %s:\t 0 0\n", imageNames[0])

	type jobResult struct {
		i      int
		motion Motion
	}

	motionWorker := func(jobs chan int, ch chan jobResult) {
		for i := range jobs {
			if motion, found := motionCache[imageNames[i]]; found {
				motionCorrection[i] = motion
				verboseOutput("Cached motion: %s\t %d %d\n", imageNames[i], motion.X, motion.Y)
				ch <- jobResult{i: i, motion: motion}
			} else {
				motion := estimateMotion(imgs[0], imgs[i])
				motionCorrection[i] = motion
				motionCache[imageNames[i]] = motion
				verboseOutput("Motion calculated: %s\t %d %d\n", imageNames[i], motionCorrection[i].X, motionCorrection[i].Y)
				ch <- jobResult{i: i, motion: motion}
			}
		}
	}

	jobQueue := make(chan int, len(imgs))
	resultQueue := make(chan jobResult, len(imgs))

	for w := 0; w < parallelism; w++ {
		go motionWorker(jobQueue, resultQueue)
	}

	for i := 1; i < len(imageNames); i++ {
		jobQueue <- i
	}
	close(jobQueue)

	for i := 1; i < len(imageNames); i++ {
		result := <-resultQueue
		motionCorrection[result.i] = result.motion
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

func downscale(img *image.NRGBA) *image.NRGBA {
	bounds := img.Bounds()
	width := bounds.Max.X / 2
	height := bounds.Max.Y / 2

	return imaging.Resize(img, width, height, imaging.CatmullRom)
}
