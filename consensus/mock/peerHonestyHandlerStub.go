package mock

// PeerHonestyHandlerStub -
type PeerHonestyHandlerStub struct {
	ChangeScoreCalled func(pk string, topic string, units int)
	DecreaseCalled    func(pk string, topic string, value float64)
}

// ChangeScore -
func (phhs *PeerHonestyHandlerStub) ChangeScore(pk string, topic string, units int) {
	if phhs.ChangeScoreCalled != nil {
		phhs.ChangeScoreCalled(pk, topic, units)
		return
	}
}

// IsInterfaceNil -
func (phhs *PeerHonestyHandlerStub) IsInterfaceNil() bool {
	return phhs == nil
}
