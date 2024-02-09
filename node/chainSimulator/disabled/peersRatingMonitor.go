package disabled

import "github.com/multiversx/mx-chain-go/p2p"

type peersRatingMonitor struct {
}

// NewPeersRatingMonitor will create a new disabled peersRatingMonitor instance
func NewPeersRatingMonitor() *peersRatingMonitor {
	return &peersRatingMonitor{}
}

// GetConnectedPeersRatings returns an empty string since it is a disabled component
func (monitor *peersRatingMonitor) GetConnectedPeersRatings(_ p2p.ConnectionsHandler) (string, error) {
	return "", nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (monitor *peersRatingMonitor) IsInterfaceNil() bool {
	return monitor == nil
}
