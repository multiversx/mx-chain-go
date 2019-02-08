package spos

import (
	"sync"
)

// RoundThreshold defines the minimum agreements needed for each subround to consider the subround finished.
// (Ex: PBFT threshold has 2 / 3 + 1 agreements)
type RoundThreshold struct {
	threshold map[int]int
	mut       sync.RWMutex
}

// NewRoundThreshold creates a new RoundThreshold object
func NewRoundThreshold() *RoundThreshold {
	rthr := RoundThreshold{}
	rthr.threshold = make(map[int]int)
	return &rthr
}

// Threshold returns the threshold of agreements needed in the given subround id
func (rthr *RoundThreshold) Threshold(subroundId int) int {
	rthr.mut.RLock()
	retcode := rthr.threshold[subroundId]
	rthr.mut.RUnlock()
	return retcode
}

// SetThreshold sets the threshold of agreements needed in the given subround id
func (rthr *RoundThreshold) SetThreshold(subroundId int, threshold int) {
	rthr.mut.Lock()
	rthr.threshold[subroundId] = threshold
	rthr.mut.Unlock()
}
