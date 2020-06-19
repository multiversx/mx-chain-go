package mock

// PeerHonestyHandlerStub -
type PeerHonestyHandlerStub struct {
	IncreaseCalled func(pk string, topic string, value float64)
	DecreaseCalled func(pk string, topic string, value float64)
}

// Increase -
func (phhs *PeerHonestyHandlerStub) Increase(pk string, topic string, value float64) {
	if phhs.IncreaseCalled != nil {
		phhs.IncreaseCalled(pk, topic, value)
		return
	}
}

// Decrease -
func (phhs *PeerHonestyHandlerStub) Decrease(pk string, topic string, value float64) {
	if phhs.DecreaseCalled != nil {
		phhs.DecreaseCalled(pk, topic, value)
		return
	}
}

// IsInterfaceNil -
func (phhs *PeerHonestyHandlerStub) IsInterfaceNil() bool {
	return phhs == nil
}
