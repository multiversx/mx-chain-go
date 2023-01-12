package syncer

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/state"
)

// TransactionsSyncHandlerMock -
type TransactionsSyncHandlerMock struct {
	SyncTransactionsForCalled func(miniBlocks map[string]*block.MiniBlock, epoch uint32, ctx context.Context) error
	GetTransactionsCalled     func() (map[string]data.TransactionHandler, error)
	GetValidatorsInfoCalled   func() (map[string]*state.ShardValidatorInfo, error)
	ClearFieldsCalled         func()
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

// GetValidatorsInfo -
func (et *TransactionsSyncHandlerMock) GetValidatorsInfo() (map[string]*state.ShardValidatorInfo, error) {
	if et.GetValidatorsInfoCalled != nil {
		return et.GetValidatorsInfoCalled()
	}
	return nil, nil
}

// ClearFields -
func (et *TransactionsSyncHandlerMock) ClearFields() {
	if et.ClearFieldsCalled != nil {
		et.ClearFieldsCalled()
	}
}

// IsInterfaceNil -
func (et *TransactionsSyncHandlerMock) IsInterfaceNil() bool {
	return et == nil
}
