package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// SCProcessorMock -
type SCProcessorMock struct {
	ComputeTransactionTypeCalled           func(tx data.TransactionHandler) (process.TransactionType, process.TransactionType)
	ExecuteSmartContractTransactionCalled  func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error)
	ExecuteBuiltInFunctionCalled           func(tx data.TransactionHandler, acntSrc, acntDst state.UserAccountHandler) (vmcommon.ReturnCode, error)
	DeploySmartContractCalled              func(tx data.TransactionHandler, acntSrc state.UserAccountHandler) (vmcommon.ReturnCode, error)
	ProcessSmartContractResultCalled       func(scr *smartContractResult.SmartContractResult) (vmcommon.ReturnCode, error)
	ProcessIfErrorCalled                   func(acntSnd state.UserAccountHandler, txHash []byte, tx data.TransactionHandler, returnCode string, returnMessage []byte, snapshot int, gasLocked uint64) error
	IsPayableCalled                        func(sndAddress, recvAddress []byte) (bool, error)
	CheckBuiltinFunctionIsExecutableCalled func(expectedBuiltinFunction string, tx data.TransactionHandler) error
	ArgsParserCalled                       func() process.ArgumentsParser
	TxTypeHandlerCalled                    func() process.TxTypeHandler
	CheckSCRBeforeProcessingCalled         func(scr *smartContractResult.SmartContractResult) (process.ScrProcessingDataHandler, error)
}

// IsPayable -
func (sc *SCProcessorMock) IsPayable(sndAddress []byte, recvAddress []byte) (bool, error) {
	if sc.IsPayableCalled != nil {
		return sc.IsPayableCalled(sndAddress, recvAddress)
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
		return process.MoveBalance, process.MoveBalance
	}

	return sc.ComputeTransactionTypeCalled(tx)
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

// CheckBuiltinFunctionIsExecutable -
func (sc *SCProcessorMock) CheckBuiltinFunctionIsExecutable(expectedBuiltinFunction string, tx data.TransactionHandler) error {
	if sc.CheckBuiltinFunctionIsExecutableCalled == nil {
		return nil
	}

	return sc.CheckBuiltinFunctionIsExecutableCalled(expectedBuiltinFunction, tx)
}

// ArgsParser -
func (sc *SCProcessorMock) ArgsParser() process.ArgumentsParser {
	if sc.ArgsParserCalled == nil {
		return nil
	}
	return sc.ArgsParserCalled()
}

// TxTypeHandler -
func (sc *SCProcessorMock) TxTypeHandler() process.TxTypeHandler {
	if sc.TxTypeHandlerCalled == nil {
		return nil
	}
	return sc.TxTypeHandlerCalled()
}

// CheckSCRBeforeProcessing -
func (sc *SCProcessorMock) CheckSCRBeforeProcessing(scr *smartContractResult.SmartContractResult) (process.ScrProcessingDataHandler, error) {
	if sc.CheckSCRBeforeProcessingCalled == nil {
		return nil, nil
	}

	return sc.CheckSCRBeforeProcessingCalled(scr)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sc *SCProcessorMock) IsInterfaceNil() bool {
	return sc == nil
}
