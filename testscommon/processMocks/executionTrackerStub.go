package processMocks

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// ExecutionTrackerStub -
type ExecutionTrackerStub struct {
	AddExecutionResultCalled               func(executionResult data.BaseExecutionResultHandler) error
	GetPendingExecutionResultsCalled       func() ([]data.BaseExecutionResultHandler, error)
	GetPendingExecutionResultByHashCalled  func(hash []byte) (data.BaseExecutionResultHandler, error)
	GetPendingExecutionResultByNonceCalled func(nonce uint64) (data.BaseExecutionResultHandler, error)
	GetLastNotarizedExecutionResultCalled  func() (data.BaseExecutionResultHandler, error)
	SetLastNotarizedResultCalled           func(executionResult data.BaseExecutionResultHandler) error
	RemoveFromNonceCalled                  func(nonce uint64) error
	CleanCalled                            func(lastNotarizedResult data.BaseExecutionResultHandler) error
	CleanConfirmedExecutionResultsCalled   func(header data.HeaderHandler) error
}

// AddExecutionResult -
func (e *ExecutionTrackerStub) AddExecutionResult(executionResult data.BaseExecutionResultHandler) error {
	if e.AddExecutionResultCalled != nil {
		return e.AddExecutionResultCalled(executionResult)
	}

	return nil
}

// GetPendingExecutionResults -
func (e *ExecutionTrackerStub) GetPendingExecutionResults() ([]data.BaseExecutionResultHandler, error) {
	if e.GetPendingExecutionResultsCalled != nil {
		return e.GetPendingExecutionResultsCalled()
	}

	return nil, nil
}

// GetPendingExecutionResultByHash -
func (e *ExecutionTrackerStub) GetPendingExecutionResultByHash(hash []byte) (data.BaseExecutionResultHandler, error) {
	if e.GetPendingExecutionResultByHashCalled != nil {
		return e.GetPendingExecutionResultByHashCalled(hash)
	}

	return nil, nil
}

// GetPendingExecutionResultByNonce -
func (e *ExecutionTrackerStub) GetPendingExecutionResultByNonce(nonce uint64) (data.BaseExecutionResultHandler, error) {
	if e.GetPendingExecutionResultByNonceCalled != nil {
		return e.GetPendingExecutionResultByNonceCalled(nonce)
	}

	return nil, nil
}

// GetLastNotarizedExecutionResult -
func (e *ExecutionTrackerStub) GetLastNotarizedExecutionResult() (data.BaseExecutionResultHandler, error) {
	if e.GetLastNotarizedExecutionResultCalled != nil {
		return e.GetLastNotarizedExecutionResultCalled()
	}

	return nil, nil
}

// SetLastNotarizedResult -
func (e *ExecutionTrackerStub) SetLastNotarizedResult(executionResult data.BaseExecutionResultHandler) error {
	if e.SetLastNotarizedResultCalled != nil {
		return e.SetLastNotarizedResultCalled(executionResult)
	}

	return nil
}

// RemoveFromNonce -
func (e *ExecutionTrackerStub) RemoveFromNonce(nonce uint64) error {
	if e.RemoveFromNonceCalled != nil {
		return e.RemoveFromNonceCalled(nonce)
	}

	return nil
}

// Clean -
func (e *ExecutionTrackerStub) Clean(lastNotarizedResult data.BaseExecutionResultHandler) error {
	if e.CleanCalled != nil {
		return e.CleanCalled(lastNotarizedResult)
	}
	return nil
}

// CleanConfirmedExecutionResults -
func (e *ExecutionTrackerStub) CleanConfirmedExecutionResults(header data.HeaderHandler) error {
	if e.CleanConfirmedExecutionResultsCalled != nil {
		return e.CleanConfirmedExecutionResultsCalled(header)
	}

	return nil
}

// IsInterfaceNil -
func (e *ExecutionTrackerStub) IsInterfaceNil() bool {
	return e == nil
}
