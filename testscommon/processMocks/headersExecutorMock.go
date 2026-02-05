package processMocks

// HeadersExecutorMock -
type HeadersExecutorMock struct {
	StartExecutionCalled  func()
	PauseExecutionCalled  func()
	ResumeExecutionCalled func()
	CloseCalled           func() error
}

// StartExecution -
func (mock *HeadersExecutorMock) StartExecution() {
	if mock.StartExecutionCalled != nil {
		mock.StartExecutionCalled()
	}
}

// PauseExecution -
func (mock *HeadersExecutorMock) PauseExecution() {
	if mock.PauseExecutionCalled != nil {
		mock.PauseExecutionCalled()
	}
}

// ResumeExecution -
func (mock *HeadersExecutorMock) ResumeExecution() {
	if mock.ResumeExecutionCalled != nil {
		mock.ResumeExecutionCalled()
	}
}

// Close -
func (mock *HeadersExecutorMock) Close() error {
	if mock.CloseCalled != nil {
		return mock.CloseCalled()
	}
	return nil
}

// IsInterfaceNil -
func (mock *HeadersExecutorMock) IsInterfaceNil() bool {
	return mock == nil
}
