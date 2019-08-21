package throttler

import (
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go/process"
)

// NumThrottler implements InterceptorThrottler and limits the message processing if current counter > max
type NumThrottler struct {
	max     int32
	counter int32
}

// NewNumThrottler creates a new num throttler instance
func NewNumThrottler(max int32) (*NumThrottler, error) {
	if max <= 0 {
		return nil, process.ErrNotPositiveValue
	}

	return &NumThrottler{
		max: max,
	}, nil
}

// CanProcess returns true if current counter is less than max
func (nt *NumThrottler) CanProcess() bool {
	valCounter := atomic.LoadInt32(&nt.counter)

	return valCounter < nt.max
}

// StartProcessing will increment current counter
func (nt *NumThrottler) StartProcessing() {
	atomic.AddInt32(&nt.counter, 1)
}

// EndProcessing will decrement current counter
func (nt *NumThrottler) EndProcessing() {
	atomic.AddInt32(&nt.counter, -1)
}
