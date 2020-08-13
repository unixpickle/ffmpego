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
	testVideoReader(t, reader, 24)
}

func TestVideoReaderResampled(t *testing.T) {
	reader, err := NewVideoReaderResampled(filepath.Join("test_data", "test_video.mp4"), 20)
	if err != nil {
		t.Fatal(err)
	}
	testVideoReader(t, reader, 40)
}

func testVideoReader(t *testing.T, reader *VideoReader, expectedFrames int) {
	defer func() {
		if err := reader.Close(); err != nil {
			t.Error(err)
		}
	}()
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
	if numFrames != expectedFrames {
		t.Errorf("incorrect number of frames: %d", numFrames)
	}
}
