package resolvers

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// ArgBaseResolver is the argument structure used as base to create a new a resolver instance
type ArgBaseResolver struct {
	SenderResolver   dataRetriever.TopicResolverSender
	Marshaller       marshal.Marshalizer
	AntifloodHandler dataRetriever.P2PAntifloodHandler
	Throttler        dataRetriever.ResolverThrottler
}

type baseResolver struct {
	dataRetriever.TopicResolverSender
}

func checkArgBase(arg ArgBaseResolver) error {
	if check.IfNil(arg.SenderResolver) {
		return dataRetriever.ErrNilResolverSender
	}
	if check.IfNil(arg.Marshaller) {
		return dataRetriever.ErrNilMarshaller
	}
	if check.IfNil(arg.AntifloodHandler) {
		return dataRetriever.ErrNilAntifloodHandler
	}
	if check.IfNil(arg.Throttler) {
		return dataRetriever.ErrNilThrottler
	}
	return nil
}

// SetResolverDebugHandler will set a resolver debug handler
func (res *baseResolver) SetResolverDebugHandler(handler dataRetriever.ResolverDebugHandler) error {
	return res.TopicResolverSender.SetResolverDebugHandler(handler)
}

// Close returns nil
func (res *baseResolver) Close() error {
	return nil
}
