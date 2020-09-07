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
	filter := NewGaussianKernel(blurRadius, float32(blurSigma))
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
	Data   []float32
}

func NewGaussianKernel(radius int, sigma float32) *GaussianKernel {
	res := &GaussianKernel{
		Radius: radius,
		Data:   make([]float32, 0, radius*2+1),
	}
	for x := -radius; x <= radius; x++ {
		intensity := math.Exp(-float64(x*x) / float64(sigma*sigma))
		res.Data = append(res.Data, float32(intensity))
	}
	return res
}

func (g *GaussianKernel) Filter(img image.Image) image.Image {
	b := img.Bounds()
	width := b.Dx()
	height := b.Dy()

	var colors [][3]float32
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			colors = append(colors, [3]float32{
				float32(r) / 0xffff,
				float32(g) / 0xffff,
				float32(b) / 0xffff,
			})
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

	sums := make([]float32, len(colors))
	intermediate := make([][3]float32, len(colors))
	mapY(func(y int) {
		for x := 0; x < width; x++ {
			var kernelSum float32
			var colorSum [3]float32
			for dy := -g.Radius; dy <= g.Radius; dy++ {
				sumY := y + dy
				if sumY >= 0 && sumY < height {
					k := g.Data[dy+g.Radius]
					c := colors[x+width*sumY]
					for i, x := range c {
						colorSum[i] += x * k
					}
					kernelSum += k
				}
			}
			idx := x + y*b.Dx()
			intermediate[idx] = colorSum
			sums[idx] = kernelSum
		}
	})

	result := image.NewRGBA(image.Rect(0, 0, width, height))
	mapY(func(y int) {
		for x := 0; x < b.Dx(); x++ {
			var kernelSum float32
			var colorSum [3]float32
			for dx := -g.Radius; dx <= g.Radius; dx++ {
				sumX := x + dx
				if sumX >= 0 && sumX < width {
					k := g.Data[dx+g.Radius]
					c := intermediate[sumX+width*y]
					for i, x := range c {
						colorSum[i] += x * k
					}
					kernelSum += k * sums[x+width*y]
				}
			}
			for i := range colorSum {
				colorSum[i] /= kernelSum
			}
			result.SetRGBA(x, y, color.RGBA{
				R: uint8(colorSum[0] * 255.999),
				G: uint8(colorSum[1] * 255.999),
				B: uint8(colorSum[2] * 255.999),
				A: 0xff,
			})
		}
	})

	return result
}
