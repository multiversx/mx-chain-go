package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
)

type SCProcessorMock struct {
	ComputeTransactionTypeCalled          func(tx data.TransactionHandler) (process.TransactionType, error)
	ExecuteSmartContractTransactionCalled func(tx data.TransactionHandler, acntSrc, acntDst state.AccountHandler, round uint64) error
	DeploySmartContractCalled             func(tx data.TransactionHandler, acntSrc state.AccountHandler, round uint64) error
	ProcessSmartContractResultCalled      func(scr *smartContractResult.SmartContractResult) error
}

func (sc *SCProcessorMock) ComputeTransactionType(
	tx data.TransactionHandler,
) (process.TransactionType, error) {
	if sc.ComputeTransactionTypeCalled == nil {
		return process.MoveBalance, nil
	}

	return sc.ComputeTransactionTypeCalled(tx)
}

func (sc *SCProcessorMock) ExecuteSmartContractTransaction(
	tx data.TransactionHandler,
	acntSrc, acntDst state.AccountHandler,
	round uint64,
) error {
	if sc.ExecuteSmartContractTransactionCalled == nil {
		return nil
	}

	return sc.ExecuteSmartContractTransactionCalled(tx, acntSrc, acntDst, round)
}

func (sc *SCProcessorMock) DeploySmartContract(
	tx data.TransactionHandler,
	acntSrc state.AccountHandler,
	round uint64,
) error {
	if sc.DeploySmartContractCalled == nil {
		return nil
	}

	return sc.DeploySmartContractCalled(tx, acntSrc, round)
}

func (sc *SCProcessorMock) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) error {
	if sc.ProcessSmartContractResultCalled == nil {
		return nil
	}

	return sc.ProcessSmartContractResultCalled(scr)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sc *SCProcessorMock) IsInterfaceNil() bool {
	if sc == nil {
		return true
	}
	return false
}
