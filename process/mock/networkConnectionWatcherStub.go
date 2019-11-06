package mock

type NetworkConnectionWatcherStub struct {
	IsConnectedToTheNetworkCalled func() bool
}

func (cws *NetworkConnectionWatcherStub) IsConnectedToTheNetwork() bool {
	return cws.IsConnectedToTheNetworkCalled()
}

func (cws *NetworkConnectionWatcherStub) IsInterfaceNil() bool {
	return cws == nil
}
