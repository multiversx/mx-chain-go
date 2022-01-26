package transactionAPI

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
)

func checkNilArgs(args *APITransactionProcessorArgs) error {
	if args == nil {
		return ErrNilAPITransactionProcessorArg
	}
	if check.IfNil(args.Marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(args.DataPool) {
		return process.ErrNilDataPoolHolder
	}
	if check.IfNil(args.HistoryRepository) {
		return process.ErrNilHistoryRepository
	}
	if check.IfNil(args.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(args.AddressPubKeyConverter) {
		return process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.StorageService) {
		return process.ErrNilStorage
	}
	if check.IfNil(args.Uint64ByteSliceConverter) {
		return process.ErrNilUint64Converter
	}

	return nil
}
