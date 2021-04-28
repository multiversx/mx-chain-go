package mock

// CurrentNetworkEpochProviderStub -
type CurrentNetworkEpochProviderStub struct {
	EpochIsActiveInNetworkCalled func(epoch uint32) bool
	EpochConfirmedCalled         func(newEpoch uint32, newTimestamp uint64)
}

// EpochIsActiveInNetwork -
func (cneps *CurrentNetworkEpochProviderStub) EpochIsActiveInNetwork(epoch uint32) bool {
	if cneps.EpochIsActiveInNetworkCalled != nil {
		return cneps.EpochIsActiveInNetworkCalled(epoch)
	}

	return true
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
