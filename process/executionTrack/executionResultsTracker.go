package executionTrack

import (
	"github.com/multiversx/mx-chain-core-go/data/block"
	"sync"
)

type executionResultConfirmed struct {
	executionResult *block.ExecutionResult
	confirmed       bool
}

type executionResultsTracker struct {
	mutex            sync.RWMutex
	executionResults map[string]*executionResultConfirmed
}

// NewExecutionResultsTracker will create a new instance of *executionResultsTracker
func NewExecutionResultsTracker() (*executionResultsTracker, error) {
	return &executionResultsTracker{
		executionResults: make(map[string]*executionResultConfirmed),
	}, nil
}

// AddExecutionResult will add the provided execution result in tracker
// It will return true if the execution result was added in the tracker
func (est *executionResultsTracker) AddExecutionResult(executionResult *block.ExecutionResult) bool {
	if executionResult == nil {
		return false
	}

	est.mutex.Lock()
	defer est.mutex.Unlock()
	est.executionResults[string(executionResult.HeaderHash)] = &executionResultConfirmed{
		executionResult: executionResult,
	}

	return true
}

// ConfirmExecutionResult will mark as confirmed the execution results that belongs to the provided hash
func (est *executionResultsTracker) ConfirmExecutionResult(headerHash []byte) error {
	est.mutex.Lock()
	defer est.mutex.Unlock()

	executionResult, found := est.executionResults[string(headerHash)]
	if !found {
		return ErrCannotFindExecutionResult
	}

	executionResult.confirmed = true

	return nil
}

// PopConfirmedExecutionResults will pop from the tracker all the confirmed execution results
func (est *executionResultsTracker) PopConfirmedExecutionResults() ([]*block.ExecutionResult, error) {
	confirmedExecutionResults := make([]*block.ExecutionResult, 0)
	est.mutex.Lock()
	defer est.mutex.Unlock()

	for hash, executionResult := range est.executionResults {
		if !executionResult.confirmed {
			continue
		}
		confirmedExecutionResults = append(confirmedExecutionResults, executionResult.executionResult)
		delete(est.executionResults, hash)
	}

	return confirmedExecutionResults, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (est *executionResultsTracker) IsInterfaceNil() bool {
	return est == nil
}
