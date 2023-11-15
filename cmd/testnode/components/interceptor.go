package components

import (
	"sync"

	"github.com/multiversx/mx-chain-communication-go/p2p"
	"github.com/multiversx/mx-chain-core-go/core"
)

type interceptor struct {
	mut   sync.Mutex
	total int
	delta int
}

// NewInterceptor -
func NewInterceptor() *interceptor {
	return &interceptor{}
}

// ProcessReceivedMessage -
func (i *interceptor) ProcessReceivedMessage(_ p2p.MessageP2P, _ core.PeerID, _ p2p.MessageHandler) error {
	i.mut.Lock()
	i.total++
	i.delta++
	i.mut.Unlock()

	return nil
}

// GetNumMessages -
func (i *interceptor) GetNumMessages() (int, int) {
	i.mut.Lock()
	valTotal := i.total
	valDelta := i.delta
	i.delta = 0
	i.mut.Unlock()

	return valTotal, valDelta
}

// IsInterfaceNil -
func (i *interceptor) IsInterfaceNil() bool {
	return i == nil
}
