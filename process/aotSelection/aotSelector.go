package aotSelection

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/cache"
)

var log = logger.GetOrCreate("process/aotSelection")

const (
	defaultSelectionTimeout          = 150 * time.Millisecond
	defaultMaxTxsPerBlock            = 30000
	defaultLoopDurationCheckInterval = 100
	defaultCacheSize                 = 10
)

// AOTSelectorArgs holds the arguments needed to create an AOT selector
type AOTSelectorArgs struct {
	NodesCoordinator     nodesCoordinator.NodesCoordinator
	ShardCoordinator     sharding.Coordinator
	KeysHandler          consensus.KeysHandler
	NodeRedundancy       consensus.NodeRedundancyHandler
	TxCache              preprocess.TxCache
	AccountsAdapter      state.AccountsAdapter
	TransactionProcessor process.TransactionProcessor
	TxVersionChecker     process.TxVersionCheckerHandler
	BlockChain           data.ChainHandler
	EconomicsDataHandler process.EconomicsDataHandler

	// Configuration
	SelectionTimeout time.Duration
	CacheSize        int
	MaxGasPerBlock   uint64
	MaxTxsPerBlock   int
}

type aotSelector struct {
	nodesCoordinator     nodesCoordinator.NodesCoordinator
	shardCoordinator     sharding.Coordinator
	keysHandler          consensus.KeysHandler
	nodeRedundancy       consensus.NodeRedundancyHandler
	txCache              preprocess.TxCache
	accountsAdapter      state.AccountsAdapter
	transactionProcessor process.TransactionProcessor
	txVersionChecker     process.TxVersionCheckerHandler
	blockChain           data.ChainHandler
	economicsDataHandler process.EconomicsDataHandler

	cache            storage.Cacher
	selectionTimeout time.Duration
	maxGasPerBlock   uint64
	maxTxsPerBlock   int

	// Synchronization
	cancelChan   chan struct{}
	selectionMut sync.Mutex
	ongoingNonce uint64
	resultChan   chan *process.AOTSelectionResult
	lastEpoch    uint32
}

// NewAOTSelector creates a new AOT selector instance
func NewAOTSelector(args AOTSelectorArgs) (*aotSelector, error) {
	if check.IfNil(args.NodesCoordinator) {
		return nil, process.ErrNilNodesCoordinator
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.KeysHandler) {
		return nil, process.ErrNilKeysHandler
	}
	if check.IfNil(args.NodeRedundancy) {
		return nil, process.ErrNilNodeRedundancyHandler
	}
	if check.IfNil(args.TxCache) {
		return nil, process.ErrNilTxCache
	}
	if check.IfNil(args.AccountsAdapter) {
		return nil, process.ErrNilAccountsAdapter
	}
	if check.IfNil(args.TransactionProcessor) {
		return nil, process.ErrNilTxProcessor
	}
	if check.IfNil(args.TxVersionChecker) {
		return nil, process.ErrNilTransactionVersionChecker
	}
	if check.IfNil(args.BlockChain) {
		return nil, process.ErrNilBlockChain
	}
	if check.IfNil(args.EconomicsDataHandler) {
		return nil, process.ErrNilEconomicsData
	}

	selectionTimeout := args.SelectionTimeout
	if selectionTimeout <= 0 {
		selectionTimeout = defaultSelectionTimeout
	}

	cacheSize := args.CacheSize
	if cacheSize <= 0 {
		cacheSize = defaultCacheSize
	}

	maxTxsPerBlock := args.MaxTxsPerBlock
	if maxTxsPerBlock <= 0 {
		maxTxsPerBlock = defaultMaxTxsPerBlock
	}

	// Use existing LRU cache from storage/cache
	lruCache, err := cache.NewLRUCache(cacheSize)
	if err != nil {
		return nil, err
	}

	return &aotSelector{
		nodesCoordinator:     args.NodesCoordinator,
		shardCoordinator:     args.ShardCoordinator,
		keysHandler:          args.KeysHandler,
		nodeRedundancy:       args.NodeRedundancy,
		txCache:              args.TxCache,
		accountsAdapter:      args.AccountsAdapter,
		transactionProcessor: args.TransactionProcessor,
		txVersionChecker:     args.TxVersionChecker,
		blockChain:           args.BlockChain,
		economicsDataHandler: args.EconomicsDataHandler,
		cache:                lruCache,
		selectionTimeout:     selectionTimeout,
		maxGasPerBlock:       args.MaxGasPerBlock,
		maxTxsPerBlock:       maxTxsPerBlock,
	}, nil
}

