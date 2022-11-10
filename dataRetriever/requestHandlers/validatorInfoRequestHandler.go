package requestHandlers

import (
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// ArgValidatorInfoRequestHandler is the argument structure used to create a new a validator info request handler instance
type ArgValidatorInfoRequestHandler struct {
	ArgBaseRequestHandler
}

type validatorInfoRequestHandler struct {
	*baseRequestHandler
}

// NewValidatorInfoRequestHandler returns a new instance of validator info request handler
func NewValidatorInfoRequestHandler(args ArgValidatorInfoRequestHandler) (*validatorInfoRequestHandler, error) {
	err := checkArgBase(args.ArgBaseRequestHandler)
	if err != nil {
		return nil, err
	}

	return &validatorInfoRequestHandler{
		baseRequestHandler: createBaseRequestHandler(args.ArgBaseRequestHandler),
	}, nil
}

// RequestDataFromHashArray requests validator info from other peers by hash array
func (handler *validatorInfoRequestHandler) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	b := &batch.Batch{
		Data: hashes,
	}
	buffHashes, err := handler.marshaller.Marshal(b)
	if err != nil {
		return err
	}

	return handler.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: buffHashes,
			Epoch: epoch,
		},
		hashes,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *validatorInfoRequestHandler) IsInterfaceNil() bool {
	return handler == nil
}
