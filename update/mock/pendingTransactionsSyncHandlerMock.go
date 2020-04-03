package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"golang.org/x/net/context"
)

// PendingTransactionsSyncHandlerMock -
type PendingTransactionsSyncHandlerMock struct {
	SyncPendingTransactionsForCalled func(miniBlocks map[string]*block.MiniBlock, epoch uint32, ctx context.Context) error
	GetTransactionsCalled            func() (map[string]data.TransactionHandler, error)
}

// SyncPendingTransactionsFor -
func (et *PendingTransactionsSyncHandlerMock) SyncPendingTransactionsFor(miniBlocks map[string]*block.MiniBlock, epoch uint32, ctx context.Context) error {
	if et.SyncPendingTransactionsForCalled != nil {
		return et.SyncPendingTransactionsForCalled(miniBlocks, epoch, ctx)
	}
	return nil
}

// GetTransactions -
func (et *PendingTransactionsSyncHandlerMock) GetTransactions() (map[string]data.TransactionHandler, error) {
	if et.GetTransactionsCalled != nil {
		return et.GetTransactionsCalled()
	}
	return nil, nil
}

// IsInterfaceNil -
func (et *PendingTransactionsSyncHandlerMock) IsInterfaceNil() bool {
	return et == nil
}
