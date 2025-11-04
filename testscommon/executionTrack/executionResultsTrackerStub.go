package executionTrack

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// ExecutionResultsTrackerStub is a stub implementation of the ExecutionResultsTracker interface
type ExecutionResultsTrackerStub struct {
	AddExecutionResultCalled               func(executionResult data.BaseExecutionResultHandler) error
	GetPendingExecutionResultsCalled       func() ([]data.BaseExecutionResultHandler, error)
	GetPendingExecutionResultByHashCalled  func(hash []byte) (data.BaseExecutionResultHandler, error)
	GetPendingExecutionResultByNonceCalled func(nonce uint64) (data.BaseExecutionResultHandler, error)
	GetLastNotarizedExecutionResultCalled  func() (data.BaseExecutionResultHandler, error)
	SetLastNotarizedResultCalled           func(executionResult data.BaseExecutionResultHandler) error
	RemoveFromNonceCalled                  func(nonce uint64) error
	GetLastExecutionResultCalled           func() (data.BaseExecutionResultHandler, error)
}

// AddExecutionResult -
func (ets *ExecutionResultsTrackerStub) AddExecutionResult(executionResult data.BaseExecutionResultHandler) error {
	if ets.AddExecutionResultCalled != nil {
		return ets.AddExecutionResultCalled(executionResult)
	}
	return nil
}

// GetPendingExecutionResults -
func (ets *ExecutionResultsTrackerStub) GetPendingExecutionResults() ([]data.BaseExecutionResultHandler, error) {
	if ets.GetPendingExecutionResultsCalled != nil {
		return ets.GetPendingExecutionResultsCalled()
	}
	return nil, nil
}

// GetPendingExecutionResultByHash -
func (ets *ExecutionResultsTrackerStub) GetPendingExecutionResultByHash(hash []byte) (data.BaseExecutionResultHandler, error) {
	if ets.GetPendingExecutionResultByHashCalled != nil {
		return ets.GetPendingExecutionResultByHashCalled(hash)
	}
	return nil, nil
}

// GetPendingExecutionResultByNonce -
func (ets *ExecutionResultsTrackerStub) GetPendingExecutionResultByNonce(nonce uint64) (data.BaseExecutionResultHandler, error) {
	if ets.GetPendingExecutionResultByNonceCalled != nil {
		return ets.GetPendingExecutionResultByNonceCalled(nonce)
	}
	return nil, nil
}

// GetLastNotarizedExecutionResult -
func (ets *ExecutionResultsTrackerStub) GetLastNotarizedExecutionResult() (data.BaseExecutionResultHandler, error) {
	if ets.GetLastNotarizedExecutionResultCalled != nil {
		return ets.GetLastNotarizedExecutionResultCalled()
	}
	return nil, nil
}

// SetLastNotarizedResult -
func (ets *ExecutionResultsTrackerStub) SetLastNotarizedResult(executionResult data.BaseExecutionResultHandler) error {
	if ets.SetLastNotarizedResultCalled != nil {
		return ets.SetLastNotarizedResultCalled(executionResult)
	}
	return nil
}

// RemoveFromNonce -
func (ets *ExecutionResultsTrackerStub) RemoveFromNonce(nonce uint64) error {
	if ets.RemoveFromNonceCalled != nil {
		return ets.RemoveFromNonceCalled(nonce)
	}
	return nil
}

// GetLastExecutionResult -
func (ets *ExecutionResultsTrackerStub) GetLastExecutionResult() (data.BaseExecutionResultHandler, error) {
	if ets.GetLastExecutionResultCalled != nil {
		return ets.GetLastExecutionResultCalled()
	}

	return nil, nil
}

// IsInterfaceNil checks if the interface is nil
func (ets *ExecutionResultsTrackerStub) IsInterfaceNil() bool {
	return ets == nil
}
