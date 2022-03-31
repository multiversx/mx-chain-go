package mock

type HardforkHandlerStub struct {
	ShouldTriggerHardforkCalled func() <-chan struct{}
	ExecuteCalled               func()
}

// ShouldTriggerHardfork -
func (stub *HardforkHandlerStub) ShouldTriggerHardfork() <-chan struct{} {
	if stub.ShouldTriggerHardforkCalled != nil {
		return stub.ShouldTriggerHardforkCalled()
	}

	return nil
}

// Execute -
func (stub *HardforkHandlerStub) Execute() {
	if stub.ExecuteCalled != nil {
		stub.ExecuteCalled()
	}
}
