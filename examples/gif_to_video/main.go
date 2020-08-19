package main

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/ffmpego"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: gif_to_video <input.gif> <output.mp4>")
		os.Exit(1)
	}
	inputFile := os.Args[1]
	outputFile := os.Args[2]

	r, err := os.Open(inputFile)
	essentials.Must(err)
	defer r.Close()

	gifImage, err := gif.DecodeAll(r)
	essentials.Must(err)

	delay := 0.0
	for _, frameDelay := range gifImage.Delay {
		delay += float64(frameDelay) / float64(100*len(gifImage.Delay))
	}
	fps := 1 / delay

	bounds := BoundsForGIF(gifImage)
	writer, err := ffmpego.NewVideoWriter(outputFile, bounds.Dx(), bounds.Dy(), fps)
	essentials.Must(err)
	defer func() {
		essentials.Must(writer.Close())
	}()

	FramesFromGIF(gifImage, func(img image.Image) {
		essentials.Must(writer.WriteFrame(img))
	})
}

func BoundsForGIF(g *gif.GIF) image.Rectangle {
	result := g.Image[0].Bounds()
	for _, frame := range g.Image {
		result = result.Union(frame.Bounds())
	}
	return result
}

func FramesFromGIF(g *gif.GIF, f func(image.Image)) {
	out := image.NewRGBA(BoundsForGIF(g))
	previous := image.NewRGBA(BoundsForGIF(g))
	for i, frame := range g.Image {
		disposal := g.Disposal[i]
		switch disposal {
		case 0:
			clearImage(out)
			clearImage(previous)
			drawImageWithBackground(out, frame, out)
			drawImageWithBackground(previous, frame, previous)
		case gif.DisposalNone:
			drawImageWithBackground(out, frame, out)
		case gif.DisposalPrevious:
			drawImageWithBackground(out, frame, previous)
		case gif.DisposalBackground:
			bgColor := g.Config.ColorModel.(color.Palette)[g.BackgroundIndex]
			fillImage(out, bgColor)
			drawImageWithBackground(out, frame, out)
		}
		f(out)
	}
}

func drawImageWithBackground(dst *image.RGBA, src, bg image.Image) {
	b := src.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			px := src.At(x, y)
			_, _, _, a := px.RGBA()
			if a == 0 {
				px = bg.At(x, y)
			}
			dst.Set(x, y, px)
		}
	}
}

func clearImage(dst *image.RGBA) {
	for i := range dst.Pix {
		dst.Pix[i] = 0
	}
}

func fillImage(dst *image.RGBA, c color.Color) {
	b := dst.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(x, y, c)
		}
	}
}
