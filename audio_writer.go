package ffmpego

import (
	"encoding/binary"
	"io"
	"os/exec"
	"strconv"

	"github.com/pkg/errors"
)

// An AudioWriter encodes an audio file using ffmpeg.
type AudioWriter struct {
	command *exec.Cmd
	writer  io.WriteCloser
}

// NewAudioWriter creates a AudioWriter which is encoding
// mono-channel audio to the given file.
func NewAudioWriter(path string, frequency int) (*AudioWriter, error) {
	vw, err := newAudioWriter(path, frequency)
	if err != nil {
		err = errors.Wrap(err, "write audio")
	}
	return vw, err
}

func newAudioWriter(path string, frequency int) (*AudioWriter, error) {
	stream, err := CreateChildStream(false)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(
		"ffmpeg",
		"-y",
		// Audio format
		"-ar", strconv.Itoa(frequency), "-ac", "1", "-f", "s16le",
		// Audio parameters
		"-probesize", "32", "-thread_queue_size", "60", "-i", stream.ResourceURL(),
		// Output parameters
		"-pix_fmt", "yuv420p", path,
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
	return &AudioWriter{
		command: cmd,
		writer:  writer,
	}, nil
}

// WriteSamples writes audio samples to the file.
//
// The samples should be in the range [-1, 1].
func (v *AudioWriter) WriteSamples(samples []float64) error {
	intData := make([]int16, len(samples))
	for i, x := range samples {
		intData[i] = int16(x * (1<<15 - 1))
	}
	if err := binary.Write(v.writer, binary.LittleEndian, intData); err != nil {
		return errors.Wrap(err, "write samples")
	}
	return nil
}

// Close closes the audio file and waits for encoding to
// complete.
func (v *AudioWriter) Close() error {
	v.writer.Close()
	err := v.command.Wait()
	if err != nil {
		return errors.Wrap(err, "close audio writer")
	}
	return nil
}
