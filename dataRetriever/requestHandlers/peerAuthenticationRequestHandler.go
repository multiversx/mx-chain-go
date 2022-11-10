package requestHandlers

import (
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// ArgPeerAuthenticationRequestHandler is the argument structure used to create a new a peer authentication request handler instance
type ArgPeerAuthenticationRequestHandler struct {
	ArgBaseRequestHandler
}

type peerAuthenticationRequestHandler struct {
	*baseRequestHandler
}

// NewPeerAuthenticationRequestHandler returns a new instance of peer authentication request handler
func NewPeerAuthenticationRequestHandler(args ArgPeerAuthenticationRequestHandler) (*peerAuthenticationRequestHandler, error) {
	err := checkArgBase(args.ArgBaseRequestHandler)
	if err != nil {
		return nil, err
	}

	return &peerAuthenticationRequestHandler{
		baseRequestHandler: createBaseRequestHandler(args.ArgBaseRequestHandler),
	}, nil
}

// RequestDataFromHashArray requests peer authentication data from other peers having input multiple public key hashes
func (handler *peerAuthenticationRequestHandler) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
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
func (handler *peerAuthenticationRequestHandler) IsInterfaceNil() bool {
	return handler == nil
}
