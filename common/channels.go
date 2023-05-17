package common

import "github.com/multiversx/mx-chain-core-go/core"

// GetClosedUnbufferedChannel returns an instance of a 'chan struct{}' that is already closed
func GetClosedUnbufferedChannel() chan struct{} {
	ch := make(chan struct{})
	close(ch)

	return ch
}

// SafelyCloseKeyValueHolderChan will close the channel if not nil
func SafelyCloseKeyValueHolderChan(ch chan core.KeyValueHolder) {
	if ch != nil {
		close(ch)
	}
}
