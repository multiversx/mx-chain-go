package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type ResolverStub struct {
	RequestDataFromHashCalled    func(hash []byte) error
	ProcessReceivedMessageCalled func(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error
}

func (rs *ResolverStub) RequestDataFromHash(hash []byte) error {
	return rs.RequestDataFromHashCalled(hash)
}

func (rs *ResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error {
	return rs.ProcessReceivedMessageCalled(message, broadcastHandler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rs *ResolverStub) IsInterfaceNil() bool {
	if rs == nil {
		return true
	}
	return false
}
