package ffmpego

import (
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestAudioWriter(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-audio-writer")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	aw, err := NewAudioWriter(filepath.Join(dir, "out.wav"), 44100)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 1; i++ {
		samples := make([]float64, 8000)
		for t := range samples {
			arg := math.Pi * 2 * 400 * float64(t) / float64(len(samples))
			samples[t] = math.Sin(arg)
		}
		if err := aw.WriteSamples(samples); err != nil {
			aw.Close()
			t.Fatal(err)
		}
	}
	if err := aw.Close(); err != nil {
		t.Fatal(err)
	}
}
