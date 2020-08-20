package main

import (
	"fmt"
	"image"
	"io"
	"log"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/ffmpego"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: reverse_video <input.mp4> <output.mp4>")
		os.Exit(1)
	}
	inputFile := os.Args[1]
	outputFile := os.Args[2]

	reader, err := ffmpego.NewVideoReader(inputFile)
	essentials.Must(err)
	defer reader.Close()
	info := reader.VideoInfo()

	log.Println("Reading video...")
	var frames []image.Image
	for {
		frame, err := reader.ReadFrame()
		if err == io.EOF {
			break
		}
		essentials.Must(err)
		frames = append(frames, frame)
	}

	log.Println("Encoding video...")
	writer, err := ffmpego.NewVideoWriter(outputFile, info.Width, info.Height, info.FPS)
	essentials.Must(err)
	defer writer.Close()
	for i := len(frames) - 1; i >= 0; i-- {
		essentials.Must(writer.WriteFrame(frames[i]))
	}
}
