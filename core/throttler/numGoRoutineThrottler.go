package throttler

import (
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go/core"
)

// NumGoRoutineThrottler can limit the number of go routines launched
type NumGoRoutineThrottler struct {
	max     int32
	counter int32
}

// NewNumGoRoutineThrottler creates a new num go routine throttler instance
func NewNumGoRoutineThrottler(max int32) (*NumGoRoutineThrottler, error) {
	if max <= 0 {
		return nil, core.ErrNotPositiveValue
	}

	return &NumGoRoutineThrottler{
		max: max,
	}, nil
}

// CanProcess returns true if current counter is less than max
func (ngrt *NumGoRoutineThrottler) CanProcess() bool {
	valCounter := atomic.LoadInt32(&ngrt.counter)

	return valCounter < ngrt.max
}

// StartProcessing will increment current counter
func (ngrt *NumGoRoutineThrottler) StartProcessing() {
	atomic.AddInt32(&ngrt.counter, 1)
}

// EndProcessing will decrement current counter
func (ngrt *NumGoRoutineThrottler) EndProcessing() {
	atomic.AddInt32(&ngrt.counter, -1)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ngrt *NumGoRoutineThrottler) IsInterfaceNil() bool {
	if ngrt == nil {
		return true
	}
	return false
}
