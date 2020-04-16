package throttler

import (
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ dataRetriever.ResolverThrottler = (*NumGoRoutinesThrottler)(nil)
var _ process.InterceptorThrottler = (*NumGoRoutinesThrottler)(nil)

// NumGoRoutinesThrottler can limit the number of go routines launched
type NumGoRoutinesThrottler struct {
	max     int32
	counter int32
}

// NewNumGoRoutinesThrottler creates a new num go routine throttler instance
func NewNumGoRoutinesThrottler(max int32) (*NumGoRoutinesThrottler, error) {
	if max <= 0 {
		return nil, core.ErrNotPositiveValue
	}

	return &NumGoRoutinesThrottler{
		max: max,
	}, nil
}

// CanProcess returns true if current counter is less than max
func (ngrt *NumGoRoutinesThrottler) CanProcess() bool {
	valCounter := atomic.LoadInt32(&ngrt.counter)

	return valCounter < ngrt.max
}

// StartProcessing will increment current counter
func (ngrt *NumGoRoutinesThrottler) StartProcessing() {
	atomic.AddInt32(&ngrt.counter, 1)
}

// EndProcessing will decrement current counter
func (ngrt *NumGoRoutinesThrottler) EndProcessing() {
	atomic.AddInt32(&ngrt.counter, -1)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ngrt *NumGoRoutinesThrottler) IsInterfaceNil() bool {
	return ngrt == nil
}
