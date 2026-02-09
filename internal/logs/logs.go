package logs

import (
	"io"
	"os"
)

// LogStreamer defines the interface for streaming detailed logs
type LogStreamer interface {
	io.Writer
	Close() error
}

// ConsoleStreamer simple streamer that writes to stdout
type ConsoleStreamer struct{}

func (c *ConsoleStreamer) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

func (c *ConsoleStreamer) Close() error {
	return nil
}
