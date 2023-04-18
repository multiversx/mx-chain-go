package testscommon

// SCQueryServiceDebuggerStub -
type SCQueryServiceDebuggerStub struct {
	NotifyExecutionStartedCalled  func(index int)
	NotifyExecutionFinishedCalled func(index int)
	CloseCalled                   func() error
}

// NotifyExecutionStarted -
func (stub *SCQueryServiceDebuggerStub) NotifyExecutionStarted(index int) {
	if stub.NotifyExecutionStartedCalled != nil {
		stub.NotifyExecutionStartedCalled(index)
	}
}

// NotifyExecutionFinished -
func (stub *SCQueryServiceDebuggerStub) NotifyExecutionFinished(index int) {
	if stub.NotifyExecutionFinishedCalled != nil {
		stub.NotifyExecutionFinishedCalled(index)
	}
}

// Close -
func (stub *SCQueryServiceDebuggerStub) Close() error {
	if stub.CloseCalled != nil {
		return stub.CloseCalled()
	}

	return nil
}

// IsInterfaceNil -
func (stub *SCQueryServiceDebuggerStub) IsInterfaceNil() bool {
	return stub == nil
}
