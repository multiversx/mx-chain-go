package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/core/fullHistory"
	"github.com/ElrondNetwork/elrond-go/data"
)

// HistoryRepositoryStub -
type HistoryRepositoryStub struct {
	RecordBlockCalled                  func(blockHeaderHash []byte, blockHeader data.HeaderHandler, blockBody data.BodyHandler) error
	GetTransactionsGroupMetadataCalled func(hash []byte) (*fullHistory.TransactionsGroupMetadataWithEpoch, error)
	GetEpochByHashCalled               func(hash []byte) (uint32, error)
	IsEnabledCalled                    func() bool
}

// RecordBlock -
func (hp *HistoryRepositoryStub) RecordBlock(blockHeaderHash []byte, blockHeader data.HeaderHandler, blockBody data.BodyHandler) error {
	if hp.RecordBlockCalled != nil {
		return hp.RecordBlockCalled(blockHeaderHash, blockHeader, blockBody)
	}
	return nil
}

// GetTransactionsGroupMetadata -
func (hp *HistoryRepositoryStub) GetTransactionsGroupMetadata(hash []byte) (*fullHistory.TransactionsGroupMetadataWithEpoch, error) {
	if hp.GetTransactionsGroupMetadataCalled != nil {
		return hp.GetTransactionsGroupMetadataCalled(hash)
	}
	return nil, nil
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
