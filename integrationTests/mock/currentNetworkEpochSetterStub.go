package mock

// CurrentNetworkEpochSetterStub -
type CurrentNetworkEpochSetterStub struct {
	SetCurrentEpochCalled func(epoch uint32)
}

// SetNetworkEpochAtBootstrap -
func (c *CurrentNetworkEpochSetterStub) SetNetworkEpochAtBootstrap(epoch uint32) {
	if c.SetCurrentEpochCalled != nil {
		c.SetCurrentEpochCalled(epoch)
	}
}

// IsInterfaceNil -
func (c *CurrentNetworkEpochSetterStub) IsInterfaceNil() bool {
	return c == nil
}
