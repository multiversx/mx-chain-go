package mock

// HardforkHandlerStub -
type HardforkHandlerStub struct {
	ShouldTriggerHardforkCalled func() <-chan struct{}
	ExecuteCalled               func()
	CloseCalled                 func()
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

// Close -
func (stub *HardforkHandlerStub) Close() {
	if stub.CloseCalled != nil {
		stub.CloseCalled()
	}
}
