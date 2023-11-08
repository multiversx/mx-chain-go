package processorV2

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-go/process"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

func (sc *scProcessor) createVMDeployInput(tx data.TransactionHandler) (*vmcommon.ContractCreateInput, []byte, error) {
	deployData, err := sc.argsParser.ParseDeployData(string(tx.GetData()))
	if err != nil {
		return nil, nil, err
	}

	vmCreateInput := &vmcommon.ContractCreateInput{}
	vmCreateInput.ContractCode = deployData.Code
	// when executing SC deploys we should always apply the flags
	codeMetadata := sc.blockChainHook.ApplyFiltersOnSCCodeMetadata(deployData.CodeMetadata)
	vmCreateInput.ContractCodeMetadata = codeMetadata.ToBytes()
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
	vmInput.OriginalCallerAddr = GetOriginalSenderForTx(tx)
	vmInput.CallerAddr = tx.GetSndAddr()
	vmInput.CallValue = new(big.Int).Set(tx.GetValue())
	vmInput.GasPrice = tx.GetGasPrice()
	vmInput.GasProvided, err = sc.prepareGasProvided(tx)
	if err != nil {
		return err
	}

	return nil
}

func isSmartContractResult(tx data.TransactionHandler) bool {
	_, isScr := tx.(*smartContractResult.SmartContractResult)
	return isScr
}

func (sc *scProcessor) prepareGasProvided(tx data.TransactionHandler) (uint64, error) {
	if isSmartContractResult(tx) {
		return tx.GetGasLimit(), nil
	}

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
		txData = string(prependLegacyCallbackFunctionNameToTxDataIfAsyncCallBack(tx.GetData(), callType))
	}

	function, arguments, err := sc.argsParser.ParseCallData(txData)
	if err != nil {
		return nil, err
	}

	finalArguments, gasLocked := getAsyncCallGasLockFromTxData(callType, arguments)

	asyncArguments, callArguments := separateAsyncArguments(callType, finalArguments)

	vmCallInput := &vmcommon.ContractCallInput{}
	vmCallInput.VMInput = vmcommon.VMInput{}
	vmCallInput.CallType = callType
	vmCallInput.RecipientAddr = tx.GetRcvAddr()
	vmCallInput.Function = function
	vmCallInput.CurrentTxHash = txHash
	vmCallInput.GasLocked = gasLocked

	gtx, ok := tx.(data.GuardedTransactionHandler)
	if ok {
		vmCallInput.TxGuardian = gtx.GetGuardianAddr()
	}

	scr, isSCR := tx.(*smartContractResult.SmartContractResult)
	if isSCR {
		vmCallInput.OriginalTxHash = scr.GetOriginalTxHash()
		vmCallInput.PrevTxHash = scr.PrevTxHash
	} else {
		vmCallInput.OriginalTxHash = txHash
		vmCallInput.PrevTxHash = txHash
	}

	vmCallInput.ReturnCallAfterError = isSCR && len(scr.ReturnMessage) > 0

	err = sc.initializeVMInputFromTx(&vmCallInput.VMInput, tx)
	if err != nil {
		return nil, err
	}

	vmCallInput.VMInput.AsyncArguments = buildAsyncArgumentsObject(callType, asyncArguments)
	vmCallInput.VMInput.Arguments = callArguments
	if vmCallInput.GasProvided > tx.GetGasLimit() {
		return nil, process.ErrInvalidVMInputGasComputation
	}

	vmCallInput.GasProvided, err = core.SafeSubUint64(vmCallInput.GasProvided, gasLocked)
	if err != nil {
		return nil, err
	}

	return vmCallInput, nil
}

func getAsyncCallGasLockFromTxData(callType vm.CallType, arguments [][]byte) ([][]byte, uint64) {
	if callType != vm.AsynchronousCall {
		return arguments, 0
	}
	lenArgs := len(arguments)
	if lenArgs == 0 {
		return arguments, 0
	}

	lastArg := arguments[lenArgs-1]
	gasLocked := big.NewInt(0).SetBytes(lastArg).Uint64()

	argsWithoutGasLocked := make([][]byte, lenArgs-1)
	copy(argsWithoutGasLocked, arguments[:lenArgs-1])

	return argsWithoutGasLocked, gasLocked
}

func separateAsyncArguments(callType vm.CallType, arguments [][]byte) ([][]byte, [][]byte) {
	if callType == vm.DirectCall || callType == vm.ESDTTransferAndExecute {
		return nil, arguments
	}

	var noOfAsyncArguments int
	if callType == vm.AsynchronousCall {
		noOfAsyncArguments = 2
	} else {
		noOfAsyncArguments = 4
	}

	noOfCallArguments := len(arguments) - noOfAsyncArguments
	if noOfCallArguments < 0 {
		log.Error("noOfCallArguments < 0")
	}
	asyncArguments := make([][]byte, noOfAsyncArguments)
	callArguments := make([][]byte, noOfCallArguments)

	copy(callArguments, arguments[:noOfCallArguments])
	copy(asyncArguments, arguments[noOfCallArguments:])

	return asyncArguments, callArguments
}

func buildAsyncArgumentsObject(callType vm.CallType, asyncArguments [][]byte) *vmcommon.AsyncArguments {
	switch callType {
	case vm.AsynchronousCall:
		return &vmcommon.AsyncArguments{
			CallID:       asyncArguments[0],
			CallerCallID: asyncArguments[1],
		}
	case vm.AsynchronousCallBack:
		return &vmcommon.AsyncArguments{
			CallID:                       asyncArguments[0],
			CallerCallID:                 asyncArguments[1],
			CallbackAsyncInitiatorCallID: asyncArguments[2],
			GasAccumulated:               big.NewInt(0).SetBytes(asyncArguments[3]).Uint64(),
		}
	default:
		return &vmcommon.AsyncArguments{}
	}
}

func determineCallType(tx data.TransactionHandler) vm.CallType {
	scr, isSCR := tx.(*smartContractResult.SmartContractResult)
	if isSCR {
		return scr.CallType
	}

	return vm.DirectCall
}

func prependLegacyCallbackFunctionNameToTxDataIfAsyncCallBack(txData []byte, callType vm.CallType) []byte {
	if callType == vm.AsynchronousCallBack {
		return append([]byte("callBack"), txData...)
	}

	return txData
}
