package fullHistory

import (
	"errors"
)

type nilHistoryRepository struct {
}

// NewNilHistoryProcessor returns a not implemented history repository
func NewNilHistoryProcessor() (*nilHistoryRepository, error) {
	return new(nilHistoryRepository), nil
}

var errNilHistoryProcessorImplementation = errors.New("this a nil implementation of history processor")

// PutTransactionsData returns a not implemented error
func (nhr *nilHistoryRepository) PutTransactionsData(_ *HistoryTransactionsData) error {
	return errNilHistoryProcessorImplementation
}

// GetTransaction returns a not implemented error
func (nhr *nilHistoryRepository) GetTransaction(_ []byte) (*HistoryTransactionWithEpoch, error) {
	return nil, errNilHistoryProcessorImplementation
}

// GetEpochForHash returns a not implemented error
func (nhr *nilHistoryRepository) GetEpochForHash(_ []byte) (uint32, error) {
	return 0, errNilHistoryProcessorImplementation
}

// IsEnabled -
func (nhr *nilHistoryRepository) IsEnabled() bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (nhr *nilHistoryRepository) IsInterfaceNil() bool {
	return nhr == nil
}
