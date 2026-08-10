package testscommon

// ProcessStatusHandlerStub -
type ProcessStatusHandlerStub struct {
	SetBusyCalled                      func(reason string)
	TrySetBusyCalled                   func(reason string) bool
	SetIdleCalled                      func()
	BlockBackgroundJobsCalled          func(reason string)
	UnblockBackgroundJobsCalled        func()
	SuspendBackgroundJobBlockingCalled func(reason string)
	ResumeBackgroundJobBlockingCalled  func()
	IsIdleCalled                       func() bool
}

// SetBusy -
func (stub *ProcessStatusHandlerStub) SetBusy(reason string) {
	if stub.SetBusyCalled != nil {
		stub.SetBusyCalled(reason)
	}
}

// TrySetBusy -
func (stub *ProcessStatusHandlerStub) TrySetBusy(reason string) bool {
	if stub.TrySetBusyCalled != nil {
		return stub.TrySetBusyCalled(reason)
	}

	return true
}

// SetIdle -
func (stub *ProcessStatusHandlerStub) SetIdle() {
	if stub.SetIdleCalled != nil {
		stub.SetIdleCalled()
	}
}

// BlockBackgroundJobs -
func (stub *ProcessStatusHandlerStub) BlockBackgroundJobs(reason string) {
	if stub.BlockBackgroundJobsCalled != nil {
		stub.BlockBackgroundJobsCalled(reason)
	}
}

// UnblockBackgroundJobs -
func (stub *ProcessStatusHandlerStub) UnblockBackgroundJobs() {
	if stub.UnblockBackgroundJobsCalled != nil {
		stub.UnblockBackgroundJobsCalled()
	}
}

// SuspendBackgroundJobBlocking -
func (stub *ProcessStatusHandlerStub) SuspendBackgroundJobBlocking(reason string) {
	if stub.SuspendBackgroundJobBlockingCalled != nil {
		stub.SuspendBackgroundJobBlockingCalled(reason)
	}
}

// ResumeBackgroundJobBlocking -
func (stub *ProcessStatusHandlerStub) ResumeBackgroundJobBlocking() {
	if stub.ResumeBackgroundJobBlockingCalled != nil {
		stub.ResumeBackgroundJobBlockingCalled()
	}
}

// IsIdle -
func (stub *ProcessStatusHandlerStub) IsIdle() bool {
	if stub.IsIdleCalled != nil {
		return stub.IsIdleCalled()
	}

	return true
}

// IsInterfaceNil -
func (stub *ProcessStatusHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
