package smartContract

import (
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

func (sc *scProcessor) createVMDeployInput(tx data.TransactionHandler) (*vmcommon.ContractCreateInput, []byte, error) {
	vmInput, err := sc.createVMInput(tx)
	if err != nil {
		return nil, nil, err
	}

	vmType, err := sc.argsParser.GetVMType()
	if err != nil {
		return nil, nil, err
	}

	codeMetadata, err := sc.argsParser.GetCodeMetadata()
	if err != nil {
		return nil, nil, err
	}

	vmInput.Arguments, err = sc.argsParser.GetConstructorArguments()
	if err != nil {
		return nil, nil, err
	}

	vmCreateInput := &vmcommon.ContractCreateInput{}
	vmCreateInput.VMInput = *vmInput
	vmCreateInput.ContractCode, err = sc.getContractCode()
	if err != nil {
		return nil, nil, err
	}

	vmCreateInput.ContractCodeMetadata = codeMetadata.ToBytes()

	return vmCreateInput, vmType, nil
}

func (sc *scProcessor) createVMInput(tx data.TransactionHandler) (*vmcommon.VMInput, error) {
	vmInput := &vmcommon.VMInput{}
	vmInput.CallType = determineCallType(tx)
	vmInput.CallerAddr = tx.GetSndAddr()
	vmInput.CallValue = new(big.Int).Set(tx.GetValue())
	vmInput.GasPrice = tx.GetGasPrice()

	txData := prependCallbackToTxDataIfAsyncCall(tx.GetData(), vmInput.CallType)

	err := sc.argsParser.ParseData(string(txData))
	if err != nil {
		return nil, err
	}

	vmInput.GasProvided, err = sc.prepareGasProvided(tx)
	if err != nil {
		return nil, err
	}

	return vmInput, nil
}

func determineCallType(tx data.TransactionHandler) vmcommon.CallType {
	scr, isSCR := tx.(*smartContractResult.SmartContractResult)
	if isSCR {
		return scr.CallType
	}

	return vmcommon.DirectCall
}

// TODO: Check if this is still needed (and if needed, it does not seem entirely correct: "callBack" + "@")
func prependCallbackToTxDataIfAsyncCall(txData []byte, callType vmcommon.CallType) []byte {
	if callType == vmcommon.AsynchronousCallBack {
		return append([]byte("callBack"), txData...)
	}

	return txData
}

func (sc *scProcessor) prepareGasProvided(tx data.TransactionHandler) (uint64, error) {
	gasForTxData := sc.economicsFee.ComputeGasLimit(tx)
	if tx.GetGasLimit() < gasForTxData {
		return 0, process.ErrNotEnoughGas
	}

	return tx.GetGasLimit() - gasForTxData, nil
}

// TODO: move to argsParser
func (sc *scProcessor) getContractCode() ([]byte, error) {
	codeHex, err := sc.argsParser.GetCode()
	if err != nil {
		return nil, err
	}

	code, err := hex.DecodeString(string(codeHex))
	if err != nil {
		return nil, err
	}

	return code, err
}

func (sc *scProcessor) createVMCallInput(tx data.TransactionHandler) (*vmcommon.ContractCallInput, error) {
	vmInput, err := sc.createVMInput(tx)
	if err != nil {
		return nil, err
	}
	vmInput.Arguments, err = sc.argsParser.GetFunctionArguments()
	if err != nil {
		return nil, err
	}

	vmCallInput := &vmcommon.ContractCallInput{}
	vmCallInput.VMInput = *vmInput
	vmCallInput.Function, err = sc.argsParser.GetFunction()
	if err != nil {
		return nil, err
	}

	vmCallInput.RecipientAddr = tx.GetRcvAddr()

	return vmCallInput, nil
}
