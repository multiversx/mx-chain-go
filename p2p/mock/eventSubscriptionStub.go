package mock

// EventSubscriptionStub -
type EventSubscriptionStub struct {
	CloseCalled func() error
	OutCalled   func() <-chan interface{}
}

// Close -
func (ess *EventSubscriptionStub) Close() error {
	if ess.CloseCalled != nil {
		return ess.CloseCalled()
	}

	return nil
}

// Out -
func (ess *EventSubscriptionStub) Out() <-chan interface{} {
	if ess.OutCalled != nil {
		return ess.OutCalled()
	}

	return make(chan interface{})
}
