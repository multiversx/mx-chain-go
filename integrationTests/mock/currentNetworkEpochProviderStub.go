package mock

// CurrentNetworkEpochProviderStub -
type CurrentNetworkEpochProviderStub struct {
	SetNetworkEpochAtBootstrapCalled func(epoch uint32)
	EpochIsActiveInNetworkCalled     func(epoch uint32) bool
	EpochConfirmedCalled             func(epoch uint32, timestamp uint64)
}

// EpochIsActiveInNetwork -
func (cneps *CurrentNetworkEpochProviderStub) EpochIsActiveInNetwork(epoch uint32) bool {
	if cneps.EpochIsActiveInNetworkCalled != nil {
		return cneps.EpochIsActiveInNetworkCalled(epoch)
	}

	return true
}

// EpochConfirmed -
func (cneps *CurrentNetworkEpochProviderStub) EpochConfirmed(epoch uint32, timestamp uint64) {
	if cneps.EpochConfirmedCalled != nil {
		cneps.EpochConfirmedCalled(epoch, timestamp)
	}
}

// IsInterfaceNil -
func (cneps *CurrentNetworkEpochProviderStub) IsInterfaceNil() bool {
	return cneps == nil
}
