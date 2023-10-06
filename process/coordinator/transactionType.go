package coordinator

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var _ process.TxTypeHandler = (*txTypeHandler)(nil)

type txTypeHandler struct {
	pubkeyConv          core.PubkeyConverter
	shardCoordinator    sharding.Coordinator
	builtInFunctions    vmcommon.BuiltInFunctionContainer
	argumentParser      process.CallArgumentsParser
	esdtTransferParser  vmcommon.ESDTTransferParser
	enableEpochsHandler common.EnableEpochsHandler
}

// ArgNewTxTypeHandler defines the arguments needed to create a new tx type handler
type ArgNewTxTypeHandler struct {
	PubkeyConverter     core.PubkeyConverter
	ShardCoordinator    sharding.Coordinator
	BuiltInFunctions    vmcommon.BuiltInFunctionContainer
	ArgumentParser      process.CallArgumentsParser
	ESDTTransferParser  vmcommon.ESDTTransferParser
	EnableEpochsHandler common.EnableEpochsHandler
}

// NewTxTypeHandler creates a transaction type handler
func NewTxTypeHandler(
	args ArgNewTxTypeHandler,
) (*txTypeHandler, error) {
	if check.IfNil(args.PubkeyConverter) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.ArgumentParser) {
		return nil, process.ErrNilArgumentParser
	}
	if check.IfNil(args.BuiltInFunctions) {
		return nil, process.ErrNilBuiltInFunction
	}
	if check.IfNil(args.ESDTTransferParser) {
		return nil, process.ErrNilESDTTransferParser
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}

	tc := &txTypeHandler{
		pubkeyConv:          args.PubkeyConverter,
		shardCoordinator:    args.ShardCoordinator,
		argumentParser:      args.ArgumentParser,
		builtInFunctions:    args.BuiltInFunctions,
		esdtTransferParser:  args.ESDTTransferParser,
		enableEpochsHandler: args.EnableEpochsHandler,
	}

	return tc, nil
}

// ComputeTransactionType calculates the transaction type
func (tth *txTypeHandler) ComputeTransactionType(tx data.TransactionHandler) (process.TransactionType, process.TransactionType) {
	err := tth.checkTxValidity(tx)
	if err != nil {
		return process.InvalidTransaction, process.InvalidTransaction
	}

	isEmptyAddress := tth.isDestAddressEmpty(tx)
	if isEmptyAddress {
		if len(tx.GetData()) > 0 {
			return process.SCDeployment, process.SCDeployment
		}
		return process.InvalidTransaction, process.InvalidTransaction
	}

	if tth.isRelayedTransactionV3(tx) {
		return process.RelayedTxV3, process.RelayedTxV3
	}

	if len(tx.GetData()) == 0 {
		return process.MoveBalance, process.MoveBalance
	}

	funcName, args := tth.getFunctionFromArguments(tx.GetData())
	isBuiltInFunction := tth.isBuiltInFunctionCall(funcName)
	if isBuiltInFunction {
		if tth.isSCCallAfterBuiltIn(funcName, args, tx) {
			return process.BuiltInFunctionCall, process.SCInvoking
		}

		return process.BuiltInFunctionCall, process.BuiltInFunctionCall
	}

	if isCallOfType(tx, vm.AsynchronousCallBack) {
		return process.SCInvoking, process.SCInvoking
	}

	if len(funcName) == 0 {
		return process.MoveBalance, process.MoveBalance
	}

	if tth.isRelayedTransactionV1(funcName) {
		return process.RelayedTx, process.RelayedTx
	}

	if tth.isRelayedTransactionV2(funcName) {
		return process.RelayedTxV2, process.RelayedTxV2
	}

	isDestInSelfShard := tth.isDestAddressInSelfShard(tx.GetRcvAddr())
	if isDestInSelfShard && core.IsSmartContractAddress(tx.GetRcvAddr()) {
		return process.SCInvoking, process.SCInvoking
	}

	if core.IsSmartContractAddress(tx.GetRcvAddr()) {
		return process.MoveBalance, process.SCInvoking
	}

	return process.MoveBalance, process.MoveBalance
}

func isCallOfType(tx data.TransactionHandler, callType vm.CallType) bool {
	scr, ok := tx.(*smartContractResult.SmartContractResult)
	if !ok {
		return false
	}

	return scr.CallType == callType
}

func (tth *txTypeHandler) isSCCallAfterBuiltIn(function string, args [][]byte, tx data.TransactionHandler) bool {
	isTransferAndAsyncCallbackFixFlagSet := tth.enableEpochsHandler.IsESDTMetadataContinuousCleanupFlagEnabled()
	if isTransferAndAsyncCallbackFixFlagSet && isCallOfType(tx, vm.AsynchronousCallBack) {
		return true
	}
	if len(args) <= 2 {
		return false
	}

	parsedTransfer, err := tth.esdtTransferParser.ParseESDTTransfers(tx.GetSndAddr(), tx.GetRcvAddr(), function, args)
	if err != nil {
		return false
	}
	if len(parsedTransfer.CallFunction) == 0 {
		return false
	}
	return core.IsSmartContractAddress(parsedTransfer.RcvAddr)
}

func (tth *txTypeHandler) getFunctionFromArguments(txData []byte) (string, [][]byte) {
	if len(txData) == 0 {
		return "", nil
	}

	function, args, err := tth.argumentParser.ParseData(string(txData))
	if err != nil {
		return "", nil
	}

	return function, args
}

func (tth *txTypeHandler) isBuiltInFunctionCall(functionName string) bool {
	function, err := tth.builtInFunctions.Get(functionName)
	if err != nil {
		return false
	}

	return function.IsActive()
}

func (tth *txTypeHandler) isRelayedTransactionV1(functionName string) bool {
	return functionName == core.RelayedTransaction
}

func (tth *txTypeHandler) isRelayedTransactionV2(functionName string) bool {
	return functionName == core.RelayedTransactionV2
}

func (tth *txTypeHandler) isRelayedTransactionV3(tx data.TransactionHandler) bool {
	return len(tx.GetUserTransactions()) > 0
}

func (tth *txTypeHandler) isDestAddressEmpty(tx data.TransactionHandler) bool {
	isEmptyAddress := bytes.Equal(tx.GetRcvAddr(), make([]byte, tth.pubkeyConv.Len()))
	return isEmptyAddress
}

func (tth *txTypeHandler) isDestAddressInSelfShard(address []byte) bool {
	shardForCurrentNode := tth.shardCoordinator.SelfId()
	shardForSrc := tth.shardCoordinator.ComputeId(address)
	return shardForCurrentNode == shardForSrc
}

func (tth *txTypeHandler) checkTxValidity(tx data.TransactionHandler) error {
	if check.IfNil(tx) {
		return process.ErrNilTransaction
	}

	recvAddressIsInvalid := tth.pubkeyConv.Len() != len(tx.GetRcvAddr())
	if recvAddressIsInvalid {
		return process.ErrWrongTransaction
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tth *txTypeHandler) IsInterfaceNil() bool {
	return tth == nil
}
