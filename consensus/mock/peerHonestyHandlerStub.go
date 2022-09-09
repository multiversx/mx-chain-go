package mock

// PeerHonestyHandlerStub -
type PeerHonestyHandlerStub struct {
	ChangeScoreCalled func(pk string, topic string, units int)
}

// ChangeScore -
func (phhs *PeerHonestyHandlerStub) ChangeScore(pk string, topic string, units int) {
	if phhs.ChangeScoreCalled != nil {
		phhs.ChangeScoreCalled(pk, topic, units)
	}
}

// Close -
func (phhs *PeerHonestyHandlerStub) Close() error {
	return nil
}

// IsInterfaceNil -
func (phhs *PeerHonestyHandlerStub) IsInterfaceNil() bool {
	return phhs == nil
}
