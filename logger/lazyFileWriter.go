package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
)

// LazyFileWriter is a custom io.Writer implementation suited for our application.
// While a writer is not provided, it will write to standard output and record all entries in a slice.
// After the writer will be provided, the buffer slice will be set to nil and the provided writer will be
// called from this point on.
type LazyFileWriter struct {
	writer io.Writer
	buffer [][]byte
	lock sync.RWMutex
}

// Writer will return the current writer
func (zfw *LazyFileWriter) Writer() io.Writer {
	zfw.lock.RLock()
	defer zfw.lock.RUnlock()
	return zfw.writer
}

// SetWriter will set the writer
func (zfw *LazyFileWriter) SetWriter(writer io.Writer) {
	zfw.lock.Lock()
	defer zfw.lock.Unlock()
	zfw.writer = writer
}

// Write hadles the output of this writer
func (zfw *LazyFileWriter) Write(p []byte) (n int, err error) {
	zfw.lock.RLock()
	writer := zfw.writer
	zfw.lock.RUnlock()

	if writer == nil {
		zfw.buffer = append(zfw.buffer, p)
		return os.Stdout.Write(p)
	}

	if zfw.buffer == nil {
		return writer.Write(p)
	}

	for _, logEntry := range zfw.buffer {
		_, err = writer.Write(logEntry)

		if err != nil {
			fmt.Println(err.Error())
		}
	}

	zfw.buffer = nil
	return writer.Write(p)
}
