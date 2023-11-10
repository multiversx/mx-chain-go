package components

import (
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
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

// ProcessMessage -
func (i *interceptor) ProcessMessage(_ *pubsub.Message) error {
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
