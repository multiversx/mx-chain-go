package mock

type ReconnecterStub struct {
	ReconnectToNetworkCalled func() <-chan struct{}
}

func (rs *ReconnecterStub) ReconnectToNetwork() <-chan struct{} {
	return rs.ReconnectToNetworkCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (rs *ReconnecterStub) IsInterfaceNil() bool {
	if rs == nil {
		return true
	}
	return false
}

func (rs *ReconnecterStub) Pause()  {}
func (rs *ReconnecterStub) Resume() {}
