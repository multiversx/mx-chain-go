package coordinator

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

var _ process.TxTypeHandler = (*txTypeHandler)(nil)

type txTypeHandler struct {
	pubkeyConv       core.PubkeyConverter
	shardCoordinator sharding.Coordinator
	builtInFuncNames map[string]struct{}
	argumentParser   process.CallArgumentsParser
}

// ArgNewTxTypeHandler defines the arguments needed to create a new tx type handler
type ArgNewTxTypeHandler struct {
	PubkeyConverter  core.PubkeyConverter
	ShardCoordinator sharding.Coordinator
	BuiltInFuncNames map[string]struct{}
	ArgumentParser   process.CallArgumentsParser
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
	if args.BuiltInFuncNames == nil {
		return nil, process.ErrNilBuiltInFunction
	}

	tc := &txTypeHandler{
		pubkeyConv:       args.PubkeyConverter,
		shardCoordinator: args.ShardCoordinator,
		argumentParser:   args.ArgumentParser,
		builtInFuncNames: args.BuiltInFuncNames,
	}

	return tc, nil
}

// ComputeTransactionType calculates the transaction type
func (tth *txTypeHandler) ComputeTransactionType(tx data.TransactionHandler) process.TransactionType {
	err := tth.checkTxValidity(tx)
	if err != nil {
		return process.InvalidTransaction
	}

	isEmptyAddress := tth.isDestAddressEmpty(tx)
	if isEmptyAddress {
		if len(tx.GetData()) > 0 {
			return process.SCDeployment
		}
		return process.InvalidTransaction
	}

	if len(tx.GetData()) == 0 {
		return process.MoveBalance
	}

	funcName := tth.getFunctionFromArguments(tx.GetData())
	isBuiltInFunction := tth.isBuiltInFunctionCall(funcName)
	if isAsynchronousCallBack(tx) {
		if isBuiltInFunction {
			return process.BuiltInFunctionCall
		}
		return process.SCInvoking
	}

	if len(funcName) == 0 {
		return process.MoveBalance
	}

	if tth.isRelayedTransaction(funcName) {
		return process.RelayedTx
	}

	isDestInSelfShard := tth.isDestAddressInSelfShard(tx.GetRcvAddr())
	if !isBuiltInFunction && !isDestInSelfShard {
		return process.MoveBalance
	}

	if isBuiltInFunction {
		return process.BuiltInFunctionCall
	}

	if core.IsSmartContractAddress(tx.GetRcvAddr()) {
		return process.SCInvoking
	}

	return process.MoveBalance
}

func isAsynchronousCallBack(tx data.TransactionHandler) bool {
	scr, ok := tx.(*smartContractResult.SmartContractResult)
	if !ok {
		return false
	}

	return scr.CallType == vmcommon.AsynchronousCallBack
}

func (tth *txTypeHandler) getFunctionFromArguments(txData []byte) string {
	if len(txData) == 0 {
		return ""
	}

	function, _, err := tth.argumentParser.ParseData(string(txData))
	if err != nil {
		return ""
	}

	return function
}

func (tth *txTypeHandler) isBuiltInFunctionCall(functionName string) bool {
	if len(tth.builtInFuncNames) == 0 {
		return false
	}

	_, ok := tth.builtInFuncNames[functionName]
	return ok
}

func (tth *txTypeHandler) isRelayedTransaction(functionName string) bool {
	return functionName == core.RelayedTransaction
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
