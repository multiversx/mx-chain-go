package processMocks

import (
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/cache"
)

// ExecutionManagerMock is a mock implementation of the ExecutionManager interface
type ExecutionManagerMock struct {
	StartExecutionCalled                         func()
	SetHeadersExecutorCalled                     func(executor process.HeadersExecutor) error
	AddPairForExecutionCalled                    func(pair cache.HeaderBodyPair) error
	GetPendingExecutionResultsCalled             func() ([]data.BaseExecutionResultHandler, error)
	CleanConfirmedExecutionResultsCalled         func(header data.HeaderHandler) error
	CleanOnConsensusReachedCalled                func(headerHash []byte, nonce uint64)
	SetLastNotarizedResultCalled                 func(executionResult data.BaseExecutionResultHandler) error
	RemoveAtNonceAndHigherCalled                 func(nonce uint64) error
	ResetAndResumeExecutionCalled                func(lastNotarizedResult data.BaseExecutionResultHandler) error
	GetLastNotarizedExecutionResultCalled        func() (data.BaseExecutionResultHandler, error)
	RemovePendingExecutionResultsFromNonceCalled func(nonce uint64) error
	CloseCalled                                  func() error
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
func (emm *ExecutionManagerMock) AddPairForExecution(pair cache.HeaderBodyPair) error {
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

// CleanOnConsensusReached -
func (emm *ExecutionManagerMock) CleanOnConsensusReached(headerHash []byte, nonce uint64) {
	if emm.CleanOnConsensusReachedCalled != nil {
		emm.CleanOnConsensusReachedCalled(headerHash, nonce)
	}
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
func (emm *ExecutionManagerMock) ResetAndResumeExecution(lastNotarizedResult data.BaseExecutionResultHandler) error {
	if emm.ResetAndResumeExecutionCalled != nil {
		return emm.ResetAndResumeExecutionCalled(lastNotarizedResult)
	}
	return nil
}

// GetLastNotarizedExecutionResult -
func (emm *ExecutionManagerMock) GetLastNotarizedExecutionResult() (data.BaseExecutionResultHandler, error) {
	if emm.GetLastNotarizedExecutionResultCalled != nil {
		return emm.GetLastNotarizedExecutionResultCalled()
	}
	return nil, nil
}

// RemovePendingExecutionResultsFromNonce -
func (emm *ExecutionManagerMock) RemovePendingExecutionResultsFromNonce(nonce uint64) error {
	if emm.RemovePendingExecutionResultsFromNonceCalled != nil {
		return emm.RemovePendingExecutionResultsFromNonceCalled(nonce)
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
