package mock

// ReconnecterStub -
type ReconnecterStub struct {
	ReconnectToNetworkCalled func()
	PauseCall                func()
	ResumeCall               func()
}

// ReconnectToNetwork -
func (rs *ReconnecterStub) ReconnectToNetwork() {
	rs.ReconnectToNetworkCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (rs *ReconnecterStub) IsInterfaceNil() bool {
	return rs == nil
}

// Pause -
func (rs *ReconnecterStub) Pause() { rs.PauseCall() }

// Resume -
func (rs *ReconnecterStub) Resume() { rs.ResumeCall() }
