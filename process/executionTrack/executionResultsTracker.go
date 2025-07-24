package executionTrack

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data/block"
)

type executionResultsTracker struct {
	lastNotarizedResult    *block.ExecutionResult
	mutex                  sync.RWMutex
	executionResultsByHash map[string]*block.ExecutionResult
	nonceHash              *nonceHash
	lastExecutedResultHash []byte
}

// NewExecutionResultsTracker will create a new instance of *executionResultsTracker
func NewExecutionResultsTracker() (*executionResultsTracker, error) {
	return &executionResultsTracker{
		executionResultsByHash: make(map[string]*block.ExecutionResult),
		nonceHash:              newNonceHash(),
	}, nil
}

// AddExecutionResult will add the provided execution result in tracker
// It will return true if the execution result was added in the tracker
func (est *executionResultsTracker) AddExecutionResult(executionResult *block.ExecutionResult) error {
	if executionResult == nil {
		return ErrNilExecutionResult
	}

	est.mutex.Lock()
	defer est.mutex.Unlock()
	if est.lastNotarizedResult == nil {
		return ErrNilLastNotarizedExecutionResult
	}

	if est.lastNotarizedResult.Nonce > executionResult.Nonce {
		return fmt.Errorf("execution results nonce(%d) is lower than last notarized nonce(%d)", executionResult.Nonce, est.lastNotarizedResult.Nonce)
	}

	lastExecutedResult, err := est.getLastExecutedResult()
	if err != nil {
		return err
	}

	if lastExecutedResult.Nonce != executionResult.Nonce-1 {
		return fmt.Errorf("execution results nonce(%d) should be equal to the subsequent nonce after last executed(%d)", executionResult.Nonce, lastExecutedResult.Nonce)
	}

	est.executionResultsByHash[string(executionResult.HeaderHash)] = executionResult
	est.nonceHash.addNonceHash(executionResult.Nonce, string(executionResult.HeaderHash))

	est.lastExecutedResultHash = executionResult.HeaderHash

	return nil
}

func (est *executionResultsTracker) getLastExecutedResult() (*block.ExecutionResult, error) {
	if bytes.Equal(est.lastExecutedResultHash, est.lastNotarizedResult.HeaderHash) {
		return est.lastNotarizedResult, nil
	}

	lastExecutedResults, found := est.executionResultsByHash[string(est.lastExecutedResultHash)]
	if !found {
		return nil, fmt.Errorf("last executed result not found hash=(%s)", hex.EncodeToString(est.lastExecutedResultHash))
	}

	return lastExecutedResults, nil
}

// GetPendingExecutionResults will return the pending execution results
func (est *executionResultsTracker) GetPendingExecutionResults() ([]*block.ExecutionResult, error) {
	est.mutex.RLock()
	defer est.mutex.RUnlock()

	return est.getPendingExecutionResults()
}

func (est *executionResultsTracker) getPendingExecutionResults() ([]*block.ExecutionResult, error) {
	executionResults := make([]*block.ExecutionResult, 0, len(est.executionResultsByHash))
	for _, executionResult := range est.executionResultsByHash {
		executionResults = append(executionResults, executionResult)
	}

	if len(executionResults) == 0 {
		return executionResults, nil
	}

	sort.Slice(executionResults, func(i, j int) bool {
		return executionResults[i].Nonce < executionResults[j].Nonce
	})

	firstElementHasCorrectNonce := executionResults[0].Nonce == est.lastNotarizedResult.Nonce+1
	if !firstElementHasCorrectNonce {
		return nil, ErrDifferentNoncesConfirmedExecutionResults
	}

	for idx := 0; idx < len(executionResults)-1; idx++ {
		hasConsecutiveNonces := executionResults[idx].Nonce+1 == executionResults[idx+1].Nonce
		if !hasConsecutiveNonces {
			return nil, ErrDifferentNoncesConfirmedExecutionResults
		}
	}

	return executionResults, nil
}

// GetPendingExecutionResultByHash will return the execution results by hash
func (est *executionResultsTracker) GetPendingExecutionResultByHash(hash []byte) (*block.ExecutionResult, error) {
	est.mutex.RLock()
	defer est.mutex.RUnlock()

	result, found := est.executionResultsByHash[string(hash)]
	if !found {
		return nil, fmt.Errorf("%w with hash: '%s'", ErrCannotFindExecutionResult, hex.EncodeToString(hash))
	}

	return result, nil
}

