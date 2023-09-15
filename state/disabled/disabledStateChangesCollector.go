package disabled

import (
	"github.com/multiversx/mx-chain-go/state"
)

// DisabledStateChangesCollector is a state changes collector that does nothing
type DisabledStateChangesCollector struct {
}

// NewDisabledStateChangesCollector creates a new DisabledStateChangesCollector
func NewDisabledStateChangesCollector() *DisabledStateChangesCollector {
	return &DisabledStateChangesCollector{}
}

// AddStateChange does nothing
func (d *DisabledStateChangesCollector) AddStateChange(_ state.StateChangeDTO) {

}

// GetStateChanges returns an empty slice
func (d *DisabledStateChangesCollector) GetStateChanges() []state.StateChangeDTO {
	return []state.StateChangeDTO{}
}

// Reset does nothing
func (d *DisabledStateChangesCollector) Reset() {

}

// IsInterfaceNil returns true if there is no value under the interface
func (d *DisabledStateChangesCollector) IsInterfaceNil() bool {
	return d == nil
}
