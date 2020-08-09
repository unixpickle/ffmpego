package ffmpego

import (
	"path/filepath"
	"testing"
)

func TestVideoInfo(t *testing.T) {
	info, err := GetVideoInfo(filepath.Join("test_data", "test_video.mp4"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Width != 64 {
		t.Errorf("expected width 64 but got %d", info.Width)
	}
	if info.Height != 32 {
		t.Errorf("expected height 32 but got %d", info.Height)
	}
	if info.FPS != 12 {
		t.Errorf("expected FPS 12 but got %f", info.FPS)
	}
}
