package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

// TxTypeHandlerMock -
type TxTypeHandlerMock struct {
	ComputeTransactionTypeCalled func(tx data.TransactionHandler) process.TransactionType
}

// ComputeTransactionType -
func (th *TxTypeHandlerMock) ComputeTransactionType(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
	if th.ComputeTransactionTypeCalled == nil {
		return process.MoveBalance, 0
	}

	return th.ComputeTransactionTypeCalled(tx), 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (th *TxTypeHandlerMock) IsInterfaceNil() bool {
	return th == nil
}
