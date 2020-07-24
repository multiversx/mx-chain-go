package fullHistory

type nilHistoryRepository struct {
}

// NewNilHistoryRepository returns a not implemented history repository
func NewNilHistoryRepository() (*nilHistoryRepository, error) {
	return new(nilHistoryRepository), nil
}

// PutTransactionsData returns a not implemented error
func (nhr *nilHistoryRepository) PutTransactionsData(_ *HistoryTransactionsData) error {
	return nil
}

// GetTransaction returns a not implemented error
func (nhr *nilHistoryRepository) GetTransaction(_ []byte) (*HistoryTransactionWithEpoch, error) {
	return nil, nil
}

// GetEpochForHash returns a not implemented error
func (nhr *nilHistoryRepository) GetEpochForHash(_ []byte) (uint32, error) {
	return 0, nil
}

// IsEnabled -
func (nhr *nilHistoryRepository) IsEnabled() bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (nhr *nilHistoryRepository) IsInterfaceNil() bool {
	return nhr == nil
}
