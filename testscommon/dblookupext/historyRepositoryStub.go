package dblookupext

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/dblookupext/esdtSupply"
)

// HistoryRepositoryStub -
type HistoryRepositoryStub struct {
	RecordBlockCalled                  func(blockHeaderHash []byte, blockHeader data.HeaderHandler, blockBody data.BodyHandler, scrsPool map[string]data.TransactionHandler, receipts map[string]data.TransactionHandler, createdIntraMiniBlocks []*block.MiniBlock, logs []*data.LogData) error
	OnNotarizedBlocksCalled            func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte)
	GetMiniblockMetadataByTxHashCalled func(hash []byte) (*dblookupext.MiniblockMetadata, error)
	GetEpochByHashCalled               func(hash []byte) (uint32, error)
	GetEventsHashesByTxHashCalled      func(hash []byte, epoch uint32) (*dblookupext.ResultsHashesByTxHash, error)
	GetESDTSupplyCalled                func(token string) (*esdtSupply.SupplyESDT, error)
	IsEnabledCalled                    func() bool
}

// RecordBlock -
func (hp *HistoryRepositoryStub) RecordBlock(
	blockHeaderHash []byte,
	blockHeader data.HeaderHandler,
	blockBody data.BodyHandler,
	scrsPool map[string]data.TransactionHandler,
	receipts map[string]data.TransactionHandler,
	createdIntraMiniBlocks []*block.MiniBlock,
	logs []*data.LogData,
) error {
	if hp.RecordBlockCalled != nil {
		return hp.RecordBlockCalled(blockHeaderHash, blockHeader, blockBody, scrsPool, receipts, createdIntraMiniBlocks, logs)
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
	if hp.GetEpochByHashCalled != nil {
		return hp.GetEpochByHashCalled(hash)
	}
	return 0, nil
}

// IsEnabled -
func (hp *HistoryRepositoryStub) IsEnabled() bool {
	if hp.IsEnabledCalled != nil {
		return hp.IsEnabledCalled()
	}
	return true
}

// GetResultsHashesByTxHash -
func (hp *HistoryRepositoryStub) GetResultsHashesByTxHash(hash []byte, epoch uint32) (*dblookupext.ResultsHashesByTxHash, error) {
	if hp.GetEventsHashesByTxHashCalled != nil {
		return hp.GetEventsHashesByTxHashCalled(hash, epoch)
	}
	return nil, nil
}

// RevertBlock -
func (hp *HistoryRepositoryStub) RevertBlock(_ data.HeaderHandler, _ data.BodyHandler) error {
	return nil
}

// GetESDTSupply -
func (hp *HistoryRepositoryStub) GetESDTSupply(token string) (*esdtSupply.SupplyESDT, error) {
	if hp.GetESDTSupplyCalled != nil {
		return hp.GetESDTSupplyCalled(token)
	}

	return nil, nil
}

// IsInterfaceNil -
func (hp *HistoryRepositoryStub) IsInterfaceNil() bool {
	return hp == nil
}
