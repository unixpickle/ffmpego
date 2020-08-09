package ffmpego

import (
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// VideoInfo is information about an encoded video.
type VideoInfo struct {
	Width  int
	Height int
	FPS    float64
}

// GetVideoInfo gets information about a video file.
func GetVideoInfo(path string) (info *VideoInfo, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "get video info")
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

	var foundFPS, foundSize bool
	result := &VideoInfo{}

	fpsExp := regexp.MustCompilePOSIX(" ([0-9\\.]*) fps,")
	sizeExp := regexp.MustCompilePOSIX(" ([0-9]+)x([0-9]+)(,| )")
	for _, line := range lines {
		if !strings.Contains(line, "Video:") {
			continue
		}
		if match := fpsExp.FindStringSubmatch(line); match != nil {
			foundFPS = true
			fps, err := strconv.ParseFloat(match[1], 0)
			if err != nil {
				return nil, errors.Wrap(err, "parse FPS")
			}
			result.FPS = fps
		}
		if match := sizeExp.FindStringSubmatch(line); match != nil {
			foundSize = true
			var size [2]int
			for i, s := range match[1:3] {
				n, err := strconv.Atoi(s)
				if err != nil {
					return nil, errors.Wrap(err, "parse dimensions")
				}
				size[i] = n
			}
			result.Width = size[0]
			result.Height = size[1]
		}
	}

	if !foundFPS {
		return nil, errors.New("could not find fps in output")
	}
	if !foundSize {
		return nil, errors.New("could not find dimensions in output")
	}
	return result, nil
}

func infoOutputLines(path string) ([]string, error) {
	cmd := exec.Command("ffmpeg", "-i", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// An error exit status is expected, since we didn't do any
		// transcoding, we are just using the video info.
		err = errors.Cause(err)
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, err
		}
	}
	return strings.Split(string(out), "\n"), nil
}
