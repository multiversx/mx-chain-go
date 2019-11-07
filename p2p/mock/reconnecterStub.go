package mock

type ReconnecterStub struct {
	ReconnectToNetworkCalled func() <-chan struct{}
	PauseCall                func()
	ResumeCall               func()
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

func (rs *ReconnecterStub) Pause()  { rs.PauseCall() }
func (rs *ReconnecterStub) Resume() { rs.ResumeCall() }
