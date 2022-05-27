package logs

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgsNewFeeComputer holds the arguments for constructing a logsRepository
type ArgsNewLogsRepository struct {
	Storer          storage.Storer
	Marshalizer     marshal.Marshalizer
	PubKeyConverter core.PubkeyConverter
}

func (args *ArgsNewLogsRepository) check() error {
	if check.IfNil(args.Storer) {
		return ErrNilStorer
	}
	if check.IfNil(args.Marshalizer) {
		return core.ErrNilMarshalizer
	}
	if check.IfNil(args.PubKeyConverter) {
		return ErrNilPubkeyConverter
	}

	return nil
}
