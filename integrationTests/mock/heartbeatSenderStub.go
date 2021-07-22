package mock

// HeartbeatSenderStub -
type HeartbeatSenderStub struct {
	SendHeartbeatCalled func() error
}

// SendHeartbeat -
func (hbss *HeartbeatSenderStub) SendHeartbeat() error {
	if hbss.SendHeartbeatCalled != nil {
		return hbss.SendHeartbeatCalled()
	}

	return nil
}

// IsInterfaceNil -
func (hbss *HeartbeatSenderStub) IsInterfaceNil() bool {
	return hbss == nil
}
