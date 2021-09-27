package disabled

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/dblookupext"
)

var errorDisabledHistoryRepository = errors.New("history repository is disabled")

type nilHistoryRepository struct {
}

// NewNilHistoryRepository returns a not implemented history repository
func NewNilHistoryRepository() (*nilHistoryRepository, error) {
	return new(nilHistoryRepository), nil
}

// RecordBlock returns a not implemented error
func (nhr *nilHistoryRepository) RecordBlock(_ []byte, _ data.HeaderHandler, _ data.BodyHandler, _, _ map[string]data.TransactionHandler, _ map[string]data.LogHandler) error {
	return nil
}

// OnNotarizedBlocks does nothing
func (nhr *nilHistoryRepository) OnNotarizedBlocks(_ uint32, _ []data.HeaderHandler, _ [][]byte) {
}

// GetMiniblockMetadataByTxHash does nothing
func (nhr *nilHistoryRepository) GetMiniblockMetadataByTxHash(_ []byte) (*dblookupext.MiniblockMetadata, error) {
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

// RevertBlock -
func (nhr *nilHistoryRepository) RevertBlock(_ data.HeaderHandler, _ data.BodyHandler) error {
	return nil
}

// GetESDTSupply -
func (nhr *nilHistoryRepository) GetESDTSupply(_ string) (string, error) {
	return "", errorDisabledHistoryRepository
}

// GetResultsHashesByTxHash -
func (nhr *nilHistoryRepository) GetResultsHashesByTxHash(_ []byte, _ uint32) (*dblookupext.ResultsHashesByTxHash, error) {
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (nhr *nilHistoryRepository) IsInterfaceNil() bool {
	return nhr == nil
}
