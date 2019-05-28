package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type SCProcessorMock struct {
	ComputeTransactionTypeCalled          func(tx *transaction.Transaction, acntSrc, acntDst state.AccountHandler) (process.TransactionType, error)
	ExecuteSmartContractTransactionCalled func(tx *transaction.Transaction, acntSrc, acntDst state.AccountHandler) error
}

func (sc *SCProcessorMock) ComputeTransactionType(
	tx *transaction.Transaction,
	acntSrc, acntDst state.AccountHandler,
) (process.TransactionType, error) {
	if sc.ComputeTransactionTypeCalled == nil {
		return process.MoveBalance, nil
	}

	return sc.ComputeTransactionTypeCalled(tx, acntSrc, acntDst)
}

func (sc *SCProcessorMock) ExecuteSmartContractTransaction(
	tx *transaction.Transaction,
	acntSrc, acntDst state.AccountHandler,
) error {
	if sc.ComputeTransactionTypeCalled == nil {
		return nil
	}

	return sc.ExecuteSmartContractTransactionCalled(tx, acntSrc, acntDst)
}
