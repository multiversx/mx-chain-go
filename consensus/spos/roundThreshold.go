package spos

import (
	"sync"
)

// roundThreshold defines the minimum agreements needed for each subround to consider the subround finished.
// (Ex: PBFT threshold has 2 / 3 + 1 agreements)
type roundThreshold struct {
	threshold         map[int]int
	fallbackThreshold map[int]int
	mut               sync.RWMutex
}

// NewRoundThreshold creates a new roundThreshold object
func NewRoundThreshold() *roundThreshold {
	rthr := roundThreshold{}
	rthr.threshold = make(map[int]int)
	rthr.fallbackThreshold = make(map[int]int)
	return &rthr
}

// Threshold returns the threshold of agreements needed in the given subround id
func (rthr *roundThreshold) Threshold(subroundId int) int {
	rthr.mut.RLock()
	retcode := rthr.threshold[subroundId]
	rthr.mut.RUnlock()
	return retcode
}

// SetThreshold sets the threshold of agreements needed in the given subround id
func (rthr *roundThreshold) SetThreshold(subroundId int, threshold int) {
	rthr.mut.Lock()
	rthr.threshold[subroundId] = threshold
	rthr.mut.Unlock()
}

// FallbackThreshold returns the fallback threshold of agreements needed in the given subround id
func (rthr *roundThreshold) FallbackThreshold(subroundId int) int {
	rthr.mut.RLock()
	retcode := rthr.fallbackThreshold[subroundId]
	rthr.mut.RUnlock()
	return retcode
}

// SetFallbackThreshold sets the fallback threshold of agreements needed in the given subround id
func (rthr *roundThreshold) SetFallbackThreshold(subroundId int, threshold int) {
	rthr.mut.Lock()
	rthr.fallbackThreshold[subroundId] = threshold
	rthr.mut.Unlock()
}
