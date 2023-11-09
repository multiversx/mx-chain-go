package components

import (
	"sync/atomic"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/p2p"
)

type interceptor struct {
	numMessages uint64
}

// NewInterceptor -
func NewInterceptor() *interceptor {
	return &interceptor{}
}

// ProcessReceivedMessage -
func (i *interceptor) ProcessReceivedMessage(_ p2p.MessageP2P, _ core.PeerID) error {
	atomic.AddUint64(&i.numMessages, 1)

	return nil
}

// GetNumMessages -
func (i *interceptor) GetNumMessages() uint64 {
	return atomic.LoadUint64(&i.numMessages)
}

// IsInterfaceNil -
func (i *interceptor) IsInterfaceNil() bool {
	return i == nil
}
