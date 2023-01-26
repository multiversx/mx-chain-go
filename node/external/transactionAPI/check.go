package transactionAPI

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
)

func checkNilArgs(arg *ArgAPITransactionProcessor) error {
	if arg == nil {
		return ErrNilAPITransactionProcessorArg
	}
	if check.IfNil(arg.Marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(arg.DataPool) {
		return process.ErrNilDataPoolHolder
	}
	if check.IfNil(arg.HistoryRepository) {
		return process.ErrNilHistoryRepository
	}
	if check.IfNil(arg.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(arg.AddressPubKeyConverter) {
		return process.ErrNilPubkeyConverter
	}
	if check.IfNil(arg.StorageService) {
		return process.ErrNilStorage
	}
	if check.IfNil(arg.Uint64ByteSliceConverter) {
		return process.ErrNilUint64Converter
	}
	if check.IfNil(arg.FeeComputer) {
		return ErrNilFeeComputer
	}
	if check.IfNil(arg.TxTypeHandler) {
		return process.ErrNilTxTypeHandler
	}
	if check.IfNil(arg.LogsFacade) {
		return ErrNilLogsFacade
	}
	if check.IfNilReflect(arg.DataFieldParser) {
		return ErrNilDataFieldParser
	}

	return nil
}
