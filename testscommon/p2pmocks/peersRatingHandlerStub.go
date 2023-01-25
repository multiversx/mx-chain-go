package p2pmocks

import "github.com/multiversx/mx-chain-core-go/core"

// PeersRatingHandlerStub -
type PeersRatingHandlerStub struct {
	AddPeersCalled                 func(pids []core.PeerID)
	IncreaseRatingCalled           func(pid core.PeerID)
	DecreaseRatingCalled           func(pid core.PeerID)
	GetTopRatedPeersFromListCalled func(peers []core.PeerID, numOfPeers int) []core.PeerID
	CloseCalled                    func() error
}

// AddPeers -
func (stub *PeersRatingHandlerStub) AddPeers(pids []core.PeerID) {
	if stub.AddPeersCalled != nil {
		stub.AddPeersCalled(pids)
	}
}

// IncreaseRating -
func (stub *PeersRatingHandlerStub) IncreaseRating(pid core.PeerID) {
	if stub.IncreaseRatingCalled != nil {
		stub.IncreaseRatingCalled(pid)
	}
}

// DecreaseRating -
func (stub *PeersRatingHandlerStub) DecreaseRating(pid core.PeerID) {
	if stub.DecreaseRatingCalled != nil {
		stub.DecreaseRatingCalled(pid)
	}
}

// GetTopRatedPeersFromList -
func (stub *PeersRatingHandlerStub) GetTopRatedPeersFromList(peers []core.PeerID, numOfPeers int) []core.PeerID {
	if stub.GetTopRatedPeersFromListCalled != nil {
		return stub.GetTopRatedPeersFromListCalled(peers, numOfPeers)
	}

	return peers
}

// Close -
func (stub *PeersRatingHandlerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *PeersRatingHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
