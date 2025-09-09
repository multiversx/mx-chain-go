package disabled

import (
	data "github.com/multiversx/mx-chain-core-go/data/stateChange"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"

	"github.com/multiversx/mx-chain-go/state"
)

// disabledStateAccessesCollector is a state accesses collector that does nothing
type disabledStateAccessesCollector struct {
}

// NewDisabledStateAccessesCollector creates a new disabledStateAccessesCollector
func NewDisabledStateAccessesCollector() state.StateAccessesCollector {
	return &disabledStateAccessesCollector{}
}

// GetAccountChanges returns the constant marking NoChange
func (d *disabledStateAccessesCollector) GetAccountChanges(_, _ vmcommon.AccountHandler) uint32 {
	return data.NoChange
}

// AddStateAccess does nothing
func (d *disabledStateAccessesCollector) AddStateAccess(_ *data.StateAccess) {
}

// Reset does nothing
func (d *disabledStateAccessesCollector) Reset() {
}

// AddTxHashToCollectedStateAccesses does nothing
func (d *disabledStateAccessesCollector) AddTxHashToCollectedStateAccesses(_ []byte) {
}

// SetIndexToLatestStateAccesses -
func (d *disabledStateAccessesCollector) SetIndexToLatestStateAccesses(_ int) error {
	return nil
}

// RevertToIndex -
func (d *disabledStateAccessesCollector) RevertToIndex(_ int) error {
	return nil
}

// GetCollectedAccesses -
func (d *disabledStateAccessesCollector) GetCollectedAccesses() map[string]*data.StateAccesses {
	return nil
}

// Store -
func (d *disabledStateAccessesCollector) Store() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledStateAccessesCollector) IsInterfaceNil() bool {
	return d == nil
}
