package components

import (
	"sync/atomic"
	"time"
)

type manualRoundHandler struct {
	index int64
}

// NewManualRoundHandler returns a manual round handler instance
func NewManualRoundHandler() *manualRoundHandler {
	return &manualRoundHandler{}
}

// IncrementIndex will increment the current round index
func (handler *manualRoundHandler) IncrementIndex() {
	atomic.AddInt64(&handler.index, 1)
}

// Index returns the current index
func (handler *manualRoundHandler) Index() int64 {
	return atomic.LoadInt64(&handler.index)
}

// BeforeGenesis returns false
func (handler *manualRoundHandler) BeforeGenesis() bool {
	return false
}

// UpdateRound does nothing as this implementation does not work with real timers
func (handler *manualRoundHandler) UpdateRound(_ time.Time, _ time.Time) {
}

// TimeStamp returns the empty time.Time value
func (handler *manualRoundHandler) TimeStamp() time.Time {
	return time.Time{}
}

// TimeDuration returns a hardcoded value
func (handler *manualRoundHandler) TimeDuration() time.Duration {
	return 0
}

// RemainingTime returns the max time as the start time is not taken into account
func (handler *manualRoundHandler) RemainingTime(_ time.Time, maxTime time.Duration) time.Duration {
	return maxTime
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *manualRoundHandler) IsInterfaceNil() bool {
	return handler == nil
}
