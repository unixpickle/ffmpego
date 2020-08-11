package ffmpego

import (
	"io"
	"math/rand"
	"path/filepath"
	"testing"
)

func TestAudioReader(t *testing.T) {
	reader, err := NewAudioReader(filepath.Join("test_data", "test_audio.wav"))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	numSamples := 0
	for {
		chunk := make([]float64, rand.Intn(100)+100)
		n, err := reader.ReadSamples(chunk)
		if n != len(chunk) {
			if err == nil {
				t.Error("expected error if fewer bytes are read")
			}
		}
		numSamples += n
		if err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
	}
	if numSamples != 8000 {
		t.Errorf("incorrect number of samples: %d", numSamples)
	}
}
