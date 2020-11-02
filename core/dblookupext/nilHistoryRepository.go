package dblookupext

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type nilHistoryRepository struct {
}

// NewNilHistoryRepository returns a not implemented history repository
func NewNilHistoryRepository() (*nilHistoryRepository, error) {
	return new(nilHistoryRepository), nil
}

// RecordBlock returns a not implemented error
func (nhr *nilHistoryRepository) RecordBlock(_ []byte, _ data.HeaderHandler, _ data.BodyHandler, _, _ map[string]data.TransactionHandler) error {
	return nil
}

// OnNotarizedBlocks does nothing
func (nhr *nilHistoryRepository) OnNotarizedBlocks(_ uint32, _ []data.HeaderHandler, _ [][]byte) {
}

// GetMiniblockMetadataByTxHash does nothing
func (nhr *nilHistoryRepository) GetMiniblockMetadataByTxHash(_ []byte) (*MiniblockMetadata, error) {
	return nil, nil
}

// GetEpochByHash returns a not implemented error
func (nhr *nilHistoryRepository) GetEpochByHash(_ []byte) (uint32, error) {
	return 0, nil
}

// IsEnabled returns false
func (nhr *nilHistoryRepository) IsEnabled() bool {
	return false
}

// GetEventsHashesByTxHash -
func (nhr *nilHistoryRepository) GetEventsHashesByTxHash(_ []byte, _ uint32) (*EventsHashesByTxHash, error) {
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (nhr *nilHistoryRepository) IsInterfaceNil() bool {
	return nhr == nil
}
