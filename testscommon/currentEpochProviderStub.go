package testscommon

// CurrentEpochProviderStub -
type CurrentEpochProviderStub struct {
	EpochIsActiveInNetworkCalled func(epoch uint32) bool
}

// EpochIsActiveInNetwork -
func (c *CurrentEpochProviderStub) EpochIsActiveInNetwork(epoch uint32) bool {
	if c.EpochIsActiveInNetworkCalled != nil {
		return c.EpochIsActiveInNetworkCalled(epoch)
	}

	return true
}

// IsInterfaceNil -
func (c *CurrentEpochProviderStub) IsInterfaceNil() bool {
	return c == nil
}
