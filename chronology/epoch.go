package chronology

import (
	"time"

	"github.com/davecgh/go-spew/spew"
)

// Epoch defines the data needed by the epoch
type Epoch struct {
	index       int
	genesisTime time.Time
}

// NewEpoch defines a new Epoch object
func NewEpoch(index int, genesisTime time.Time) Epoch {
	epc := Epoch{index, genesisTime}
	return epc
}

// Print method just spew to the console the Epoch object in some pretty format
func (epc *Epoch) Print() {
	spew.Dump(epc)
}
