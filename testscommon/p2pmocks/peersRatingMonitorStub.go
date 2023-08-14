package p2pmocks

import "github.com/multiversx/mx-chain-go/p2p"

// PeersRatingMonitorStub -
type PeersRatingMonitorStub struct {
	GetConnectedPeersRatingsCalled func(connectionsHandler p2p.ConnectionsHandler) (string, error)
}

// GetConnectedPeersRatings -
func (stub *PeersRatingMonitorStub) GetConnectedPeersRatings(connectionsHandler p2p.ConnectionsHandler) (string, error) {
	if stub.GetConnectedPeersRatingsCalled != nil {
		return stub.GetConnectedPeersRatingsCalled(connectionsHandler)
	}
	return "", nil
}

// IsInterfaceNil -
func (stub *PeersRatingMonitorStub) IsInterfaceNil() bool {
	return stub == nil
}
