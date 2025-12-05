package disabled

type headersExecutor struct {
}

// NewHeadersExecutor returns a new instance of disabled headers executor
func NewHeadersExecutor() *headersExecutor {
	return &headersExecutor{}
}

// StartExecution does nothing as it is disabled
func (he *headersExecutor) StartExecution() {
}

// PauseExecution does nothing as it is disabled
func (he *headersExecutor) PauseExecution() {
}

// ResumeExecution does nothing as it is disabled
func (he *headersExecutor) ResumeExecution() {
}

// Close always returns nil as it is disabled
func (he *headersExecutor) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (he *headersExecutor) IsInterfaceNil() bool {
	return he == nil
}
