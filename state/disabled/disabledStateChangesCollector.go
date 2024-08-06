package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/state"
)

// disabledStateChangesCollector is a state changes collector that does nothing
type disabledStateChangesCollector struct {
}

// NewDisabledStateChangesCollector creates a new disabledStateChangesCollector
func NewDisabledStateChangesCollector() state.StateChangesCollector {
	return &disabledStateChangesCollector{}
}

// AddStateChange does nothing
func (d *disabledStateChangesCollector) AddStateChange(_ state.StateChangeDTO) {
}

// GetStateChanges returns an empty slice
func (d *disabledStateChangesCollector) GetStateChanges() []state.StateChangesForTx {
	return make([]state.StateChangesForTx, 0)
}

// Reset does nothing
func (d *disabledStateChangesCollector) Reset() {
}

// AddTxHashToCollectedStateChanges does nothing
func (d *disabledStateChangesCollector) AddTxHashToCollectedStateChanges(_ []byte, _ *transaction.Transaction) {
}

// SetIndexToLastStateChange -
func (d *disabledStateChangesCollector) SetIndexToLastStateChange(index int) error {
	return nil
}

// RevertToIndex -
func (d *disabledStateChangesCollector) RevertToIndex(index int) error {
	return nil
}

// DumpToJSONFile returns nil
func (d *disabledStateChangesCollector) DumpToJSONFile() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledStateChangesCollector) IsInterfaceNil() bool {
	return d == nil
}
