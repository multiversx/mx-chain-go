package common

import "github.com/multiversx/mx-chain-core-go/core"

// GetClosedUnbufferedChannel returns an instance of a 'chan struct{}' that is already closed
func GetClosedUnbufferedChannel() chan struct{} {
	ch := make(chan struct{})
	close(ch)

	return ch
}

// CloseKeyValueHolderChan will close the channel if not nil
func CloseKeyValueHolderChan(ch chan core.KeyValueHolder) {
	if ch != nil {
		close(ch)
	}
}
