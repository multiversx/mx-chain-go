package mock

// ConsensusRatingStub -
type ConsensusRatingStub struct {
	IncreaseCalled func(round int64, pk string, topic string, value float64)
	DecreaseCalled func(round int64, pk string, topic string, value float64)
}

// Increase -
func (crs *ConsensusRatingStub) Increase(round int64, pk string, topic string, value float64) {
	if crs.IncreaseCalled != nil {
		crs.IncreaseCalled(round, pk, topic, value)
		return
	}
}

// Decrease -
func (crs *ConsensusRatingStub) Decrease(round int64, pk string, topic string, value float64) {
	if crs.DecreaseCalled != nil {
		crs.DecreaseCalled(round, pk, topic, value)
		return
	}
}

// IsInterfaceNil -
func (crs *ConsensusRatingStub) IsInterfaceNil() bool {
	return crs == nil
}
