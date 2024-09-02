package disabled

import (
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/stateChanges"
)

// disabledStateChangesCollector is a state changes collector that does nothing
type disabledStateChangesCollector struct {
}

// NewDisabledStateChangesCollector creates a new disabledStateChangesCollector
func NewDisabledStateChangesCollector() state.StateChangesCollector {
	return &disabledStateChangesCollector{}
}

// AddSaveAccountStateChange -
func (d *disabledStateChangesCollector) AddSaveAccountStateChange(oldAccount, account vmcommon.AccountHandler, stateChange stateChanges.StateChange) {
}

// AddStateChange does nothing
func (d *disabledStateChangesCollector) AddStateChange(_ stateChanges.StateChange) {
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

// Publish returns nil
func (d *disabledStateChangesCollector) Publish() error {
	return nil
}

func (d *disabledStateChangesCollector) GetStateChanges() []stateChanges.StateChange {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledStateChangesCollector) IsInterfaceNil() bool {
	return d == nil
}
