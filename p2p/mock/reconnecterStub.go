package mock

// ReconnecterStub -
type ReconnecterStub struct {
	ReconnectToNetworkCalled func() <-chan struct{}
	PauseCall                func()
	ResumeCall               func()
}

// ReconnectToNetwork -
func (rs *ReconnecterStub) ReconnectToNetwork() <-chan struct{} {
	return rs.ReconnectToNetworkCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (rs *ReconnecterStub) IsInterfaceNil() bool {
	return rs == nil
}
