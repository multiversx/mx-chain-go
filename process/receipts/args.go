package receipts

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
)

// ArgsNewReceiptsRepository holds arguments for creating a receiptsRepository
type ArgsNewReceiptsRepository struct {
	Marshaller marshal.Marshalizer
	Hasher     hashing.Hasher
	Store      dataRetriever.StorageService
}

func (args *ArgsNewReceiptsRepository) check() error {
	if check.IfNil(args.Marshaller) {
		return core.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return core.ErrNilHasher
	}
	if check.IfNil(args.Store) {
		return core.ErrNilStore
	}

	return nil
}
