package executionManager

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/disabled"
	"github.com/multiversx/mx-chain-go/sharding"

	"github.com/multiversx/mx-chain-go/process/asyncExecution/cache"
)

var log = logger.GetOrCreate("process/asyncExecution/executionManager")

// ArgsExecutionManager holds all the components needed to create a new instance of executionManager
type ArgsExecutionManager struct {
	BlocksQueue             process.BlocksCache
	ExecutionResultsTracker process.ExecutionResultsTracker
	BlockChain              data.ChainHandler
	Headers                 dataRetriever.HeadersPool
	StorageService          dataRetriever.StorageService
	Marshaller              marshal.Marshalizer
	ShardCoordinator        sharding.Coordinator
}

type executionManager struct {
	mut                     sync.RWMutex
	headersExecutor         process.HeadersExecutor
	blocksCache             process.BlocksCache
	executionResultsTracker process.ExecutionResultsTracker
	blockChain              data.ChainHandler
	headers                 dataRetriever.HeadersPool
	storageService          dataRetriever.StorageService
	marshaller              marshal.Marshalizer
	shardCoordinator        sharding.Coordinator
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
	if check.IfNil(args.StorageService) {
		return nil, process.ErrNilStorage
	}
	if check.IfNil(args.Marshaller) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	instance := &executionManager{
		headersExecutor:         disabled.NewHeadersExecutor(),
		blocksCache:             args.BlocksQueue,
		executionResultsTracker: args.ExecutionResultsTracker,
		blockChain:              args.BlockChain,
		headers:                 args.Headers,
		storageService:          args.StorageService,
		marshaller:              args.Marshaller,
		shardCoordinator:        args.ShardCoordinator,
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
func (em *executionManager) AddPairForExecution(pair cache.HeaderBodyPair) error {
	// lock the internal mutex to avoid any concurrent removal requests
	em.mut.Lock()
	defer em.mut.Unlock()

	lastExecutedBlock := em.blockChain.GetLastExecutedBlockHeader()
	if !check.IfNil(lastExecutedBlock) &&
		lastExecutedBlock.GetNonce() >= pair.Header.GetNonce() {
		err := process.UpdateContextForReplacedHeader(
			pair.Header,
			em,
			em.blockChain,
			em.headers,
			em.storageService,
			em.marshaller,
			em.shardCoordinator.SelfId(),
		)
		if err != nil {
			return err
		}
	}

	return em.blocksCache.AddOrReplace(pair)
}

// GetPendingExecutionResults calls the same method from executionResultsTracker
func (em *executionManager) GetPendingExecutionResults() ([]data.BaseExecutionResultHandler, error) {
	return em.executionResultsTracker.GetPendingExecutionResults()
}

// GetLastNotarizedExecutionResult will return the last notarized execution result
func (em *executionManager) GetLastNotarizedExecutionResult() (data.BaseExecutionResultHandler, error) {
	return em.executionResultsTracker.GetLastNotarizedExecutionResult()
}

// SetLastNotarizedResult calls the same method from executionResultsTracker
func (em *executionManager) SetLastNotarizedResult(executionResult data.BaseExecutionResultHandler) error {
	return em.executionResultsTracker.SetLastNotarizedResult(executionResult)
}

// CleanConfirmedExecutionResults calls the same method from executionResultsTracker
func (em *executionManager) CleanConfirmedExecutionResults(header data.HeaderHandler) error {
	for _, executionResult := range header.GetExecutionResultsHandlers() {
		em.blocksCache.Remove(executionResult.GetHeaderNonce())
	}

	return em.executionResultsTracker.CleanConfirmedExecutionResults(header)
}

// CleanOnConsensusReached calls the same method from executionResultsTracker
func (em *executionManager) CleanOnConsensusReached(headerHash []byte, headerNonce uint64) {
	em.executionResultsTracker.CleanOnConsensusReached(headerHash, headerNonce)
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
	removedNonces := em.blocksCache.RemoveAtNonceAndHigher(nonceToRemove)
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

// RemovePendingExecutionResultsFromNonce will remove the execution result with the provided nonce and all execution results with higher nonces
func (em *executionManager) RemovePendingExecutionResultsFromNonce(nonce uint64) error {
	return em.executionResultsTracker.RemoveFromNonce(nonce)
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

	em.blocksCache.Clean()

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

	header, err := process.GetHeader(lastExecutedHeaderHash, em.headers, em.storageService, em.marshaller, em.shardCoordinator.SelfId())
	if err != nil {
		log.Debug("executionmanager.updateBlockchainAfterRemoval: could not find header in pool or storage",
			"hash", lastExecutedHeaderHash,
			"nonce", lastExecutedHeaderNonce,
			"error", err,
		)
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

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (em *executionManager) IsInterfaceNil() bool {
	return em == nil
}
