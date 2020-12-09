package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
)

// SCProcessorMock -
type SCProcessorMock struct {
	ComputeTransactionTypeCalled          func(tx data.TransactionHandler) process.TransactionType
	ExecuteSmartContractTransactionCalled func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error)
	ExecuteBuiltInFunctionCalled          func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error)
	DeploySmartContractCalled             func(tx data.TransactionHandler, acntSrc state.UserAccountHandler) (vmcommon.ReturnCode, error)
	ProcessSmartContractResultCalled      func(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error)
	ProcessIfErrorCalled                  func(acntSnd state.UserAccountHandler, txHash []byte, tx data.TransactionHandler, returnCode string, returnMessage []byte, snapshot int, gasLocked uint64) error
	IsPayableCalled                       func(address []byte) (bool, error)
}

// IsPayable -
func (sc *SCProcessorMock) IsPayable(address []byte) (bool, error) {
	if sc.IsPayableCalled != nil {
		return sc.IsPayable(address)
	}
	return true, nil
}

// ProcessIfError -
func (sc *SCProcessorMock) ProcessIfError(
	acntSnd state.UserAccountHandler,
	txHash []byte,
	tx data.TransactionHandler,
	returnCode string,
	returnMessage []byte,
	snapshot int,
	gasLocked uint64,
) error {
	if sc.ProcessIfErrorCalled != nil {
		return sc.ProcessIfErrorCalled(acntSnd, txHash, tx, returnCode, returnMessage, snapshot, gasLocked)
	}
	return nil
}

// ComputeTransactionType -
func (sc *SCProcessorMock) ComputeTransactionType(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
	if sc.ComputeTransactionTypeCalled == nil {
		return process.MoveBalance, 0
	}

	return sc.ComputeTransactionTypeCalled(tx), 0
}

// ExecuteSmartContractTransaction -
func (sc *SCProcessorMock) ExecuteSmartContractTransaction(
	tx data.TransactionHandler,
	acntSrc, acntDst state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	if sc.ExecuteSmartContractTransactionCalled == nil {
		return 0, nil
	}

	return sc.ExecuteSmartContractTransactionCalled(tx, acntSrc, acntDst)
}

// ExecuteBuiltInFunction -
func (sc *SCProcessorMock) ExecuteBuiltInFunction(
	tx data.TransactionHandler,
	acntSrc, acntDst state.UserAccountHandler,
) (vmcommon.ReturnCode, error) {
	if sc.ExecuteBuiltInFunctionCalled == nil {
		return 0, nil
	}

	return sc.ExecuteBuiltInFunctionCalled(tx, acntSrc, acntDst)
}

// DeploySmartContract -
func (sc *SCProcessorMock) DeploySmartContract(tx data.TransactionHandler, acntSrc state.UserAccountHandler) (vmcommon.ReturnCode, error) {
	if sc.DeploySmartContractCalled == nil {
		return 0, nil
	}

	return sc.DeploySmartContractCalled(tx, acntSrc)
}

// ProcessSmartContractResult -
func (sc *SCProcessorMock) ProcessSmartContractResult(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error) {
	if sc.ProcessSmartContractResultCalled == nil {
		return 0, nil
	}

	return sc.ProcessSmartContractResultCalled(scr)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sc *SCProcessorMock) IsInterfaceNil() bool {
	return sc == nil
}
