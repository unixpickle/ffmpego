package ffmpego

import (
	"io"
	"path/filepath"
	"testing"
)

func TestVideoReader(t *testing.T) {
	reader, err := NewVideoReader(filepath.Join("test_data", "test_video.mp4"))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	numFrames := 0
	for {
		frame, err := reader.ReadFrame()
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		numFrames++
		if frame.Bounds().Dx() != 64 || frame.Bounds().Dy() != 32 {
			t.Error("bad video bounds:", frame.Bounds())
		}
	}
	if numFrames != 24 {
		t.Errorf("incorrect number of frames: %d", numFrames)
	}
}
