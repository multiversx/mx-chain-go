package testscommon

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core/fullHistory"
	"github.com/ElrondNetwork/elrond-go/data"
)

// HistoryRepositoryStub -
type HistoryRepositoryStub struct {
	RegisterToBlockTrackerCalled       func(blockTracker fullHistory.BlockTracker)
	RecordBlockCalled                  func(blockHeaderHash []byte, blockHeader data.HeaderHandler, blockBody data.BodyHandler) error
	GetMiniblockMetadataByTxHashCalled func(hash []byte) (*fullHistory.MiniblockMetadata, error)
	GetEpochByHashCalled               func(hash []byte) (uint32, error)
	IsEnabledCalled                    func() bool
}

// RegisterToBlockTracker -
func (hp *HistoryRepositoryStub) RegisterToBlockTracker(blockTracker fullHistory.BlockTracker) {
	if hp.RegisterToBlockTrackerCalled != nil {
		hp.RegisterToBlockTrackerCalled(blockTracker)
	}
}

// RecordBlock -
func (hp *HistoryRepositoryStub) RecordBlock(blockHeaderHash []byte, blockHeader data.HeaderHandler, blockBody data.BodyHandler) error {
	if hp.RecordBlockCalled != nil {
		return hp.RecordBlockCalled(blockHeaderHash, blockHeader, blockBody)
	}
	return nil
}

// GetMiniblockMetadataByTxHash -
func (hp *HistoryRepositoryStub) GetMiniblockMetadataByTxHash(hash []byte) (*fullHistory.MiniblockMetadata, error) {
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

// IsInterfaceNil -
func (hp *HistoryRepositoryStub) IsInterfaceNil() bool {
	return hp == nil
}
