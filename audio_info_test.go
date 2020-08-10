package ffmpego

import (
	"path/filepath"
	"testing"
)

func TestAudioInfo(t *testing.T) {
	info, err := GetAudioInfo(filepath.Join("test_data", "test_audio.wav"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Frequency != 8000 {
		t.Errorf("expected frequency 8000 but got %d", info.Frequency)
	}
}
