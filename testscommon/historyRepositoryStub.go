package testscommon

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/dblookupext"
	"github.com/ElrondNetwork/elrond-go/data"
)

// HistoryRepositoryStub -
type HistoryRepositoryStub struct {
	RecordBlockCalled                  func(blockHeaderHash []byte, blockHeader data.HeaderHandler, blockBody data.BodyHandler, scrsPool map[string]data.TransactionHandler, receipts map[string]data.TransactionHandler) error
	OnNotarizedBlocksCalled            func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)
	GetMiniblockMetadataByTxHashCalled func(hash []byte) (*dblookupext.MiniblockMetadata, error)
	GetEpochByHashCalled               func(hash []byte) (uint32, error)
	GetEventsHashesByTxHashCalled      func(hash []byte, epoch uint32) (*dblookupext.EventsHashesByTxHash, error)
	IsEnabledCalled                    func() bool
}

// RecordBlock -
func (hp *HistoryRepositoryStub) RecordBlock(
	blockHeaderHash []byte,
	blockHeader data.HeaderHandler,
	blockBody data.BodyHandler,
	scrsPool map[string]data.TransactionHandler,
	receipts map[string]data.TransactionHandler,
) error {
	if hp.RecordBlockCalled != nil {
		return hp.RecordBlockCalled(blockHeaderHash, blockHeader, blockBody, scrsPool, receipts)
	}
	return nil
}

// OnNotarizedBlocks -
func (hp *HistoryRepositoryStub) OnNotarizedBlocks(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte) {
	if hp.OnNotarizedBlocksCalled != nil {
		hp.OnNotarizedBlocksCalled(shardID, headers, headersHashes)
	}
}

// GetMiniblockMetadataByTxHash -
func (hp *HistoryRepositoryStub) GetMiniblockMetadataByTxHash(hash []byte) (*dblookupext.MiniblockMetadata, error) {
	if hp.GetMiniblockMetadataByTxHashCalled != nil {
		return hp.GetMiniblockMetadataByTxHashCalled(hash)
	}
	return nil, fmt.Errorf("miniblock metadata not found")
}

// GetEpochByHash -
func (hp *HistoryRepositoryStub) GetEpochByHash(hash []byte) (uint32, error) {
	return hp.GetEpochByHashCalled(hash)
}

// IsEnabled -
func (hp *HistoryRepositoryStub) IsEnabled() bool {
	if hp.IsEnabledCalled != nil {
		return hp.IsEnabledCalled()
	}
	return true
}

// GetEventsHashesByTxHash -
func (hp *HistoryRepositoryStub) GetEventsHashesByTxHash(hash []byte, epoch uint32) (*dblookupext.EventsHashesByTxHash, error) {
	if hp.GetEventsHashesByTxHashCalled != nil {
		return hp.GetEventsHashesByTxHashCalled(hash, epoch)
	}
	return nil, nil
}

// IsInterfaceNil -
func (hp *HistoryRepositoryStub) IsInterfaceNil() bool {
	return hp == nil
}
