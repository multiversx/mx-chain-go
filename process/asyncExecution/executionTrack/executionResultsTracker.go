package executionTrack

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("process/asyncExecution/executionTrack")

type executionResultsTracker struct {
	lastNotarizedResult    data.ExecutionResultHandler
	mutex                  sync.RWMutex
	executionResultsByHash map[string]data.ExecutionResultHandler
	nonceHash              *nonceHash
	lastExecutedResultHash []byte
	hashToRemoveOnAdd      []byte
}

// NewExecutionResultsTracker will create a new instance of *executionResultsTracker
func NewExecutionResultsTracker() *executionResultsTracker {
	return &executionResultsTracker{
		executionResultsByHash: make(map[string]data.ExecutionResultHandler),
		nonceHash:              newNonceHash(),
	}
}

// AddExecutionResult will add the provided execution result in tracker
// It will return true if the execution result was added in the tracker
func (est *executionResultsTracker) AddExecutionResult(executionResult data.ExecutionResultHandler) error {
	if executionResult == nil {
		return ErrNilExecutionResult
	}

	est.mutex.Lock()
	defer est.mutex.Unlock()
	if est.lastNotarizedResult == nil {
		return ErrNilLastNotarizedExecutionResult
	}

	shouldIgnoreExecutionResult := bytes.Equal(est.hashToRemoveOnAdd, executionResult.GetHeaderHash())
	if shouldIgnoreExecutionResult {
		est.hashToRemoveOnAdd = nil
		log.Debug("est.AddExecutionResult ignored execution result", "hash", hex.EncodeToString(executionResult.GetHeaderHash()))
		return nil
	}

	if est.lastNotarizedResult.GetHeaderNonce() >= executionResult.GetHeaderNonce() {
		return fmt.Errorf("%w nonce(%d) is lower than last notarized nonce(%d)", ErrWrongExecutionResultNonce, executionResult.GetHeaderNonce(), est.lastNotarizedResult.GetHeaderNonce())
	}

	lastExecutedResult, err := est.getLastExecutionResult()
	if err != nil {
		return err
	}

	last := lastExecutedResult.GetHeaderNonce()
	current := executionResult.GetHeaderNonce()
	isNextOrSameNonce := current == last || current == last+1
	if !isNextOrSameNonce {
		return fmt.Errorf("%w nonce(%d) should be equal to the subsequent nonce after last executed(%d)", ErrWrongExecutionResultNonce, executionResult.GetHeaderNonce(), lastExecutedResult.GetHeaderNonce())
	}

	est.executionResultsByHash[string(executionResult.GetHeaderHash())] = executionResult
	est.nonceHash.addNonceHash(executionResult.GetHeaderNonce(), string(executionResult.GetHeaderHash()))

	est.lastExecutedResultHash = executionResult.GetHeaderHash()

	return nil
}

func (est *executionResultsTracker) getLastExecutionResult() (data.ExecutionResultHandler, error) {
	if est.lastNotarizedResult == nil {
		return nil, ErrNilLastNotarizedExecutionResult
	}

	if est.lastExecutedResultHash == nil {
		return nil, fmt.Errorf("%w last executed result hash is not set", ErrCannotFindExecutionResult)
	}

	if bytes.Equal(est.lastExecutedResultHash, est.lastNotarizedResult.GetHeaderHash()) {
		return est.lastNotarizedResult, nil
	}

	lastExecutedResults, found := est.executionResultsByHash[string(est.lastExecutedResultHash)]
	if !found {
		return nil, fmt.Errorf("%w hash(%s)", ErrCannotFindExecutionResult, hex.EncodeToString(est.lastExecutedResultHash))
	}

	return lastExecutedResults, nil
}

// GetPendingExecutionResults will return the pending execution results
func (est *executionResultsTracker) GetPendingExecutionResults() ([]data.ExecutionResultHandler, error) {
	est.mutex.RLock()
	defer est.mutex.RUnlock()

	return est.getPendingExecutionResults()
}

func (est *executionResultsTracker) getPendingExecutionResults() ([]data.ExecutionResultHandler, error) {
	executionResults := make([]data.ExecutionResultHandler, 0, len(est.executionResultsByHash))
	for _, executionResult := range est.executionResultsByHash {
		executionResults = append(executionResults, executionResult)
	}

	if len(executionResults) == 0 {
		return executionResults, nil
	}

	sort.Slice(executionResults, func(i, j int) bool {
		return executionResults[i].GetHeaderNonce() < executionResults[j].GetHeaderNonce()
	})

	firstElementHasCorrectNonce := executionResults[0].GetHeaderNonce() == est.lastNotarizedResult.GetHeaderNonce()+1
	if !firstElementHasCorrectNonce {
		return nil, ErrDifferentNoncesConfirmedExecutionResults
	}

	for idx := 0; idx < len(executionResults)-1; idx++ {
		hasConsecutiveNonces := executionResults[idx].GetHeaderNonce()+1 == executionResults[idx+1].GetHeaderNonce()
		if !hasConsecutiveNonces {
			return nil, ErrDifferentNoncesConfirmedExecutionResults
		}
	}

	return executionResults, nil
}

// GetPendingExecutionResultByHash will return the execution results by hash
func (est *executionResultsTracker) GetPendingExecutionResultByHash(hash []byte) (data.ExecutionResultHandler, error) {
	est.mutex.RLock()
	defer est.mutex.RUnlock()

	result, found := est.executionResultsByHash[string(hash)]
	if !found {
		return nil, fmt.Errorf("%w with hash: '%s'", ErrCannotFindExecutionResult, hex.EncodeToString(hash))
	}

	return result, nil
}

