package mock

import (
	heartbeatData "github.com/multiversx/mx-chain-go/heartbeat/data"
)

// HeartbeatMonitorStub -
type HeartbeatMonitorStub struct {
	GetHeartbeatsCalled func() []heartbeatData.PubKeyHeartbeat
}

// GetHeartbeats -
func (hbms *HeartbeatMonitorStub) GetHeartbeats() []heartbeatData.PubKeyHeartbeat {
	if hbms.GetHeartbeatsCalled != nil {
		return hbms.GetHeartbeatsCalled()
	}
	return nil
}

// Close -
func (hbms *HeartbeatMonitorStub) Close() error {
	return nil
}

// IsInterfaceNil -
func (hbms *HeartbeatMonitorStub) IsInterfaceNil() bool {
	return hbms == nil
}
