package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type TrieNodesResolverStub struct {
	RequestDataFromHashCalled    func(hash []byte) error
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
}

func (tnrs *TrieNodesResolverStub) RequestDataFromHash(hash []byte) error {
	if tnrs.RequestDataFromHashCalled != nil {
		return tnrs.RequestDataFromHashCalled(hash)
	}

	return errNotImplemented
}

func (tnrs *TrieNodesResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	if tnrs.ProcessReceivedMessageCalled != nil {
		return tnrs.ProcessReceivedMessageCalled(message)
	}

	return errNotImplemented
}

// IsInterfaceNil returns true if there is no value under the interface
func (tnrs *TrieNodesResolverStub) IsInterfaceNil() bool {
	if tnrs == nil {
		return true
	}
	return false
}
