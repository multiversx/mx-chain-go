package mock

type ReconnecterStub struct {
	ReconnectToNetworkCalled func() <-chan struct{}
}

func (rs *ReconnecterStub) ReconnectToNetwork() <-chan struct{} {
	return rs.ReconnectToNetworkCalled()
}
