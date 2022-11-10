package requestHandlers

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// ArgBaseRequestHandler is the argument structure used as base to create a new a request handler instance
type ArgBaseRequestHandler struct {
	SenderResolver dataRetriever.TopicResolverSender
	Marshaller     marshal.Marshalizer
}

type baseRequestHandler struct {
	dataRetriever.TopicResolverSender
	marshaller marshal.Marshalizer
}

func createBaseRequestHandler(args ArgBaseRequestHandler) *baseRequestHandler {
	return &baseRequestHandler{
		TopicResolverSender: args.SenderResolver,
		marshaller:          args.Marshaller,
	}
}

func checkArgBase(arg ArgBaseRequestHandler) error {
	if check.IfNil(arg.SenderResolver) {
		return dataRetriever.ErrNilResolverSender
	}
	if check.IfNil(arg.Marshaller) {
		return dataRetriever.ErrNilMarshalizer
	}
	return nil
}

// RequestDataFromHash requests data from other peers having input a hash
func (handler *baseRequestHandler) RequestDataFromHash(hash []byte, epoch uint32) error {
	return handler.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashType,
			Value: hash,
			Epoch: epoch,
		},
		[][]byte{hash},
	)
}

// SetNumPeersToQuery sets the number of intra shard and cross shard number of peer to query
func (handler *baseRequestHandler) SetNumPeersToQuery(intra int, cross int) {
	handler.TopicResolverSender.SetNumPeersToQuery(intra, cross)
}

// NumPeersToQuery returns the number of intra shard and cross shard number of peer to query
func (handler *baseRequestHandler) NumPeersToQuery() (int, int) {
	return handler.TopicResolverSender.NumPeersToQuery()
}

// SetResolverDebugHandler sets a resolver debug handler
func (handler *baseRequestHandler) SetResolverDebugHandler(debugHandler dataRetriever.ResolverDebugHandler) error {
	return handler.TopicResolverSender.SetResolverDebugHandler(debugHandler)
}
