package disabled

// ProcessStatusHandler is the disabled implementation for the status handler that keeps track what the node is doing:
// processing blocks or idle
type ProcessStatusHandler struct {
}

// SetToBusy does nothing
func (psh *ProcessStatusHandler) SetToBusy(_ string) {}

// SetToIdle does nothing
func (psh *ProcessStatusHandler) SetToIdle() {}

// IsIdle returns true
func (psh *ProcessStatusHandler) IsIdle() bool {
	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (psh *ProcessStatusHandler) IsInterfaceNil() bool {
	return psh == nil
}
