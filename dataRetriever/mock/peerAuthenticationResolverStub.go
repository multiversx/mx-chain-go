package mock

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/p2p"
)

// PeerAuthenticationResolverStub -
type PeerAuthenticationResolverStub struct {
	RequestDataFromHashCalled      func(hash []byte, epoch uint32) error
	ProcessReceivedMessageCalled   func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error
	SetResolverDebugHandlerCalled  func(handler dataRetriever.ResolverDebugHandler) error
	SetNumPeersToQueryCalled       func(intra int, cross int)
	NumPeersToQueryCalled          func() (int, int)
	CloseCalled                    func() error
	RequestDataFromHashArrayCalled func(hashes [][]byte, epoch uint32) error
}

// RequestDataFromHash -
func (pars *PeerAuthenticationResolverStub) RequestDataFromHash(hash []byte, epoch uint32) error {
	if pars.RequestDataFromHashCalled != nil {
		return pars.RequestDataFromHashCalled(hash, epoch)
	}

	return nil
}

// ProcessReceivedMessage -
func (pars *PeerAuthenticationResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	if pars.ProcessReceivedMessageCalled != nil {
		return pars.ProcessReceivedMessageCalled(message, fromConnectedPeer)
	}

	return nil
}

// SetResolverDebugHandler -
func (pars *PeerAuthenticationResolverStub) SetResolverDebugHandler(handler dataRetriever.ResolverDebugHandler) error {
	if pars.SetResolverDebugHandlerCalled != nil {
		return pars.SetResolverDebugHandlerCalled(handler)
	}

	return nil
}

// SetNumPeersToQuery -
func (pars *PeerAuthenticationResolverStub) SetNumPeersToQuery(intra int, cross int) {
	if pars.SetNumPeersToQueryCalled != nil {
		pars.SetNumPeersToQueryCalled(intra, cross)
	}
}

// NumPeersToQuery -
func (pars *PeerAuthenticationResolverStub) NumPeersToQuery() (int, int) {
	if pars.NumPeersToQueryCalled != nil {
		return pars.NumPeersToQueryCalled()
	}

	return 0, 0
}

func (pars *PeerAuthenticationResolverStub) Close() error {
	if pars.CloseCalled != nil {
		return pars.CloseCalled()
	}

	return nil
}

// RequestDataFromHashArray -
func (pars *PeerAuthenticationResolverStub) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	if pars.RequestDataFromHashArrayCalled != nil {
		return pars.RequestDataFromHashArrayCalled(hashes, epoch)
	}

	return nil
}

// IsInterfaceNil -
func (pars *PeerAuthenticationResolverStub) IsInterfaceNil() bool {
	return pars == nil
}
