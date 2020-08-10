package ffmpego

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// AudioInfo stores information about an audio file.
type AudioInfo struct {
	// Frequency stores the frequency in Hz.
	Frequency int
}

// GetAudioInfo gets information about a audio file.
func GetAudioInfo(path string) (info *AudioInfo, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "get audio info")
		}
	}()

	// Make sure file exists so we can give a clean error
	// message in this case, instead of depending on ffmpeg.
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

	lines, err := infoOutputLines(path)
	if err != nil {
		return nil, err
	}

	var foundFreq bool
	result := &AudioInfo{}

	freqExp := regexp.MustCompilePOSIX(" ([0-9\\.]*) Hz,")
	for _, line := range lines {
		if !strings.Contains(line, "Audio:") {
			continue
		}
		if match := freqExp.FindStringSubmatch(line); match != nil {
			foundFreq = true
			freq, err := strconv.Atoi(match[1])
			if err != nil {
				return nil, errors.Wrap(err, "parse frequency")
			}
			result.Frequency = freq
		}
	}

	if !foundFreq {
		return nil, errors.New("could not find frequency in output")
	}
	return result, nil
}
