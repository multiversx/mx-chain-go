package syncer

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"golang.org/x/net/context"
)

// TransactionsSyncHandlerMock -
type TransactionsSyncHandlerMock struct {
	SyncTransactionsForCalled func(miniBlocks map[string]*block.MiniBlock, epoch uint32, ctx context.Context) error
	GetTransactionsCalled     func() (map[string]data.TransactionHandler, error)
}

// SyncTransactionsFor -
func (et *TransactionsSyncHandlerMock) SyncTransactionsFor(miniBlocks map[string]*block.MiniBlock, epoch uint32, ctx context.Context) error {
	if et.SyncTransactionsForCalled != nil {
		return et.SyncTransactionsForCalled(miniBlocks, epoch, ctx)
	}
	return nil
}

// GetTransactions -
func (et *TransactionsSyncHandlerMock) GetTransactions() (map[string]data.TransactionHandler, error) {
	if et.GetTransactionsCalled != nil {
		return et.GetTransactionsCalled()
	}
	return nil, nil
}

// IsInterfaceNil -
func (et *TransactionsSyncHandlerMock) IsInterfaceNil() bool {
	return et == nil
}
