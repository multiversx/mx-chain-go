package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
)

// SCProcessorMock -
type SCProcessorMock struct {
	ComputeTransactionTypeCalled          func(tx data.TransactionHandler) (process.TransactionType, error)
	ExecuteSmartContractTransactionCalled func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) error
	DeploySmartContractCalled             func(tx data.TransactionHandler, acntSrc state.UserAccountHandler) error
	ProcessSmartContractResultCalled      func(scr *smartContractResult.SmartContractResult) error
}

// ComputeTransactionType -
func (sc *SCProcessorMock) ComputeTransactionType(
	tx data.TransactionHandler,
) (process.TransactionType, error) {
	if sc.ComputeTransactionTypeCalled == nil {
		return process.MoveBalance, nil
	}

	return sc.ComputeTransactionTypeCalled(tx)
}

// ExecuteSmartContractTransaction -
func (sc *SCProcessorMock) ExecuteSmartContractTransaction(
	tx data.TransactionHandler,
	acntSrc, acntDst state.UserAccountHandler,
) error {
	if sc.ExecuteSmartContractTransactionCalled == nil {
		return nil
	}

	return sc.ExecuteSmartContractTransactionCalled(tx, acntSrc, acntDst)
}

// DeploySmartContract -
func (sc *SCProcessorMock) DeploySmartContract(
	tx data.TransactionHandler,
	acntSrc state.UserAccountHandler,
) error {
	if sc.DeploySmartContractCalled == nil {
		return nil
	}

	return sc.DeploySmartContractCalled(tx, acntSrc)
}

// ProcessSmartContractResult -
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
