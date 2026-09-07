package processMocks

import (
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/process/asyncExecution/executionTrack"
)

// ExecutionTrackerStub -
type ExecutionTrackerStub struct {
	AddExecutionResultCalled               func(executionResult data.BaseExecutionResultHandler) (bool, error)
	GetPendingExecutionResultsCalled       func() ([]data.BaseExecutionResultHandler, error)
	GetPendingExecutionResultByHashCalled  func(hash []byte) (data.BaseExecutionResultHandler, error)
	GetPendingExecutionResultByNonceCalled func(nonce uint64) (data.BaseExecutionResultHandler, error)
	GetLastNotarizedExecutionResultCalled  func() (data.BaseExecutionResultHandler, error)
	SetLastNotarizedResultCalled           func(executionResult data.BaseExecutionResultHandler) error
	RemoveFromNonceCalled                  func(nonce uint64) error
	CleanCalled                            func(lastNotarizedResult data.BaseExecutionResultHandler)
	CleanConfirmedExecutionResultsCalled   func(header data.HeaderHandler) error
	CleanOnConsensusReachedCalled          func(headerHash []byte, header data.HeaderHandler)
	PopDismissedResultsCalled              func() []executionTrack.DismissedBatch
}

// PopDismissedResults -
func (e *ExecutionTrackerStub) PopDismissedResults() []executionTrack.DismissedBatch {
	if e.PopDismissedResultsCalled != nil {
		return e.PopDismissedResultsCalled()
	}

	return nil
}

// AddExecutionResult -
func (e *ExecutionTrackerStub) AddExecutionResult(executionResult data.BaseExecutionResultHandler) (bool, error) {
	if e.AddExecutionResultCalled != nil {
		return e.AddExecutionResultCalled(executionResult)
	}

	return true, nil
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
func (e *ExecutionTrackerStub) Clean(lastNotarizedResult data.BaseExecutionResultHandler) {
	if e.CleanCalled != nil {
		e.CleanCalled(lastNotarizedResult)
	}
}

// CleanConfirmedExecutionResults -
func (e *ExecutionTrackerStub) CleanConfirmedExecutionResults(header data.HeaderHandler) error {
	if e.CleanConfirmedExecutionResultsCalled != nil {
		return e.CleanConfirmedExecutionResultsCalled(header)
	}

	return nil
}

// CleanOnConsensusReached -
func (e *ExecutionTrackerStub) CleanOnConsensusReached(headerHash []byte, header data.HeaderHandler) {
	if e.CleanOnConsensusReachedCalled != nil {
		e.CleanOnConsensusReachedCalled(headerHash, header)
	}
}

// IsInterfaceNil -
func (e *ExecutionTrackerStub) IsInterfaceNil() bool {
	return e == nil
}
