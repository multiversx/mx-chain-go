package requestHandlers

import (
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// ArgMiniblockRequestHandler is the argument structure used to create a new a miniblock request handler instance
type ArgMiniblockRequestHandler struct {
	ArgBaseRequestHandler
}

type miniblockRequestHandler struct {
	*baseRequestHandler
}

// NewMiniblockRequestHandler returns a new instance of miniblock request handler
func NewMiniblockRequestHandler(args ArgMiniblockRequestHandler) (*miniblockRequestHandler, error) {
	err := checkArgBase(args.ArgBaseRequestHandler)
	if err != nil {
		return nil, err
	}

	return &miniblockRequestHandler{
		baseRequestHandler: createBaseRequestHandler(args.ArgBaseRequestHandler),
	}, nil
}

// RequestDataFromHashArray requests a block body from other peers having input the block body hash
func (handler *miniblockRequestHandler) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	b := &batch.Batch{
		Data: hashes,
	}
	batchBytes, err := handler.marshaller.Marshal(b)
	if err != nil {
		return err
	}

	return handler.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: batchBytes,
			Epoch: epoch,
		},
		hashes,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *miniblockRequestHandler) IsInterfaceNil() bool {
	return handler == nil
}
