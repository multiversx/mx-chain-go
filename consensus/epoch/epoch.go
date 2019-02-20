package epoch

import (
	"time"
)

// epoch defines the data needed by the epoch
type epoch struct {
	index       int
	genesisTime time.Time
}

// NewEpoch defines a new epoch object
func NewEpoch(index int, genesisTime time.Time) *epoch {
	epc := epoch{index, genesisTime}
	return &epc
}

// Index returns the index of the epoch
func (epc *epoch) Index() int {
	return epc.index
}

// GenesisTime returns the creation time of epoch
func (epc *epoch) GenesisTime() time.Time {
	return epc.genesisTime
}
