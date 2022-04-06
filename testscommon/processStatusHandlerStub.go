package testscommon

// ProcessStatusHandlerStub -
type ProcessStatusHandlerStub struct {
	SetToBusyCalled func(reason string)
	SetToIdleCalled func()
	IsIdleCalled    func() bool
}

// SetToBusy -
func (stub *ProcessStatusHandlerStub) SetToBusy(reason string) {
	if stub.SetToBusyCalled != nil {
		stub.SetToBusyCalled(reason)
	}
}

// SetToIdle -
func (stub *ProcessStatusHandlerStub) SetToIdle() {
	if stub.SetToIdleCalled != nil {
		stub.SetToIdleCalled()
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
