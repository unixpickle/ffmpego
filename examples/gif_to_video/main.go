package main

import (
	"fmt"
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

	bounds := gifImage.Image[0].Bounds()
	writer, err := ffmpego.NewVideoWriter(outputFile, bounds.Dx(), bounds.Dy(), fps)
	essentials.Must(err)
	defer func() {
		essentials.Must(writer.Close())
	}()

	for _, frame := range gifImage.Image {
		essentials.Must(writer.WriteFrame(frame))
	}
}
