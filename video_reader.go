package ffmpego

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"os/exec"

	"github.com/pkg/errors"
)

// A VideoReader decodes a video file using ffmpeg.
type VideoReader struct {
	command *exec.Cmd
	reader  io.ReadCloser
	info    *VideoInfo
}

func NewVideoReader(path string) (*VideoReader, error) {
	vr, err := newVideoReader(path, -1)
	if err != nil {
		err = errors.Wrap(err, "read video")
	}
	return vr, err
}

// NewVideoReaderResampled creates a VideoReader that
// automatically changes the input frame rate.
func NewVideoReaderResampled(path string, fps float64) (*VideoReader, error) {
	if fps <= 0 {
		panic("FPS must be positive")
	}
	vr, err := newVideoReader(path, fps)
	if err != nil {
		err = errors.Wrap(err, "read video")
	}
	return vr, err
}

func newVideoReader(path string, resampleFPS float64) (*VideoReader, error) {
	info, err := GetVideoInfo(path)
	if err != nil {
		return nil, err
	}

	if resampleFPS > 0 {
		info.FPS = resampleFPS
	}

	stream, err := CreateChildStream(true)
	if err != nil {
		return nil, err
	}

	args := []string{
		"-i", path,
		"-f", "rawvideo", "-pix_fmt", "rgb24",
	}
	if resampleFPS > 0 {
		args = append(args, "-filter:v", fmt.Sprintf("fps=fps=%f", resampleFPS))
	}
	args = append(args, stream.ResourceURL())

	cmd := exec.Command("ffmpeg", args...)
	cmd.ExtraFiles = stream.ExtraFiles()
	if err := cmd.Start(); err != nil {
		stream.Cancel()
		return nil, err
	}
	reader, err := stream.Connect()
	if err != nil {
		cmd.Process.Kill()
		return nil, err
	}
	return &VideoReader{
		command: cmd,
		reader:  reader,
		info:    info,
	}, nil
}

// VideoInfo gets information about the current video.
func (v *VideoReader) VideoInfo() *VideoInfo {
	return v.info
}

// ReadFrame reads the next frame from the video.
//
// If the video is finished decoding, nil will be returned
// along with io.EOF.
func (v *VideoReader) ReadFrame() (image.Image, error) {
	buf := make([]byte, 3*v.info.Width*v.info.Height)
	if _, err := io.ReadFull(v.reader, buf); err != nil {
		return nil, err
	}
	img := image.NewRGBA(image.Rect(0, 0, v.info.Width, v.info.Height))
	for y := 0; y < v.info.Height; y++ {
		for x := 0; x < v.info.Width; x++ {
			rgb := buf[:3]
			buf = buf[3:]
			img.Set(x, y, &color.RGBA{
				R: rgb[0],
				G: rgb[1],
				B: rgb[2],
				A: 0xff,
			})
		}
	}
	return img, nil
}

// Close stops the decoding process and closes all
// associated files.
func (v *VideoReader) Close() error {
	// When we close the pipe, the subprocess should terminate
	// (possibly with an error) because it cannot write.
	v.reader.Close()
	v.command.Wait()
	return nil
}
