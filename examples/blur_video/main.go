package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"math"
	"os"
	"runtime"
	"sync"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/ffmpego"
)

const MaxKernelSum = 0x800000

func main() {
	var blurRadius int
	var blurSigma float64
	flag.IntVar(&blurRadius, "radius", 5, "the number of pixels for the filter to span")
	flag.Float64Var(&blurSigma, "sigma", 2.0, "the blurring standard deviation")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: blur_video [flags] <input.mp4> <output.mp4>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()

	if len(flag.Args()) != 2 {
		flag.Usage()
	}
	inputFile := flag.Args()[0]
	outputFile := flag.Args()[1]

	reader, err := ffmpego.NewVideoReader(inputFile)
	essentials.Must(err)
	defer reader.Close()
	info := reader.VideoInfo()

	writer, err := ffmpego.NewVideoWriterWithAudio(
		outputFile,
		info.Width,
		info.Height,
		info.FPS,
		inputFile,
	)
	essentials.Must(err)
	defer writer.Close()

	log.Println("Copying and blurring frames...")
	filter := NewGaussianKernel(blurRadius, blurSigma)
	for i := 0; true; i++ {
		log.Printf("Blurring frame %d...", i+1)
		frame, err := reader.ReadFrame()
		if err == io.EOF {
			break
		}
		frame = filter.Filter(frame)
		essentials.Must(writer.WriteFrame(frame))
	}
}

type GaussianKernel struct {
	Radius int

	// Data stores coefficients for a 1D gaussian, which
	// can be applied in both axes to make a 2D gaussian.
	Data []uint32
}

func NewGaussianKernel(radius int, sigma float64) *GaussianKernel {
	res := &GaussianKernel{
		Radius: radius,
		Data:   make([]uint32, radius*2+1),
	}
	floatGaussian := make([]float64, 0, radius*2+1)
	for x := -radius; x <= radius; x++ {
		intensity := math.Exp(-float64(x*x) / (sigma * sigma))
		floatGaussian = append(floatGaussian, intensity)
	}

	// Make sure we don't overflow uint32
	var floatSum float64
	for _, x := range floatGaussian {
		for _, y := range floatGaussian {
			floatSum += x * y
		}
	}
	scale := math.Sqrt(float64(MaxKernelSum)/floatSum) * 0.9999
	for i, x := range floatGaussian {
		res.Data[i] = uint32(x * scale)
	}

	return res
}

func (g *GaussianKernel) Filter(img image.Image) image.Image {
	b := img.Bounds()
	width := b.Dx()
	height := b.Dy()

	colors := make([][3]uint32, 0, width*height)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			colors = append(colors, [3]uint32{r >> 8, g >> 8, b >> 8})
		}
	}

	mapY := func(f func(y int)) {
		var wg sync.WaitGroup
		numGos := runtime.GOMAXPROCS(0)
		for i := 0; i < numGos; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				for y := i; y < height; y += numGos {
					f(y)
				}
			}(i)
		}
		wg.Wait()
	}

	sums := make([]uint32, len(colors))
	intermediate := make([][3]uint32, len(colors))
	mapY(func(y int) {
		for x := 0; x < width; x++ {
			var kernelSum uint32
			var colorSum [3]uint32
			colorIdx := x + width*(y-g.Radius)
			for _, k := range g.Data {
				if colorIdx >= 0 && colorIdx < len(colors) {
					c := colors[colorIdx]
					for i, x := range c {
						colorSum[i] += x * k
					}
					kernelSum += k
				}
				colorIdx += width
			}
			idx := x + y*width
			intermediate[idx] = colorSum
			sums[idx] = kernelSum
		}
	})

	result := image.NewRGBA(image.Rect(0, 0, width, height))
	mapY(func(y int) {
		startIdx := y * width
		endIdx := (y + 1) * width
		for x := 0; x < b.Dx(); x++ {
			var kernelSum uint32
			var colorSum [3]uint32
			colorIdx := (x - g.Radius) + y*width
			for _, k := range g.Data {
				if colorIdx >= startIdx && colorIdx < endIdx {
					c := intermediate[colorIdx]
					for i, x := range c {
						colorSum[i] += x * k
					}
					kernelSum += k * sums[colorIdx]
				}
				colorIdx++
			}
			for i := range colorSum {
				colorSum[i] /= kernelSum
			}
			result.SetRGBA(x, y, color.RGBA{
				R: uint8(colorSum[0]),
				G: uint8(colorSum[1]),
				B: uint8(colorSum[2]),
				A: 0xff,
			})
		}
	})

	return result
}
