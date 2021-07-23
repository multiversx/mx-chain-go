package mock

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MiniBlocksResolverStub -
type MiniBlocksResolverStub struct {
	RequestDataFromHashCalled      func(hash []byte, epoch uint32) error
	RequestDataFromHashArrayCalled func(hashes [][]byte, epoch uint32) error
	ProcessReceivedMessageCalled   func(message p2p.MessageP2P) error
	SetNumPeersToQueryCalled       func(intra int, cross int)
	NumPeersToQueryCalled          func() (int, int)
	SetResolverDebugHandlerCalled  func(handler dataRetriever.ResolverDebugHandler) error
}

// SetNumPeersToQuery -
func (mbrs *MiniBlocksResolverStub) SetNumPeersToQuery(intra int, cross int) {
	if mbrs.SetNumPeersToQueryCalled != nil {
		mbrs.SetNumPeersToQueryCalled(intra, cross)
	}
}

// NumPeersToQuery -
func (mbrs *MiniBlocksResolverStub) NumPeersToQuery() (int, int) {
	if mbrs.NumPeersToQueryCalled != nil {
		return mbrs.NumPeersToQueryCalled()
	}

	return 2, 2
}

// RequestDataFromHash -
func (mbrs *MiniBlocksResolverStub) RequestDataFromHash(hash []byte, epoch uint32) error {
	return mbrs.RequestDataFromHashCalled(hash, epoch)
}

// RequestDataFromHashArray -
func (mbrs *MiniBlocksResolverStub) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	return mbrs.RequestDataFromHashArrayCalled(hashes, epoch)
}

// ProcessReceivedMessage -
func (mbrs *MiniBlocksResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID) error {
	return mbrs.ProcessReceivedMessageCalled(message)
}

// SetResolverDebugHandler -
func (mbrs *MiniBlocksResolverStub) SetResolverDebugHandler(handler dataRetriever.ResolverDebugHandler) error {
	if mbrs.SetResolverDebugHandlerCalled != nil {
		return mbrs.SetResolverDebugHandlerCalled(handler)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mbrs *MiniBlocksResolverStub) IsInterfaceNil() bool {
	return mbrs == nil
}
