package mock

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MiniBlocksResolverMock -
type MiniBlocksResolverMock struct {
	RequestDataFromHashCalled      func(hash []byte, epoch uint32) error
	RequestDataFromHashArrayCalled func(hashes [][]byte, epoch uint32) error
	ProcessReceivedMessageCalled   func(message p2p.MessageP2P) error
	SetNumPeersToQueryCalled       func(intra int, cross int)
	NumPeersToQueryCalled          func() (int, int)
}

// SetNumPeersToQuery -
func (hrm *MiniBlocksResolverMock) SetNumPeersToQuery(intra int, cross int) {
	if hrm.SetNumPeersToQueryCalled != nil {
		hrm.SetNumPeersToQueryCalled(intra, cross)
	}
}

// NumPeersToQuery -
func (hrm *MiniBlocksResolverMock) NumPeersToQuery() (int, int) {
	if hrm.NumPeersToQueryCalled != nil {
		return hrm.NumPeersToQueryCalled()
	}

	return 2, 2
}

// RequestDataFromHash -
func (hrm *MiniBlocksResolverMock) RequestDataFromHash(hash []byte, epoch uint32) error {
	return hrm.RequestDataFromHashCalled(hash, epoch)
}

// RequestDataFromHashArray -
func (hrm *MiniBlocksResolverMock) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	return hrm.RequestDataFromHashArrayCalled(hashes, epoch)
}

// ProcessReceivedMessage -
func (hrm *MiniBlocksResolverMock) ProcessReceivedMessage(message p2p.MessageP2P, _ core.PeerID) error {
	return hrm.ProcessReceivedMessageCalled(message)
}

// SetResolverDebugHandler -
func (hrm *MiniBlocksResolverMock) SetResolverDebugHandler(_ dataRetriever.ResolverDebugHandler) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hrm *MiniBlocksResolverMock) IsInterfaceNil() bool {
	return hrm == nil
}
