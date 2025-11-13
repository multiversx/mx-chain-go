package processMocks

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/queue"
)

// ExecutionManagerMock is a mock implementation of the ExecutionManager interface
type ExecutionManagerMock struct {
	StartExecutionCalled                 func()
	SetHeadersExecutorCalled             func(executor process.HeadersExecutor) error
	AddPairForExecutionCalled            func(pair queue.HeaderBodyPair) error
	GetPendingExecutionResultsCalled     func() ([]data.BaseExecutionResultHandler, error)
	CleanConfirmedExecutionResultsCalled func(header data.HeaderHandler) error
	SetLastNotarizedResultCalled         func(executionResult data.BaseExecutionResultHandler) error
	RemoveAtNonceAndHigherCalled         func(nonce uint64) error
	ResetAndResumeExecutionCalled        func() error
	CloseCalled                          func() error
}

// StartExecution -
func (emm *ExecutionManagerMock) StartExecution() {
	if emm.StartExecutionCalled != nil {
		emm.StartExecutionCalled()
	}
}

// SetHeadersExecutor -
func (emm *ExecutionManagerMock) SetHeadersExecutor(executor process.HeadersExecutor) error {
	if emm.SetHeadersExecutorCalled != nil {
		return emm.SetHeadersExecutorCalled(executor)
	}
	return nil
}

// AddPairForExecution -
func (emm *ExecutionManagerMock) AddPairForExecution(pair queue.HeaderBodyPair) error {
	if emm.AddPairForExecutionCalled != nil {
		return emm.AddPairForExecutionCalled(pair)
	}
	return nil
}

// GetPendingExecutionResults -
func (emm *ExecutionManagerMock) GetPendingExecutionResults() ([]data.BaseExecutionResultHandler, error) {
	if emm.GetPendingExecutionResultsCalled != nil {
		return emm.GetPendingExecutionResultsCalled()
	}
	return nil, nil
}

// CleanConfirmedExecutionResults -
func (emm *ExecutionManagerMock) CleanConfirmedExecutionResults(header data.HeaderHandler) error {
	if emm.CleanConfirmedExecutionResultsCalled != nil {
		return emm.CleanConfirmedExecutionResultsCalled(header)
	}
	return nil
}

// SetLastNotarizedResult -
func (emm *ExecutionManagerMock) SetLastNotarizedResult(executionResult data.BaseExecutionResultHandler) error {
	if emm.SetLastNotarizedResultCalled != nil {
		return emm.SetLastNotarizedResultCalled(executionResult)
	}
	return nil
}

// RemoveAtNonceAndHigher -
func (emm *ExecutionManagerMock) RemoveAtNonceAndHigher(nonce uint64) error {
	if emm.RemoveAtNonceAndHigherCalled != nil {
		return emm.RemoveAtNonceAndHigherCalled(nonce)
	}
	return nil
}

// ResetAndResumeExecution -
func (emm *ExecutionManagerMock) ResetAndResumeExecution() error {
	if emm.ResetAndResumeExecutionCalled != nil {
		return emm.ResetAndResumeExecutionCalled()
	}
	return nil
}

// Close -
func (emm *ExecutionManagerMock) Close() error {
	if emm.CloseCalled != nil {
		return emm.CloseCalled()
	}
	return nil
}

// IsInterfaceNil -
func (emm *ExecutionManagerMock) IsInterfaceNil() bool {
	return emm == nil
}
