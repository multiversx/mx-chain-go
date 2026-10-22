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

// TrySetBusy returns true
func (psh *processStatusHandler) TrySetBusy(_ string) bool { return true }

// SetIdle does nothing
func (psh *processStatusHandler) SetIdle() {}

// BlockBackgroundJobs does nothing
func (psh *processStatusHandler) BlockBackgroundJobs(_ string) {}

// UnblockBackgroundJobs does nothing
func (psh *processStatusHandler) UnblockBackgroundJobs() {}

// SuspendBackgroundJobBlocking does nothing
func (psh *processStatusHandler) SuspendBackgroundJobBlocking(_ string) {}

// ResumeBackgroundJobBlocking does nothing
func (psh *processStatusHandler) ResumeBackgroundJobBlocking() {}

// IsIdle returns true
func (psh *processStatusHandler) IsIdle() bool {
	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (psh *processStatusHandler) IsInterfaceNil() bool {
	return psh == nil
}
