package closing

import (
	"sync"
)

type safeChanCloser struct {
	mut     sync.Mutex
	chClose chan struct{}
}

// NewSafeChanCloser returns a safe chan closer instance
func NewSafeChanCloser() *safeChanCloser {
	return &safeChanCloser{
		chClose: make(chan struct{}),
	}
}

// Close will close the channel in a safe concurrent manner
func (closer *safeChanCloser) Close() {
	closer.mut.Lock()
	defer closer.mut.Unlock()

	select {
	case <-closer.chClose:
		return
	default:
		close(closer.chClose)
	}
}

// ChanClose returns the closing channel
func (closer *safeChanCloser) ChanClose() <-chan struct{} {
	return closer.chClose
}

// IsInterfaceNil returns true if there is no value under the interface
func (closer *safeChanCloser) IsInterfaceNil() bool {
	return closer == nil
}
