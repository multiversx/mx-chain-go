package mock

// PeerHonestyHandlerStub -
type PeerHonestyHandlerStub struct {
	IncreaseCalled func(round int64, pk string, topic string, value float64)
	DecreaseCalled func(round int64, pk string, topic string, value float64)
}

// Increase -
func (phhs *PeerHonestyHandlerStub) Increase(round int64, pk string, topic string, value float64) {
	if phhs.IncreaseCalled != nil {
		phhs.IncreaseCalled(round, pk, topic, value)
		return
	}
}

// Decrease -
func (phhs *PeerHonestyHandlerStub) Decrease(round int64, pk string, topic string, value float64) {
	if phhs.DecreaseCalled != nil {
		phhs.DecreaseCalled(round, pk, topic, value)
		return
	}
}

// IsInterfaceNil -
func (phhs *PeerHonestyHandlerStub) IsInterfaceNil() bool {
	return phhs == nil
}
