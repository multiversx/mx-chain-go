package errChan

import "sync"

type errChan struct {
	ch         chan error
	closed     bool
	closeMutex sync.Mutex
}

// NewErrChan creates a new errChan
func NewErrChan() *errChan {
	return &errChan{
		ch:     make(chan error, 1),
		closed: false,
	}
}

// WriteInChanNonBlocking will send the given error on the channel if the chan is not blocked
func (ec *errChan) WriteInChanNonBlocking(err error) {
	select {
	case ec.ch <- err:
	default:
	}
}

// ReadFromChanNonBlocking will read from the channel, or return nil if no error was sent on the channel
func (ec *errChan) ReadFromChanNonBlocking() error {
	select {
	case err := <-ec.ch:
		return err
	default:
		return nil
	}
}

// Close will close the channel
func (ec *errChan) Close() {
	ec.closeMutex.Lock()
	defer ec.closeMutex.Unlock()

	if ec.closed {
		return
	}

	if ec.ch == nil {
		return
	}

	close(ec.ch)
	ec.closed = true
}

// Len returns the length of the channel
func (ec *errChan) Len() int {
	return len(ec.ch)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ec *errChan) IsInterfaceNil() bool {
	return ec == nil
}
