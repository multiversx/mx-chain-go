package mock

import "time"

type ReconnecterWithPauseAndResumeStub struct {
	ReconnectToNetworkCalled func() <-chan struct{}
	PauseCall                func()
	ResumeCall               func()
}

func (rs *ReconnecterWithPauseAndResumeStub) ReconnectToNetwork() <-chan struct{} {
	return rs.ReconnectToNetworkCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (rs *ReconnecterWithPauseAndResumeStub) IsInterfaceNil() bool {
	if rs == nil {
		return true
	}
	return false
}

func (rs *ReconnecterWithPauseAndResumeStub) Pause()  { rs.PauseCall() }
func (rs *ReconnecterWithPauseAndResumeStub) Resume() { rs.ResumeCall() }

func (rs *ReconnecterWithPauseAndResumeStub) StartWatchdog(time.Duration) error { return nil }
func (rs *ReconnecterWithPauseAndResumeStub) StopWatchdog() error               { return nil }
func (rs *ReconnecterWithPauseAndResumeStub) KickWatchdog() error               { return nil }
