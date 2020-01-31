package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type ResolverStub struct {
	RequestDataFromHashCalled    func(hash []byte, epoch uint32) error
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
}

func (rs *ResolverStub) RequestDataFromHash(hash []byte, epoch uint32) error {
	return rs.RequestDataFromHashCalled(hash, epoch)
}

func (rs *ResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, _ p2p.PeerID) error {
	return rs.ProcessReceivedMessageCalled(message)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rs *ResolverStub) IsInterfaceNil() bool {
	return rs == nil
}
