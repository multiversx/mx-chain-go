package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// HashSliceResolverStub -
type HashSliceResolverStub struct {
	RequestDataFromHashCalled      func(hash []byte, epoch uint32) error
	ProcessReceivedMessageCalled   func(message p2p.MessageP2P) error
	RequestDataFromHashArrayCalled func(hashes [][]byte, epoch uint32) error
	SetNumPeersToQueryCalled       func(intra int, cross int)
	NumPeersToQueryCalled          func() (int, int)
	SetResolverDebugHandlerCalled  func(handler dataRetriever.ResolverDebugHandler) error
	CloseCalled                    func() error
}

// SetNumPeersToQuery -
func (hsrs *HashSliceResolverStub) SetNumPeersToQuery(intra int, cross int) {
	if hsrs.SetNumPeersToQueryCalled != nil {
		hsrs.SetNumPeersToQueryCalled(intra, cross)
	}
}

// NumPeersToQuery -
func (hsrs *HashSliceResolverStub) NumPeersToQuery() (int, int) {
	if hsrs.NumPeersToQueryCalled != nil {
		return hsrs.NumPeersToQueryCalled()
	}

	return 2, 2
}

// RequestDataFromHash -
func (hsrs *HashSliceResolverStub) RequestDataFromHash(hash []byte, epoch uint32) error {
	if hsrs.RequestDataFromHashCalled != nil {
		return hsrs.RequestDataFromHashCalled(hash, epoch)
	}

	return errNotImplemented
}

// ProcessReceivedMessage -
func (hsrs *HashSliceResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID) error {
	if hsrs.ProcessReceivedMessageCalled != nil {
		return hsrs.ProcessReceivedMessageCalled(message)
	}

	return errNotImplemented
}

// RequestDataFromHashArray -
func (hsrs *HashSliceResolverStub) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	if hsrs.RequestDataFromHashArrayCalled != nil {
		return hsrs.RequestDataFromHashArrayCalled(hashes, epoch)
	}

	return errNotImplemented
}

// SetResolverDebugHandler -
func (hsrs *HashSliceResolverStub) SetResolverDebugHandler(handler dataRetriever.ResolverDebugHandler) error {
	if hsrs.SetResolverDebugHandlerCalled != nil {
		return hsrs.SetResolverDebugHandlerCalled(handler)
	}

	return nil
}

// Close -
func (hsrs *HashSliceResolverStub) Close() error {
	if hsrs.CloseCalled != nil {
		return hsrs.CloseCalled()
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hsrs *HashSliceResolverStub) IsInterfaceNil() bool {
	return hsrs == nil
}
