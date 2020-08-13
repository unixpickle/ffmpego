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
	testAudioReader(t, reader, 8000)
}

func TestAudioReaderResampled(t *testing.T) {
	reader, err := NewAudioReaderResampled(filepath.Join("test_data", "test_audio.wav"), 16000)
	if err != nil {
		t.Fatal(err)
	}
	testAudioReader(t, reader, 16000)
}

func testAudioReader(t *testing.T, reader *AudioReader, expectedSamples int) {
	defer func() {
		if err := reader.Close(); err != nil {
			t.Error(err)
		}
	}()
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
	if numSamples != expectedSamples {
		t.Errorf("incorrect number of samples: %d", numSamples)
	}
}
