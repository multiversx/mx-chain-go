package mock

// ReconnecterStub -
type ReconnecterStub struct {
	ReconnectToNetworkCalled func() <-chan struct{}
	PauseCall                func()
	ResumeCall               func()
}

// ReconnectToNetwork -
func (rs *ReconnecterStub) ReconnectToNetwork() <-chan struct{} {
	if rs.ReconnectToNetworkCalled != nil {
		return rs.ReconnectToNetworkCalled()
	}

	return make(chan struct{}, 1000)
}

// IsInterfaceNil returns true if there is no value under the interface
func (rs *ReconnecterStub) IsInterfaceNil() bool {
	return rs == nil
}
