package mock

// CurrentNetworkEpochProviderStub -
type CurrentNetworkEpochProviderStub struct {
	SetNetworkEpochAtBootstrapCalled func(epoch uint32)
	EpochIsActiveInNetworkCalled     func(epoch uint32) bool
	CurrentEpochCalled               func() uint32
}

// SetNetworkEpochAtBootstrap -
func (cneps *CurrentNetworkEpochProviderStub) SetNetworkEpochAtBootstrap(epoch uint32) {
	if cneps.SetNetworkEpochAtBootstrapCalled != nil {
		cneps.SetNetworkEpochAtBootstrapCalled(epoch)
	}
}

// EpochIsActiveInNetwork -
func (cneps *CurrentNetworkEpochProviderStub) EpochIsActiveInNetwork(epoch uint32) bool {
	if cneps.EpochIsActiveInNetworkCalled != nil {
		return cneps.EpochIsActiveInNetworkCalled(epoch)
	}

	return true
}

// CurrentEpoch returns the current network epoch
func (cneps *CurrentNetworkEpochProviderStub) CurrentEpoch() uint32 {
	if cneps.CurrentEpochCalled != nil {
		return cneps.CurrentEpochCalled()
	}

	return 0
}

// IsInterfaceNil -
func (cneps *CurrentNetworkEpochProviderStub) IsInterfaceNil() bool {
	return cneps == nil
}
