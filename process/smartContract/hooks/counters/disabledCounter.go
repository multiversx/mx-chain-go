package counters

import vmcommon "github.com/multiversx/mx-chain-vm-common-go"

type disabledCounter struct {
}

// NewDisabledCounter will create a new instance of type disabledCounter
func NewDisabledCounter() *disabledCounter {
	return &disabledCounter{}
}

// ProcessCrtNumberOfTrieReadsCounter returns nil
func (counter *disabledCounter) ProcessCrtNumberOfTrieReadsCounter() error {
	return nil
}

// ProcessMaxBuiltInCounters returns nil
func (counter *disabledCounter) ProcessMaxBuiltInCounters(_ *vmcommon.ContractCallInput) error {
	return nil
}

// ResetCounters does nothing
func (counter *disabledCounter) ResetCounters() {}

// SetMaximumValues does nothing
func (counter *disabledCounter) SetMaximumValues(_ map[string]uint64) {}

// GetCounterValues returns an empty map
func (counter *disabledCounter) GetCounterValues() map[string]uint64 {
	return make(map[string]uint64)
}

// IsInterfaceNil returns true if there is no value under the interface
func (counter *disabledCounter) IsInterfaceNil() bool {
	return counter == nil
}
