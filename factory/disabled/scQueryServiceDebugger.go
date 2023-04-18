package disabled

type scQueryServiceDebugger struct {
}

// NewDisabledSCQueryServiceDebugger returns a new instance of disabled scQueryServiceDebugger
func NewDisabledSCQueryServiceDebugger() *scQueryServiceDebugger {
	return &scQueryServiceDebugger{}
}

// NotifyExecutionStarted does nothing
func (debugger *scQueryServiceDebugger) NotifyExecutionStarted(_ int) {
}

// NotifyExecutionFinished does nothing
func (debugger *scQueryServiceDebugger) NotifyExecutionFinished(_ int) {
}

// Close does nothing and returns nil
func (debugger *scQueryServiceDebugger) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (debugger *scQueryServiceDebugger) IsInterfaceNil() bool {
	return debugger == nil
}
