package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type PendingTransactionsSyncHandlerMock struct {
	SyncPendingTransactionsForCalled func(miniBlocks map[string]*block.MiniBlock, epoch uint32, waitTime time.Duration) error
	GetTransactionsCalled            func() (map[string]data.TransactionHandler, error)
}

func (et *PendingTransactionsSyncHandlerMock) SyncPendingTransactionsFor(miniBlocks map[string]*block.MiniBlock, epoch uint32, waitTime time.Duration) error {
	if et.SyncPendingTransactionsForCalled != nil {
		return et.SyncPendingTransactionsForCalled(miniBlocks, epoch, waitTime)
	}
	return nil
}

func (et *PendingTransactionsSyncHandlerMock) GetTransactions() (map[string]data.TransactionHandler, error) {
	if et.GetTransactionsCalled != nil {
		return et.GetTransactionsCalled()
	}
	return nil, nil
}

func (et *PendingTransactionsSyncHandlerMock) IsInterfaceNil() bool {
	return et == nil
}
