package executionManager

import (
	"bytes"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/disabled"

	"github.com/multiversx/mx-chain-go/process/asyncExecution/queue"
)

var log = logger.GetOrCreate("process/asyncExecution/executionManager")

// ArgsExecutionManager holds all the components needed to create a new instance of executionManager
type ArgsExecutionManager struct {
	BlocksQueue             process.BlocksQueue
	ExecutionResultsTracker process.ExecutionResultsTracker
	BlockChain              data.ChainHandler
	Headers                 common.HeadersPool
}

type executionManager struct {
	mut                     sync.RWMutex
	headersExecutor         process.HeadersExecutor
	blocksQueue             process.BlocksQueue
	executionResultsTracker process.ExecutionResultsTracker
	blockChain              data.ChainHandler
	headers                 common.HeadersPool
}

// NewExecutionManager creates a new instance of executionManager
func NewExecutionManager(args ArgsExecutionManager) (*executionManager, error) {
	if check.IfNil(args.BlocksQueue) {
		return nil, ErrNilBlocksQueue
	}
	if check.IfNil(args.ExecutionResultsTracker) {
		return nil, ErrNilExecutionResultsTracker
	}
	if check.IfNil(args.BlockChain) {
		return nil, ErrNilBlockchain
	}
	if check.IfNil(args.Headers) {
		return nil, ErrNilHeadersPool
	}

	instance := &executionManager{
		headersExecutor:         disabled.NewHeadersExecutor(),
		blocksQueue:             args.BlocksQueue,
		executionResultsTracker: args.ExecutionResultsTracker,
		blockChain:              args.BlockChain,
		headers:                 args.Headers,
	}

	return instance, nil
}

// StartExecution starts the executor to begin processing blocks from the queue
func (em *executionManager) StartExecution() {
	em.mut.Lock()
	defer em.mut.Unlock()

	log.Debug("starting headers execution...")
	em.headersExecutor.StartExecution()
}

// SetHeadersExecutor sets the headers executor
func (em *executionManager) SetHeadersExecutor(executor process.HeadersExecutor) error {
	if check.IfNil(executor) {
		return ErrNilHeadersExecutor
	}

	em.mut.Lock()
	defer em.mut.Unlock()

	em.headersExecutor = executor

	return nil
}

// AddPairForExecution adds or replaces a header-body pair in the blocks queue
func (em *executionManager) AddPairForExecution(pair queue.HeaderBodyPair) error {
	// lock the internal mutex to avoid any concurrent removal requests
	em.mut.Lock()
	defer em.mut.Unlock()

	lastExecutedBlock := em.blockChain.GetLastExecutedBlockHeader()
	if areSameHeaders(pair.Header, lastExecutedBlock) {
		log.Warn("header already executed", "nonce", pair.Header.GetNonce(), "round", pair.Header.GetRound())
		return nil
	}
	// todo: remove pending execution result on same nonce if any
	// todo: make sure the lastExecutedBlockHeader from blockchain is only set in blockchain if
	// the block has passed consensus (was committed)
	return em.blocksQueue.AddOrReplace(pair)
}

func areSameHeaders(header1, header2 data.HeaderHandler) bool {
	if check.IfNil(header1) || check.IfNil(header2) {
		return false
	}

	sameNonce := header1.GetNonce() == header2.GetNonce()
	sameRound := header1.GetRound() == header2.GetRound()
	samePreviousHash := bytes.Equal(header1.GetPrevHash(), header2.GetPrevHash())

	return sameNonce && sameRound && samePreviousHash
}

// GetPendingExecutionResults calls the same method from executionResultsTracker
func (em *executionManager) GetPendingExecutionResults() ([]data.BaseExecutionResultHandler, error) {
	return em.executionResultsTracker.GetPendingExecutionResults()
}

// SetLastNotarizedResult calls the same method from executionResultsTracker
func (em *executionManager) SetLastNotarizedResult(executionResult data.BaseExecutionResultHandler) error {
	return em.executionResultsTracker.SetLastNotarizedResult(executionResult)
}

// CleanConfirmedExecutionResults calls the same method from executionResultsTracker
func (em *executionManager) CleanConfirmedExecutionResults(header data.HeaderHandler) error {
	return em.executionResultsTracker.CleanConfirmedExecutionResults(header)
}

