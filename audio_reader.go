package ffmpego

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/pkg/errors"
)

// A AudioReader decodes an audio file using ffmpeg.
type AudioReader struct {
	command *exec.Cmd
	inPipe  *os.File
	info    *AudioInfo
}

func NewAudioReader(path string) (*AudioReader, error) {
	vr, err := newAudioReader(path, -1)
	if err != nil {
		err = errors.Wrap(err, "read audio")
	}
	return vr, err
}

// NewAudioReaderResampled creates an AudioReader that
// automatically changes the input frequency.
func NewAudioReaderResampled(path string, frequency int) (*AudioReader, error) {
	if frequency <= 0 {
		panic("frequency must be positive")
	}
	vr, err := newAudioReader(path, frequency)
	if err != nil {
		err = errors.Wrap(err, "read audio")
	}
	return vr, err
}

func newAudioReader(path string, forceFrequency int) (*AudioReader, error) {
	info, err := GetAudioInfo(path)
	if err != nil {
		return nil, err
	}
	if forceFrequency > 0 {
		info.Frequency = forceFrequency
	}

	inPipe, childPipe, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(
		"ffmpeg",
		"-i", path,
		"-f", "s16le",
		"-ar", strconv.Itoa(info.Frequency),
		"-ac", "1",
		"pipe:3",
	)
	cmd.ExtraFiles = []*os.File{childPipe}
	if err := cmd.Start(); err != nil {
		inPipe.Close()
		childPipe.Close()
	}
	childPipe.Close()
	return &AudioReader{
		command: cmd,
		inPipe:  inPipe,
		info:    info,
	}, nil
}

// AudioInfo gets information about the current video.
func (a *AudioReader) AudioInfo() *AudioInfo {
	return a.info
}

// ReadSamples reads up to len(samples) from the file.
//
// Returns the number of samples actually read, along with
// an error if one was encountered.
//
// If fewer samples than len(out) are read, an error must
// be returned.
// At the end of decoding, io.EOF is returned.
func (a *AudioReader) ReadSamples(out []float64) (int, error) {
	buf := make([]byte, 2*len(out))
	n, err := io.ReadFull(a.inPipe, buf)
	if err != nil {
		if err == io.ErrUnexpectedEOF || err == io.EOF {
			if n%2 == 0 {
				err = io.EOF
			} else {
				err = io.ErrUnexpectedEOF
				n -= 1
			}
		}
	}
	if n%2 != 0 {
		n -= 1
	}
	data := make([]int16, n/2)
	binary.Read(bytes.NewReader(buf[:n]), binary.LittleEndian, data)
	for i, x := range data {
		out[i] = float64(x) / float64(1<<15-1)
	}
	return len(data), err
}

// Close stops the decoding process and closes all
// associated files.
func (a *AudioReader) Close() error {
	// When we close the pipe, the subprocess should terminate
	// (possibly with an error) because it cannot write.
	a.inPipe.Close()
	a.command.Wait()
	return nil
}
