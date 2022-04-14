package disabled

// ProcessStatusHandler is the disabled implementation for the status handler that keeps track what the node is doing:
// processing blocks or idle
type ProcessStatusHandler struct {
}

// SetBusy does nothing
func (psh *ProcessStatusHandler) SetBusy(_ string) {}

// SetIdle does nothing
func (psh *ProcessStatusHandler) SetIdle() {}

// IsIdle returns true
func (psh *ProcessStatusHandler) IsIdle() bool {
	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (psh *ProcessStatusHandler) IsInterfaceNil() bool {
	return psh == nil
}
