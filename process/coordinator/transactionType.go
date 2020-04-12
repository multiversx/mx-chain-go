package coordinator

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type txTypeHandler struct {
	adrConv          state.AddressConverter
	shardCoordinator sharding.Coordinator
	builtInFuncNames map[string]struct{}
	argumentParser   process.ArgumentsParser
}

// ArgNewTxTypeHandler defines the arguments needed to create a new tx type handler
type ArgNewTxTypeHandler struct {
	AddressConverter state.AddressConverter
	ShardCoordinator sharding.Coordinator
	BuiltInFuncNames map[string]struct{}
	ArgumentParser   process.ArgumentsParser
}

// NewTxTypeHandler creates a transaction type handler
func NewTxTypeHandler(
	args ArgNewTxTypeHandler,
) (*txTypeHandler, error) {
	if check.IfNil(args.AddressConverter) {
		return nil, process.ErrNilAddressConverter
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
		adrConv:          args.AddressConverter,
		shardCoordinator: args.ShardCoordinator,
		argumentParser:   args.ArgumentParser,
		builtInFuncNames: args.BuiltInFuncNames,
	}

	return tc, nil
}

// ComputeTransactionType calculates the transaction type
func (tth *txTypeHandler) ComputeTransactionType(tx data.TransactionHandler) (process.TransactionType, error) {
	err := tth.checkTxValidity(tx)
	if err != nil {
		return process.InvalidTransaction, err
	}

	isEmptyAddress := tth.isDestAddressEmpty(tx)
	if isEmptyAddress {
		if len(tx.GetData()) > 0 {
			return process.SCDeployment, nil
		}
		return process.InvalidTransaction, process.ErrWrongTransaction
	}

	isDestInSelfShard, err := tth.isDestAddressInSelfShard(tx.GetRcvAddr())
	if err != nil {
		return process.InvalidTransaction, err
	}

	if !isDestInSelfShard || len(tx.GetData()) == 0 {
		return process.MoveBalance, nil
	}

	if core.IsSmartContractAddress(tx.GetRcvAddr()) || tth.isBuiltInFunctionCall(tx.GetData()) {
		return process.SCInvoking, nil
	}

	return process.MoveBalance, nil
}

func (tth *txTypeHandler) isBuiltInFunctionCall(txData []byte) bool {
	if len(tth.builtInFuncNames) == 0 {
		return false
	}

	err := tth.argumentParser.ParseData(string(txData))
	if err != nil {
		return false
	}

	function, err := tth.argumentParser.GetFunction()
	if err != nil {
		return false
	}

	_, ok := tth.builtInFuncNames[function]
	return ok
}

func (tth *txTypeHandler) isDestAddressEmpty(tx data.TransactionHandler) bool {
	isEmptyAddress := bytes.Equal(tx.GetRcvAddr(), make([]byte, tth.adrConv.AddressLen()))
	return isEmptyAddress
}

func (tth *txTypeHandler) isDestAddressInSelfShard(address []byte) (bool, error) {
	adrSrc, err := tth.adrConv.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return false, err
	}

	shardForCurrentNode := tth.shardCoordinator.SelfId()
	shardForSrc := tth.shardCoordinator.ComputeId(adrSrc)
	if shardForCurrentNode != shardForSrc {
		return false, nil
	}

	return true, nil
}

func (tth *txTypeHandler) checkTxValidity(tx data.TransactionHandler) error {
	if check.IfNil(tx) {
		return process.ErrNilTransaction
	}

	recvAddressIsInvalid := tth.adrConv.AddressLen() != len(tx.GetRcvAddr())
	if recvAddressIsInvalid {
		return process.ErrWrongTransaction
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (tth *txTypeHandler) IsInterfaceNil() bool {
	return tth == nil
}
