package aotSelection

import (
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/process"
)

type disabledAOTSelector struct{}

// NewDisabledAOTSelector creates a new disabled AOT selector
// Used when AOT selection is explicitly disabled by configuration
func NewDisabledAOTSelector() *disabledAOTSelector {
	return &disabledAOTSelector{}
}

// TriggerAOTSelection does nothing for disabled selector
func (d *disabledAOTSelector) TriggerAOTSelection(_ data.HeaderHandler, _ uint64) {
}

// GetPreSelectedTransactions returns false for disabled selector
func (d *disabledAOTSelector) GetPreSelectedTransactions(_ uint64) (*process.AOTSelectionResult, bool) {
	return nil, false
}

// CancelOngoingSelection does nothing for disabled selector
func (d *disabledAOTSelector) CancelOngoingSelection() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledAOTSelector) IsInterfaceNil() bool {
	return d == nil
}
