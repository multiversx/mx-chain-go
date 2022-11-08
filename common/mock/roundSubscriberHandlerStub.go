package mock

// RoundSubscriberHandlerStub -
type RoundSubscriberHandlerStub struct {
	RoundConfirmedCalled func(round uint64, timestamp uint64)
}

// RoundConfirmed -
func (rsh *RoundSubscriberHandlerStub) RoundConfirmed(round uint64, timestamp uint64) {
	if rsh.RoundConfirmedCalled != nil {
		rsh.RoundConfirmedCalled(round, timestamp)
	}
}

// IsInterfaceNil -
func (rsh *RoundSubscriberHandlerStub) IsInterfaceNil() bool {
	return rsh == nil
}
