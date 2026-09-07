package mock

// CurrentNetworkEpochProviderStub -
type CurrentNetworkEpochProviderStub struct {
	EpochIsActiveInNetworkCalled      func(epoch uint32) bool
	EpochIsAvailableOnMainPeersCalled func(epoch uint32) bool
	EpochConfirmedCalled              func(newEpoch uint32, newTimestamp uint64)
}

// EpochIsActiveInNetwork -
func (cneps *CurrentNetworkEpochProviderStub) EpochIsActiveInNetwork(epoch uint32) bool {
	if cneps.EpochIsActiveInNetworkCalled != nil {
		return cneps.EpochIsActiveInNetworkCalled(epoch)
	}

	return true
}

// EpochIsAvailableOnMainPeers -
func (cneps *CurrentNetworkEpochProviderStub) EpochIsAvailableOnMainPeers(epoch uint32) bool {
	if cneps.EpochIsAvailableOnMainPeersCalled != nil {
		return cneps.EpochIsAvailableOnMainPeersCalled(epoch)
	}

	// default mirrors EpochIsActiveInNetwork to keep pre-existing tests' band semantics
	return cneps.EpochIsActiveInNetwork(epoch)
}

// EpochConfirmed -
func (cneps *CurrentNetworkEpochProviderStub) EpochConfirmed(newEpoch uint32, newTimestamp uint64) {
	if cneps.EpochConfirmedCalled != nil {
		cneps.EpochConfirmedCalled(newEpoch, newTimestamp)
	}
}

// IsInterfaceNil -
func (cneps *CurrentNetworkEpochProviderStub) IsInterfaceNil() bool {
	return cneps == nil
}
