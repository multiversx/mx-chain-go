package fullHistory

import (
	"errors"
)

type nilHistoryProcessor struct {
}

// NewNilHistoryProcessor -
func NewNilHistoryProcessor() (*nilHistoryProcessor, error) {
	return new(nilHistoryProcessor), nil
}

var errNilHistoryProcessorImplementation = errors.New("this a nil implementation of history processor")

// PutTransactionsData -
func (nhs *nilHistoryProcessor) PutTransactionsData(_ *HistoryTransactionsData) error {
	return errNilHistoryProcessorImplementation
}

// GetTransaction -
func (nhs *nilHistoryProcessor) GetTransaction(_ []byte) (*HistoryTransaction, error) {
	return nil, errNilHistoryProcessorImplementation
}

// IsEnabled -
func (nhs *nilHistoryProcessor) IsEnabled() bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (nhs *nilHistoryProcessor) IsInterfaceNil() bool {
	return nhs == nil
}
