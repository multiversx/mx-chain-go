package executionTrack

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/multiversx/mx-chain-core-go/data"
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data/block"
)

// ArgsExecutionResultsTracker holds all the components needed to create a new instance of executionResultsTracker
type ArgsExecutionResultsTracker struct {
}

type executionResultsTracker struct {
	lastNotarizedNonce     uint64
	mutex                  sync.RWMutex
	executionResultsByHash map[string]*block.ExecutionResult
	nonceHashes            *nonceHashes
}

// NewExecutionResultsTracker will create a new instance of *executionResultsTracker
func NewExecutionResultsTracker(_ ArgsExecutionResultsTracker) (*executionResultsTracker, error) {
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
	if est.lastNotarizedNonce >= executionResult.Nonce {
		return fmt.Errorf("execution results nonce(%d) is lower then last notarized nonce(%d)", executionResult.Nonce, est.lastNotarizedNonce)
	}

	est.executionResultsByHash[string(executionResult.HeaderHash)] = executionResult
	est.nonceHashes.addNonceHash(executionResult.Nonce, string(executionResult.HeaderHash))

	return nil
}

func (est *executionResultsTracker) cleanExecutionResultsWithSameNonce(nonce uint64, confirmedHash string) {
	hashes := est.nonceHashes.popDifferentHashes(nonce, confirmedHash)
	for _, hash := range hashes {
		delete(est.executionResultsByHash, hash)
	}
}

// GetPendingExecutionResults will return the pending execution results
func (est *executionResultsTracker) GetPendingExecutionResults() ([]*block.ExecutionResult, error) {
	executionResults := make([]*block.ExecutionResult, 0)
	est.mutex.Lock()
	defer est.mutex.Unlock()

	for _, executionResult := range est.executionResultsByHash {
		executionResults = append(executionResults, executionResult)
	}

	if len(executionResults) == 0 {
		return executionResults, nil
	}

	sort.Slice(executionResults, func(i, j int) bool {
		if executionResults[i].Nonce < executionResults[j].Nonce {
			return true
		}
		return false
	})

	firstElementHasCorrectNonce := executionResults[0].Nonce+1 == est.lastNotarizedNonce
	if !firstElementHasCorrectNonce {
		return nil, errors.New("confirmed execution results have different nonces")
	}

	for idx := 0; idx < len(executionResults)-1; idx++ {
		hasConsecutiveNonces := executionResults[idx].Nonce+1 == executionResults[idx+1].Nonce
		if !hasConsecutiveNonces {
			return nil, errors.New("confirmed execution results have different nonces")
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
		return nil, errors.New("cannot find execution results by provided hash	")
	}

	return result, nil
}

// GetExecutionResultByNonce will return the execution results by nonce
func (est *executionResultsTracker) GetExecutionResultByNonce(nonce uint64) ([]*block.ExecutionResult, error) {
	executionResults := make([]*block.ExecutionResult, 0)

	est.mutex.RLock()
	defer est.mutex.RUnlock()

	hashes := est.nonceHashes.getNonceHashes(nonce)
	for _, hash := range hashes {
		result, found := est.executionResultsByHash[hash]
		if !found {
			return nil, fmt.Errorf("cannot find execution results by provided hash: %s", hex.EncodeToString([]byte(hash)))
		}

		executionResults = append(executionResults, result)
	}

	return executionResults, nil
}

// CleanConfirmedExecutionResults will clean the confirmed execution results
func (est *executionResultsTracker) CleanConfirmedExecutionResults(headerHash []byte, header data.HeaderHandler) error {
	est.mutex.Lock()
	defer est.mutex.Unlock()

	// extend header

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (est *executionResultsTracker) IsInterfaceNil() bool {
	return est == nil
}
