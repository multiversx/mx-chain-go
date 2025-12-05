package disabled

import (
	data "github.com/multiversx/mx-chain-core-go/data/stateChange"
	"github.com/multiversx/mx-chain-go/state"
)

type stateAccessesStorer struct {
}

// NewDisabledStateAccessesStorer creates a new disabled state accesses storer
func NewDisabledStateAccessesStorer() state.StateAccessesStorer {
	return &stateAccessesStorer{}
}

// Store does nothing
func (dsas *stateAccessesStorer) Store(_ map[string]*data.StateAccesses) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (dsas *stateAccessesStorer) IsInterfaceNil() bool {
	return dsas == nil
}
