package requestHandlers

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// ArgHeaderRequestHandler is the argument structure used to create a new a header request handler instance
type ArgHeaderRequestHandler struct {
	ArgBaseRequestHandler
	NonceConverter typeConverters.Uint64ByteSliceConverter
}

type headerRequestHandler struct {
	*baseRequestHandler
	nonceConverter typeConverters.Uint64ByteSliceConverter
}

// NewHeaderRequestHandler returns a new instance of header request handler
func NewHeaderRequestHandler(args ArgHeaderRequestHandler) (*headerRequestHandler, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &headerRequestHandler{
		baseRequestHandler: createBaseRequestHandler(args.ArgBaseRequestHandler),
		nonceConverter:     args.NonceConverter,
	}, nil
}

func checkArgs(args ArgHeaderRequestHandler) error {
	err := checkArgBase(args.ArgBaseRequestHandler)
	if err != nil {
		return err
	}
	if check.IfNilReflect(args.NonceConverter) {
		return dataRetriever.ErrNilUint64ByteSliceConverter
	}

	return nil
}

// RequestDataFromNonce requests a header from other peers having input the hdr nonce
func (handler *headerRequestHandler) RequestDataFromNonce(nonce uint64, epoch uint32) error {
	return handler.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.NonceType,
			Value: handler.nonceConverter.ToByteSlice(nonce),
			Epoch: epoch,
		},
		[][]byte{handler.nonceConverter.ToByteSlice(nonce)},
	)
}

// RequestDataFromEpoch requests a header from other peers having input the epoch
func (handler *headerRequestHandler) RequestDataFromEpoch(identifier []byte) error {
	return handler.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.EpochType,
			Value: identifier,
		},
		[][]byte{identifier},
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *headerRequestHandler) IsInterfaceNil() bool {
	return handler == nil
}
