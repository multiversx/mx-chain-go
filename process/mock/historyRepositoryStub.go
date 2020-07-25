package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/fullHistory"
)

// HistoryRepositoryStub -
type HistoryRepositoryStub struct {
	PutTransactionsDataCalled func(htd *fullHistory.HistoryTransactionsData) error
	GetTransactionCalled      func(hash []byte) (*fullHistory.HistoryTransactionWithEpoch, error)
	GetEpochForHashCalled     func(hash []byte) (uint32, error)
	IsEnabledCalled           func() bool
}

// PutTransactionsData -
func (hr *HistoryRepositoryStub) PutTransactionsData(historyTxsData *fullHistory.HistoryTransactionsData) error {
	if hr.PutTransactionsDataCalled != nil {
		return hr.PutTransactionsDataCalled(historyTxsData)
	}
	return nil
}

// GetTransaction -
func (hr *HistoryRepositoryStub) GetTransaction(hash []byte) (*fullHistory.HistoryTransactionWithEpoch, error) {
	if hr.GetTransactionCalled != nil {
		return hr.GetTransactionCalled(hash)
	}
	return nil, nil
}

// GetEpochForHash -
func (hr *HistoryRepositoryStub) GetEpochForHash(hash []byte) (uint32, error) {
	return hr.GetEpochForHashCalled(hash)
}

// IsEnabled -
func (hr *HistoryRepositoryStub) IsEnabled() bool {
	return true
}

// IsInterfaceNil -
func (hr *HistoryRepositoryStub) IsInterfaceNil() bool {
	return hr == nil
}
