package mock

import "context"

// ReconnecterStub -
type ReconnecterStub struct {
	ReconnectToNetworkCalled func(ctx context.Context)
	PauseCall                func()
	ResumeCall               func()
}

// ReconnectToNetwork -
func (rs *ReconnecterStub) ReconnectToNetwork(ctx context.Context) {
	if rs.ReconnectToNetworkCalled != nil {
		rs.ReconnectToNetworkCalled(ctx)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (rs *ReconnecterStub) IsInterfaceNil() bool {
	return rs == nil
}
