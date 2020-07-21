package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/fullHistory"
)

// HistoryProcessorStub -
type HistoryProcessorStub struct {
	PutTransactionsDataCalled func(htd *fullHistory.HistoryTransactionsData) error
	GetTransactionCalled      func(hash []byte) (*fullHistory.HistoryTransactionWithEpoch, error)
	GetEpochForHashCalled     func(hash []byte) (uint32, error)
	IsEnabledCalled           func() bool
}

// PutTransactionsData -
func (hp *HistoryProcessorStub) PutTransactionsData(historyTxsData *fullHistory.HistoryTransactionsData) error {
	if hp.PutTransactionsDataCalled != nil {
		return hp.PutTransactionsDataCalled(historyTxsData)
	}
	return nil
}

// GetTransaction -
func (hp *HistoryProcessorStub) GetTransaction(hash []byte) (*fullHistory.HistoryTransactionWithEpoch, error) {
	if hp.GetTransactionCalled != nil {
		return hp.GetTransactionCalled(hash)
	}
	return nil, nil
}

// GetEpochForHash -
func (hp *HistoryProcessorStub) GetEpochForHash(hash []byte) (uint32, error) {
	return hp.GetEpochForHashCalled(hash)
}

// IsEnabled -
func (hp *HistoryProcessorStub) IsEnabled() bool {
	return true
}

// IsInterfaceNil -
func (hp *HistoryProcessorStub) IsInterfaceNil() bool {
	return hp == nil
}
