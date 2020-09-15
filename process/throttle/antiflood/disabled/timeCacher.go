package disabled

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.TimeCacher = (*TimeCache)(nil)

// TimeCache is a mock implementation of TimeCacher
type TimeCache struct {
}

// Add does nothing
func (tc *TimeCache) Add(_ string) error {
	return nil
}

// Upsert does nothing
func (tc *TimeCache) Upsert(_ string, _ time.Duration) error {
	return nil
}

// Sweep does nothing
func (tc *TimeCache) Sweep() {
}

// Has outputs false (all keys are white listed)
func (tc *TimeCache) Has(_ string) bool {
	return false
}

// Len does nothing
func (tc *TimeCache) Len() int {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (tc *TimeCache) IsInterfaceNil() bool {
	return tc == nil
}
