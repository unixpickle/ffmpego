package ffmpego

import (
	"image"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestVideoWriter(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-video-writer")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	vw, err := NewVideoWriter(filepath.Join(dir, "out.mp4"), 50, 50, 12)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 24; i++ {
		frame := image.NewGray(image.Rect(0, 0, 50, 50))
		for j := 0; j < (len(frame.Pix)*i)/24; j++ {
			frame.Pix[j] = 0xff
		}
		if err := vw.WriteFrame(frame); err != nil {
			vw.Close()
			t.Fatal(err)
		}
	}
	if err := vw.Close(); err != nil {
		t.Fatal(err)
	}
}
