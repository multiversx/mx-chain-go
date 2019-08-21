package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type InterceptorStub struct {
	ProcessReceivedMessageCalled func(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error
}

func (is *InterceptorStub) ProcessReceivedMessage(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error {
	return is.ProcessReceivedMessageCalled(message, broadcastHandler)
}
