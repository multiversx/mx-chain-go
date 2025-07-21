package executionTrack

import (
	"encoding/hex"
	"fmt"
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
)

type executionResultsTracker struct {
	lastNotarizedResult    *block.ExecutionResult
	mutex                  sync.RWMutex
	executionResultsByHash map[string]*block.ExecutionResult
	nonceHashes            *nonceHashes
}

// NewExecutionResultsTracker will create a new instance of *executionResultsTracker
func NewExecutionResultsTracker() (*executionResultsTracker, error) {
	return &executionResultsTracker{
		executionResultsByHash: make(map[string]*block.ExecutionResult),
		nonceHashes:            newNonceHashes(),
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

	if est.lastNotarizedResult.Nonce >= executionResult.Nonce {
		return fmt.Errorf("execution results nonce(%d) is lower than last notarized nonce(%d)", executionResult.Nonce, est.lastNotarizedResult.Nonce)
	}

	est.executionResultsByHash[string(executionResult.HeaderHash)] = executionResult
	est.nonceHashes.addNonceHash(executionResult.Nonce, string(executionResult.HeaderHash))

	return nil
}

// GetPendingExecutionResults will return the pending execution results
func (est *executionResultsTracker) GetPendingExecutionResults() ([]*block.ExecutionResult, error) {
	executionResults := make([]*block.ExecutionResult, 0, len(est.executionResultsByHash))
	est.mutex.RLock()
	defer est.mutex.RUnlock()

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

// GetExecutionResultByHash will return the execution results by hash
func (est *executionResultsTracker) GetExecutionResultByHash(hash []byte) (*block.ExecutionResult, error) {
	est.mutex.RLock()
	defer est.mutex.RUnlock()

	result, found := est.executionResultsByHash[string(hash)]
	if !found {
		return nil, fmt.Errorf("%w with hash: '%s'", ErrCannotFindExecutionResult, hex.EncodeToString(hash))
	}

	return result, nil
}

// GetExecutionResultsByNonce will return the execution results by nonce
func (est *executionResultsTracker) GetExecutionResultsByNonce(nonce uint64) ([]*block.ExecutionResult, error) {
	est.mutex.RLock()
	defer est.mutex.RUnlock()

	hashes := est.nonceHashes.getNonceHashes(nonce)
	executionResults := make([]*block.ExecutionResult, 0, len(hashes))
	for _, hash := range hashes {
		result, found := est.executionResultsByHash[hash]
		if !found {
			return nil, fmt.Errorf("%w with hash: '%s'", ErrCannotFindExecutionResult, hex.EncodeToString([]byte(hash)))
		}

		executionResults = append(executionResults, result)
	}

	return executionResults, nil
}

// CleanConfirmedExecutionResults will clean the confirmed execution results
// TODO add unit tests when new header structure is available
func (est *executionResultsTracker) CleanConfirmedExecutionResults(headerHash []byte, header data.HeaderHandler) error {
	est.mutex.Lock()
	defer est.mutex.Unlock()

	headerExecutionResults := make([]*block.ExecutionResult, 0) // TODO get from header
	if len(headerExecutionResults) == 0 {
		return nil
	}

	for _, executionResultFromHeader := range headerExecutionResults {
		executionResultFromTracker, found := est.executionResultsByHash[string(executionResultFromHeader.HeaderHash)]
		if !found {
			return fmt.Errorf("%w with for header hash %s", ErrCannotFindExecutionResult, hex.EncodeToString(executionResultFromHeader.HeaderHash))
		}

		areEqual := executionResultFromTracker.Equal(executionResultFromHeader)
		if !areEqual {
			return fmt.Errorf("$%w for header hash", ErrDifferentExecutionResults)
		}

		delete(est.executionResultsByHash, string(executionResultFromHeader.HeaderHash))
		est.nonceHashes.removeByNonce(executionResultFromHeader.Nonce)
	}

	est.lastNotarizedResult = headerExecutionResults[len(headerExecutionResults)-1]

	return nil
}

// SetLastNotarizedResult will set the last notarized execution result
func (est *executionResultsTracker) SetLastNotarizedResult(executionResult *block.ExecutionResult) error {
	if executionResult == nil {
		return ErrNilExecutionResult
	}

	est.mutex.Lock()
	est.lastNotarizedResult = executionResult
	est.mutex.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (est *executionResultsTracker) IsInterfaceNil() bool {
	return est == nil
}
