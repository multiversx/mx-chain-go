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
	lastNotarizedResult    data.BaseExecutionResultHandler
	mutex                  sync.RWMutex
	executionResultsByHash map[string]data.BaseExecutionResultHandler
	nonceHash              *nonceHash
	lastExecutedResultHash []byte
	hashToRemoveOnAdd      []byte
}

// NewExecutionResultsTracker will create a new instance of *executionResultsTracker
func NewExecutionResultsTracker() *executionResultsTracker {
	return &executionResultsTracker{
		executionResultsByHash: make(map[string]data.BaseExecutionResultHandler),
		nonceHash:              newNonceHash(),
	}
}

// AddExecutionResult will add the provided execution result in tracker
// It will return true if the execution result was added in the tracker
func (ert *executionResultsTracker) AddExecutionResult(executionResult data.BaseExecutionResultHandler) error {
	if executionResult == nil {
		return ErrNilExecutionResult
	}

	ert.mutex.Lock()
	defer ert.mutex.Unlock()
	if ert.lastNotarizedResult == nil {
		return ErrNilLastNotarizedExecutionResult
	}

	shouldIgnoreExecutionResult := bytes.Equal(ert.hashToRemoveOnAdd, executionResult.GetHeaderHash())
	if shouldIgnoreExecutionResult {
		ert.hashToRemoveOnAdd = nil
		log.Debug("ert.AddExecutionResult ignored execution result", "hash", executionResult.GetHeaderHash())
		return nil
	}

	if ert.lastNotarizedResult.GetHeaderNonce() >= executionResult.GetHeaderNonce() {
		return fmt.Errorf("%w nonce(%d) is lower than last notarized nonce(%d)", ErrWrongExecutionResultNonce, executionResult.GetHeaderNonce(), ert.lastNotarizedResult.GetHeaderNonce())
	}

	lastExecutedResult, err := ert.getLastExecutionResult()
	if err != nil {
		return err
	}

	last := lastExecutedResult.GetHeaderNonce()
	current := executionResult.GetHeaderNonce()
	isNextOrSameNonce := current == last || current == last+1
	if !isNextOrSameNonce {
		return fmt.Errorf("%w nonce(%d) should be equal to the subsequent nonce after last executed(%d)", ErrWrongExecutionResultNonce, executionResult.GetHeaderNonce(), lastExecutedResult.GetHeaderNonce())
	}

	ert.executionResultsByHash[string(executionResult.GetHeaderHash())] = executionResult
	ert.nonceHash.addNonceHash(executionResult.GetHeaderNonce(), string(executionResult.GetHeaderHash()))

	ert.lastExecutedResultHash = executionResult.GetHeaderHash()

	return nil
}

func (ert *executionResultsTracker) getLastExecutionResult() (data.BaseExecutionResultHandler, error) {
	if ert.lastNotarizedResult == nil {
		return nil, ErrNilLastNotarizedExecutionResult
	}

	if ert.lastExecutedResultHash == nil {
		return nil, fmt.Errorf("%w last executed result hash is not set", ErrCannotFindExecutionResult)
	}

	if bytes.Equal(ert.lastExecutedResultHash, ert.lastNotarizedResult.GetHeaderHash()) {
		return ert.lastNotarizedResult, nil
	}

	lastExecutedResults, found := ert.executionResultsByHash[string(ert.lastExecutedResultHash)]
	if !found {
		return nil, fmt.Errorf("%w hash(%s)", ErrCannotFindExecutionResult, ert.lastExecutedResultHash)
	}

	return lastExecutedResults, nil
}

// GetPendingExecutionResults will return the pending execution results
func (ert *executionResultsTracker) GetPendingExecutionResults() ([]data.BaseExecutionResultHandler, error) {
	ert.mutex.RLock()
	defer ert.mutex.RUnlock()

	return ert.getPendingExecutionResults()
}

