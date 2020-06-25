package smartContract

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

func (sc *scProcessor) createVMDeployInput(tx data.TransactionHandler) (*vmcommon.ContractCreateInput, []byte, error) {
	err := sc.argsParser.ParseData(string(tx.GetData()))
	if err != nil {
		return nil, nil, err
	}

	vmCreateInput := &vmcommon.ContractCreateInput{}
	vmCreateInput.ContractCode, err = sc.argsParser.GetCodeDecoded()
	if err != nil {
		return nil, nil, err
	}

	vmCreateInput.ContractCodeMetadata, err = sc.getCodeMetadata()
	if err != nil {
		return nil, nil, err
	}

	vmCreateInput.VMInput = vmcommon.VMInput{}
	err = sc.initializeVMInputFromTx(&vmCreateInput.VMInput, tx)
	if err != nil {
		return nil, nil, err
	}

	vmCreateInput.VMInput.Arguments, err = sc.argsParser.GetConstructorArguments()
	if err != nil {
		return nil, nil, err
	}

	vmType, err := sc.getVMType()
	if err != nil {
		return nil, nil, err
	}

	return vmCreateInput, vmType, nil
}

func (sc *scProcessor) getCodeMetadata() ([]byte, error) {
	codeMetadata, err := sc.argsParser.GetCodeMetadata()
	if err != nil {
		return nil, err
	}

	return codeMetadata.ToBytes(), nil
}

func (sc *scProcessor) initializeVMInputFromTx(vmInput *vmcommon.VMInput, tx data.TransactionHandler) error {
	var err error

	vmInput.CallerAddr = tx.GetSndAddr()
	vmInput.CallValue = new(big.Int).Set(tx.GetValue())
	vmInput.GasPrice = tx.GetGasPrice()
	vmInput.GasProvided, err = sc.prepareGasProvided(tx)
	if err != nil {
		return err
	}

	return nil
}

func (sc *scProcessor) prepareGasProvided(tx data.TransactionHandler) (uint64, error) {
	if sc.shardCoordinator.ComputeId(tx.GetSndAddr()) == core.MetachainShardId {
		return tx.GetGasLimit(), nil
	}

	gasForTxData := sc.economicsFee.ComputeGasLimit(tx)
	if tx.GetGasLimit() < gasForTxData {
		return 0, process.ErrNotEnoughGas
	}

	return tx.GetGasLimit() - gasForTxData, nil
}

func (sc *scProcessor) getVMType() ([]byte, error) {
	vmType, err := sc.argsParser.GetVMType()
	if err != nil {
		return nil, err
	}

	if len(vmType) != core.VMTypeLen {
		return nil, process.ErrVMTypeLengthInvalid
	}

	return vmType, nil
}

func (sc *scProcessor) createVMCallInput(tx data.TransactionHandler) (*vmcommon.ContractCallInput, error) {
	callType := determineCallType(tx)
	txData := prependCallbackToTxDataIfAsyncCall(tx.GetData(), callType)

	err := sc.argsParser.ParseData(string(txData))
	if err != nil {
		return nil, err
	}

	vmCallInput := &vmcommon.ContractCallInput{}
	vmCallInput.VMInput = vmcommon.VMInput{}
	vmCallInput.CallType = callType
	vmCallInput.RecipientAddr = tx.GetRcvAddr()
	vmCallInput.Function, err = sc.argsParser.GetFunction()
	if err != nil {
		return nil, err
	}

	err = sc.initializeVMInputFromTx(&vmCallInput.VMInput, tx)
	if err != nil {
		return nil, err
	}

	vmCallInput.VMInput.Arguments, err = sc.argsParser.GetFunctionArguments()
	if err != nil {
		return nil, err
	}

	return vmCallInput, nil
}

func determineCallType(tx data.TransactionHandler) vmcommon.CallType {
	scr, isSCR := tx.(*smartContractResult.SmartContractResult)
	if isSCR {
		return scr.CallType
	}

	return vmcommon.DirectCall
}

func prependCallbackToTxDataIfAsyncCall(txData []byte, callType vmcommon.CallType) []byte {
	if callType == vmcommon.AsynchronousCallBack {
		return append([]byte("callBack"), txData...)
	}

	return txData
}
