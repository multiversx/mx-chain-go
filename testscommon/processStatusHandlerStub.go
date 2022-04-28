package testscommon

// ProcessStatusHandlerStub -
type ProcessStatusHandlerStub struct {
	SetBusyCalled func(reason string)
	SetIdleCalled func()
	IsIdleCalled  func() bool
}

// SetBusy -
func (stub *ProcessStatusHandlerStub) SetBusy(reason string) {
	if stub.SetBusyCalled != nil {
		stub.SetBusyCalled(reason)
	}
}

// SetIdle -
func (stub *ProcessStatusHandlerStub) SetIdle() {
	if stub.SetIdleCalled != nil {
		stub.SetIdleCalled()
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