// RemoveAtNonceAndHigher removes the header-body pair at the specified nonce
// and all pairs with higher nonces from the queue.
// if anything was processed after that nonce, remove the execution results
// and rollback blockchain state
func (em *executionManager) RemoveAtNonceAndHigher(nonce uint64) error {
	em.mut.Lock()
	defer em.mut.Unlock()

	lastNotarizedResult, err := em.executionResultsTracker.GetLastNotarizedExecutionResult()
	if err != nil {
		return err
	}

	lastNotarizedNonce := lastNotarizedResult.GetHeaderNonce()
	nonceToRemove := nonce
	if nonce <= lastNotarizedNonce {
		nonceToRemove = lastNotarizedNonce + 1
		log.Debug("attempting to remove nonce lower than last notarized execution result, updating it",
			"nonce provided", nonce,
			"last notarized nonce", lastNotarizedNonce,
			"removing from nonce", nonceToRemove,
		)
	}

	// first pause the headers executor which will block it from popping new headers
	// but allow it to finish anything currently processing
	em.headersExecutor.PauseExecution()

	// remove from queue
	removedNonces := em.blocksQueue.RemoveAtNonceAndHigher(nonceToRemove)
	if len(removedNonces) > 0 && removedNonces[0] == nonceToRemove {
		// if the first nonce removed is the initial one,
		// it means it was still in queue and was not processed.
		// no matter how many were removed, safe to resume execution
		em.headersExecutor.ResumeExecution()

		return nil
	}

	// if the initial nonce was not returned as removed from the queue,
	// it means that it was already popped for execution (and perhaps not the only one).
	// inform executionResultsTracker to remove all nonces >= than the provided one
	err = em.executionResultsTracker.RemoveFromNonce(nonceToRemove)
	if err != nil {
		return err
	}

	// update blockchain with the last executed header, similar to headersExecution
	err = em.updateBlockchainAfterRemoval(lastNotarizedResult)
	if err != nil {
		return err
	}

	// resume execution
	em.headersExecutor.ResumeExecution()

	return nil
}

// ResetAndResumeExecution resets the managed components to the last notarized result and resumes execution
func (em *executionManager) ResetAndResumeExecution(lastNotarizedResult data.BaseExecutionResultHandler) error {
	if check.IfNil(lastNotarizedResult) {
		return process.ErrNilLastExecutionResultHandler
	}

	em.mut.Lock()
	defer em.mut.Unlock()

	// even though the headers executor might already be paused, safe to try it one more time
	em.headersExecutor.PauseExecution()

	em.executionResultsTracker.Clean(lastNotarizedResult)

	lastNotarizedNonce := lastNotarizedResult.GetHeaderNonce()
	em.blocksQueue.Clean(lastNotarizedNonce)

	em.headersExecutor.ResumeExecution()

	return nil
}

func (em *executionManager) updateBlockchainAfterRemoval(lastNotarizedResult data.BaseExecutionResultHandler) error {
	lastExecutedHeaderHash := lastNotarizedResult.GetHeaderHash()
	lastExecutedHeaderNonce := lastNotarizedResult.GetHeaderNonce()
	lastExecutedHeaderRootHash := lastNotarizedResult.GetRootHash()
	pendingExecutionResults, err := em.executionResultsTracker.GetPendingExecutionResults()
	if err != nil {
		return err
	}

	// if there are still pending execution results, use the last one that was executed
	lastExecutionResult := lastNotarizedResult
	if len(pendingExecutionResults) > 0 {
		lastPending := pendingExecutionResults[len(pendingExecutionResults)-1]
		lastExecutedHeaderHash = lastPending.GetHeaderHash()
		lastExecutedHeaderNonce = lastPending.GetHeaderNonce()
		lastExecutedHeaderRootHash = lastPending.GetRootHash()

		lastExecutionResult = lastPending
	}

	header, err := em.headers.GetHeaderByHash(lastExecutedHeaderHash)
	if err != nil {
		return err
	}

	// update blockchain
	em.blockChain.SetFinalBlockInfo(
		lastExecutedHeaderNonce,
		lastExecutedHeaderHash,
		lastExecutedHeaderRootHash,
	)

	em.blockChain.SetLastExecutedBlockHeaderAndRootHash(header, lastExecutedHeaderHash, lastExecutedHeaderRootHash)
	em.blockChain.SetLastExecutionResult(lastExecutionResult)

	return nil
}

// Close closes the execution manager and all its components
func (em *executionManager) Close() error {
	log.Debug("closing execution manager")

	err := em.headersExecutor.Close()
	if err != nil {
		log.Warn("executionManager.Close - failed to close headers executor", "error", err)
	}

	em.blocksQueue.Close()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (em *executionManager) IsInterfaceNil() bool {
	return em == nil
}
