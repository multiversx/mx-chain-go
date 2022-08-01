package blockAPI

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
)

var (
	errNilArgAPIBlockProcessor = errors.New("nil arg api block processor")
	errNilTransactionHandler   = errors.New("nil API transaction handler")
	errNilLogsFacade           = errors.New("nil logs facade")
	errNilReceiptsRepository   = errors.New("nil receipts repository")
)

func checkNilArg(arg *ArgAPIBlockProcessor) error {
	if arg == nil {
		return errNilArgAPIBlockProcessor
	}
	if check.IfNil(arg.Uint64ByteSliceConverter) {
		return process.ErrNilUint64Converter
	}
	if check.IfNil(arg.Store) {
		return process.ErrNilStorage
	}
	if check.IfNil(arg.Marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(arg.HistoryRepo) {
		return process.ErrNilHistoryRepository
	}
	if check.IfNil(arg.APITransactionHandler) {
		return errNilTransactionHandler
	}
	if check.IfNil(arg.Hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(arg.AddressPubkeyConverter) {
		return process.ErrNilPubkeyConverter
	}
	if check.IfNil(arg.LogsFacade) {
		return errNilLogsFacade
	}
	if check.IfNil(arg.ReceiptsRepository) {
		return errNilReceiptsRepository
	}

	return nil
}
