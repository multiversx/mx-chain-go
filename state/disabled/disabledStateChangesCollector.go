package disabled

import (
	coreData "github.com/multiversx/mx-chain-core-go/data"
	data "github.com/multiversx/mx-chain-core-go/data/stateChange"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/state"
)

// disabledStateChangesCollector is a state changes collector that does nothing
type disabledStateChangesCollector struct {
}

// NewDisabledStateChangesCollector creates a new disabledStateChangesCollector
func NewDisabledStateChangesCollector() state.StateChangesCollector {
	return &disabledStateChangesCollector{}
}

// AddSaveAccountStateChange -
func (d *disabledStateChangesCollector) AddSaveAccountStateChange(_, _ vmcommon.AccountHandler, stateChange state.StateChange) {
}

// AddStateChange does nothing
func (d *disabledStateChangesCollector) AddStateChange(_ state.StateChange) {
}

// Reset does nothing
func (d *disabledStateChangesCollector) Reset() {
}

// AddTxHashToCollectedStateChanges does nothing
func (d *disabledStateChangesCollector) AddTxHashToCollectedStateChanges(_ []byte, _ coreData.TransactionHandler) {
}

// SetIndexToLastStateChange -
func (d *disabledStateChangesCollector) SetIndexToLastStateChange(_ int) error {
	return nil
}

// RevertToIndex -
func (d *disabledStateChangesCollector) RevertToIndex(_ int) error {
	return nil
}

// Publish -
func (d *disabledStateChangesCollector) Publish() (map[string]*data.StateChanges, error) {
	return nil, nil
}

// Store -
func (d *disabledStateChangesCollector) Store() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledStateChangesCollector) IsInterfaceNil() bool {
	return d == nil
}
