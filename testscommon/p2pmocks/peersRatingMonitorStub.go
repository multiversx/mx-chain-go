package p2pmocks

// PeersRatingMonitorStub -
type PeersRatingMonitorStub struct {
	GetConnectedPeersRatingsCalled func() string
}

// GetConnectedPeersRatings -
func (stub *PeersRatingMonitorStub) GetConnectedPeersRatings() string {
	if stub.GetConnectedPeersRatingsCalled != nil {
		return stub.GetConnectedPeersRatingsCalled()
	}
	return ""
}

// IsInterfaceNil -
func (stub *PeersRatingMonitorStub) IsInterfaceNil() bool {
	return stub == nil
}