// TriggerAOTSelection triggers AOT for the next block after commit
// Checks if node is leader in round N+1 OR N+2 (if N+1 fails)
func (s *aotSelector) TriggerAOTSelection(committedHeader data.HeaderHandler, currentRound uint64) {
	if check.IfNil(committedHeader) {
		log.Debug("TriggerAOTSelection: nil header, skipping")
		return
	}

	// Only process for shard nodes (metachain does not select transactions)
	if s.shardCoordinator.SelfId() == core.MetachainShardId {
		log.Trace("TriggerAOTSelection: metachain node, skipping")
		return
	}

	epoch := committedHeader.GetEpoch()
	// Detect epoch change and invalidate cache
	s.selectionMut.Lock()
	previousEpoch := s.lastEpoch
	epochChanged := s.lastEpoch != epoch && s.lastEpoch != 0
	s.lastEpoch = epoch
	if epochChanged {
		s.cache.Clear()
		if s.cancelChan != nil {
			close(s.cancelChan)
			s.cancelChan = nil
		}
	}
	s.selectionMut.Unlock()

	if epochChanged {
		log.Debug("TriggerAOTSelection: epoch changed, cleared cache and cancelled ongoing selection",
			"previousEpoch", previousEpoch,
			"newEpoch", epoch)
	}

	randomness := committedHeader.GetRandSeed()
	nextNonce := committedHeader.GetNonce() + 1
	nextRound := currentRound + 1

	// Check if we could be leader in either of the next two rounds
	isLeaderNextRound := s.isSelfLeaderForRound(randomness, nextRound, epoch)
	isLeaderRoundAfter := s.isSelfLeaderForRound(randomness, nextRound+1, epoch)

	if !isLeaderNextRound && !isLeaderRoundAfter {
		log.Trace("TriggerAOTSelection: not a potential leader, skipping",
			"nextRound", nextRound,
			"nextNonce", nextNonce)
		return
	}

	log.Debug("TriggerAOTSelection: starting AOT selection",
		"nextNonce", nextNonce,
		"nextRound", nextRound,
		"isLeaderNextRound", isLeaderNextRound,
		"isLeaderRoundAfter", isLeaderRoundAfter)

	// Spawn goroutine for background selection
	go s.runAOTSelection(nextNonce, randomness)
}

// isSelfLeaderForRound checks if self is leader for a specific round using nodesCoordinator directly
func (s *aotSelector) isSelfLeaderForRound(randomness []byte, round uint64, epoch uint32) bool {
	shardId := s.shardCoordinator.SelfId()

	// Use nodesCoordinator directly - no separate LeaderPredictor
	leader, _, err := s.nodesCoordinator.ComputeConsensusGroup(randomness, round, shardId, epoch)
	if err != nil {
		log.Trace("isSelfLeaderForRound: ComputeConsensusGroup failed", "error", err)
		return false
	}

	// Check single key
	selfPubKey := s.keysHandler.GetHandledPrivateKey(leader.PubKey())
	isSelfLeader := selfPubKey != nil && s.shouldConsiderSelfKeyInConsensus()

	// Check multi-key
	isMultiKeyLeader := s.keysHandler.IsKeyManagedByCurrentNode(leader.PubKey())

	return isSelfLeader || isMultiKeyLeader
}

// shouldConsiderSelfKeyInConsensus checks if the node should consider its self key in consensus
// Returns true if not a redundancy node, or if the main machine is not active
func (s *aotSelector) shouldConsiderSelfKeyInConsensus() bool {
	if check.IfNil(s.nodeRedundancy) {
		log.Warn("shouldConsiderSelfKeyInConsensus: redundancy handler is nil")
		return false
	}

	isMainMachine := !s.nodeRedundancy.IsRedundancyNode()
	if isMainMachine {
		return true
	}
	return !s.nodeRedundancy.IsMainMachineActive()
}

