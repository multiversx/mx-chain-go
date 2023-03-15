package errChan

import "sync"

type errChanWrapper struct {
	ch         chan error
	closed     bool
	closeMutex sync.RWMutex
}

// NewErrChanWrapper creates a new errChanWrapper
func NewErrChanWrapper() *errChanWrapper {
	return &errChanWrapper{
		ch:     make(chan error, 1),
		closed: false,
	}
}

// WriteInChanNonBlocking will send the given error on the channel if the chan is not blocked
func (ec *errChanWrapper) WriteInChanNonBlocking(err error) {
	ec.closeMutex.RLock()
	defer ec.closeMutex.RUnlock()

	if ec.closed {
		return
	}

	select {
	case ec.ch <- err:
	default:
	}
}

// ReadFromChanNonBlocking will read from the channel, or return nil if no error was sent on the channel
func (ec *errChanWrapper) ReadFromChanNonBlocking() error {
	select {
	case err := <-ec.ch:
		return err
	default:
		return nil
	}
}

// Close will close the channel
func (ec *errChanWrapper) Close() {
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
func (ec *errChanWrapper) Len() int {
	return len(ec.ch)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ec *errChanWrapper) IsInterfaceNil() bool {
	return ec == nil
}
