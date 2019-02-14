package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type InterceptorStub struct {
	NameCalled                      func() string
	SetReceivedMessageHandlerCalled func(handler p2p.TopicValidatorHandler)
	ReceivedMessageHandlerCalled    func() p2p.TopicValidatorHandler
	MarshalizerCalled               func() marshal.Marshalizer
}

func (is *InterceptorStub) Name() string {
	return is.NameCalled()
}

func (is *InterceptorStub) SetReceivedMessageHandler(handler p2p.TopicValidatorHandler) {
	is.SetReceivedMessageHandlerCalled(handler)
}

func (is *InterceptorStub) ReceivedMessageHandler() p2p.TopicValidatorHandler {
	return is.ReceivedMessageHandlerCalled()
}

func (is *InterceptorStub) Marshalizer() marshal.Marshalizer {
	return is.MarshalizerCalled()
}
