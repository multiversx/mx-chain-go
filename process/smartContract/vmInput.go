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
	deployData, err := sc.argsParser.ParseDeployData(string(tx.GetData()))
	if err != nil {
		return nil, nil, err
	}

	vmCreateInput := &vmcommon.ContractCreateInput{}
	vmCreateInput.ContractCode = deployData.Code
	vmCreateInput.ContractCodeMetadata = deployData.CodeMetadata.ToBytes()
	vmCreateInput.VMInput = vmcommon.VMInput{}
	err = sc.initializeVMInputFromTx(&vmCreateInput.VMInput, tx)
	if err != nil {
		return nil, nil, err
	}

	vmCreateInput.VMInput.Arguments = deployData.Arguments

	return vmCreateInput, deployData.VMType, nil
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

func (sc *scProcessor) createVMCallInput(
	tx data.TransactionHandler,
	txHash []byte,
	builtInFuncCall bool,
) (*vmcommon.ContractCallInput, error) {
	callType := determineCallType(tx)
	txData := string(tx.GetData())
	if !builtInFuncCall {
		txData = string(prependCallbackToTxDataIfAsyncCall(tx.GetData(), callType))
	}

	function, arguments, err := sc.argsParser.ParseCallData(txData)
	if err != nil {
		return nil, err
	}

	vmCallInput := &vmcommon.ContractCallInput{}
	vmCallInput.VMInput = vmcommon.VMInput{}
	vmCallInput.CallType = callType
	vmCallInput.RecipientAddr = tx.GetRcvAddr()
	vmCallInput.Function = function
	vmCallInput.CurrentTxHash = txHash

	scr, isSCR := tx.(*smartContractResult.SmartContractResult)
	if isSCR {
		vmCallInput.OriginalTxHash = scr.GetOriginalTxHash()
	} else {
		vmCallInput.OriginalTxHash = txHash
	}

	err = sc.initializeVMInputFromTx(&vmCallInput.VMInput, tx)
	if err != nil {
		return nil, err
	}

	vmCallInput.VMInput.Arguments = arguments

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