func (ert *executionResultsTracker) getPendingExecutionResults() ([]data.BaseExecutionResultHandler, error) {
	executionResults := make([]data.BaseExecutionResultHandler, 0, len(ert.executionResultsByHash))
	for _, executionResult := range ert.executionResultsByHash {
		executionResults = append(executionResults, executionResult)
	}

	if len(executionResults) == 0 {
		return executionResults, nil
	}

	sort.Slice(executionResults, func(i, j int) bool {
		return executionResults[i].GetHeaderNonce() < executionResults[j].GetHeaderNonce()
	})

	firstElementHasCorrectNonce := executionResults[0].GetHeaderNonce() == ert.lastNotarizedResult.GetHeaderNonce()+1
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
func (ert *executionResultsTracker) GetPendingExecutionResultByHash(hash []byte) (data.BaseExecutionResultHandler, error) {
	ert.mutex.RLock()
	defer ert.mutex.RUnlock()

	result, found := ert.executionResultsByHash[string(hash)]
	if !found {
		return nil, fmt.Errorf("%w with hash: '%s'", ErrCannotFindExecutionResult, hex.EncodeToString(hash))
	}

	return result, nil
}

// GetPendingExecutionResultByNonce will return the execution results by nonce
func (ert *executionResultsTracker) GetPendingExecutionResultByNonce(nonce uint64) (data.BaseExecutionResultHandler, error) {
	ert.mutex.RLock()
	defer ert.mutex.RUnlock()

	return ert.getPendingExecutionResultsByNonce(nonce)
}

func (ert *executionResultsTracker) getPendingExecutionResultsByNonce(nonce uint64) (data.BaseExecutionResultHandler, error) {
	hash := ert.nonceHash.getHashByNonce(nonce)
	result, found := ert.executionResultsByHash[hash]
	if !found {
		return nil, fmt.Errorf("%w with hash: '%s'", ErrCannotFindExecutionResult, hex.EncodeToString([]byte(hash)))
	}

	return result, nil
}

// CleanConfirmedExecutionResults will clean the confirmed execution results
func (ert *executionResultsTracker) CleanConfirmedExecutionResults(header data.HeaderHandler) error {
	ert.mutex.Lock()
	defer ert.mutex.Unlock()

	headerExecutionResults := header.GetExecutionResultsHandlers()

	return ert.cleanConfirmedExecutionResults(headerExecutionResults)
}

func (ert *executionResultsTracker) cleanConfirmedExecutionResults(headerExecutionResults []data.BaseExecutionResultHandler) error {
	if len(headerExecutionResults) == 0 {
		return nil
	}

	pendingExecutionResult, err := ert.getPendingExecutionResults()
	if err != nil {
		return err
	}

	lastMatchingResultNonce := ert.lastNotarizedResult.GetHeaderNonce()
	lastMatchingHash := ert.lastNotarizedResult.GetHeaderHash()
	for idx, executionResultFromHeader := range headerExecutionResults {
		if idx >= len(pendingExecutionResult) {
			// missing execution result
			return fmt.Errorf("%w, executon result nod found in peding executon results, last matching result nonce: %d", ErrCannotFindExecutionResult, lastMatchingResultNonce)
		}

		executionResultFromTracker := pendingExecutionResult[idx]
		sameHash := string(executionResultFromTracker.GetHeaderHash()) == string(executionResultFromHeader.GetHeaderHash())
		sameNonce := executionResultFromTracker.GetHeaderNonce() == executionResultFromHeader.GetHeaderNonce()
		sameRound := executionResultFromTracker.GetHeaderRound() == executionResultFromHeader.GetHeaderRound()
		sameRootHash := string(executionResultFromTracker.GetRootHash()) == string(executionResultFromHeader.GetRootHash())
		areEqual := sameHash && sameNonce && sameRound && sameRootHash
		if !areEqual {
			ert.lastExecutedResultHash = lastMatchingHash

			// different execution result should clean everything starting from this execution result and return CleanResultMismatch
			ert.cleanExecutionResults(pendingExecutionResult[idx:])

			return fmt.Errorf("%w, last matching result nonce: %d", ErrExecutionResultMissmatch, lastMatchingResultNonce)
		}

		lastMatchingResultNonce = executionResultFromHeader.GetHeaderNonce()
		lastMatchingHash = executionResultFromHeader.GetHeaderHash()
	}

	ert.cleanExecutionResults(headerExecutionResults)
	ert.lastNotarizedResult = headerExecutionResults[len(headerExecutionResults)-1]

	return nil
}

func (ert *executionResultsTracker) cleanExecutionResults(executionResult []data.BaseExecutionResultHandler) {
	for _, result := range executionResult {
		delete(ert.executionResultsByHash, string(result.GetHeaderHash()))
		ert.nonceHash.removeByNonce(result.GetHeaderNonce())
	}
}

// GetLastNotarizedExecutionResult will return the last notarized execution result
func (ert *executionResultsTracker) GetLastNotarizedExecutionResult() (data.BaseExecutionResultHandler, error) {
	ert.mutex.RLock()
	defer ert.mutex.RUnlock()
	if ert.lastNotarizedResult == nil {
		return nil, ErrNilLastNotarizedExecutionResult
	}

	return ert.lastNotarizedResult, nil
}

// SetLastNotarizedResult will set the last notarized execution result
func (ert *executionResultsTracker) SetLastNotarizedResult(executionResult data.BaseExecutionResultHandler) error {
	if executionResult == nil {
		return ErrNilExecutionResult
	}

	ert.mutex.Lock()
	ert.lastNotarizedResult = executionResult
	ert.lastExecutedResultHash = executionResult.GetHeaderHash()
	ert.mutex.Unlock()

	return nil
}

// RemoveByHash will remove the execution result by header hash
func (ert *executionResultsTracker) RemoveByHash(hash []byte) error {
	ert.mutex.Lock()
	defer ert.mutex.Unlock()

	executionResult, found := ert.executionResultsByHash[string(hash)]
	if !found {
		ert.hashToRemoveOnAdd = hash
		return nil
	}

	pendingExecutionResult, err := ert.getPendingExecutionResults()
	if err != nil {
		return err
	}

	// check that the provided is the last execution result
	lastElement := pendingExecutionResult[len(pendingExecutionResult)-1]
	if !bytes.Equal(lastElement.GetHeaderHash(), hash) {
		return fmt.Errorf("%w between last pending execution result hash and the provided hash", ErrHashMismatch)
	}

	delete(ert.executionResultsByHash, string(hash))
	ert.nonceHash.removeByNonce(executionResult.GetHeaderNonce())

	// remove the last element
	pendingExecutionResult = pendingExecutionResult[:len(pendingExecutionResult)-1]

	if len(pendingExecutionResult) == 0 {
		// set last execution result with last notarized
		ert.lastExecutedResultHash = ert.lastNotarizedResult.GetHeaderHash()
		return nil
	}

	ert.lastExecutedResultHash = pendingExecutionResult[len(pendingExecutionResult)-1].GetHeaderHash()
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ert *executionResultsTracker) IsInterfaceNil() bool {
	return ert == nil
}
