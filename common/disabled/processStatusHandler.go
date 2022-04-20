package disabled

// processStatusHandler is the disabled implementation for the status handler that keeps track what the node is doing:
// processing blocks or idle
type processStatusHandler struct {
}

// NewProcessStatusHandler creates a new instance of type processStatusHandler
func NewProcessStatusHandler() *processStatusHandler {
	return &processStatusHandler{}
}

// SetBusy does nothing
func (psh *processStatusHandler) SetBusy(_ string) {}

// SetIdle does nothing
func (psh *processStatusHandler) SetIdle() {}

// IsIdle returns true
func (psh *processStatusHandler) IsIdle() bool {
	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (psh *processStatusHandler) IsInterfaceNil() bool {
	return psh == nil
}
