package mock

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// ResolverStub -
type ResolverStub struct {
	RequestDataFromHashCalled    func(hash []byte, epoch uint32) error
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
	SetNumPeersToQueryCalled     func(intra int, cross int)
	GetNumPeersToQueryCalled     func() (int, int)
}

// SetNumPeersToQuery -
func (rs *ResolverStub) SetNumPeersToQuery(intra int, cross int) {
	if rs.SetNumPeersToQueryCalled != nil {
		rs.SetNumPeersToQueryCalled(intra, cross)
	}
}

// GetNumPeersToQuery -
func (rs *ResolverStub) GetNumPeersToQuery() (int, int) {
	if rs.GetNumPeersToQueryCalled != nil {
		return rs.GetNumPeersToQueryCalled()
	}

	return 2, 2
}

// RequestDataFromHash -
func (rs *ResolverStub) RequestDataFromHash(hash []byte, epoch uint32) error {
	return rs.RequestDataFromHashCalled(hash, epoch)
}

// ProcessReceivedMessage -
func (rs *ResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, _ p2p.PeerID) error {
	return rs.ProcessReceivedMessageCalled(message)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rs *ResolverStub) IsInterfaceNil() bool {
	return rs == nil
}