// GetPendingExecutionResultByNonce will return the execution results by nonce
func (est *executionResultsTracker) GetPendingExecutionResultByNonce(nonce uint64) (data.ExecutionResultHandler, error) {
	est.mutex.RLock()
	defer est.mutex.RUnlock()

	return est.getPendingExecutionResultsByNonce(nonce)
}

func (est *executionResultsTracker) getPendingExecutionResultsByNonce(nonce uint64) (data.ExecutionResultHandler, error) {
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

	return est.cleanConfirmedExecutionResults(headerExecutionResults)
}

func (est *executionResultsTracker) cleanConfirmedExecutionResults(headerExecutionResults []data.ExecutionResultHandler) (*CleanInfo, error) {
	if len(headerExecutionResults) == 0 {
		return &CleanInfo{
			CleanResult:             CleanResultOK,
			LastMatchingResultNonce: est.lastNotarizedResult.GetHeaderNonce(),
		}, nil
	}

	pendingExecutionResult, err := est.getPendingExecutionResults()
	if err != nil {
		return nil, err
	}

	lastMatchingResultNonce := est.lastNotarizedResult.GetHeaderNonce()
	lastMatchingHash := est.lastNotarizedResult.GetHeaderHash()
	for idx, executionResultFromHeader := range headerExecutionResults {
		if idx >= len(pendingExecutionResult) {
			// missing execution result
			return &CleanInfo{
				CleanResult:             CleanResultNotFound,
				LastMatchingResultNonce: lastMatchingResultNonce,
			}, nil
		}

		executionResultFromTracker := pendingExecutionResult[idx]
		sameHash := string(executionResultFromTracker.GetHeaderHash()) == string(executionResultFromHeader.GetHeaderHash())
		sameNonce := executionResultFromTracker.GetHeaderNonce() == executionResultFromHeader.GetHeaderNonce()
		sameRound := executionResultFromTracker.GetHeaderRound() == executionResultFromHeader.GetHeaderRound()
		sameRootHash := string(executionResultFromTracker.GetRootHash()) == string(executionResultFromHeader.GetRootHash())
		areEqual := sameHash && sameNonce && sameRound && sameRootHash
		if !areEqual {
			est.lastExecutedResultHash = lastMatchingHash

			// different execution result should clean everything starting from this execution result and return CleanResultMismatch
			est.cleanExecutionResults(pendingExecutionResult[idx:])

			return &CleanInfo{
				CleanResult:             CleanResultMismatch,
				LastMatchingResultNonce: lastMatchingResultNonce,
			}, nil
		}

		lastMatchingResultNonce = executionResultFromHeader.GetHeaderNonce()
		lastMatchingHash = executionResultFromHeader.GetHeaderHash()
	}

	est.cleanExecutionResults(headerExecutionResults)
	est.lastNotarizedResult = headerExecutionResults[len(headerExecutionResults)-1]

	return &CleanInfo{
		CleanResult:             CleanResultOK,
		LastMatchingResultNonce: est.lastNotarizedResult.GetHeaderNonce(),
	}, nil
}

func (est *executionResultsTracker) cleanExecutionResults(executionResult []data.ExecutionResultHandler) {
	for _, result := range executionResult {
		delete(est.executionResultsByHash, string(result.GetHeaderHash()))
		est.nonceHash.removeByNonce(result.GetHeaderNonce())
	}
}

// GetLastNotarizedExecutionResult will return the last notarized execution result
func (est *executionResultsTracker) GetLastNotarizedExecutionResult() (data.ExecutionResultHandler, error) {
	est.mutex.RLock()
	defer est.mutex.RUnlock()
	if est.lastNotarizedResult == nil {
		return nil, ErrNilLastNotarizedExecutionResult
	}

	return est.lastNotarizedResult, nil
}

// SetLastNotarizedResult will set the last notarized execution result
func (est *executionResultsTracker) SetLastNotarizedResult(executionResult data.ExecutionResultHandler) error {
	if executionResult == nil {
		return ErrNilExecutionResult
	}

	est.mutex.Lock()
	est.lastNotarizedResult = executionResult
	est.lastExecutedResultHash = executionResult.GetHeaderHash()
	est.mutex.Unlock()

	return nil
}

// RemoveByHash will remove the execution result by header hash
func (est *executionResultsTracker) RemoveByHash(hash []byte) error {
	est.mutex.Lock()
	defer est.mutex.Unlock()

	executionResult, found := est.executionResultsByHash[string(hash)]
	if !found {
		est.hashToRemoveOnAdd = hash
		return nil
	}

	delete(est.executionResultsByHash, string(hash))
	est.nonceHash.removeByNonce(executionResult.GetHeaderNonce())

	pendingExecutionResult, err := est.getPendingExecutionResults()
	if err != nil {
		return err
	}

	if len(pendingExecutionResult) == 0 {
		// set last execution result with last notarized
		est.lastExecutedResultHash = est.lastNotarizedResult.GetHeaderHash()
		return nil
	}

	est.lastExecutedResultHash = pendingExecutionResult[len(pendingExecutionResult)-1].GetHeaderHash()
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (est *executionResultsTracker) IsInterfaceNil() bool {
	return est == nil
}
