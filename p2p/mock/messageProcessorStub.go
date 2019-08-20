package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type MessageProcessorStub struct {
	ProcessMessageCalled func(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error
}

func (mps *MessageProcessorStub) ProcessReceivedMessage(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error {
	return mps.ProcessMessageCalled(message, broadcastHandler)
}
