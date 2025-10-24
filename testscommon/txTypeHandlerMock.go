package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
)

// TxTypeHandlerMock -
type TxTypeHandlerMock struct {
	ComputeTransactionTypeCalled        func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType, bool)
	ComputeTransactionTypeInEpochCalled func(tx data.TransactionHandler, epoch uint32) (process.TransactionType, process.TransactionType, bool)
}

// ComputeTransactionTypeInEpoch -
func (th *TxTypeHandlerMock) ComputeTransactionTypeInEpoch(tx data.TransactionHandler, epoch uint32) (process.TransactionType, process.TransactionType, bool) {
	if th.ComputeTransactionTypeInEpochCalled == nil {
		return process.MoveBalance, process.MoveBalance, false
	}

	return th.ComputeTransactionTypeInEpochCalled(tx, epoch)
}

// ComputeTransactionType -
func (th *TxTypeHandlerMock) ComputeTransactionType(tx data.TransactionHandler) (process.TransactionType, process.TransactionType, bool) {
	if th.ComputeTransactionTypeCalled == nil {
		return process.MoveBalance, process.MoveBalance, false
	}

	return th.ComputeTransactionTypeCalled(tx)
}

// IsInterfaceNil -
func (th *TxTypeHandlerMock) IsInterfaceNil() bool {
	return th == nil
}
