package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/core/fullHistory"
)

// HistoryRepositoryStub -
type HistoryRepositoryStub struct {
	PutTransactionsDataCalled          func(htd *fullHistory.HistoryTransactionsData) error
	GetTransactionsGroupMetadataCalled func(hash []byte) (*fullHistory.HistoryTransactionWithEpoch, error)
	GetEpochForHashCalled              func(hash []byte) (uint32, error)
	IsEnabledCalled                    func() bool
}

// PutTransactionsData -
func (hp *HistoryRepositoryStub) PutTransactionsData(historyTxsData *fullHistory.HistoryTransactionsData) error {
	if hp.PutTransactionsDataCalled != nil {
		return hp.PutTransactionsDataCalled(historyTxsData)
	}
	return nil
}

// GetTransactionsGroupMetadata -
func (hp *HistoryRepositoryStub) GetTransactionsGroupMetadata(hash []byte) (*fullHistory.HistoryTransactionWithEpoch, error) {
	if hp.GetTransactionsGroupMetadataCalled != nil {
		return hp.GetTransactionsGroupMetadataCalled(hash)
	}
	return nil, nil
}

// GetEpochForHash -
func (hp *HistoryRepositoryStub) GetEpochForHash(hash []byte) (uint32, error) {
	return hp.GetEpochForHashCalled(hash)
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
