package ffmpego

import (
	"fmt"
	"image"
	"io"
	"os/exec"

	"github.com/pkg/errors"
)

// A VideoWriter encodes a video file using ffmpeg.
type VideoWriter struct {
	command *exec.Cmd
	writer  io.WriteCloser
	width   int
	height  int
}

// NewVideoWriter creates a VideoWriter which is encoding
// to the given file.
func NewVideoWriter(path string, width, height int, fps float64) (*VideoWriter, error) {
	vw, err := newVideoWriter(path, width, height, fps)
	if err != nil {
		err = errors.Wrap(err, "write video")
	}
	return vw, err
}

func newVideoWriter(path string, width, height int, fps float64) (*VideoWriter, error) {
	stream, err := CreateChildStream(false)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		// Video format
		"-r", fmt.Sprintf("%f", fps),
		"-s", fmt.Sprintf("%dx%d", width, height),
		"-pix_fmt", "rgb24", "-f", "rawvideo",
		// Video parameters
		"-probesize", "32", "-thread_queue_size", "10000", "-i", stream.ResourceURL(),
		// Output parameters
		"-c:v", "libx264", "-preset", "fast", "-crf", "18",
		"-pix_fmt", "yuv420p", "-vf", "pad=ceil(iw/2)*2:ceil(ih/2)*2",
		path,
	)
	cmd.ExtraFiles = stream.ExtraFiles()
	if err := cmd.Start(); err != nil {
		stream.Cancel()
		return nil, err
	}
	writer, err := stream.Connect()
	if err != nil {
		cmd.Process.Kill()
		return nil, err
	}
	return &VideoWriter{
		command: cmd,
		writer:  writer,
		width:   width,
		height:  height,
	}, nil
}

// WriteFrame adds a frame to the current video.
func (v *VideoWriter) WriteFrame(img image.Image) error {
	bounds := img.Bounds()
	if bounds.Dx() != v.width || bounds.Dy() != v.height {
		return fmt.Errorf("write frame: image size (%dx%d) does not match video size (%dx%d)",
			bounds.Dx(), bounds.Dy(), v.width, v.height)
	}
	data := make([]byte, 0, 3*v.width*v.height)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			data = append(data, uint8(r>>8), uint8(g>>8), uint8(b>>8))
		}
	}
	_, err := v.writer.Write(data)
	if err != nil {
		return errors.Wrap(err, "write frame")
	}
	return nil
}

// Close closes the video file and waits for encoding to
// complete.
func (v *VideoWriter) Close() error {
	v.writer.Close()
	err := v.command.Wait()
	if err != nil {
		return errors.Wrap(err, "close video writer")
	}
	return nil
}
