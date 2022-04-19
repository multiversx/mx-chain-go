package p2pmocks

import "github.com/ElrondNetwork/elrond-go-core/core"

// PeersRatingHandlerStub -
type PeersRatingHandlerStub struct {
	AddPeerCalled                  func(pid core.PeerID)
	IncreaseRatingCalled           func(pid core.PeerID)
	DecreaseRatingCalled           func(pid core.PeerID)
	GetTopRatedPeersFromListCalled func(peers []core.PeerID, numOfPeers int) []core.PeerID
}

// AddPeer -
func (prhs *PeersRatingHandlerStub) AddPeer(pid core.PeerID) {
	if prhs.AddPeerCalled != nil {
		prhs.AddPeerCalled(pid)
	}
}

// IncreaseRating -
func (prhs *PeersRatingHandlerStub) IncreaseRating(pid core.PeerID) {
	if prhs.IncreaseRatingCalled != nil {
		prhs.IncreaseRatingCalled(pid)
	}
}

// DecreaseRating -
func (prhs *PeersRatingHandlerStub) DecreaseRating(pid core.PeerID) {
	if prhs.DecreaseRatingCalled != nil {
		prhs.DecreaseRatingCalled(pid)
	}
}

// GetTopRatedPeersFromList -
func (prhs *PeersRatingHandlerStub) GetTopRatedPeersFromList(peers []core.PeerID, numOfPeers int) []core.PeerID {
	if prhs.GetTopRatedPeersFromListCalled != nil {
		return prhs.GetTopRatedPeersFromListCalled(peers, numOfPeers)
	}

	return peers
}

// IsInterfaceNil returns true if there is no value under the interface
func (prhs *PeersRatingHandlerStub) IsInterfaceNil() bool {
	return prhs == nil
}
