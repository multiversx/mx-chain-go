package processMocks

import (
	"github.com/multiversx/mx-chain-core-go/data"
)

// ExecutionTrackerStub -
type ExecutionTrackerStub struct {
	AddExecutionResultCalled func(executionResult data.ExecutionResultHandler) error
}

// AddExecutionResult -
func (e *ExecutionTrackerStub) AddExecutionResult(executionResult data.ExecutionResultHandler) error {
	if e.AddExecutionResultCalled != nil {
		return e.AddExecutionResultCalled(executionResult)
	}

	return nil
}

// IsInterfaceNil -
func (e *ExecutionTrackerStub) IsInterfaceNil() bool {
	return e == nil
}
