package mock

// NetworkConnectionWatcherStub -
type NetworkConnectionWatcherStub struct {
	IsConnectedToTheNetworkCalled func() bool
}

// IsConnectedToTheNetwork -
func (cws *NetworkConnectionWatcherStub) IsConnectedToTheNetwork() bool {
	return cws.IsConnectedToTheNetworkCalled()
}

// IsInterfaceNil -
func (cws *NetworkConnectionWatcherStub) IsInterfaceNil() bool {
	return cws == nil
}