// GetPendingExecutionResultByNonce will return the execution results by nonce
func (est *executionResultsTracker) GetPendingExecutionResultByNonce(nonce uint64) (*block.ExecutionResult, error) {
	est.mutex.RLock()
	defer est.mutex.RUnlock()

	return est.getPendingExecutionResultsByNonce(nonce)
}

func (est *executionResultsTracker) getPendingExecutionResultsByNonce(nonce uint64) (*block.ExecutionResult, error) {
	hash := est.nonceHash.getHashByNonce(nonce)
	result, found := est.executionResultsByHash[hash]
	if !found {
		return nil, fmt.Errorf("%w with hash: '%s'", ErrCannotFindExecutionResult, hex.EncodeToString([]byte(hash)))
	}

	return result, nil
}

// CleanConfirmedExecutionResults will clean the confirmed execution results
func (est *executionResultsTracker) CleanConfirmedExecutionResults(header HeaderWithExecutionResults) (*CleanInfo, error) {
	est.mutex.Lock()
	defer est.mutex.Unlock()

	headerExecutionResults := header.GetExecutionResults()
	if len(headerExecutionResults) == 0 {
		return &CleanInfo{
			CleanResult:             CleanResultOK,
			LastMatchingResultNonce: est.lastNotarizedResult.Nonce,
		}, nil
	}

	return est.cleanConfirmedExecutionResults(headerExecutionResults)
}

func (est *executionResultsTracker) cleanConfirmedExecutionResults(headerExecutionResults []*block.ExecutionResult) (*CleanInfo, error) {
	pendingExecutionResult, err := est.getPendingExecutionResults()
	if err != nil {
		return nil, err
	}

	lastMatchingResultNonce := est.lastNotarizedResult.Nonce
	lastMatchingHash := est.lastNotarizedResult.HeaderHash
	for idx, executionResultFromHeader := range headerExecutionResults {
		if idx > len(pendingExecutionResult)-1 {
			// missing  execution result
			return &CleanInfo{
				CleanResult:             CleanResultNotFound,
				LastMatchingResultNonce: lastMatchingResultNonce,
			}, nil
		}

		executionResultFromTracker := pendingExecutionResult[idx]

		areEqual := executionResultFromTracker.Equal(executionResultFromHeader)
		if !areEqual {
			est.lastExecutedResultHash = lastMatchingHash

			// different execution result should clean everything starting from this execution result and return CleanResultMismatch
			est.cleanExecutionResults(pendingExecutionResult[idx:])

			return &CleanInfo{
				CleanResult:             CleanResultMismatch,
				LastMatchingResultNonce: lastMatchingResultNonce,
			}, nil
		}

		lastMatchingResultNonce = executionResultFromHeader.Nonce
		lastMatchingHash = executionResultFromHeader.HeaderHash
	}

	est.cleanExecutionResults(headerExecutionResults)
	est.lastNotarizedResult = headerExecutionResults[len(headerExecutionResults)-1]

	return &CleanInfo{
		CleanResult:             CleanResultOK,
		LastMatchingResultNonce: est.lastNotarizedResult.Nonce,
	}, nil
}

func (est *executionResultsTracker) cleanExecutionResults(executionResult []*block.ExecutionResult) {
	for _, result := range executionResult {
		delete(est.executionResultsByHash, string(result.HeaderHash))
		est.nonceHash.removeByNonce(result.Nonce)
	}
}

// GetLastNotarizedExecutionResult will return the last notarized execution result
func (est *executionResultsTracker) GetLastNotarizedExecutionResult() (*block.ExecutionResult, error) {
	est.mutex.RLock()
	defer est.mutex.RUnlock()
	if est.lastNotarizedResult == nil {
		return nil, ErrNilLastNotarizedExecutionResult
	}

	return est.lastNotarizedResult, nil
}

// SetLastNotarizedResult will set the last notarized execution result
func (est *executionResultsTracker) SetLastNotarizedResult(executionResult *block.ExecutionResult) error {
	if executionResult == nil {
		return ErrNilExecutionResult
	}

	est.mutex.Lock()
	est.lastNotarizedResult = executionResult
	est.lastExecutedResultHash = executionResult.HeaderHash
	est.mutex.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (est *executionResultsTracker) IsInterfaceNil() bool {
	return est == nil
}
