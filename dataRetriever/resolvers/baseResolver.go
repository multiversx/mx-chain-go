package resolvers

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
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
		return dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(arg.AntifloodHandler) {
		return dataRetriever.ErrNilAntifloodHandler
	}
	if check.IfNil(arg.Throttler) {
		return dataRetriever.ErrNilThrottler
	}
	return nil
}

// SetDebugHandler will set a debug handler
func (res *baseResolver) SetDebugHandler(handler dataRetriever.DebugHandler) error {
	return res.TopicResolverSender.SetDebugHandler(handler)
}

// Close returns nil
func (res *baseResolver) Close() error {
	return nil
}
