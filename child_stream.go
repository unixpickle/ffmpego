package ffmpego

import (
	"io"
	"net"
	"os"
	"runtime"
	"time"
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
	if runtime.GOOS == "windows" {
		return NewChildSocketStream()
	} else {
		return NewChildPipeStream(reading)
	}
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
		c.parentPipe.Close()
		return nil, err
	}
	return c.parentPipe, nil
}

func (c *ChildPipeStream) Cancel() error {
	c.childPipe.Close()
	return c.parentPipe.Close()
}

// A ChildSocketStream uses a TCP socket to communicate
// with subprocesses.
//
// This should be supported on all operating systems, but
// some may prevent process from listening on sockets.
type ChildSocketStream struct {
	listener *net.TCPListener
}

// NewChildSocketStream creates a ChildSocketStream.
func NewChildSocketStream() (*ChildSocketStream, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &ChildSocketStream{
		listener: listener,
	}, nil
}

func (c *ChildSocketStream) ExtraFiles() []*os.File {
	return nil
}

func (c *ChildSocketStream) ResourceURL() string {
	return "tcp://" + c.listener.Addr().String()
}

func (c *ChildSocketStream) Connect() (io.ReadWriteCloser, error) {
	if err := c.listener.SetDeadline(time.Now().Add(time.Second * 10)); err != nil {
		c.listener.Close()
		return nil, err
	}
	conn, err := c.listener.Accept()
	c.listener.Close()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *ChildSocketStream) Cancel() error {
	return c.listener.Close()
}
