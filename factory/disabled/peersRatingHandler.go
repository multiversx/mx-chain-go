package disabled

import "github.com/multiversx/mx-chain-core-go/core"

type peersRatingHandler struct {
}

// NewPeersRatingHandler returns a new disabled PeersRatingHandler implementation
func NewPeersRatingHandler() *peersRatingHandler {
	return &peersRatingHandler{}
}

// IncreaseRating does nothing as it is disabled
func (handler *peersRatingHandler) IncreaseRating(_ core.PeerID) {
}

// DecreaseRating does nothing as it is disabled
func (handler *peersRatingHandler) DecreaseRating(_ core.PeerID) {
}

// GetTopRatedPeersFromList returns the provided peers list as it is disabled
func (handler *peersRatingHandler) GetTopRatedPeersFromList(peers []core.PeerID, _ int) []core.PeerID {
	return peers
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *peersRatingHandler) IsInterfaceNil() bool {
	return handler == nil
}
