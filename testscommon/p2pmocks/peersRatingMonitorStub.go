package p2pmocks

// PeersRatingMonitorStub -
type PeersRatingMonitorStub struct {
	GetPeersRatingsCalled func() string
}

// GetPeersRatings -
func (stub *PeersRatingMonitorStub) GetPeersRatings() string {
	if stub.GetPeersRatingsCalled != nil {
		return stub.GetPeersRatingsCalled()
	}
	return "{}"
}

// IsInterfaceNil -
func (stub *PeersRatingMonitorStub) IsInterfaceNil() bool {
	return stub == nil
}
