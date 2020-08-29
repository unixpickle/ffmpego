package ffmpego

import (
	"io"
	"os"
)

// A ChildStream is a connection to an ffmpeg process.
// Typically a ChildStream can only be used for either
// reading or writing, but not both.
//
// If you create a ChildStream, you must either call
// Connect() or Cancel() in order to properly dispose of
// system resources.
//
// Pass the ExtraFiles() to the executed command to ensure
// that it can access the stream.
type ChildStream interface {
	// ExtraFiles returns the files that should be passed
	// to the executed command in order for it to be able
	// to access the stream.
	ExtraFiles() []*os.File

	// ResourceURL gets the URL or filename that the child
	// process can use to access this stream.
	//
	// It is intended to be passed as a CLI option.
	ResourceURL() string

	// Connect should be called once the sub-process is
	// running. If successful, it will return an object
	// that maps to the subprocess.
	//
	// While the return value is a ReadWriter, only either
	// Read or Write should be used.
	//
	// After Connect() is called, you needn't call Cancel()
	// on the ChildStream, but must call Close() on the
	// returned io.ReadWriteCloser.
	Connect() (io.ReadWriteCloser, error)

	// Cancel disposes of resources in this process
	// associated with the stream.
	// This is only intended to be used if Connect()
	// cannot be called.
	Cancel() error
}

// CreateChildStream creates a ChildStream suitable for
// use on the current operating system.
//
// If the reading flag is true, then the stream should be
// read from. Otherwise it should be written to.
func CreateChildStream(reading bool) (ChildStream, error) {
	// TODO: support other stream types here.
	return NewChildPipeStream(reading)
}

// A ChildPipeStream uses a pipe to communicate with
// subprocesses.
//
// This is not supported on Windows.
type ChildPipeStream struct {
	parentPipe *os.File
	childPipe  *os.File
}

// NewChildPipeStream creates a ChildPipeStream.
//
// If the reading flag is true, then the stream should be
// read from. Otherwise it should be written to.
func NewChildPipeStream(reading bool) (*ChildPipeStream, error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	if reading {
		return &ChildPipeStream{
			parentPipe: reader,
			childPipe:  writer,
		}, nil
	} else {
		return &ChildPipeStream{
			parentPipe: writer,
			childPipe:  reader,
		}, nil
	}
}

func (c *ChildPipeStream) ExtraFiles() []*os.File {
	return []*os.File{c.childPipe}
}

func (c *ChildPipeStream) ResourceURL() string {
	return "pipe:3"
}

func (c *ChildPipeStream) Connect() (io.ReadWriteCloser, error) {
	if err := c.childPipe.Close(); err != nil {
		return nil, err
	}
	return c.parentPipe, nil
}

func (c *ChildPipeStream) Cancel() error {
	c.childPipe.Close()
	return c.parentPipe.Close()
}
