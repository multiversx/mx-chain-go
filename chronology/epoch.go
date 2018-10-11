package chronology

import (
	"time"

	"github.com/davecgh/go-spew/spew"
)

type Epoch struct {
	index       int
	genesisTime time.Time
}

func NewEpoch(index int, genesisTime time.Time) Epoch {
	epc := Epoch{index, genesisTime}
	return epc
}

func (epc *Epoch) Print() {
	spew.Dump(epc)
}
