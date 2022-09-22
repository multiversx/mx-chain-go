package process

type disabledDebugger struct {
}

// NewDisabledDebugger creates a disabled process debugger instance
func NewDisabledDebugger() *disabledDebugger {
	return &disabledDebugger{}
}

// SetLastCommittedBlockRound does nothing
func (debugger *disabledDebugger) SetLastCommittedBlockRound(_ uint64) {
}

// Close does nothing and returns nil
func (debugger *disabledDebugger) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (debugger *disabledDebugger) IsInterfaceNil() bool {
	return debugger == nil
}
