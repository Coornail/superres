package sampler

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestUniformSampler(t *testing.T) {
	bounds := image.Rect(0, 0, 1024, 768)
	output := image.NewNRGBA(bounds)
	us := NewUniformSampler(output, 12)

	renderSampler(us, output, "uniform.png")
}

func renderSampler(sampler ImageSampler, output *image.NRGBA, fileName string) {
	bounds := output.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			output.Set(x, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		}
	}

	for sampler.HasMore() {
		x, y := sampler.Next()
		fmt.Printf("%d %d\n", x, y)
		output.Set(x-1, y-1, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		output.Set(x-1, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		output.Set(x-1, y+1, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		output.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	}

	f, _ := os.Create(fileName)
	defer f.Close()
	png.Encode(f, output)
}
