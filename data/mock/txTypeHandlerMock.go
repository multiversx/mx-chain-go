package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type TxTypeHandlerMock struct {
	ComputeTransactionTypeCalled func(tx data.TransactionHandler) (process.TransactionType, error)
}

func (th *TxTypeHandlerMock) ComputeTransactionType(tx data.TransactionHandler) (process.TransactionType, error) {
	if th.ComputeTransactionTypeCalled == nil {
		return process.MoveBalance, nil
	}

	return th.ComputeTransactionTypeCalled(tx)
}
