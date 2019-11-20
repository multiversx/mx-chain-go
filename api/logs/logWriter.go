package logs

import (
	"sync"
)

const msgQueueSize = 100

type logWriter struct {
	mutChanClosed sync.RWMutex
	chanClosed    bool
	dataChan      chan []byte
}

// NewLogWriter creates a new chan-based
func NewLogWriter() *logWriter {
	return &logWriter{
		dataChan: make(chan []byte, msgQueueSize),
	}
}

// Write will try to output the data on the channel
func (lw *logWriter) Write(p []byte) (n int, err error) {
	if p == nil {
		return 0, nil
	}

	lw.mutChanClosed.RLock()
	defer lw.mutChanClosed.RUnlock()

	if lw.chanClosed {
		return 0, ErrWriterClosed
	}

	select {
	case lw.dataChan <- p:
		return len(p), nil
	default:
		return 0, ErrWriterBusy
	}
}

// Close closes the writer by closing the underlying chan
// Subsequent calls of this method will return ErrWriterClosed
func (lw *logWriter) Close() error {
	lw.mutChanClosed.Lock()
	defer lw.mutChanClosed.Unlock()

	if lw.chanClosed {
		return ErrWriterClosed
	}

	lw.chanClosed = true
	close(lw.dataChan)

	return nil
}

// ReadBlocking will try to read from the data channel.
// It blocks until a new byte slice have been written with write method
// or the chan is closed
func (lw *logWriter) ReadBlocking() ([]byte, bool) {
	data, ok := <-lw.dataChan

	return data, ok
}
