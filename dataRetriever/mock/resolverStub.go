package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// ResolverStub -
type ResolverStub struct {
	RequestDataFromHashCalled     func(hash []byte, epoch uint32) error
	ProcessReceivedMessageCalled  func(message p2p.MessageP2P) error
	SetNumPeersToQueryCalled      func(intra int, cross int)
	NumPeersToQueryCalled         func() (int, int)
	SetResolverDebugHandlerCalled func(handler dataRetriever.ResolverDebugHandler) error
	CloseCalled                   func() error
}

// SetNumPeersToQuery -
func (rs *ResolverStub) SetNumPeersToQuery(intra int, cross int) {
	if rs.SetNumPeersToQueryCalled != nil {
		rs.SetNumPeersToQueryCalled(intra, cross)
	}
}

// NumPeersToQuery -
func (rs *ResolverStub) NumPeersToQuery() (int, int) {
	if rs.NumPeersToQueryCalled != nil {
		return rs.NumPeersToQueryCalled()
	}

	return 2, 2
}

// RequestDataFromHash -
func (rs *ResolverStub) RequestDataFromHash(hash []byte, epoch uint32) error {
	return rs.RequestDataFromHashCalled(hash, epoch)
}

// ProcessReceivedMessage -
func (rs *ResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID) error {
	return rs.ProcessReceivedMessageCalled(message)
}

// SetResolverDebugHandler -
func (rs *ResolverStub) SetResolverDebugHandler(handler dataRetriever.ResolverDebugHandler) error {
	if rs.SetResolverDebugHandlerCalled != nil {
		return rs.SetResolverDebugHandlerCalled(handler)
	}

	return nil
}

// Close -
func (rs *ResolverStub) Close() error {
	if rs.CloseCalled != nil {
		return rs.CloseCalled()
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rs *ResolverStub) IsInterfaceNil() bool {
	return rs == nil
}
