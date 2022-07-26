package receipts

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
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
