package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type InterceptorStub struct {
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) ([]byte, error)
}

func (is *InterceptorStub) ProcessReceivedMessage(message p2p.MessageP2P) ([]byte, error) {
	return is.ProcessReceivedMessageCalled(message)
}
