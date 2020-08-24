package fullHistory

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type nilHistoryRepository struct {
}

// NewNilHistoryRepository returns a not implemented history repository
func NewNilHistoryRepository() (*nilHistoryRepository, error) {
	return new(nilHistoryRepository), nil
}

// RegisterToBlockTracker does nothing
func (nhr *nilHistoryRepository) RegisterToBlockTracker(blockTracker BlockTracker) {
}

// RecordBlock returns a not implemented error
func (nhr *nilHistoryRepository) RecordBlock(_ []byte, _ data.HeaderHandler, _ data.BodyHandler) error {
	return nil
}

// GetTransaction returns a not implemented error
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

// IsInterfaceNil returns true if there is no value under the interface
func (nhr *nilHistoryRepository) IsInterfaceNil() bool {
	return nhr == nil
}
