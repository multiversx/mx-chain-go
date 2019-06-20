package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
)

type SCProcessorMock struct {
	ComputeTransactionTypeCalled          func(tx *transaction.Transaction) (process.TransactionType, error)
	ExecuteSmartContractTransactionCalled func(tx *transaction.Transaction, acntSrc, acntDst state.AccountHandler, round uint32) ([]data.TransactionHandler, error)
	DeploySmartContractCalled             func(tx *transaction.Transaction, acntSrc state.AccountHandler, round uint32) ([]data.TransactionHandler, error)
	ProcessSmartContractResultCalled      func(scr *smartContractResult.SmartContractResult) error
}

func (sc *SCProcessorMock) ComputeTransactionType(
	tx *transaction.Transaction,
) (process.TransactionType, error) {
	if sc.ComputeTransactionTypeCalled == nil {
		return process.MoveBalance, nil
	}

	return sc.ComputeTransactionTypeCalled(tx)
}

func (sc *SCProcessorMock) ExecuteSmartContractTransaction(
	tx *transaction.Transaction,
	acntSrc, acntDst state.AccountHandler,
	round uint32,
) ([]data.TransactionHandler, error) {
	if sc.ExecuteSmartContractTransactionCalled == nil {
		return nil, nil
	}

	return sc.ExecuteSmartContractTransactionCalled(tx, acntSrc, acntDst, round)
}

func (sc *SCProcessorMock) DeploySmartContract(
	tx *transaction.Transaction,
	acntSrc state.AccountHandler,
	round uint32,
) ([]data.TransactionHandler, error) {
	if sc.DeploySmartContractCalled == nil {
		return nil, nil
	}

	return sc.DeploySmartContractCalled(tx, acntSrc, round)
}

func (sc *SCProcessorMock) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) error {
	if sc.ProcessSmartContractResultCalled == nil {
		return nil
	}

	return sc.ProcessSmartContractResultCalled(scr)
}
