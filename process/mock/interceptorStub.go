package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type InterceptorStub struct {
	NameCalled                      func() string
	SetReceivedMessageHandlerCalled func(func(message p2p.MessageP2P) error)
	ReceivedMessageHandlerCalled    func() func(message p2p.MessageP2P) error
	MarshalizerCalled               func() marshal.Marshalizer
}

func (is *InterceptorStub) Name() string {
	return is.NameCalled()
}

func (is *InterceptorStub) SetReceivedMessageHandler(handler func(message p2p.MessageP2P) error) {
	is.SetReceivedMessageHandlerCalled(handler)
}

func (is *InterceptorStub) ReceivedMessageHandler() func(message p2p.MessageP2P) error {
	return is.ReceivedMessageHandlerCalled()
}

func (is *InterceptorStub) Marshalizer() marshal.Marshalizer {
	return is.MarshalizerCalled()
}
