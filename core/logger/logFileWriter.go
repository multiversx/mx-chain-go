package logger

import (
	"fmt"
	"io"
	"sync"
)

// LogFileWriter is a custom io.Writer implementation suited for our application.
// While a writer is not provided, it will write to standard output and record all entries in a slice.
// After the writer will be provided, the buffer slice will be set to nil and the provided writer will be
// called from this point on.
type LogFileWriter struct {
	writer     io.Writer
	buffer     [][]byte
	writerLock sync.RWMutex
	bufferLock sync.Mutex
}

// Writer will return the current writer
func (lfw *LogFileWriter) Writer() io.Writer {
	lfw.writerLock.RLock()
	defer lfw.writerLock.RUnlock()
	return lfw.writer
}

// SetWriter will set the writer
func (lfw *LogFileWriter) SetWriter(writer io.Writer) {
	lfw.writerLock.Lock()
	defer lfw.writerLock.Unlock()
	lfw.writer = writer
}

// Write handles the output of this writer
func (lfw *LogFileWriter) Write(p []byte) (n int, err error) {
	lfw.writerLock.RLock()
	writer := lfw.writer
	lfw.writerLock.RUnlock()

	if writer == nil {
		lfw.bufferLock.Lock()
		lfw.buffer = append(lfw.buffer, p)
		lfw.bufferLock.Unlock()
		return 0, nil
	}

	if lfw.buffer == nil {
		return writer.Write(p)
	}

	lfw.bufferLock.Lock()
	for _, logEntry := range lfw.buffer {
		_, err = writer.Write(logEntry)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	lfw.buffer = nil
	lfw.bufferLock.Unlock()
	return writer.Write(p)
}