// GetPreSelectedTransactions retrieves cached selection for the given nonce
// If selection is ongoing, WAITS for completion (we need the result anyway)
func (s *aotSelector) GetPreSelectedTransactions(blockNonce uint64) (*process.AOTSelectionResult, bool) {
	s.selectionMut.Lock()
	if s.ongoingNonce == blockNonce && s.resultChan != nil {
		resultChan := s.resultChan
		s.selectionMut.Unlock()
		// Wait for ongoing selection to complete
		result := <-resultChan
		if result != nil {
			return result, true
		}
		return nil, false
	}
	s.selectionMut.Unlock()

	// Check cache
	cacheKey := []byte(fmt.Sprintf("%d", blockNonce))
	if val, ok := s.cache.Get(cacheKey); ok {
		if result, isResult := val.(*process.AOTSelectionResult); isResult {
			return result, true
		}
	}
	return nil, false
}

// CancelOngoingSelection cancels any ongoing AOT selection
// Called before OnProposed/OnExecuted to avoid conflicts
func (s *aotSelector) CancelOngoingSelection() {
	s.selectionMut.Lock()
	defer s.selectionMut.Unlock()
	if s.cancelChan != nil {
		close(s.cancelChan)
		s.cancelChan = nil
	}
}

// runAOTSelection performs the actual transaction selection in a background goroutine
func (s *aotSelector) runAOTSelection(targetNonce uint64, randomness []byte) {
	s.selectionMut.Lock()
	s.cancelChan = make(chan struct{})
	s.resultChan = make(chan *process.AOTSelectionResult, 1)
	s.ongoingNonce = targetNonce
	cancelChan := s.cancelChan
	resultChan := s.resultChan
	s.selectionMut.Unlock()

	defer func() {
		s.selectionMut.Lock()
		s.ongoingNonce = 0
		s.resultChan = nil
		s.cancelChan = nil
		s.selectionMut.Unlock()
		close(resultChan)
		close(cancelChan)
	}()

	startTime := time.Now()

	log.Debug("runAOTSelection: starting", "targetNonce", targetNonce)

	// Create selection session
	session, err := preprocess.NewSelectionSession(preprocess.ArgsSelectionSession{
		AccountsAdapter:         s.accountsAdapter,
		TransactionsProcessor:   s.transactionProcessor,
		TxVersionCheckerHandler: s.txVersionChecker,
	})
	if err != nil {
		log.Debug("runAOTSelection: failed to create selection session", "error", err)
		resultChan <- nil
		return
	}

	// Get gas limit
	maxGas := s.maxGasPerBlock
	if maxGas == 0 {
		maxGas = s.economicsDataHandler.MaxGasLimitPerBlock(s.shardCoordinator.SelfId())
	}

	timeout := s.selectionTimeout
	options, err := holders.NewTxSelectionOptions(
		maxGas,
		s.maxTxsPerBlock,
		defaultLoopDurationCheckInterval,
		func() bool {
			if time.Since(startTime) >= timeout {
				return false
			}
			select {
			case <-cancelChan:
				return false
			default:
				return true
			}
		},
	)
	if err != nil {
		log.Debug("runAOTSelection: failed to create selection options", "error", err)
		resultChan <- nil
		return
	}

	// Perform simulated selection
	wrappedTxs, accumulatedGas, err := s.txCache.SimulateSelectTransactions(session, options)
	if err != nil {
		log.Debug("runAOTSelection: selection failed", "error", err)
		resultChan <- nil
		return
	}

	// Check if cancelled
	select {
	case <-cancelChan:
		log.Debug("runAOTSelection: cancelled", "targetNonce", targetNonce)
		resultChan <- nil
		return
	default:
	}

	// Extract transaction hashes
	txHashes := make([][]byte, len(wrappedTxs))
	for i, wrappedTx := range wrappedTxs {
		txHashes[i] = bytes.Clone(wrappedTx.TxHash)
	}

	result := &process.AOTSelectionResult{
		TxHashes:            txHashes,
		GasProvided:         accumulatedGas,
		PredictedBlockNonce: targetNonce,
		Randomness:          bytes.Clone(randomness),
		SelectionTimestamp:  time.Now(),
	}

	// Store in cache
	cacheKey := []byte(fmt.Sprintf("%d", targetNonce))
	s.cache.Put(cacheKey, result, 1)

	log.Info("runAOTSelection: completed",
		"targetNonce", targetNonce,
		"numTxs", len(txHashes),
		"gasProvided", accumulatedGas,
		"duration", time.Since(startTime))

	resultChan <- result
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *aotSelector) IsInterfaceNil() bool {
	return s == nil
}
