package mock

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

var errNotImplemented = errors.New("not implemented")

// TrieNodesResolverStub -
type TrieNodesResolverStub struct {
	RequestDataFromHashCalled    func(hash []byte) error
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
}

// RequestDataFromHash -
func (tnrs *TrieNodesResolverStub) RequestDataFromHash(hash []byte, _ uint32) error {
	if tnrs.RequestDataFromHashCalled != nil {
		return tnrs.RequestDataFromHashCalled(hash)
	}

	return errNotImplemented
}

// ProcessReceivedMessage -
func (tnrs *TrieNodesResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID) error {
	if tnrs.ProcessReceivedMessageCalled != nil {
		return tnrs.ProcessReceivedMessageCalled(message)
	}

	return errNotImplemented
}

// IsInterfaceNil returns true if there is no value under the interface
func (tnrs *TrieNodesResolverStub) IsInterfaceNil() bool {
	return tnrs == nil
}
