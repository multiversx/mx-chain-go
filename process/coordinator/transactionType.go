package coordinator

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var _ process.TxTypeHandler = (*txTypeHandler)(nil)

type txTypeHandler struct {
	pubkeyConv             core.PubkeyConverter
	shardCoordinator       sharding.Coordinator
	builtInFunctions       vmcommon.BuiltInFunctionContainer
	argumentParser         process.CallArgumentsParser
	flagRelayedTxV2        atomic.Flag
	relayedTxV2EnableEpoch uint32
	esdtTransferParser     vmcommon.ESDTTransferParser
}

// ArgNewTxTypeHandler defines the arguments needed to create a new tx type handler
type ArgNewTxTypeHandler struct {
	PubkeyConverter        core.PubkeyConverter
	ShardCoordinator       sharding.Coordinator
	BuiltInFunctions       vmcommon.BuiltInFunctionContainer
	ArgumentParser         process.CallArgumentsParser
	RelayedTxV2EnableEpoch uint32
	EpochNotifier          process.EpochNotifier
	ESDTTransferParser     vmcommon.ESDTTransferParser
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
	if check.IfNil(args.EpochNotifier) {
		return nil, process.ErrNilEpochNotifier
	}
	if check.IfNil(args.ESDTTransferParser) {
		return nil, process.ErrNilESDTTransferParser
	}

	tc := &txTypeHandler{
		pubkeyConv:             args.PubkeyConverter,
		shardCoordinator:       args.ShardCoordinator,
		argumentParser:         args.ArgumentParser,
		builtInFunctions:       args.BuiltInFunctions,
		relayedTxV2EnableEpoch: args.RelayedTxV2EnableEpoch,
		esdtTransferParser:     args.ESDTTransferParser,
	}

	args.EpochNotifier.RegisterNotifyHandler(tc)
	log.Debug("txTypeHandler: enable epoch for relayed transactions v2", "epoch", args.RelayedTxV2EnableEpoch)

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

	if isAsynchronousCallBack(tx) {
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

func isAsynchronousCallBack(tx data.TransactionHandler) bool {
	scr, ok := tx.(*smartContractResult.SmartContractResult)
	if !ok {
		return false
	}

	return scr.CallType == vmcommon.AsynchronousCallBack
}

func (tth *txTypeHandler) isSCCallAfterBuiltIn(function string, args [][]byte, tx data.TransactionHandler) bool {
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
	if !tth.flagRelayedTxV2.IsSet() {
		return false
	}
	return functionName == core.RelayedTransactionV2
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

// EpochConfirmed is called whenever a new epoch is confirmed
func (tth *txTypeHandler) EpochConfirmed(epoch uint32, _ uint64) {
	tth.flagRelayedTxV2.Toggle(epoch >= tth.relayedTxV2EnableEpoch)
	log.Debug("txTypeHandler: relayed transactions v2", "enabled", tth.flagRelayedTxV2.IsSet())
}

// IsInterfaceNil returns true if there is no value under the interface
func (tth *txTypeHandler) IsInterfaceNil() bool {
	return tth == nil
}
