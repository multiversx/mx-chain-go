package sync

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/closing"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/cache"
	"github.com/multiversx/mx-chain-go/update"
	updateSync "github.com/multiversx/mx-chain-go/update/sync"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dblookupext"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/sync/storageBootstrap/metricsLoader"
	"github.com/multiversx/mx-chain-go/process/sync/trieIterators"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/trie/storageMarker"
)

var log = logger.GetOrCreate("process/sync")

type currentBlockTxProtector interface {
	ProtectSetOfDataAgainstEvictionForCurrentBlock(keys [][]byte, cacheID string) func()
}

type txSizeHandler interface {
	Size() int
}

type notarizedHeaderSelector interface {
	getNotarizedHeaderSelection(nonce uint64) notarizedHeaderSelection
	getHeaderVersion(nonce uint64, hash []byte) (bool, bool)
}

type notarizedHeaderAuthority interface {
	notarizedHeaderSelector
	hasUnresolvedNotarizedAmbiguity() bool
	getLowestAmbiguousNotarizedHeaderSelection() (notarizedHeaderSelection, bool)
	applyNotarizedHeaderSelection(nonce uint64, selectedHash []byte) notarizedHeaderResolution
	getProcessedHeaderHash(nonce uint64) []byte
}

var _ closing.Closer = (*baseBootstrap)(nil)

// sleepTime defines the time in milliseconds between each iteration made in syncBlocks method
const sleepTime = 5 * time.Millisecond
const sleepTimeOnFail = 400 * time.Millisecond
const minimumProcessWaitTime = time.Millisecond * 100
const defaultTimeToWaitForRequestedData = 5 * time.Minute

// defaultExecutionResultsRecoveryCooldown is the minimum time between two recovery attempts for
// the same diverging execution result nonce
const defaultExecutionResultsRecoveryCooldown = time.Minute

// maxFetchFailuresBeforeDroppingUnprovenHeader bounds how many consecutive failed sync iterations
// an unproven header may sit in the pool before being dropped for a re-request; counted only while
// the header is present, so a re-fetched header always gets the full window for its proof
const maxFetchFailuresBeforeDroppingUnprovenHeader = 10

// hdrInfo hold the data related to a header
type hdrInfo struct {
	Nonce uint64
	Hash  []byte
}

type notarizedInfo struct {
	lastNotarized           map[uint32]*hdrInfo
	finalNotarized          map[uint32]*hdrInfo
	blockWithLastNotarized  map[uint32]uint64
	blockWithFinalNotarized map[uint32]uint64
	startNonce              uint64
}

type nonceRecoveryInfo struct {
	numAttempts uint32
	lastAttempt time.Time
}

type reconcileEvidence struct {
	nonce               uint64
	localHash           []byte
	competitorHash      []byte
	lastEvaluatedRound  int64
	scanCursor          uint64
	selectedByAuthority bool
}

type ambiguityRecoveryState struct {
	nonce              uint64
	scanCursor         uint64
	lastEvaluatedRound int64
}

type baseBootstrap struct {
	historyRepo dblookupext.HistoryRepository
	headers     dataRetriever.HeadersPool
	proofs      dataRetriever.ProofsPool
	dataPool    dataRetriever.PoolsHolder

	chainHandler     data.ChainHandler
	blockProcessor   process.BlockProcessor
	executionManager process.ExecutionManager
	store            dataRetriever.StorageService

	roundHandler        consensus.RoundHandler
	hasher              hashing.Hasher
	marshalizer         marshal.Marshalizer
	epochHandler        dataRetriever.EpochHandler
	forkDetector        process.ForkDetector
	requestHandler      process.RequestHandler
	shardCoordinator    sharding.Coordinator
	accounts            state.AccountsAdapter
	blockBootstrapper   blockBootstrapper
	settlementChecker   settlementChecker
	blackListHandler    process.TimeCacher
	enableEpochsHandler common.EnableEpochsHandler
	enableRoundsHandler common.EnableRoundsHandler

	mutHeader     sync.RWMutex
	headerNonce   *uint64
	headerhash    []byte
	chRcvHdrNonce chan bool
	chRcvHdrHash  chan bool

	requestedHashes process.RequiredDataPool

	statusHandler core.AppStatusHandler

	chStopSync chan bool

	mutNodeState          sync.RWMutex
	isNodeSynchronized    bool
	isNodeStateCalculated bool
	nodeStateHasAmbiguity bool
	hasLastBlock          bool
	roundIndex            int64

	forkInfo *process.ForkInfo

	mutReconcile      sync.Mutex
	pendingReconcile  *reconcileEvidence
	mutAmbiguity      sync.Mutex
	ambiguityRecovery ambiguityRecoveryState
	mutRecovery       sync.Mutex
	recoveryState     resyncRecoveryState
	recoveryActive    atomic.Bool
	recoveryBypass    atomic.Bool
	recoveryEvalSet   atomic.Bool
	recoveryEvalRound atomic.Int64

	// only touched from the sync goroutine, no lock needed
	divergenceEvaluatedRound int64
	epochStartTrigger        process.EpochStartTriggerHandler
	epochStartDisarmer       epochStartTriggerDisarmer

	mutRcvHdrNonce           sync.RWMutex
	mutRcvHdrHash            sync.RWMutex
	syncStateListeners       []func(bool)
	mutSyncStateListeners    sync.RWMutex
	uint64Converter          typeConverters.Uint64ByteSliceConverter
	mapNonceSyncedWithErrors map[uint64]uint32
	mapNonceRecoveryAttempts map[uint64]*nonceRecoveryInfo // guarded by mutNonceSyncedWithErrors
	mutNonceSyncedWithErrors sync.RWMutex

	// owned by the single sync goroutine (doJobOnSyncBlockFail and the post-commit cleanup), no mutex
	blockingUnprovenHdrHash     []byte
	blockingUnprovenHdrFailures uint32

	executionResultsRecoveryCooldown time.Duration

	requestMiniBlocks func(headerHandler data.HeaderHandler)

	networkWatcher process.NetworkConnectionWatcher

	headerStore          storage.Storer
	headerNonceHashStore storage.Storer
	syncStarter          syncStarter
	bootStorer           process.BootStorer
	storageBootstrapper  process.BootstrapperFromStorage
	currentEpochProvider process.CurrentNetworkEpochProviderHandler

	outportHandler        outport.OutportHandler
	accountsDBSyncer      process.AccountsDBSyncer
	processConfigsHandler common.ProcessConfigsHandler

	chRcvMiniBlocks              chan bool
	mutRcvMiniBlocks             sync.Mutex
	miniBlocksProvider           process.MiniBlockProvider
	poolsHolder                  dataRetriever.PoolsHolder
	mutRequestHeaders            sync.Mutex
	cancelFunc                   func()
	isInImportMode               bool
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	processWaitTime              time.Duration
	processWaitTimeSupernova     time.Duration
	preparedForSync              bool
	preparedForSyncAtBootstrap   bool
	pendingV3Realign             bool
	pendingV3RollBack            *pendingV3RollBack

	repopulateTokensSupplies bool

	miniBlocksSyncer epochStart.PendingMiniBlocksSyncHandler
	txSyncer         update.TransactionsSyncHandler

	signalProcessCompletionChan chan uint64
}

// pendingV3RollBack tracks a v3 roll back interrupted mid-way, so the sync loop can complete or
// abandon it before any other work; accessed only from the sync goroutine
type pendingV3RollBack struct {
	currHeaderHash       []byte
	currHeader           data.HeaderHandler
	prevHeaderHash       []byte
	prevHeader           data.HeaderHandler
	currBody             data.BodyHandler
	restoreDone          bool
	executionPruned      bool
	unrestoredMetaBlocks []process.MovedMetaBlock
}

func (boot *baseBootstrap) getProcessWaitTime(round uint64) time.Duration {
	if boot.enableRoundsHandler.IsFlagEnabledInRound(common.SupernovaRoundFlag, round) {
		return boot.processWaitTimeSupernova
	}

	return boot.processWaitTime
}

// setRequestedHeaderNonce method sets the header nonce requested by the sync mechanism
func (boot *baseBootstrap) setRequestedHeaderNonce(nonce *uint64) {
	boot.mutHeader.Lock()
	boot.headerNonce = nonce
	boot.mutHeader.Unlock()
}

// setRequestedHeaderHash method sets the header hash requested by the sync mechanism
func (boot *baseBootstrap) setRequestedHeaderHash(hash []byte) {
	boot.mutHeader.Lock()
	boot.headerhash = hash
	boot.mutHeader.Unlock()
}

// requestedHeaderNonce method gets the header nonce requested by the sync mechanism
func (boot *baseBootstrap) requestedHeaderNonce() *uint64 {
	boot.mutHeader.RLock()
	defer boot.mutHeader.RUnlock()
	return boot.headerNonce
}

// requestedHeaderHash method gets the header hash requested by the sync mechanism
func (boot *baseBootstrap) requestedHeaderHash() []byte {
	boot.mutHeader.RLock()
	defer boot.mutHeader.RUnlock()
	return boot.headerhash
}

func (boot *baseBootstrap) processReceivedProof(headerProof data.HeaderProofHandler) {
	if boot.shardCoordinator.SelfId() != headerProof.GetHeaderShardId() {
		return
	}

	boot.forkDetector.ReceivedProof(headerProof)
	boot.enrichForkDetectorWithProofHeader(headerProof)
	boot.clearRecoveryAfterProgress()

	boot.checkProofCorrespondsToRequestedHash(headerProof)
	boot.checkProofCorrespondsToRequestedNonce(headerProof)
}

func (boot *baseBootstrap) enrichForkDetectorWithProofHeader(headerProof data.HeaderProofHandler) {
	if !common.IsAsyncExecutionEnabledForEpochAndRound(
		boot.enableEpochsHandler,
		boot.enableRoundsHandler,
		headerProof.GetHeaderEpoch(),
		headerProof.GetHeaderRound(),
	) {
		return
	}

	header, err := boot.getHeaderFromPool(headerProof.GetHeaderHash())
	if err != nil {
		return
	}

	err = boot.forkDetector.AddHeader(header, headerProof.GetHeaderHash(), process.BHReceived, nil, nil)
	if err != nil {
		log.Trace("failed to enrich fork detector with proof header", "error", err)
	}
}

func (boot *baseBootstrap) checkProofCorrespondsToRequestedHash(headerProof data.HeaderProofHandler) {
	boot.mutRcvHdrHash.RLock()
	hash := boot.requestedHeaderHash()
	wasHashRequested := hash != nil && bytes.Equal(hash, headerProof.GetHeaderHash())
	if !wasHashRequested {
		boot.mutRcvHdrHash.RUnlock()
		return
	}

	// if header is also received, release the chan and set requested to nil
	// otherwise wait for the header
	_, err := boot.getHeader(headerProof.GetHeaderHash())
	hasHeader := err == nil
	if hasHeader {
		boot.setRequestedHeaderHash(nil)
		boot.mutRcvHdrHash.RUnlock()

		boot.chRcvHdrHash <- true

		return
	}

	boot.mutRcvHdrHash.RUnlock()
}

func (boot *baseBootstrap) checkProofCorrespondsToRequestedNonce(headerProof data.HeaderProofHandler) {
	boot.mutRcvHdrNonce.RLock()
	n := boot.requestedHeaderNonce()
	wasNonceRequested := n != nil && *n == headerProof.GetHeaderNonce()
	if !wasNonceRequested {
		boot.mutRcvHdrNonce.RUnlock()
		return
	}

	// if header is also received, release the chan and set requested to nil
	// otherwise wait for the header
	_, err := boot.getHeader(headerProof.GetHeaderHash())
	hasHeader := err == nil
	if hasHeader {
		boot.setRequestedHeaderNonce(nil)
		boot.mutRcvHdrNonce.RUnlock()

		boot.chRcvHdrNonce <- true

		return
	}

	boot.mutRcvHdrNonce.RUnlock()
}

func (boot *baseBootstrap) processReceivedHeader(headerHandler data.HeaderHandler, headerHash []byte) {
	if boot.shardCoordinator.SelfId() != headerHandler.GetShardID() {
		return
	}

	log.Debug("sync: received header from network",
		"shard", headerHandler.GetShardID(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
		"hash", headerHash,
	)

	err := boot.forkDetector.AddHeader(headerHandler, headerHash, process.BHReceived, nil, nil)
	if err != nil {
		log.Debug("forkDetector.AddHeader", "error", err.Error())
	}

	boot.observeRecoveryHeader(headerHandler)

	go boot.requestMiniBlocks(headerHandler)

	boot.confirmHeaderReceivedByNonce(headerHandler, headerHash)
	boot.confirmHeaderReceivedByHash(headerHandler, headerHash)
}

func (boot *baseBootstrap) confirmHeaderReceivedByNonce(headerHandler data.HeaderHandler, hdrHash []byte) {
	boot.mutRcvHdrNonce.Lock()
	n := boot.requestedHeaderNonce()
	if n != nil && *n == headerHandler.GetNonce() {
		log.Debug("received requested header from network",
			"shard", headerHandler.GetShardID(),
			"round", headerHandler.GetRound(),
			"nonce", headerHandler.GetNonce(),
			"hash", hdrHash,
		)

		// if flag is not active for the header, do not check the proof and release chan
		isFlagActive := common.IsProofsFlagEnabledForHeader(boot.enableEpochsHandler, headerHandler)
		if !isFlagActive {
			boot.setRequestedHeaderNonce(nil)
			boot.mutRcvHdrNonce.Unlock()

			boot.chRcvHdrNonce <- true

			return
		}

		// if proof is also received, release chan and set requested to nil
		// otherwise, wait for the proof too
		hasProof := boot.proofs.HasProof(headerHandler.GetShardID(), hdrHash)
		if hasProof {
			log.Debug("received requested proof from network",
				"shard", headerHandler.GetShardID(),
				"round", headerHandler.GetRound(),
				"nonce", headerHandler.GetNonce(),
				"hash", hdrHash,
			)
			boot.setRequestedHeaderNonce(nil)
		}
		boot.mutRcvHdrNonce.Unlock()

		if hasProof {
			boot.chRcvHdrNonce <- true
			return
		}

		boot.requestHandler.SetEpoch(headerHandler.GetEpoch())
		boot.requestHandler.RequestEquivalentProofByHashForEpoch(headerHandler.GetShardID(), hdrHash, headerHandler.GetEpoch())

		return
	}

	boot.mutRcvHdrNonce.Unlock()
}

func (boot *baseBootstrap) confirmHeaderReceivedByHash(headerHandler data.HeaderHandler, hdrHash []byte) {
	boot.mutRcvHdrHash.Lock()
	hash := boot.requestedHeaderHash()
	if hash != nil && bytes.Equal(hash, hdrHash) {
		log.Debug("received requested header from network",
			"shard", headerHandler.GetShardID(),
			"round", headerHandler.GetRound(),
			"nonce", headerHandler.GetNonce(),
			"hash", hash,
		)

		// if flag is not active for the header, do not check the proof and release chan
		isFlagActive := common.IsProofsFlagEnabledForHeader(boot.enableEpochsHandler, headerHandler)
		if !isFlagActive {
			boot.setRequestedHeaderHash(nil)
			boot.mutRcvHdrHash.Unlock()

			boot.chRcvHdrHash <- true

			return
		}

		// if proof is also received, release chan and set requested to nil
		// otherwise, wait for the proof too
		hasProof := boot.proofs.HasProof(headerHandler.GetShardID(), hash)
		if hasProof {
			log.Debug("received requested proof from network",
				"shard", headerHandler.GetShardID(),
				"round", headerHandler.GetRound(),
				"nonce", headerHandler.GetNonce(),
				"hash", hash,
			)
			boot.setRequestedHeaderHash(nil)
		}
		boot.mutRcvHdrHash.Unlock()

		if hasProof {
			boot.chRcvHdrHash <- true
			return
		}

		boot.requestHandler.SetEpoch(headerHandler.GetEpoch())
		boot.requestHandler.RequestEquivalentProofByHashForEpoch(headerHandler.GetShardID(), hdrHash, headerHandler.GetEpoch())

		return
	}

	boot.mutRcvHdrHash.Unlock()
}

func (boot *baseBootstrap) hasProof(hash []byte, header data.HeaderHandler) bool {
	if !common.IsProofsFlagEnabledForHeader(boot.enableEpochsHandler, header) {
		return true
	}

	return boot.proofs.HasProof(boot.shardCoordinator.SelfId(), hash)
}

// AddSyncStateListener adds a syncStateListener that get notified each time the sync status of the node changes
func (boot *baseBootstrap) AddSyncStateListener(syncStateListener func(isSyncing bool)) {
	boot.mutSyncStateListeners.Lock()
	boot.syncStateListeners = append(boot.syncStateListeners, syncStateListener)
	boot.mutSyncStateListeners.Unlock()
}

func (boot *baseBootstrap) notifySyncStateListeners(isNodeSynchronized bool) {
	boot.mutSyncStateListeners.RLock()
	for i := 0; i < len(boot.syncStateListeners); i++ {
		go boot.syncStateListeners[i](isNodeSynchronized)
	}
	boot.mutSyncStateListeners.RUnlock()
}

// getNonceForNextBlock will get the nonce for the next block
func (boot *baseBootstrap) getNonceForNextBlock() uint64 {
	nonce := boot.chainHandler.GetGenesisHeader().GetNonce() + 1 // first block nonce after genesis block
	currentBlockHeader := boot.chainHandler.GetCurrentBlockHeader()
	if !check.IfNil(currentBlockHeader) {
		nonce = currentBlockHeader.GetNonce() + 1
	}
	return nonce
}

// getCurrentBlock will get the current block
func (boot *baseBootstrap) getCurrentBlock() data.HeaderHandler {
	currentBlockHeader := boot.chainHandler.GetCurrentBlockHeader()
	if !check.IfNil(currentBlockHeader) {
		return currentBlockHeader
	}
	return boot.chainHandler.GetGenesisHeader()
}

// getCurrentRootHashLegacy will get the current root hash
func (boot *baseBootstrap) getCurrentRootHashLegacy() []byte {
	currentRootHash := boot.chainHandler.GetCurrentBlockRootHash()
	if len(currentRootHash) != 0 {
		return currentRootHash
	}
	genesisHeader := boot.chainHandler.GetGenesisHeader()
	return genesisHeader.GetRootHash()
}

// getCurrentBlockHash will get the current block hash
func (boot *baseBootstrap) getCurrentBlockHash() []byte {
	currentHash := boot.chainHandler.GetCurrentBlockHeaderHash()
	if len(currentHash) != 0 {
		return currentHash
	}
	return boot.chainHandler.GetGenesisHeaderHash()
}

// getNonceForCurrentBlock will get the nonce for the current block
func (boot *baseBootstrap) getNonceForCurrentBlock() uint64 {
	nonce := boot.chainHandler.GetGenesisHeader().GetNonce() // genesis block nonce
	currentBlockHeader := boot.chainHandler.GetCurrentBlockHeader()
	if !check.IfNil(currentBlockHeader) {
		nonce = currentBlockHeader.GetNonce()
	}
	return nonce
}

// getEpochOfCurrentBlock will get the epoch for the current block as stored in the chain handler implementation
func (boot *baseBootstrap) getEpochOfCurrentBlock() uint32 {
	epoch := boot.chainHandler.GetGenesisHeader().GetEpoch()
	currentBlockHeader := boot.chainHandler.GetCurrentBlockHeader()
	if !check.IfNil(currentBlockHeader) {
		epoch = currentBlockHeader.GetEpoch()
	}
	return epoch
}

func (boot *baseBootstrap) getWaitTime() time.Duration {
	return boot.roundHandler.TimeDuration()
}

// waitForHeaderAndProofByNonce method wait for header with the requested nonce to be received
func (boot *baseBootstrap) waitForHeaderAndProofByNonce() error {
	select {
	case <-boot.chRcvHdrNonce:
		return nil
	case <-time.After(boot.getWaitTime()):
		return process.ErrTimeIsOut
	}
}

// waitForHeaderAndProofByHash method wait for header with the requested hash to be received
func (boot *baseBootstrap) waitForHeaderAndProofByHash() error {
	select {
	case <-boot.chRcvHdrHash:
		return nil
	case <-time.After(boot.getWaitTime()):
		return process.ErrTimeIsOut
	}
}

func (boot *baseBootstrap) computeNodeState(round int64) {
	boot.tryResolveNotarizedAmbiguity(round)
	hasUnresolvedAuthority := boot.hasUnresolvedNotarizedAmbiguity()

	boot.mutNodeState.Lock()
	defer boot.mutNodeState.Unlock()

	isNodeStateCalculatedInCurrentRound := boot.roundIndex == round && boot.isNodeStateCalculated
	if isNodeStateCalculatedInCurrentRound && boot.nodeStateHasAmbiguity == hasUnresolvedAuthority {
		return
	}

	boot.forkInfo = boot.forkDetector.CheckFork()
	hasUnresolvedAuthority = hasUnresolvedAuthority || boot.hasUnresolvedNotarizedAmbiguity()

	genesisNonce := boot.chainHandler.GetGenesisHeader().GetNonce()
	currentHeader := boot.chainHandler.GetCurrentBlockHeader()
	if check.IfNil(currentHeader) {
		boot.hasLastBlock = boot.forkDetector.ProbableHighestNonce() == genesisNonce
		log.Debug("computeNodeState",
			"probableHighestNonce", boot.forkDetector.ProbableHighestNonce(),
			"currentBlockNonce", nil,
			"boot.hasLastBlock", boot.hasLastBlock)
	} else {
		boot.hasLastBlock = boot.forkDetector.ProbableHighestNonce() <= boot.chainHandler.GetCurrentBlockHeader().GetNonce()
		log.Debug("computeNodeState",
			"probableHighestNonce", boot.forkDetector.ProbableHighestNonce(),
			"currentBlockNonce", boot.chainHandler.GetCurrentBlockHeader().GetNonce(),
			"boot.hasLastBlock", boot.hasLastBlock)
	}

	isNodeConnectedToTheNetwork := boot.networkWatcher.IsConnectedToTheNetwork()
	isNodeSynchronized := !boot.forkInfo.IsDetected && !hasUnresolvedAuthority && boot.hasLastBlock && isNodeConnectedToTheNetwork
	if isNodeSynchronized != boot.isNodeSynchronized {
		log.Debug("node has changed its synchronized state",
			"state", isNodeSynchronized,
		)
	}

	boot.isNodeSynchronized = isNodeSynchronized
	boot.isNodeStateCalculated = true
	boot.nodeStateHasAmbiguity = hasUnresolvedAuthority
	boot.roundIndex = round
	boot.notifySyncStateListeners(isNodeSynchronized)

	result := uint64(1)
	if isNodeSynchronized {
		result = uint64(0)
	}

	boot.statusHandler.SetUInt64Value(common.MetricIsSyncing, result)
	log.Debug("computeNodeState",
		"isNodeStateCalculated", boot.isNodeStateCalculated,
		"isNodeSynchronized", boot.isNodeSynchronized)

	shouldRequest, bypassGeneration := boot.shouldTryToRequestHeaders()
	if shouldRequest {
		go boot.requestHeadersIfSyncIsStuckForGeneration(bypassGeneration)
	}
}

func (boot *baseBootstrap) shouldTryToRequestHeaders() (bool, uint64) {
	if boot.roundHandler.BeforeGenesis() {
		return false, 0
	}
	if boot.isForcedRollBackOneBlock() {
		return false, 0
	}
	if boot.isForcedRollBackToNonce() {
		return false, 0
	}
	if !boot.isNodeSynchronized {
		// normal sync handles requests while the probable nonce is ahead
		hasKnownBacklog := boot.forkDetector.ProbableHighestNonce() > boot.currentCommittedNonce()
		return !hasKnownBacklog, 0
	}

	roundIndex := boot.roundHandler.Index()
	useBypass, generation := boot.usePostBootstrapWatchdogBypass(roundIndex)
	if useBypass {
		return true, generation
	}

	roundModulusTriggerWhenSyncIsStuck := boot.processConfigsHandler.GetRoundModulusTriggerWhenSyncIsStuck(uint64(roundIndex))

	return roundIndex%int64(roundModulusTriggerWhenSyncIsStuck) == 0, 0
}

func (boot *baseBootstrap) requestHeadersIfSyncIsStuckForGeneration(generation uint64) {
	if generation != 0 && !boot.isWatchdogBypassGenerationActive(generation) {
		return
	}

	boot.requestHeadersIfSyncIsStuck()
}

func (boot *baseBootstrap) requestHeadersIfSyncIsStuck() {
	lastSyncedRound := boot.chainHandler.GetGenesisHeader().GetRound()
	currHeader := boot.chainHandler.GetCurrentBlockHeader()
	if !check.IfNil(currHeader) {
		lastSyncedRound = currHeader.GetRound()
	}

	currentRound := boot.roundHandler.Index()
	if currentRound < 0 || uint64(currentRound) <= lastSyncedRound {
		return
	}

	roundDiff := uint64(currentRound) - lastSyncedRound
	if roundDiff <= boot.getMaxRoundsWithoutBlockReceived(lastSyncedRound) {
		return
	}

	fromNonce := boot.getNonceForNextBlock()
	numHeadersToRequest := core.MinUint64(process.MaxHeadersToRequestInAdvance, roundDiff-1)
	toNonce := fromNonce + numHeadersToRequest - 1

	if fromNonce > toNonce {
		return
	}

	log.Debug("requestHeadersIfSyncIsStuck",
		"from nonce", fromNonce,
		"to nonce", toNonce,
		"probable highest nonce", boot.forkDetector.ProbableHighestNonce())

	boot.requestHeaders(fromNonce, toNonce)
}

func (boot *baseBootstrap) getMaxRoundsWithoutBlockReceived(round uint64) uint64 {
	return uint64(boot.processConfigsHandler.GetMaxRoundsWithoutNewBlockReceivedByRound(round))
}

func (boot *baseBootstrap) removeHeaderFromPools(header data.HeaderHandler) []byte {
	hash, err := core.CalculateHash(boot.marshalizer, boot.hasher, header)
	if err != nil {
		log.Debug("CalculateHash", "error", err.Error())
		return nil
	}

	log.Debug("removeHeaderFromPools",
		"shard", header.GetShardID(),
		"epoch", header.GetEpoch(),
		"round", header.GetRound(),
		"nonce", header.GetNonce(),
		"hash", hash)

	boot.headers.RemoveHeaderByHash(hash)

	return hash
}

func (boot *baseBootstrap) removeHeadersHigherThanNonceFromPool(nonce uint64) {
	shardID := boot.shardCoordinator.SelfId()
	log.Debug("removeHeadersHigherThanNonceFromPool",
		"shard", shardID,
		"nonce", nonce)

	nonces := boot.headers.Nonces(shardID)
	for _, currentNonce := range nonces {
		if currentNonce <= nonce {
			continue
		}

		boot.headers.RemoveHeaderByNonceAndShardId(currentNonce, shardID)
	}
}

func (boot *baseBootstrap) cleanCachesAndStorageOnRollback(header data.HeaderHandler) {
	hash := boot.removeHeaderFromPools(header)
	boot.forkDetector.RemoveHeader(header.GetNonce(), hash)
	nonceToByteSlice := boot.uint64Converter.ToByteSlice(header.GetNonce())
	_ = boot.headerNonceHashStore.Remove(nonceToByteSlice)
}

// checkBaseBootstrapParameters will check the correctness of the provided parameters
func checkBaseBootstrapParameters(arguments ArgBaseBootstrapper) error {
	if check.IfNil(arguments.ChainHandler) {
		return process.ErrNilBlockChain
	}
	if check.IfNil(arguments.EpochStartTrigger) {
		return process.ErrNilEpochStartTrigger
	}
	if check.IfNil(arguments.RoundHandler) {
		return process.ErrNilRoundHandler
	}
	if check.IfNil(arguments.BlockProcessor) {
		return process.ErrNilBlockProcessor
	}
	if check.IfNil(arguments.ExecutionManager) {
		return process.ErrNilExecutionManager
	}
	if check.IfNil(arguments.Hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(arguments.Marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(arguments.ForkDetector) {
		return process.ErrNilForkDetector
	}
	if check.IfNil(arguments.RequestHandler) {
		return process.ErrNilRequestHandler
	}
	if check.IfNil(arguments.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(arguments.Accounts) {
		return process.ErrNilAccountsAdapter
	}
	if check.IfNil(arguments.Store) {
		return process.ErrNilStore
	}
	if check.IfNil(arguments.BlackListHandler) {
		return process.ErrNilBlackListCacher
	}
	if check.IfNil(arguments.NetworkWatcher) {
		return process.ErrNilNetworkWatcher
	}
	if check.IfNil(arguments.BootStorer) {
		return process.ErrNilBootStorer
	}
	if check.IfNil(arguments.MiniblocksProvider) {
		return process.ErrNilMiniBlocksProvider
	}
	if check.IfNil(arguments.AppStatusHandler) {
		return process.ErrNilAppStatusHandler
	}
	if check.IfNil(arguments.OutportHandler) {
		return process.ErrNilOutportHandler
	}
	if check.IfNil(arguments.AccountsDBSyncer) {
		return process.ErrNilAccountsDBSyncer
	}
	if check.IfNil(arguments.CurrentEpochProvider) {
		return process.ErrNilCurrentNetworkEpochProvider
	}
	if check.IfNil(arguments.HistoryRepo) {
		return process.ErrNilHistoryRepository
	}
	if check.IfNil(arguments.ScheduledTxsExecutionHandler) {
		return process.ErrNilScheduledTxsExecutionHandler
	}
	if arguments.ProcessWaitTime < minimumProcessWaitTime {
		return fmt.Errorf("%w, minimum is %v, provided is %v", process.ErrInvalidProcessWaitTime, minimumProcessWaitTime, arguments.ProcessWaitTime)
	}
	if arguments.ProcessWaitTimeSupernova < minimumProcessWaitTime {
		return fmt.Errorf("%w for Supernova, minimum is %v, provided is %v", process.ErrInvalidProcessWaitTime, minimumProcessWaitTime, arguments.ProcessWaitTimeSupernova)
	}
	if check.IfNil(arguments.EnableEpochsHandler) {
		return process.ErrNilEnableEpochsHandler
	}
	if check.IfNil(arguments.EnableRoundsHandler) {
		return process.ErrNilEnableRoundsHandler
	}

	return nil
}

func (boot *baseBootstrap) requestHeadersFromNonceIfMissing(fromNonce uint64) {
	toNonce := core.MinUint64(fromNonce+process.MaxHeadersToRequestInAdvance-1, boot.forkDetector.ProbableHighestNonce())

	if fromNonce > toNonce {
		// request at least the next header so the fork detector
		// can discover blocks beyond probableHighestNonce
		toNonce = fromNonce
	}

	log.Debug("requestHeadersFromNonceIfMissing",
		"from nonce", fromNonce,
		"to nonce", toNonce,
		"probable highest nonce", boot.forkDetector.ProbableHighestNonce())

	boot.requestHeaders(fromNonce, toNonce)
}

// syncBlocks method calls repeatedly synchronization method SyncBlock
func (boot *baseBootstrap) syncBlocks(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("bootstrap's go routine is stopping...")
			return
		case <-time.After(sleepTime):
		}

		if !boot.networkWatcher.IsConnectedToTheNetwork() {
			continue
		}
		if boot.roundHandler.BeforeGenesis() {
			continue
		}

		err := boot.syncStarter.SyncBlock(ctx)
		if err != nil {
			if common.IsContextDone(ctx) {
				log.Debug("SyncBlock finished, bootstrap's go routine is stopping...")
				return
			}

			log.Debug("SyncBlock", "error", err.Error())

			select {
			case nonce := <-boot.signalProcessCompletionChan:
				log.Debug("SyncBlock - error - notification process finished", "nonce", nonce)
			case <-time.After(sleepTimeOnFail):
			}
		} else {
			// Non-blocking drain of completion signal when sync succeeds
			select {
			case nonce := <-boot.signalProcessCompletionChan:
				log.Debug("SyncBlock - success - notification process finished", "nonce", nonce)
			default:
			}
		}
	}
}

func (boot *baseBootstrap) getMaxSyncWithErrorsAllowed(
	header data.HeaderHandler,
) uint32 {
	// no header means the sync never got one for this nonce: fall back to the wall-clock round so
	// the limit comes from the active config even when the chronology-backed index lags
	round := uint64(0)
	if !check.IfNil(header) {
		round = header.GetRound()
	} else if currentRound := boot.roundHandler.IndexForCurrentTime(); currentRound > 0 {
		round = uint64(currentRound)
	}

	return boot.processConfigsHandler.GetMaxSyncWithErrorsAllowed(round)
}

func (boot *baseBootstrap) doJobOnSyncBlockFail(bodyHandler data.BodyHandler, headerHandler data.HeaderHandler, err error) {
	if errors.Is(err, errBranchAwareSyncRetry) {
		return
	}

	if errors.Is(err, process.ErrBlockProcessorBusy) {
		// block processor is busy with another call (e.g. consensus processing the same block);
		// no processing started, nothing to track or roll back - just retry on next sync iteration
		return
	}

	if boot.tryRecoverFromExecutionResultsMismatch(headerHandler, err) {
		// nothing to roll back, the synced header stays in pool for retry
		return
	}

	processBlockStarted := !check.IfNil(bodyHandler) && !check.IfNil(headerHandler)
	// missing data is not evidence against the local chain: keep the header and the fetched txs so
	// retries accumulate; the errors limit below stays the last resort
	isMissingDataFailure := errors.Is(err, process.ErrTimeIsOut) ||
		errors.Is(err, process.ErrTxNotFound) ||
		errors.Is(err, process.ErrMissingTransaction)
	isProcessWithError := processBlockStarted && !isMissingDataFailure

	numSyncedWithErrors := boot.incrementSyncedWithErrorsForNonce(boot.getNonceForNextBlock())
	allowedSyncWithErrorsLimitReached := numSyncedWithErrors >= boot.getMaxSyncWithErrorsAllowed(headerHandler)
	isInProperRound := process.IsInProperRound(boot.roundHandler.Index())
	isSyncWithErrorsLimitReachedInProperRound := allowedSyncWithErrorsLimitReached && isInProperRound

	lastCommittedBlock := boot.chainHandler.GetCurrentBlockHeader()
	lastCommittedBlockHash := boot.chainHandler.GetCurrentBlockHeaderHash()
	shouldAllowRollback := boot.shouldAllowRollback(lastCommittedBlock, lastCommittedBlockHash)

	shouldRollBack := isProcessWithError || isSyncWithErrorsLimitReachedInProperRound
	didRollBack := shouldRollBack && shouldAllowRollback
	if didRollBack {
		if !check.IfNil(headerHandler) {
			hash := boot.removeHeaderFromPools(headerHandler)
			boot.forkDetector.RemoveHeader(headerHandler.GetNonce(), hash)
		}

		errNotCritical := boot.rollBack(false)
		if errNotCritical != nil {
			log.Debug("rollBack", "error", errNotCritical.Error())
		}

		if isSyncWithErrorsLimitReachedInProperRound {
			boot.forkDetector.ResetProbableHighestNonce()
			boot.removeHeadersHigherThanNonceFromPool(boot.getNonceForCurrentBlock())
		}
	}

	// stuck fetching the next header (no processing, no rollback): drop a non-final fork header whose proof
	// will never arrive so it gets re-requested
	if check.IfNil(headerHandler) && !didRollBack {
		boot.removeBlockingUnprovenNextHeader()
	}
}

func (boot *baseBootstrap) removeBlockingUnprovenNextHeader() {
	// paced by its own tracker, decoupled from the rollback-limit counter: counting only consecutive
	// iterations with the unproven header present guarantees a re-fetched header the full window
	nonce := boot.getNonceForNextBlock()
	hdr, hash, err := boot.getHeaderFromPoolWithNonce(nonce)
	if err != nil {
		boot.clearBlockingUnprovenHdrTracker()
		return
	}
	if boot.hasProof(hash, hdr) {
		boot.clearBlockingUnprovenHdrTracker()
		return
	}

	if !bytes.Equal(hash, boot.blockingUnprovenHdrHash) {
		// copied: the pool owns the returned slice
		boot.blockingUnprovenHdrHash = append(boot.blockingUnprovenHdrHash[:0], hash...)
		boot.blockingUnprovenHdrFailures = 1
		return
	}

	boot.blockingUnprovenHdrFailures++
	if boot.blockingUnprovenHdrFailures < maxFetchFailuresBeforeDroppingUnprovenHeader {
		return
	}

	log.Debug("removeBlockingUnprovenNextHeader: removing unproven header blocking sync, will re-request",
		"shard", hdr.GetShardID(),
		"nonce", nonce,
		"hash", hash,
	)

	boot.headers.RemoveHeaderByHash(hash)
	boot.forkDetector.RemoveHeader(nonce, hash)
	boot.clearBlockingUnprovenHdrTracker()
}

func (boot *baseBootstrap) clearBlockingUnprovenHdrTracker() {
	boot.blockingUnprovenHdrHash = boot.blockingUnprovenHdrHash[:0]
	boot.blockingUnprovenHdrFailures = 0
}

func (boot *baseBootstrap) incrementSyncedWithErrorsForNonce(nonce uint64) uint32 {
	boot.mutNonceSyncedWithErrors.Lock()
	boot.mapNonceSyncedWithErrors[nonce]++
	numSyncedWithErrors := boot.mapNonceSyncedWithErrors[nonce]
	boot.mutNonceSyncedWithErrors.Unlock()

	return numSyncedWithErrors
}

func (boot *baseBootstrap) resetSyncedWithErrorsForNonce(nonce uint64) {
	boot.mutNonceSyncedWithErrors.Lock()
	delete(boot.mapNonceSyncedWithErrors, nonce)
	boot.mutNonceSyncedWithErrors.Unlock()
}

// tryRecoverFromExecutionResultsMismatch removes the local pending execution results diverging
// from the notarized ones carried by the synced header (canonical, as the header passed consensus)
// and re-queues the affected blocks for re-execution
func (boot *baseBootstrap) tryRecoverFromExecutionResultsMismatch(headerHandler data.HeaderHandler, err error) bool {
	if !errors.Is(err, process.ErrExecutionResultDoesNotMatch) {
		return false
	}
	if check.IfNil(headerHandler) || !headerHandler.IsHeaderV3() {
		return false
	}

	rewindNonce, found := boot.getFirstDivergingExecutionResultNonce(headerHandler)
	if !found {
		lastNotarizedResult, errNotarized := boot.executionManager.GetLastNotarizedExecutionResult()
		if errNotarized != nil || check.IfNil(lastNotarizedResult) {
			log.Warn("tryRecoverFromExecutionResultsMismatch: cannot get last notarized execution result",
				"error", errNotarized,
			)
			return false
		}

		rewindNonce = lastNotarizedResult.GetHeaderNonce() + 1
	}

	numAttempts, allowed := boot.shouldAttemptRecoveryForNonce(rewindNonce)
	if !allowed {
		log.Debug("tryRecoverFromExecutionResultsMismatch: recovery cooldown not expired",
			"rewind nonce", rewindNonce,
			"synced header nonce", headerHandler.GetNonce(),
			"num recovery attempts", numAttempts,
		)
		return false
	}

	log.Warn("tryRecoverFromExecutionResultsMismatch: local execution results diverged from the "+
		"notarized ones carried by the synced header, removing local pending execution results "+
		"and re-executing the affected blocks",
		"rewind nonce", rewindNonce,
		"synced header nonce", headerHandler.GetNonce(),
		"synced header round", headerHandler.GetRound(),
		"recovery attempt", numAttempts,
	)

	errRemove := boot.executionManager.RemoveAtNonceAndHigher(rewindNonce)
	if errRemove != nil {
		log.Warn("tryRecoverFromExecutionResultsMismatch: RemoveAtNonceAndHigher failed",
			"rewind nonce", rewindNonce,
			"error", errRemove,
		)
		return false
	}

	// force the backfill to re-queue the removed blocks for execution
	boot.preparedForSync = false
	boot.resetSyncedWithErrorsForNonce(boot.getNonceForNextBlock())

	return true
}

// getFirstDivergingExecutionResultNonce returns the nonce of the first local pending execution
// result differing from the header's notarized one; matched by nonce to be immune to list misalignment
func (boot *baseBootstrap) getFirstDivergingExecutionResultNonce(headerHandler data.HeaderHandler) (uint64, bool) {
	pendingExecutionResults, err := boot.executionManager.GetPendingExecutionResults()
	if err != nil {
		log.Debug("getFirstDivergingExecutionResultNonce: cannot get pending execution results", "error", err)
		return 0, false
	}

	pendingByNonce := make(map[uint64]data.BaseExecutionResultHandler, len(pendingExecutionResults))
	for _, pendingResult := range pendingExecutionResults {
		pendingByNonce[pendingResult.GetHeaderNonce()] = pendingResult
	}

	for _, headerResult := range headerHandler.GetExecutionResultsHandlers() {
		pendingResult, ok := pendingByNonce[headerResult.GetHeaderNonce()]
		if !ok {
			continue
		}
		if !headerResult.Equal(pendingResult) {
			return headerResult.GetHeaderNonce(), true
		}
	}

	return 0, false
}

// shouldAttemptRecoveryForNonce records a new recovery attempt, unless the cooldown has not expired
func (boot *baseBootstrap) shouldAttemptRecoveryForNonce(nonce uint64) (uint32, bool) {
	boot.mutNonceSyncedWithErrors.Lock()
	defer boot.mutNonceSyncedWithErrors.Unlock()

	if boot.mapNonceRecoveryAttempts == nil {
		boot.mapNonceRecoveryAttempts = make(map[uint64]*nonceRecoveryInfo)
	}

	info, ok := boot.mapNonceRecoveryAttempts[nonce]
	if ok && time.Since(info.lastAttempt) < boot.executionResultsRecoveryCooldown {
		return info.numAttempts, false
	}

	if !ok {
		info = &nonceRecoveryInfo{}
		boot.mapNonceRecoveryAttempts[nonce] = info
	}
	info.numAttempts++
	info.lastAttempt = time.Now()

	return info.numAttempts, true
}

func (boot *baseBootstrap) prepareForSyncAtBoostrapIfNeeded() error {
	// this will be triggered only once, after a full node restart.
	// it is needed for the case when the node will go through bootstrap process and start
	// directly into execution flow, because it is already synced (ex: if the entire shard
	// was down and when the node will came back it will still be in sync, because the
	// shard did not advance while the node was down).
	// in case of shuffle out and moving to another shard, the node will not have to
	// go through this flow, it will go through sync flow directly, so it will not be
	// a problem that preparedForSyncAtBootstrap is already set

	if boot.preparedForSyncAtBootstrap {
		return nil
	}

	// at this point, current header should be the last applied header at bootstrap
	currentHeader := boot.getCurrentBlock()

	if !currentHeader.IsHeaderV3() {
		boot.preparedForSyncAtBootstrap = true
		boot.armPostBootstrapWatchdogBypass()

		return nil
	}

	// syncing nonce is taken as next nonce, this is for preparedForSyncIfNeeded to work
	// properly in this case
	syncingNonce := currentHeader.GetNonce() + 1

	log.Debug("prepareForSyncAtBoostrapIfNeeded",
		"currHeader nonce", currentHeader.GetNonce(),
	)

	err := boot.prepareForSyncIfNeeded(syncingNonce)
	if err != nil {
		return err
	}

	boot.preparedForSyncAtBootstrap = true
	boot.armPostBootstrapWatchdogBypass()

	return nil
}

func (boot *baseBootstrap) syncBlock() error {
	// an interrupted roll back leaves state no other sync work may build on; resolving it
	// (completing or abandoning) is mandatory before anything else runs
	if boot.pendingV3RollBack != nil {
		return boot.completeInterruptedV3RollBack()
	}

	// a failed post-rollback realign leaves execution state behind the tip; nothing may run on
	// top of it, so the loop stays blocked until the rewind goes through
	if boot.pendingV3Realign {
		boot.realignAfterV3RollBack()
		if boot.pendingV3Realign {
			boot.invalidateNodeState()
			return ErrExecutionRealignPending
		}
		return nil
	}

	// one round snapshot for the backstops and the state computation: a state computed across a
	// mid-tick round change stays attributed to the evaluated round, which GetNodeState rejects
	evaluationRound := boot.roundHandler.Index()

	// evaluated before the node state is computed: the tip may be dead while the node reads as
	// synchronized, and a round in which a backstop fires must never publish a synchronized state
	if boot.tryReconcileEquivocation(evaluationRound) {
		boot.invalidateNodeState()
		return nil
	}

	if boot.tryReconcileDivergence(evaluationRound) {
		boot.invalidateNodeState()
		return nil
	}

	boot.computeNodeState(evaluationRound)
	boot.clearRecoveryAfterProgress()
	boot.evaluateFastRecovery(evaluationRound)

	nodeState := boot.GetNodeState()

	if nodeState != common.NsNotSynchronized {
		err := boot.prepareForSyncAtBoostrapIfNeeded()
		if err != nil {
			return err
		}

		boot.preparedForSync = false // reset the state for next loop
		return nil
	}

	defer boot.invalidateNodeState()

	if boot.forkInfo.IsDetected {
		boot.statusHandler.Increment(common.MetricNumTimesInForkChoice)

		if boot.isForcedRollBackOneBlock() {
			log.Debug("roll back one block has been forced")
			boot.rollBackOneBlockForced()
			return nil
		}

		if boot.isForcedRollBackToNonce() {
			log.Debug("roll back to nonce has been forced", "nonce", boot.forkInfo.Nonce)
			boot.rollBackToNonceForced()
			return nil
		}

		log.Debug("fork detected",
			"nonce", boot.forkInfo.Nonce,
			"hash", boot.forkInfo.Hash,
		)
		err := boot.rollBack(true)
		if err != nil {
			return err
		}
	}
	if boot.hasUnresolvedNotarizedAmbiguity() {
		return errBranchAwareSyncRetry
	}

	var body data.BodyHandler
	var header data.HeaderHandler
	var err error

	defer func() {
		if err != nil {
			log.Debug("sync block failed", "error", err)

			boot.doJobOnSyncBlockFail(body, header, err)
		}
	}()

	var headerHash []byte
	header, headerHash, err = boot.getNextHeaderRequestingIfMissing()
	if err != nil {
		return err
	}

	go boot.requestHeadersFromNonceIfMissing(header.GetNonce() + 1)

	body, err = boot.blockBootstrapper.getBlockBodyRequestingIfMissing(header)
	if err != nil {
		return err
	}

	if header.IsHeaderV3() {
		// update err to enable the deferred treatment
		err = boot.syncBlockV3(body, header, headerHash)
		return err
	}

	// update err to enable the deferred treatment
	err = boot.syncBlockLegacy(body, header)

	return err
}

// syncBlockLegacy method actually does the synchronization. It requests the next block header from the pool
// and if it is not found there it will be requested from the network. After the header is received,
// it requests the block body in the same way(pool and then, if it is not found in the pool, from network).
// If either header and body are received the ProcessBlock and CommitBlock method will be called successively.
// These methods will execute the block and its transactions. Finally, if everything works, the block will be committed
// in the blockchain, and all this mechanism will be reiterated for the next block.
func (boot *baseBootstrap) syncBlockLegacy(body data.BodyHandler, header data.HeaderHandler) error {
	err := boot.prepareForLegacySyncIfNeeded()
	if err != nil {
		return err
	}

	startTime := time.Now()
	waitTime := boot.getProcessWaitTime(header.GetRound())
	haveTime := func() time.Duration {
		return waitTime - time.Since(startTime)
	}

	startProcessBlockTime := time.Now()
	err = boot.blockProcessor.ProcessBlock(header, body, haveTime)
	elapsedTime := time.Since(startProcessBlockTime)
	log.Debug("elapsed time to process block",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	startProcessScheduledBlockTime := time.Now()
	err = boot.blockProcessor.ProcessScheduledBlock(header, body, haveTime)
	elapsedTime = time.Since(startProcessScheduledBlockTime)
	log.Debug("elapsed time to process scheduled block",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	startCommitBlockTime := time.Now()
	err = boot.blockProcessor.CommitBlock(header, body)
	elapsedTime = time.Since(startCommitBlockTime)
	if elapsedTime >= common.CommitMaxTime {
		log.Warn("syncBlock.CommitBlock", "elapsed time", elapsedTime)
	} else {
		log.Debug("elapsed time to commit block",
			"time [s]", elapsedTime,
		)
	}
	if err != nil {
		return err
	}

	log.Debug("block has been synced successfully",
		"nonce", header.GetNonce(),
	)

	boot.cleanNoncesSyncedWithErrorsBehindFinal()
	boot.clearBlockingUnprovenHdrTracker()
	boot.cleanProofsBehindFinal(header)

	return nil
}

func (boot *baseBootstrap) prepareForLegacySyncIfNeeded() error {
	if boot.preparedForSync {
		return nil
	}

	currentHeader := boot.getCurrentBlock()
	currentRootHash := boot.getCurrentRootHashLegacy()
	txPool := boot.poolsHolder.Transactions()
	err := txPool.OnExecutedBlock(currentHeader, currentRootHash)
	if err != nil {
		txPool.ResetTracker()
		return err
	}

	boot.preparedForSync = true

	return nil
}

// syncBlockV3 method actually does the synchronization. It requests the next block header from the pool
// and if it is not found there it will be requested from the network. After the header is received,
// it requests the block body in the same way(pool and then, if it is not found in the pool, from network).
// Once received, the header is verified through VerifyBlockProposal, but not before warming up the tx pool.
// Finally, if everything works, the block will be committed and added into the processing queue.
// And all this mechanism will be reiterated for the next block.
func (boot *baseBootstrap) syncBlockV3(body data.BodyHandler, header data.HeaderHandler, headerHash []byte) error {
	err := boot.prepareForSyncIfNeeded(header.GetNonce())
	if err != nil {
		return err
	}

	startTime := time.Now()
	waitTime := boot.getProcessWaitTime(header.GetRound())
	haveTime := func() time.Duration {
		return waitTime - time.Since(startTime)
	}

	startVerifyBlockTime := time.Now()
	err = boot.blockProcessor.VerifyBlockProposal(header, body, haveTime)
	elapsedTime := time.Since(startVerifyBlockTime)
	log.Debug("elapsed time to verify block",
		"time [s]", elapsedTime,
		"nonce", header.GetNonce(),
	)
	if err != nil {
		return err
	}

	err = boot.executionManager.AddPairForExecution(cache.HeaderBodyPair{
		Header:     header,
		Body:       body,
		HeaderHash: headerHash,
	})
	if err != nil {
		return err
	}

	startCommitBlockTime := time.Now()
	err = boot.blockProcessor.CommitBlock(header, body)
	elapsedTime = time.Since(startCommitBlockTime)
	if elapsedTime >= common.CommitMaxTime {
		log.Warn("syncBlock.CommitBlock", "elapsed time", elapsedTime)
	} else {
		log.Debug("elapsed time to commit block",
			"time [s]", elapsedTime,
			"nonce", header.GetNonce(),
		)
	}
	if err != nil {
		return err
	}

	log.Debug("block has been synced successfully",
		"nonce", header.GetNonce(),
	)

	boot.cleanNoncesSyncedWithErrorsBehindFinal()
	boot.clearBlockingUnprovenHdrTracker()
	boot.cleanProofsBehindFinal(header)

	return nil
}

// getMiniBlocksToSync will check already synced miniblocks and return only miniblocks that are not in pool
func (boot *baseBootstrap) getMiniBlocksToSync(
	miniBlocks []data.MiniBlockHeaderHandler,
) []data.MiniBlockHeaderHandler {
	miniBlocksToSync := make([]data.MiniBlockHeaderHandler, 0)

	for _, mb := range miniBlocks {
		_, ok := boot.dataPool.MiniBlocks().Get(mb.GetHash())
		if ok {
			continue
		}

		miniBlocksToSync = append(miniBlocksToSync, mb)
	}

	return miniBlocksToSync
}

func (boot *baseBootstrap) syncMiniBlocksAndTxsForHeader(
	header data.HeaderHandler,
) (func(), error) {
	miniBlocksToSync := boot.getMiniBlocksToSync(header.GetMiniBlockHeaderHandlers())

	boot.miniBlocksSyncer.ClearFields()
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeToWaitForRequestedData)
	err := boot.miniBlocksSyncer.SyncPendingMiniBlocks(miniBlocksToSync, ctx)
	cancel()
	if err != nil {
		return func() {}, err
	}

	miniBlocks, err := boot.miniBlocksSyncer.GetMiniBlocks()
	if err != nil {
		return func() {}, err
	}

	bodyHandler, err := boot.blockBootstrapper.getBlockBody(header)
	if err != nil {
		return func() {}, err
	}
	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return func() {}, process.ErrWrongTypeAssertion
	}

	releaseTxProtection := func() {}
	if header.GetShardID() != core.MetachainShardId {
		if protector, ok := boot.dataPool.Transactions().(currentBlockTxProtector); ok {
			releases := make([]func(), 0, len(body.MiniBlocks))
			for _, miniBlock := range body.MiniBlocks {
				if miniBlock.Type != block.TxBlock && miniBlock.Type != block.InvalidBlock {
					continue
				}

				cacheID := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
				releases = append(releases, protector.ProtectSetOfDataAgainstEvictionForCurrentBlock(miniBlock.TxHashes, cacheID))
			}
			releaseTxProtection = func() {
				for idx := len(releases) - 1; idx >= 0; idx-- {
					releases[idx]()
				}
			}
		}
	}

	// sync all txs into pools

	boot.txSyncer.ClearFields()
	ctx, cancel = context.WithTimeout(context.Background(), defaultTimeToWaitForRequestedData)
	err = boot.txSyncer.SyncTransactionsFor(miniBlocks, header.GetEpoch(), ctx)
	cancel()
	if err != nil {
		releaseTxProtection()
		return func() {}, err
	}

	return releaseTxProtection, nil
}

func (boot *baseBootstrap) prepareForSyncIfNeeded(
	syncingNonce uint64,
) error {
	if boot.preparedForSync {
		return nil
	}

	currentHeader := boot.getCurrentBlock()
	currentHeaderHash := boot.getCurrentBlockHash()
	lastExecResultNonce, lastExecResultHash, err := boot.getExecutionResultHeaderNonceForSyncStart(syncingNonce, currentHeader, currentHeaderHash)
	if err != nil {
		return err
	}

	if currentHeader.GetNonce() <= lastExecResultNonce {
		boot.preparedForSync = true
		return nil
	}

	// Walk backward from currentHeader following PrevHash pointers to collect
	// the canonical chain of committed headers between the last execution result
	// and the syncing header. Hash-based lookups are used instead of nonce-based
	// pool lookups to avoid ambiguity when multiple headers exist for the same nonce.
	type backfillEntry struct {
		header     data.HeaderHandler
		headerHash []byte
	}

	headersToAdd := make([]backfillEntry, 0, currentHeader.GetNonce()-lastExecResultNonce)
	walker := currentHeader
	walkerHash := currentHeaderHash

	for walker.GetNonce() > lastExecResultNonce {
		headersToAdd = append(headersToAdd, backfillEntry{
			header:     walker,
			headerHash: walkerHash,
		})

		if walker.GetNonce() == lastExecResultNonce+1 {
			if len(lastExecResultHash) > 0 && !bytes.Equal(walker.GetPrevHash(), lastExecResultHash) {
				return fmt.Errorf("%w: backfill chain at nonce %d has prevHash mismatch with last execution result hash",
					process.ErrBlockHashDoesNotMatch, walker.GetNonce())
			}
			break
		}

		prevHash := walker.GetPrevHash()
		prevHeader, errGetHdr := boot.getHeader(prevHash)
		if errGetHdr != nil {
			log.Debug("prepareForSyncIfNeeded: failed to get header by hash during backfill",
				"hash", prevHash,
				"expected nonce", walker.GetNonce()-1,
				"error", errGetHdr,
			)
			return errGetHdr
		}

		expectedNonce := walker.GetNonce() - 1
		if prevHeader.GetNonce() != expectedNonce {
			return fmt.Errorf("%w: backfill walk at nonce %d resolved prevHash to nonce %d, expected %d",
				process.ErrWrongNonceInBlock, walker.GetNonce(), prevHeader.GetNonce(), expectedNonce)
		}

		walker = prevHeader
		walkerHash = prevHash
	}

	// add headers for execution in forward (ascending nonce) order
	for i := len(headersToAdd) - 1; i >= 0; i-- {
		info := headersToAdd[i]

		releaseTxProtection, errSync := boot.syncMiniBlocksAndTxsForHeader(info.header)
		if errSync != nil {
			return errSync
		}

		err = func() error {
			defer releaseTxProtection()

			body, errGetBody := boot.blockBootstrapper.getBlockBody(info.header)
			if errGetBody != nil {
				return errGetBody
			}

			errSave := boot.saveProposedTxsToPool(info.header, body)
			if errSave != nil {
				return errSave
			}

			errBackfill := boot.blockProcessor.OnBackfilledBlock(
				body,
				info.header,
				info.headerHash,
			)
			if errBackfill != nil {
				return errBackfill
			}

			return boot.executionManager.AddPairForExecution(cache.HeaderBodyPair{
				Header:     info.header,
				Body:       body,
				HeaderHash: info.headerHash,
			})
		}()
		if err != nil {
			return err
		}
	}

	boot.preparedForSync = true

	return nil
}

func (boot *baseBootstrap) saveProposedTxsToPool(
	header data.HeaderHandler,
	body data.BodyHandler,
) error {
	if !header.IsHeaderV3() {
		return nil
	}

	bodyPtr, ok := body.(*block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	separatedBodies := process.SeparateBodyByType(bodyPtr)

	for blockType, blockBody := range separatedBodies {
		dataPool, err := process.GetDataPoolByBlockType(blockType, boot.dataPool)
		if err != nil {
			return err
		}

		unit, err := process.GetStorageUnitByBlockType(blockType)
		if err != nil {
			return err
		}

		storer, err := boot.store.GetStorer(unit)
		if err != nil {
			return err
		}

		for i := 0; i < len(blockBody.MiniBlocks); i++ {
			miniBlock := blockBody.MiniBlocks[i]
			err = boot.saveTxsToPool(dataPool, storer, miniBlock, blockType)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (boot *baseBootstrap) saveTxsToPool(
	dataPool dataRetriever.ShardedDataCacherNotifier,
	storer storage.Storer,
	miniBlock *block.MiniBlock,
	blockType block.Type,
) error {
	txHashes := miniBlock.TxHashes

	for _, txHash := range txHashes {
		// continue if already in pool
		_, ok := dataPool.SearchFirstData(txHash)
		if ok {
			continue
		}

		txBuff, err := storer.Get(txHash)
		if err != nil {
			return err
		}

		tx, err := boot.unmarshalTxByBlockType(blockType, txBuff)
		if err != nil {
			return err
		}

		cacherIdentifier := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
		dataPool.AddData(
			txHash,
			tx,
			tx.Size(),
			cacherIdentifier,
		)
	}

	return nil
}

func (boot *baseBootstrap) getExecutionResultHeaderNonceForSyncStart(
	syncingNonce uint64,
	currentHeader data.HeaderHandler,
	currentHeaderHash []byte,
) (uint64, []byte, error) {
	lastNotarizedExecResult, err := process.GetPrevBlockLastExecutionResult(boot.chainHandler)
	if err != nil {
		return 0, nil, err
	}

	lastNotarizedExecResultsHandler, err := common.ExtractBaseExecutionResultHandler(lastNotarizedExecResult)
	if err != nil {
		return 0, nil, err
	}

	log.Debug("getExecutionResultHeaderNonceForSyncStart",
		"syncingNonce", syncingNonce,
		"currHeader nonce", currentHeader.GetNonce(),
		"currHeader hash", currentHeaderHash,
		"lastNotarizedExecRes nonce", lastNotarizedExecResultsHandler.GetHeaderNonce(),
		"lastNotarizedExecRes hash", lastNotarizedExecResultsHandler.GetHeaderHash(),
		"lastNotarizedExecRes rootHash", lastNotarizedExecResultsHandler.GetRootHash(),
	)

	lastNotarizedExecutedHash := lastNotarizedExecResultsHandler.GetHeaderHash()
	lastNotarizedExecutedHeader, err := boot.getHeader(lastNotarizedExecutedHash)
	if err != nil {
		return 0, nil, err
	}

	rootHash := lastNotarizedExecResultsHandler.GetRootHash()

	txPool := boot.poolsHolder.Transactions()
	err = txPool.OnExecutedBlock(lastNotarizedExecutedHeader, rootHash)
	if err != nil {
		txPool.ResetTracker()
		// the emptied tracker must be rebuilt from the notarized anchor: the realign drops the
		// pending execution results, else the backfill starts above the anchor and leaves a gap
		boot.pendingV3Realign = true
		return 0, nil, err
	}

	lastExecutionResultNonce := lastNotarizedExecutedHeader.GetNonce()
	defer func() {
		log.Debug("getExecutionResultHeaderNonceForSyncStart", "lastExecutionResultNonce", lastExecutionResultNonce)
	}()

	// check with pending execution
	pendingExecutionResults, err := boot.executionManager.GetPendingExecutionResults()
	if err != nil {
		return 0, nil, err
	}
	var pendingExecutionResult data.BaseExecutionResultHandler
	for idx := len(pendingExecutionResults) - 1; idx >= 0; idx-- {
		pendingExecutionResult = pendingExecutionResults[idx]
		if pendingExecutionResult.GetHeaderNonce() <= lastExecutionResultNonce {
			log.Warn("getExecutionResultHeaderNonceForSyncStart found pending execution result with lower or equal nonce than last executed",
				"pending nonce", pendingExecutionResult.GetHeaderNonce(),
				"lastExecutionResultNonce", lastExecutionResultNonce,
			)
			continue
		}

		if boot.hasProofInCacheOrStorage(pendingExecutionResult.GetHeaderHash()) {
			return pendingExecutionResult.GetHeaderNonce(), pendingExecutionResult.GetHeaderHash(), nil
		}
	}

	return lastExecutionResultNonce, lastNotarizedExecutedHash, nil
}

func (boot *baseBootstrap) hasProofInCacheOrStorage(hash []byte) bool {
	if boot.proofs.HasProof(boot.shardCoordinator.SelfId(), hash) {
		return true
	}

	proofsStorer, errGetStorer := boot.store.GetStorer(dataRetriever.ProofsUnit)
	if errGetStorer != nil {
		return false
	}

	proofBytes, err := proofsStorer.Get(hash)
	if err != nil {
		return false
	}

	proof := &block.HeaderProof{}
	err = boot.marshalizer.Unmarshal(proof, proofBytes)
	if err != nil {
		// return true here, since the proof exists in storer
		log.Warn("hasProofInCacheOrStorage invalid proof in storage", "error", err.Error(), "hash", hash)
		return true
	}

	boot.proofs.AddProof(proof)

	return true
}

func (boot *baseBootstrap) unmarshalTxByBlockType(
	blockType block.Type,
	txBuff []byte,
) (txSizeHandler, error) {
	var tx txSizeHandler
	var err error

	switch blockType {
	case block.TxBlock, block.InvalidBlock:
		tx = &transaction.Transaction{}
	case block.SmartContractResultBlock:
		tx = &smartContractResult.SmartContractResult{}
	case block.RewardsBlock:
		tx = &rewardTx.RewardTx{}
	case block.PeerBlock:
		tx = &state.ShardValidatorInfo{}
	default:
		return nil, fmt.Errorf("unsupported block type: %d", blockType)
	}

	err = boot.marshalizer.Unmarshal(tx, txBuff)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (boot *baseBootstrap) handleTrieSyncError(err error, ctx context.Context) {
	shouldOutputLog := err != nil && !common.IsContextDone(ctx)
	if shouldOutputLog {
		log.Debug("SyncBlock syncTrie", "error", err)
	}
}

func (boot *baseBootstrap) syncUserAccountsState(key []byte) error {
	log.Warn("base sync: started syncUserAccountsState")
	return boot.accountsDBSyncer.SyncAccounts(key, storageMarker.NewDisabledStorageMarker())
}

func (boot *baseBootstrap) cleanNoncesSyncedWithErrorsBehindFinal() {
	boot.mutNonceSyncedWithErrors.Lock()
	defer boot.mutNonceSyncedWithErrors.Unlock()

	finalNonce := boot.forkDetector.GetHighestFinalBlockNonce()
	for nonce := range boot.mapNonceSyncedWithErrors {
		if nonce < finalNonce {
			delete(boot.mapNonceSyncedWithErrors, nonce)
		}
	}

	for nonce := range boot.mapNonceRecoveryAttempts {
		if nonce < finalNonce {
			delete(boot.mapNonceRecoveryAttempts, nonce)
		}
	}
}

func (boot *baseBootstrap) cleanProofsBehindFinal(header data.HeaderHandler) {
	if !boot.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, header.GetEpoch()) {
		return
	}

	finalNonce := boot.forkDetector.GetHighestFinalBlockNonce()

	err := boot.proofs.CleanupProofsBehindNonce(header.GetShardID(), finalNonce)
	if err != nil {
		log.Warn("failed to cleanup notarized proofs behind nonce",
			"nonce", finalNonce,
			"shardID", header.GetShardID(),
			"error", err)
	}

	log.Trace("baseBootstrap.cleanProofsBehindFinal cleanup successfully", "finalNonce", finalNonce)
}

// rollBack decides if rollBackOneBlock must be called
func (boot *baseBootstrap) rollBack(revertUsingForkNonce bool) (err error) {
	var roleBackOneBlockExecuted bool
	var currHeaderHash []byte
	var currHeader data.HeaderHandler
	var prevHeader data.HeaderHandler
	var currBody data.BodyHandler

	defer func() {
		isHeaderV3 := !check.IfNil(currHeader) && currHeader.IsHeaderV3()
		if !roleBackOneBlockExecuted && !isHeaderV3 {
			errScheduled := boot.scheduledTxsExecutionHandler.RollBackToBlock(currHeaderHash)
			if errScheduled != nil {
				rootHash := boot.chainHandler.GetGenesisHeader().GetRootHash()
				if currHeader != nil {
					rootHash = currHeader.GetRootHash()
				}
				scheduledInfo := &process.ScheduledInfo{
					RootHash:        rootHash,
					IntermediateTxs: make(map[block.Type][]data.TransactionHandler),
					GasAndFees:      process.GetZeroGasAndFees(),
					MiniBlocks:      make(block.MiniBlockSlice, 0),
				}
				boot.scheduledTxsExecutionHandler.SetScheduledInfo(scheduledInfo)
			}
		}
	}()

	rolledBackV3 := false
	// runs on every exit path: a v3 tip lowered by even one block must never keep a stale
	// watermark; a rewind failure surfaces in the result, so no caller continues on top of it
	defer func() {
		if !rolledBackV3 {
			return
		}
		boot.realignAfterV3RollBack()
		if err == nil && boot.pendingV3Realign {
			err = ErrExecutionRealignPending
		}
	}()

	log.Debug("starting roll back")
	for {
		currHeaderHash = boot.chainHandler.GetCurrentBlockHeaderHash()
		currHeader, err = boot.blockBootstrapper.getCurrHeader()
		if err != nil {
			return err
		}

		allowRollBack := boot.shouldAllowRollback(currHeader, currHeaderHash)
		// a header v3 switch must never cross the final checkpoint, not even fork-driven
		isRollBackDenied := !allowRollBack && (!revertUsingForkNonce || currHeader.IsHeaderV3())
		if isRollBackDenied {
			return ErrRollBackBehindFinalHeader
		}

		shouldEndRollBack := revertUsingForkNonce && currHeader.GetNonce() < boot.forkInfo.Nonce
		if shouldEndRollBack {
			return ErrRollBackBehindForkNonce
		}

		prevHeaderHash := currHeader.GetPrevHash()
		prevHeader, err = boot.blockBootstrapper.getPrevHeader(currHeader, boot.headerStore)
		if err != nil {
			return err
		}

		log.Debug("roll back to block",
			"nonce", currHeader.GetNonce()-1,
			"hash", currHeader.GetPrevHash(),
		)
		log.Debug("highest final block nonce",
			"nonce", boot.forkDetector.GetHighestFinalBlockNonce(),
		)

		if currHeader.IsHeaderV3() {
			err = boot.checkRollBackExecutionBase(prevHeader)
			if err != nil {
				return err
			}

			currBody, err = boot.rollBackOneBlockV3(
				currHeaderHash,
				currHeader,
				prevHeaderHash,
				prevHeader,
			)
			// sticky across iterations: any completed restore phase moved the tip, so the
			// realign must run even when a later iteration fails before moving anything
			if err == nil || (boot.pendingV3RollBack != nil && boot.pendingV3RollBack.restoreDone) {
				rolledBackV3 = true
			}
		} else {
			currBody, err = boot.rollBackOneBlock(
				currHeaderHash,
				currHeader,
				prevHeaderHash,
				prevHeader,
			)
			roleBackOneBlockExecuted = true
		}
		if err != nil {
			return err
		}

		_, _ = metricsLoader.UpdateMetricsFromStorage(boot.store, boot.uint64Converter, boot.marshalizer, boot.statusHandler, prevHeader.GetNonce())

		err = boot.bootStorer.SaveLastRound(int64(prevHeader.GetRound()))
		if err != nil {
			log.Debug("save last round in storage",
				"error", err.Error(),
				"round", prevHeader.GetRound(),
			)
		}

		err = boot.historyRepo.RevertBlock(currHeader, currBody)
		if err != nil {
			log.Debug("boot.historyRepo.RevertBlock",
				"error", err.Error(),
			)

			return err
		}

		if !currHeader.IsHeaderV3() {
			err = boot.scheduledTxsExecutionHandler.RollBackToBlock(prevHeaderHash)
			if err != nil {
				scheduledInfo := &process.ScheduledInfo{
					RootHash:        prevHeader.GetRootHash(),
					IntermediateTxs: make(map[block.Type][]data.TransactionHandler),
					GasAndFees:      process.GetZeroGasAndFees(),
					MiniBlocks:      make(block.MiniBlockSlice, 0),
				}
				boot.scheduledTxsExecutionHandler.SetScheduledInfo(scheduledInfo)
			}
		}

		err = boot.outportHandler.RevertIndexedBlock(&outportcore.HeaderDataWithBody{
			Body:       currBody,
			HeaderHash: currHeaderHash,
			Header:     currHeader,
		})
		if err != nil {
			log.Warn("baseBootstrap.outportHandler.RevertIndexedBlock cannot revert indexed block", "error", err)
		}

		shouldAddHeaderToBlackList := revertUsingForkNonce && boot.blockBootstrapper.isForkTriggeredByMeta()
		if shouldAddHeaderToBlackList {
			process.AddHeaderToBlackList(boot.blackListHandler, currHeaderHash)
		}

		shouldContinueRollBack := revertUsingForkNonce && currHeader.GetNonce() > boot.forkInfo.Nonce
		if shouldContinueRollBack {
			continue
		}

		break
	}

	log.Debug("ending roll back")
	return nil
}

// completeInterruptedV3RollBack finishes or abandons a roll back interrupted mid-way; retried
// steps are idempotent or guarded, so re-driving converges
func (boot *baseBootstrap) completeInterruptedV3RollBack() error {
	pending := boot.pendingV3RollBack

	if !pending.restoreDone {
		// the restore failed atomically, so nothing needs undoing: the roll back is simply
		// dropped once the block is confirmed to stay, by a commit on top or by its own proof
		currentHash := boot.chainHandler.GetCurrentBlockHeaderHash()
		isSuperseded := !bytes.Equal(currentHash, pending.currHeaderHash)
		isFinal := pending.currHeader.GetNonce() <= boot.forkDetector.GetHighestFinalBlockNonce()
		if isSuperseded || isFinal {
			// the block stays committed, so anything a failed restore moved out of committed
			// storage must be written back before the roll back may be dropped
			if !boot.repairPartialRestore(pending) {
				boot.invalidateNodeState()
				return ErrPendingStorageRepair
			}

			log.Error("abandoning the interrupted v3 roll back, the block stays committed",
				"hash", pending.currHeaderHash,
				"nonce", pending.currHeader.GetNonce(),
				"superseded", isSuperseded,
				"final", isFinal,
			)
			boot.pendingV3RollBack = nil
			return nil
		}

		err := boot.rollBack(false)
		if err != nil {
			boot.invalidateNodeState()
			return err
		}
		return nil
	}

	siblingCommitted, err := boot.finishRollBackOneBlockV3(pending)
	if err != nil {
		boot.invalidateNodeState()
		return err
	}

	boot.postRollBackBookkeeping(pending, !siblingCommitted)

	boot.realignAfterV3RollBack()
	if boot.pendingV3Realign {
		boot.invalidateNodeState()
		return ErrExecutionRealignPending
	}

	return nil
}

// postRollBackBookkeeping mirrors the roll back loop tail for a block whose roll back was
// completed outside the loop; failures here are node-local records, not chain state
func (boot *baseBootstrap) postRollBackBookkeeping(pending *pendingV3RollBack, updateLastRound bool) {
	if updateLastRound {
		err := boot.bootStorer.SaveLastRound(int64(pending.prevHeader.GetRound()))
		if err != nil {
			log.Debug("save last round in storage",
				"error", err.Error(),
				"round", pending.prevHeader.GetRound(),
			)
		}
	}

	err := boot.historyRepo.RevertBlock(pending.currHeader, pending.currBody)
	if err != nil {
		log.Warn("postRollBackBookkeeping: cannot revert history for the rolled back block",
			"hash", pending.currHeaderHash,
			"error", err,
		)
	}

	err = boot.outportHandler.RevertIndexedBlock(&outportcore.HeaderDataWithBody{
		Body:       pending.currBody,
		HeaderHash: pending.currHeaderHash,
		Header:     pending.currHeader,
	})
	if err != nil {
		log.Warn("baseBootstrap.outportHandler.RevertIndexedBlock cannot revert indexed block", "error", err)
	}
}

// realignAfterV3RollBack rewinds the execution results state to the rolled-back tip and re-arms
// the sync prepare step; a failed rewind arms a mandatory retry that blocks the sync loop
func (boot *baseBootstrap) realignAfterV3RollBack() {
	boot.pendingV3Realign = false
	newTip := boot.chainHandler.GetCurrentBlockHeader()
	if check.IfNil(newTip) || !newTip.IsHeaderV3() {
		return
	}

	err := boot.executionManager.RewindExecutionStateToTip(newTip)
	if err != nil {
		boot.pendingV3Realign = true
		log.Warn("realignAfterV3RollBack: cannot rewind execution state, sync blocked until retried",
			"tip nonce", newTip.GetNonce(),
			"error", err,
		)
		return
	}

	boot.poolsHolder.Transactions().ResetTracker()
	boot.preparedForSync = false
	boot.resetSyncedWithErrorsForNonce(boot.getNonceForNextBlock())
}

func (boot *baseBootstrap) shouldAllowRollback(currHeader data.HeaderHandler, currHeaderHash []byte) bool {
	if check.IfNil(currHeader) {
		return false
	}
	if currHeader.IsHeaderV3() {
		return boot.shouldAllowRollbackV3(currHeader)
	}

	finalBlockNonce := boot.forkDetector.GetHighestFinalBlockNonce()
	finalBlockHash := boot.forkDetector.GetHighestFinalBlockHash()
	isRollBackBehindFinal := currHeader.GetNonce() <= finalBlockNonce
	isFinalBlockRollBack := currHeader.GetNonce() == finalBlockNonce
	canRollbackBlock := boot.canRollbackBlock(currHeader)

	headerWithScheduledMiniBlocks := currHeader.HasScheduledMiniBlocks()
	headerHashDoesNotMatchWithFinalBlockHash := !bytes.Equal(currHeaderHash, finalBlockHash)
	allowFinalBlockRollBack := (headerWithScheduledMiniBlocks || headerHashDoesNotMatchWithFinalBlockHash) && isFinalBlockRollBack && canRollbackBlock
	allowRollBack := !isRollBackBehindFinal || allowFinalBlockRollBack

	log.Debug("baseBootstrap.shouldAllowRollback",
		"isRollBackBehindFinal", isRollBackBehindFinal,
		"isFinalBlockRollBack", isFinalBlockRollBack,
		"headerWithScheduledMiniBlocks", headerWithScheduledMiniBlocks,
		"headerHashDoesNotMatchWithFinalBlockHash", headerHashDoesNotMatchWithFinalBlockHash,
		"allowFinalBlockRollBack", allowFinalBlockRollBack,
		"canRollbackBlock", canRollbackBlock,
		"allowRollBack", allowRollBack,
	)

	return allowRollBack
}

// shouldAllowRollbackV3 allows replacing a committed block only while it is not final;
// the state is never reverted through tries, the adopted sibling re-executes asynchronously
func (boot *baseBootstrap) shouldAllowRollbackV3(currHeader data.HeaderHandler) bool {
	finalBlockNonce := boot.forkDetector.GetHighestFinalBlockNonce()
	allowRollBack := currHeader.GetNonce() > finalBlockNonce

	log.Debug("baseBootstrap.shouldAllowRollbackV3",
		"nonce", currHeader.GetNonce(),
		"final block nonce", finalBlockNonce,
		"allowRollBack", allowRollBack,
	)

	return allowRollBack
}

func (boot *baseBootstrap) canRollbackBlock(currHeader data.HeaderHandler) bool {
	firstCommittedNonce := boot.blockProcessor.NonceOfFirstCommittedBlock()

	return currHeader.GetNonce() >= firstCommittedNonce.Value && firstCommittedNonce.HasValue
}

// checkRollBackExecutionBase refuses a roll back whose post-rollback execution base state cannot
// be recreated: proceeding would strand the node unable to execute forward or roll back further
func (boot *baseBootstrap) checkRollBackExecutionBase(prevHeader data.HeaderHandler) error {
	if !prevHeader.IsHeaderV3() {
		return nil
	}

	lastExecResult, err := common.GetLastBaseExecutionResultHandler(prevHeader)
	if err != nil {
		return err
	}

	rootHash := lastExecResult.GetRootHash()
	_, err = boot.accounts.GetTrie(rootHash)
	if err != nil {
		boot.statusHandler.Increment(common.MetricNumRollBacksRefusedMissingState)
		log.Error("roll back refused: execution base state is missing",
			"nonce", prevHeader.GetNonce(),
			"root hash", rootHash,
			"error", err,
		)
		return ErrRollBackExecutionBaseMissing
	}

	return nil
}

func (boot *baseBootstrap) rollBackOneBlock(
	currHeaderHash []byte,
	currHeader data.HeaderHandler,
	prevHeaderHash []byte,
	prevHeader data.HeaderHandler,
) (data.BodyHandler, error) {

	var err error

	prevHeaderRootHash := boot.getRootHashFromBlock(prevHeader, prevHeaderHash)
	currHeaderRootHash := boot.getRootHashFromBlock(currHeader, currHeaderHash)

	defer func() {
		if err != nil {
			boot.restoreState(currHeaderHash, currHeader, currHeaderRootHash)
		}
	}()

	if currHeader.GetNonce() > 1 {
		err = boot.setCurrentBlockInfo(prevHeaderHash, prevHeader, prevHeaderRootHash)
		if err != nil {
			return nil, err
		}
	} else {
		err = boot.setCurrentBlockInfo(nil, nil, nil)
		if err != nil {
			return nil, err
		}
	}

	err = boot.blockProcessor.RevertStateToBlock(prevHeader, prevHeaderRootHash)
	if err != nil {
		return nil, err
	}

	boot.blockProcessor.PruneStateOnRollback(currHeader, currHeaderHash, prevHeader, prevHeaderHash)

	currBlockBody, errNotCritical := boot.blockBootstrapper.getBlockBody(currHeader)
	if errNotCritical != nil {
		log.Debug("rollBackOneBlock getBlockBody error", "error", errNotCritical)
	}

	err = boot.blockProcessor.RestoreBlockIntoPools(currHeader, currBlockBody)
	if err != nil {
		return nil, err
	}

	boot.cleanCachesAndStorageOnRollback(currHeader)

	return currBlockBody, nil
}

// rollBackOneBlockV3 reverts a committed, not yet final header so a same-nonce sibling can be
// adopted; the trie state is not reverted, the sibling's execution results are produced async
func (boot *baseBootstrap) rollBackOneBlockV3(
	currHeaderHash []byte,
	currHeader data.HeaderHandler,
	prevHeaderHash []byte,
	prevHeader data.HeaderHandler,
) (data.BodyHandler, error) {
	currBlockBody, errNotCritical := boot.blockBootstrapper.getBlockBody(currHeader)
	if errNotCritical != nil {
		log.Debug("rollBackOneBlockV3 getBlockBody error", "error", errNotCritical)
	}

	prior := boot.pendingV3RollBack
	boot.pendingV3RollBack = &pendingV3RollBack{
		currHeaderHash: currHeaderHash,
		currHeader:     currHeader,
		prevHeaderHash: prevHeaderHash,
		prevHeader:     prevHeader,
		currBody:       currBlockBody,
	}
	// an earlier attempt may have left moved blocks unrepaired; the obligation survives the re-drive
	if prior != nil && bytes.Equal(prior.currHeaderHash, currHeaderHash) {
		boot.pendingV3RollBack.unrestoredMetaBlocks = prior.unrestoredMetaBlocks
	}

	// restore before the tip moves: roll backs run only while consensus is idle, so no commit
	// races them; a failed restore mutates nothing durable unless it says so, and can be retried
	err := boot.blockProcessor.RestoreBlockIntoPools(currHeader, currBlockBody)
	if err != nil {
		boot.recordPartialRestore(err)
		return nil, err
	}
	boot.pendingV3RollBack.restoreDone = true

	_, err = boot.finishRollBackOneBlockV3(boot.pendingV3RollBack)
	if err != nil {
		return nil, err
	}

	return currBlockBody, nil
}

// finishRollBackOneBlockV3 runs the steps after a completed restore, each idempotent or guarded
// so a re-driven run converges; returns true when it stood down to a sibling committed meanwhile
func (boot *baseBootstrap) finishRollBackOneBlockV3(pending *pendingV3RollBack) (bool, error) {
	currentHash := boot.chainHandler.GetCurrentBlockHeaderHash()
	if bytes.Equal(currentHash, pending.currHeaderHash) {
		err := boot.chainHandler.SetCurrentBlockHeaderAndHash(pending.prevHeaderHash, pending.prevHeader)
		if err != nil {
			return false, err
		}
	} else if !bytes.Equal(currentHash, pending.prevHeaderHash) {
		return true, boot.finishRollBackV3AfterSiblingCommit(pending)
	}
	boot.updateSupernovaTransitionReadiness(pending.prevHeader, pending.prevHeaderHash)

	if !pending.executionPruned {
		err := boot.executionManager.RemoveAtNonceAndHigher(pending.currHeader.GetNonce())
		if err != nil {
			return false, err
		}
		pending.executionPruned = true
	}

	// no-op unless the roll back crosses the recorded epoch start; kept after every fallible
	// step, since nothing can restore a reverted trigger on an abort
	err := boot.epochStartTrigger.RevertStateToBlock(pending.prevHeader)
	if err != nil {
		return false, err
	}

	hash := boot.removeHeaderFromPools(pending.currHeader)
	boot.forkDetector.RemoveCommittedHeader(pending.currHeader.GetNonce(), hash)
	nonceToByteSlice := boot.uint64Converter.ToByteSlice(pending.currHeader.GetNonce())
	_ = boot.headerNonceHashStore.Remove(nonceToByteSlice)
	boot.pendingV3RollBack = nil

	return false, nil
}

type supernovaTransitionReadinessUpdater interface {
	UpdateSupernovaTransitionReadiness(header data.HeaderHandler, headerHash []byte)
}

func (boot *baseBootstrap) updateSupernovaTransitionReadiness(header data.HeaderHandler, headerHash []byte) {
	updater, ok := boot.blockProcessor.(supernovaTransitionReadinessUpdater)
	if !ok {
		return
	}

	updater.UpdateSupernovaTransitionReadiness(header, headerHash)
}

// restoreWriteBackRetrier retries the storage write back of meta blocks moved by a failed restore
type restoreWriteBackRetrier interface {
	RetryRestoreWriteBack(movedMetaBlocks []process.MovedMetaBlock) []process.MovedMetaBlock
}

// repairPartialRestore writes the moved meta blocks of a failed restore back into committed
// storage; returns false while any of them is still out, keeping the roll back pending
func (boot *baseBootstrap) repairPartialRestore(pending *pendingV3RollBack) bool {
	if len(pending.unrestoredMetaBlocks) == 0 {
		return true
	}

	retrier, ok := boot.blockProcessor.(restoreWriteBackRetrier)
	if !ok {
		log.Error("repairPartialRestore: the block processor cannot retry the restore write back")
		return false
	}

	pending.unrestoredMetaBlocks = retrier.RetryRestoreWriteBack(pending.unrestoredMetaBlocks)
	if len(pending.unrestoredMetaBlocks) > 0 {
		log.Error("interrupted v3 roll back kept pending, moved meta blocks not yet written back",
			"num meta blocks", len(pending.unrestoredMetaBlocks),
			"hash", pending.currHeaderHash,
		)
		return false
	}

	return true
}

// recordPartialRestore keeps the meta blocks a failed restore could not write back, so an
// abandon can only follow a completed storage repair
func (boot *baseBootstrap) recordPartialRestore(err error) {
	var partialErr *process.PartialRestoreError
	if !errors.As(err, &partialErr) {
		return
	}

	pending := boot.pendingV3RollBack
	pending.unrestoredMetaBlocks = mergeMovedMetaBlocks(pending.unrestoredMetaBlocks, partialErr.UnrestoredMetaBlocks)
}

func mergeMovedMetaBlocks(existing []process.MovedMetaBlock, added []process.MovedMetaBlock) []process.MovedMetaBlock {
	for _, add := range added {
		found := false
		for _, have := range existing {
			if bytes.Equal(have.Hash, add.Hash) {
				found = true
				break
			}
		}
		if !found {
			existing = append(existing, add)
		}
	}

	return existing
}

// finishRollBackV3AfterSiblingCommit closes an interrupted roll back after a same-nonce sibling
// commit: trigger, execution results and nonce mapping belong to the sibling now
func (boot *baseBootstrap) finishRollBackV3AfterSiblingCommit(pending *pendingV3RollBack) error {
	newTip := boot.chainHandler.GetCurrentBlockHeader()
	isSibling := !check.IfNil(newTip) &&
		newTip.GetNonce() == pending.currHeader.GetNonce() &&
		bytes.Equal(newTip.GetPrevHash(), pending.prevHeaderHash)
	if !isSibling {
		log.Error("unexpected chain tip while completing an interrupted v3 roll back",
			"rolled back hash", pending.currHeaderHash,
			"rolled back nonce", pending.currHeader.GetNonce(),
			"tip hash", boot.chainHandler.GetCurrentBlockHeaderHash(),
		)
		return ErrInconsistentRollBackState
	}

	log.Warn("interrupted v3 roll back closed after a sibling commit",
		"rolled back hash", pending.currHeaderHash,
		"nonce", pending.currHeader.GetNonce(),
	)

	hash := boot.removeHeaderFromPools(pending.currHeader)
	boot.forkDetector.RemoveCommittedHeader(pending.currHeader.GetNonce(), hash)
	boot.pendingV3RollBack = nil

	return nil
}

func (boot *baseBootstrap) getRootHashFromBlock(hdr data.HeaderHandler, hdrHash []byte) []byte {
	hdrRootHash := hdr.GetRootHash()
	scheduledHdrRootHash, err := boot.scheduledTxsExecutionHandler.GetScheduledRootHashForHeader(hdrHash)
	if err == nil {
		hdrRootHash = scheduledHdrRootHash
	}

	return hdrRootHash
}

func (boot *baseBootstrap) getNextHeaderRequestingIfMissing() (data.HeaderHandler, []byte, error) {
	nonce := boot.getNonceForNextBlock()

	boot.setRequestedHeaderHash(nil)
	boot.setRequestedHeaderNonce(nil)

	var hash []byte
	isDirectedV3 := false
	selector, ok := boot.forkDetector.(notarizedHeaderSelector)
	if ok {
		selection := selector.getNotarizedHeaderSelection(nonce)
		hash = selection.hash
		isDirectedV3 = selection.isV3
		if len(selection.candidates) > 1 {
			round := int64(-1)
			if !check.IfNil(boot.roundHandler) {
				round = boot.roundHandler.Index()
			}
			boot.tryResolveNotarizedAmbiguity(round)
			selection = selector.getNotarizedHeaderSelection(nonce)
			hash = selection.hash
			isDirectedV3 = selection.isV3
			if len(selection.candidates) > 1 {
				return nil, nil, errBranchAwareSyncRetry
			}
		}
	} else {
		hash = boot.forkDetector.GetNotarizedHeaderHash(nonce)
		isDirectedV3 = len(hash) > 0 && boot.isAsyncExecutionEnabledForHash(hash)
	}
	if boot.forkInfo.IsDetected {
		// A unique V3 notarization takes precedence over the recovery hint.
		if !isDirectedV3 {
			hash = boot.forkInfo.Hash
			versionFound := false
			if ok && len(hash) > 0 {
				isDirectedV3, versionFound = selector.getHeaderVersion(nonce, hash)
			}
			if !versionFound {
				isDirectedV3 = len(hash) > 0 && boot.isAsyncExecutionEnabledForHash(hash)
			}
		}
	}

	selectedFromProof := false
	if !isDirectedV3 {
		proof, err := boot.proofs.GetProofByNonce(nonce, boot.shardCoordinator.SelfId())
		if err == nil {
			hash = proof.GetHeaderHash()
			selectedFromProof = true
		}
	}

	hash = boot.selectNonBlackListedHash(hash, nonce)

	if hash != nil {
		if selectedFromProof {
			return boot.getGenericProofHeaderRequestingIfMissing(nonce, hash)
		}

		header, err := boot.getHeaderWithHashRequestingIfMissing(hash)
		return header, hash, err
	}

	return boot.getHeaderWithNonceRequestingIfMissing(nonce)
}

func (boot *baseBootstrap) hasUnresolvedNotarizedAmbiguity() bool {
	if boot.shardCoordinator.SelfId() == core.MetachainShardId {
		return false
	}

	authority, ok := boot.forkDetector.(notarizedHeaderAuthority)
	return ok && authority.hasUnresolvedNotarizedAmbiguity()
}

func (boot *baseBootstrap) tryResolveNotarizedAmbiguity(round int64) bool {
	if boot.shardCoordinator.SelfId() == core.MetachainShardId {
		return false
	}

	authority, ok := boot.forkDetector.(notarizedHeaderAuthority)
	if !ok || !authority.hasUnresolvedNotarizedAmbiguity() {
		return false
	}

	selection, found := authority.getLowestAmbiguousNotarizedHeaderSelection()
	if !found || len(selection.candidates) < 2 {
		boot.clearAmbiguityRecovery(0)
		return false
	}
	if boot.settlementChecker == nil {
		return true
	}

	scanCursor, shouldEvaluate := boot.ambiguityRecoveryToEvaluate(selection.candidates[0].nonce, round)
	if !shouldEvaluate {
		return true
	}

	nonce := selection.candidates[0].nonce
	selectedHash := boot.settlementChecker.resolveNotarizedHeader(nonce, selection.candidates)
	if len(selectedHash) > 0 {
		switch authority.applyNotarizedHeaderSelection(nonce, selectedHash) {
		case notarizedHeaderApplied:
			boot.clearAmbiguityRecovery(nonce)
			return authority.hasUnresolvedNotarizedAmbiguity()
		case notarizedHeaderNeedsReconciliation:
			if boot.proofs.HasProof(boot.shardCoordinator.SelfId(), selectedHash) {
				boot.armAuthorityReconciliation(nonce, selectedHash, round)
			}
		}
	}

	boot.requestAmbiguousNotarizedCandidates(selection.candidates)
	_, _, nextCursor := boot.settlementChecker.prepareInclusionScan(scanCursor)
	boot.storeAmbiguityScanCursor(nonce, nextCursor)

	return true
}

func (boot *baseBootstrap) ambiguityRecoveryToEvaluate(nonce uint64, round int64) (uint64, bool) {
	boot.mutAmbiguity.Lock()
	defer boot.mutAmbiguity.Unlock()

	if boot.ambiguityRecovery.nonce != nonce {
		boot.ambiguityRecovery = ambiguityRecoveryState{
			nonce:              nonce,
			lastEvaluatedRound: -1,
		}
	}
	if boot.ambiguityRecovery.lastEvaluatedRound == round {
		return boot.ambiguityRecovery.scanCursor, false
	}

	boot.ambiguityRecovery.lastEvaluatedRound = round
	return boot.ambiguityRecovery.scanCursor, true
}

func (boot *baseBootstrap) storeAmbiguityScanCursor(nonce uint64, scanCursor uint64) {
	boot.mutAmbiguity.Lock()
	if boot.ambiguityRecovery.nonce == nonce {
		boot.ambiguityRecovery.scanCursor = scanCursor
	}
	boot.mutAmbiguity.Unlock()
}

func (boot *baseBootstrap) clearAmbiguityRecovery(nonce uint64) {
	boot.mutAmbiguity.Lock()
	if nonce == 0 || boot.ambiguityRecovery.nonce == nonce {
		boot.ambiguityRecovery = ambiguityRecoveryState{}
	}
	boot.mutAmbiguity.Unlock()
}

func (boot *baseBootstrap) requestAmbiguousNotarizedCandidates(candidates []notarizedHeaderCandidate) {
	shardID := boot.shardCoordinator.SelfId()
	if shardID == core.MetachainShardId {
		return
	}

	for _, candidate := range candidates {
		if _, err := boot.headers.GetHeaderByHash(candidate.hash); err != nil {
			boot.requestHandler.RequestShardHeaderForEpoch(shardID, candidate.hash, candidate.epoch)
		}
		if !boot.proofs.HasProof(shardID, candidate.hash) {
			boot.requestHandler.RequestEquivalentProofByHashForEpoch(shardID, candidate.hash, candidate.epoch)
		}
	}
}

func (boot *baseBootstrap) armAuthorityReconciliation(nonce uint64, selectedHash []byte, round int64) {
	authority, ok := boot.forkDetector.(notarizedHeaderAuthority)
	if !ok {
		return
	}
	localHash := authority.getProcessedHeaderHash(nonce)
	currentHeader := boot.chainHandler.GetCurrentBlockHeader()
	if check.IfNil(currentHeader) || currentHeader.GetNonce() < nonce || len(localHash) == 0 || bytes.Equal(localHash, selectedHash) {
		return
	}

	boot.mutReconcile.Lock()
	defer boot.mutReconcile.Unlock()

	if boot.pendingReconcile != nil && boot.pendingReconcile.nonce == nonce &&
		bytes.Equal(boot.pendingReconcile.localHash, localHash) &&
		bytes.Equal(boot.pendingReconcile.competitorHash, selectedHash) &&
		boot.pendingReconcile.selectedByAuthority {
		return
	}

	boot.pendingReconcile = &reconcileEvidence{
		nonce:               nonce,
		localHash:           append([]byte(nil), localHash...),
		competitorHash:      append([]byte(nil), selectedHash...),
		lastEvaluatedRound:  round,
		selectedByAuthority: true,
	}
}

func (boot *baseBootstrap) isAsyncExecutionEnabledForHash(hash []byte) bool {
	header, err := boot.getHeaderFromPool(hash)
	if err == nil && common.IsAsyncExecutionEnabledForEpochAndRound(
		boot.enableEpochsHandler,
		boot.enableRoundsHandler,
		header.GetEpoch(),
		header.GetRound(),
	) {
		return true
	}

	proof, err := boot.proofs.GetProof(boot.shardCoordinator.SelfId(), hash)
	if err != nil {
		return false
	}

	return common.IsAsyncExecutionEnabledForEpochAndRound(
		boot.enableEpochsHandler,
		boot.enableRoundsHandler,
		proof.GetHeaderEpoch(),
		proof.GetHeaderRound(),
	)
}

func (boot *baseBootstrap) getGenericProofHeaderRequestingIfMissing(
	nonce uint64,
	selectedHash []byte,
) (data.HeaderHandler, []byte, error) {
	currentHeader := boot.chainHandler.GetCurrentBlockHeader()
	currentHash := boot.chainHandler.GetCurrentBlockHeaderHash()
	if check.IfNil(currentHeader) || len(currentHash) == 0 ||
		!common.IsAsyncExecutionEnabledForEpochAndRound(
			boot.enableEpochsHandler,
			boot.enableRoundsHandler,
			currentHeader.GetEpoch(),
			currentHeader.GetRound(),
		) {
		header, err := boot.getHeaderWithHashRequestingIfMissing(selectedHash)
		return header, selectedHash, err
	}

	selectedHeader, err := boot.getHeaderFromPool(selectedHash)
	if err == nil && bytes.Equal(selectedHeader.GetPrevHash(), currentHash) {
		header, getErr := boot.getHeaderWithHashRequestingIfMissing(selectedHash)
		return header, selectedHash, getErr
	}

	proofs, err := boot.proofs.GetProofsByNonce(nonce, boot.shardCoordinator.SelfId())
	if err != nil {
		selectedProof, getErr := boot.proofs.GetProof(boot.shardCoordinator.SelfId(), selectedHash)
		if getErr != nil {
			boot.requestUnknownCanonicalHeader(nonce, currentHeader.GetNonce())
			return nil, nil, errBranchAwareSyncRetry
		}
		proofs = []data.HeaderProofHandler{selectedProof}
	}

	missingProofs := make([]data.HeaderProofHandler, 0)
	for _, proof := range proofs {
		hash := proof.GetHeaderHash()
		if boot.blackListHandler.Has(string(hash)) {
			continue
		}

		header, getErr := boot.getHeaderFromPool(hash)
		if getErr != nil {
			missingProofs = append(missingProofs, proof)
			continue
		}
		if bytes.Equal(header.GetPrevHash(), currentHash) {
			readyHeader, readyErr := boot.getHeaderWithHashRequestingIfMissing(hash)
			return readyHeader, hash, readyErr
		}
	}

	if len(missingProofs) > 0 {
		for _, proof := range missingProofs {
			boot.requestProofHeader(proof)
		}

		return nil, nil, errBranchAwareSyncRetry
	}

	boot.requestUnknownCanonicalHeader(nonce, currentHeader.GetNonce())
	return nil, nil, errBranchAwareSyncRetry
}

func (boot *baseBootstrap) requestProofHeader(proof data.HeaderProofHandler) {
	shardID := boot.shardCoordinator.SelfId()
	if shardID == core.MetachainShardId {
		boot.requestHandler.RequestMetaHeaderForEpoch(proof.GetHeaderHash(), proof.GetHeaderEpoch())
		return
	}

	boot.requestHandler.RequestShardHeaderForEpoch(
		shardID,
		proof.GetHeaderHash(),
		proof.GetHeaderEpoch(),
	)
}

func (boot *baseBootstrap) requestUnknownCanonicalHeader(nonce uint64, currentNonce uint64) {
	if boot.forkDetector.ProbableHighestNonce() <= currentNonce {
		return
	}

	boot.blockBootstrapper.requestHeaderByNonce(nonce)
	boot.blockBootstrapper.requestProofByNonce(nonce)
}

// selectNonBlackListedHash prevents re-adopting a hash blacklisted by a fork rollback or the
// reconcile backstop, preferring a non-blacklisted proofed sibling at the nonce
func (boot *baseBootstrap) selectNonBlackListedHash(hash []byte, nonce uint64) []byte {
	if len(hash) == 0 {
		return hash
	}

	boot.blackListHandler.Sweep()
	if !boot.blackListHandler.Has(string(hash)) {
		return hash
	}

	log.Debug("selectNonBlackListedHash: chosen header hash is blacklisted, trying proofed siblings",
		"nonce", nonce,
		"hash", hash,
	)

	proofs, err := boot.proofs.GetProofsByNonce(nonce, boot.shardCoordinator.SelfId())
	if err != nil {
		return nil
	}
	for _, siblingProof := range proofs {
		if !boot.blackListHandler.Has(string(siblingProof.GetHeaderHash())) {
			return siblingProof.GetHeaderHash()
		}
	}

	return nil
}

func (boot *baseBootstrap) getHeaderWithHashRequestingIfMissing(hash []byte) (data.HeaderHandler, error) {
	hdr, err := boot.getHeaderFromPool(hash)
	headerInPool := err == nil
	if !headerInPool {
		hdr, err = process.GetHeaderFromStorage(
			boot.shardCoordinator.SelfId(),
			hash,
			boot.marshalizer,
			boot.store,
		)
	}

	hasHeader := err == nil
	needsProof := boot.checkNeedsProofByHash(hash, hdr)
	if hasHeader && !needsProof {
		return hdr, nil
	}

	readyHeader := boot.requestHeaderAndProofByHashIfMissing(hash, hdr, !headerInPool, needsProof)
	if !check.IfNil(readyHeader) {
		return readyHeader, nil
	}

	err = boot.waitForHeaderAndProofByHash()
	if err != nil {
		return nil, err
	}

	hdr, err = boot.getHeaderFromPool(hash)
	if err != nil {
		return nil, err
	}

	if !boot.hasProof(hash, hdr) {
		return nil, process.ErrMissingHeaderProof
	}

	return hdr, nil
}

func (boot *baseBootstrap) checkNeedsProofByHash(hash []byte, header data.HeaderHandler) bool {
	// if header exists, check if it has or needs a proof
	// 		if it has a proof, do not wait
	// 		if it does not need a proof, do not wait
	// 		if it needs a proof, request and wait for the proof
	// if header does not exist
	//		if it has a proof, request the header
	//		if it does not have the proof, request the header first; its callback requests the proof
	_, errGetProof := boot.proofs.GetProof(boot.shardCoordinator.SelfId(), hash)
	hasProof := errGetProof == nil
	needsProof := !hasProof
	if check.IfNil(header) {
		return needsProof
	}

	isFlagActiveForExistingHeader := common.IsProofsFlagEnabledForHeader(boot.enableEpochsHandler, header)
	needsProof = needsProof && isFlagActiveForExistingHeader
	return needsProof
}

// getHeaderWithNonceRequestingIfMissing method gets the header with a given nonce from pool. If it is not found there, it will
// be requested from network
func (boot *baseBootstrap) getHeaderWithNonceRequestingIfMissing(nonce uint64) (data.HeaderHandler, []byte, error) {
	hdr, hash, err := boot.getHeaderFromPoolWithNonce(nonce)
	hasHeader := err == nil && !boot.blackListHandler.Has(string(hash))

	if hasHeader && boot.hasProof(hash, hdr) {
		return hdr, hash, nil
	}

	needsProof := boot.checkNeedsProofByNonce(nonce, hdr, hash)

	if hasHeader {
		boot.requestHandler.SetEpoch(hdr.GetEpoch())
	}

	// no usable header is held here, so ask for one even when the pool has an unproven fork: that
	// fork may never gain a proof, while a request by nonce is answered with the proven header
	readyHeader, readyHash := boot.requestHeaderAndProofByNonce(hash, hdr, nonce, needsProof)
	if !check.IfNil(readyHeader) {
		return readyHeader, readyHash, nil
	}

	err = boot.waitForHeaderAndProofByNonce()
	if err != nil {
		return nil, nil, err
	}

	// re-read the proven hash: the wait above is what a freshly arrived proof releases
	hdr, hash, err = boot.getHeaderFromPoolPreferProven(nonce, boot.getProvenHashForNonce(nonce))
	if err != nil {
		log.Debug("getHeaderWithNonceRequestingIfMissing: failed to get header with nonce", "nonce", nonce, "error", err)
		return nil, nil, err
	}

	if boot.blackListHandler.Has(string(hash)) {
		return nil, nil, process.ErrHeaderIsBlackListed
	}

	if !boot.hasProof(hash, hdr) {
		return nil, nil, process.ErrMissingHeaderProof
	}

	return hdr, hash, nil
}

func (boot *baseBootstrap) checkNeedsProofByNonce(
	nonce uint64,
	header data.HeaderHandler,
	headerHash []byte,
) bool {
	// if header exists, check if it has or needs a proof
	// 		if it has a proof, do not wait
	// 		if it does not need a proof, do not wait
	// 		if it needs a proof, request and wait for the proof
	// if header does not exist
	//		if it has a proof, request the header
	//		if it does not have the proof, request the header first; its callback requests the proof
	proof, errGetProof := boot.proofs.GetProofByNonce(nonce, boot.shardCoordinator.SelfId())
	hasProof := errGetProof == nil
	needsProof := !hasProof

	if check.IfNil(header) {
		return needsProof
	}

	if hasProof && !bytes.Equal(headerHash, proof.GetHeaderHash()) {
		needsProof = true
	}

	isFlagActiveForExistingHeader := common.IsProofsFlagEnabledForHeader(boot.enableEpochsHandler, header)
	needsProof = needsProof && isFlagActiveForExistingHeader

	return needsProof
}

func (boot *baseBootstrap) requestHeaderAndProofByHashIfMissing(
	hash []byte,
	header data.HeaderHandler,
	needsHeaderInPool bool,
	needsProof bool,
) data.HeaderHandler {
	_ = core.EmptyChannel(boot.chRcvHdrHash)
	if needsHeaderInPool {
		boot.mutRcvHdrHash.Lock()
		boot.setRequestedHeaderHash(hash)
		receivedHeader, err := boot.getHeaderFromPool(hash)
		if err == nil && !check.IfNil(receivedHeader) {
			if boot.hasProof(hash, receivedHeader) {
				boot.setRequestedHeaderHash(nil)
				boot.mutRcvHdrHash.Unlock()
				return receivedHeader
			}

			boot.mutRcvHdrHash.Unlock()
			boot.requestSelfShardProof(hash, receivedHeader)
			return nil
		}

		if !check.IfNil(header) && boot.hasProof(hash, header) {
			boot.setRequestedHeaderHash(nil)
			boot.mutRcvHdrHash.Unlock()
			return header
		}

		boot.mutRcvHdrHash.Unlock()
		if !check.IfNil(header) {
			boot.headers.AddHeader(hash, header)
			return nil
		}

		boot.requestHeaderByHash(hash)
		return nil
	}

	if !needsProof {
		return header
	}

	boot.mutRcvHdrHash.Lock()
	boot.setRequestedHeaderHash(hash)
	if boot.hasProof(hash, header) {
		boot.setRequestedHeaderHash(nil)
		boot.mutRcvHdrHash.Unlock()
		return header
	}
	boot.mutRcvHdrHash.Unlock()

	log.Debug("requesting equivalent proof from network",
		"hash", hex.EncodeToString(hash),
	)
	boot.requestSelfShardProof(hash, header)
	return nil
}

func (boot *baseBootstrap) requestSelfShardProof(hash []byte, header data.HeaderHandler) {
	boot.requestHandler.RequestEquivalentProofByHashForEpoch(boot.shardCoordinator.SelfId(), hash, header.GetEpoch())
}

func (boot *baseBootstrap) requestHeaderByHash(hash []byte) {
	logMsg := fmt.Sprintf("requesting %s header from network", boot.getShardLabel())
	log.Debug(logMsg,
		"hash", hash,
		"probable highest nonce", boot.forkDetector.ProbableHighestNonce(),
	)

	if boot.shardCoordinator.SelfId() == core.MetachainShardId {
		boot.requestHandler.RequestMetaHeader(hash)
		return
	}

	boot.requestHandler.RequestShardHeader(boot.shardCoordinator.SelfId(), hash)
}

func (boot *baseBootstrap) getShardLabel() string {
	shardLabel := "meta"
	if boot.shardCoordinator.SelfId() != core.MetachainShardId {
		shardLabel = "shard"
	}

	return shardLabel
}

func (boot *baseBootstrap) requestHeaderAndProofByNonce(
	hash []byte,
	header data.HeaderHandler,
	nonce uint64,
	needsProof bool,
) (data.HeaderHandler, []byte) {
	_ = core.EmptyChannel(boot.chRcvHdrNonce)
	boot.mutRcvHdrNonce.Lock()
	boot.setRequestedHeaderNonce(&nonce)
	if check.IfNil(header) {
		receivedHeader, receivedHash, err := boot.getHeaderFromPoolWithNonce(nonce)
		if err == nil && !boot.blackListHandler.Has(string(receivedHash)) {
			if boot.hasProof(receivedHash, receivedHeader) {
				boot.setRequestedHeaderNonce(nil)
				boot.mutRcvHdrNonce.Unlock()
				return receivedHeader, receivedHash
			}

			boot.mutRcvHdrNonce.Unlock()
			boot.requestHandler.SetEpoch(receivedHeader.GetEpoch())
			boot.requestHeaderByNonce(nonce)
			boot.requestSelfShardProof(receivedHash, receivedHeader)
			return nil, nil
		}

		boot.mutRcvHdrNonce.Unlock()
		boot.requestHeaderByNonce(nonce)
		return nil, nil
	}
	if !boot.blackListHandler.Has(string(hash)) && boot.hasProof(hash, header) {
		boot.setRequestedHeaderNonce(nil)
		boot.mutRcvHdrNonce.Unlock()
		return header, hash
	}
	boot.mutRcvHdrNonce.Unlock()
	boot.requestHeaderByNonce(nonce)

	if !needsProof {
		return nil, nil
	}

	if len(hash) == 0 {
		log.Debug("requesting equivalent proof from network",
			"nonce", nonce,
		)

		boot.requestHandler.RequestEquivalentProofByNonce(boot.shardCoordinator.SelfId(), nonce)
		return nil, nil
	}

	log.Debug("requesting equivalent proof from network",
		"hash", hex.EncodeToString(hash),
	)

	boot.requestSelfShardProof(hash, header)
	return nil, nil
}

func (boot *baseBootstrap) requestHeaderByNonce(nonce uint64) {
	logMsg := fmt.Sprintf("requesting %s header by nonce from network", boot.getShardLabel())
	log.Debug(logMsg,
		"nonce", nonce,
		"probable highest nonce", boot.forkDetector.ProbableHighestNonce(),
	)

	if boot.shardCoordinator.SelfId() == core.MetachainShardId {
		boot.requestHandler.RequestMetaHeaderByNonce(nonce)
		return
	}

	boot.requestHandler.RequestShardHeaderByNonce(boot.shardCoordinator.SelfId(), nonce)
}

func (boot *baseBootstrap) getHeader(hash []byte) (data.HeaderHandler, error) {
	if boot.shardCoordinator.SelfId() == core.MetachainShardId {
		return process.GetMetaHeader(hash, boot.headers, boot.marshalizer, boot.store)
	}

	return process.GetShardHeader(hash, boot.headers, boot.marshalizer, boot.store)
}

// getHeaderFromPool will try to get the header from pool
func (boot *baseBootstrap) getHeaderFromPool(hash []byte) (data.HeaderHandler, error) {
	if boot.shardCoordinator.SelfId() == core.MetachainShardId {
		return process.GetMetaHeaderFromPool(hash, boot.headers)
	}

	return process.GetShardHeaderFromPool(hash, boot.headers)
}

func (boot *baseBootstrap) getHeaderFromPoolWithNonce(
	nonce uint64,
) (data.HeaderHandler, []byte, error) {
	if boot.shardCoordinator.SelfId() == core.MetachainShardId {
		return process.GetMetaHeaderFromPoolWithNonce(nonce, boot.headers)
	}

	return process.GetShardHeaderFromPoolWithNonce(nonce, boot.shardCoordinator.SelfId(), boot.headers)
}

// getProvenHashForNonce returns the hash a proof at the given nonce attests, if such a proof is known
func (boot *baseBootstrap) getProvenHashForNonce(nonce uint64) []byte {
	proof, err := boot.proofs.GetProofByNonce(nonce, boot.shardCoordinator.SelfId())
	if err != nil {
		return nil
	}

	return proof.GetHeaderHash()
}

// getHeaderFromPoolPreferProven returns the header at the given nonce, preferring the proven one: the
// pool keeps every header received at a nonce and a later fork header must not shadow the proven one
func (boot *baseBootstrap) getHeaderFromPoolPreferProven(
	nonce uint64,
	provenHash []byte,
) (data.HeaderHandler, []byte, error) {
	if len(provenHash) > 0 {
		hdr, err := boot.getHeaderFromPool(provenHash)
		if err == nil {
			return hdr, provenHash, nil
		}
	}

	return boot.getHeaderFromPoolWithNonce(nonce)
}

// onEquivocationEvidence records reconcile evidence, an equivocation proof at the final
// chain tip nonce; the evidence is verified and acted upon from the sync loop
func (boot *baseBootstrap) onEquivocationEvidence(headerProof data.HeaderProofHandler, competingProofs []data.HeaderProofHandler) {
	if check.IfNil(headerProof) || headerProof.GetHeaderShardId() != boot.shardCoordinator.SelfId() {
		return
	}

	nonce := headerProof.GetHeaderNonce()
	if nonce != boot.forkDetector.GetHighestFinalBlockNonce() {
		return
	}

	currentHeader := boot.chainHandler.GetCurrentBlockHeader()
	localHash := boot.chainHandler.GetCurrentBlockHeaderHash()
	isFinalHead := !check.IfNil(currentHeader) && currentHeader.GetNonce() == nonce && currentHeader.IsHeaderV3()
	if !isFinalHead {
		return
	}

	competitorHash := pickCompetitorHash(localHash, headerProof, competingProofs)
	if len(competitorHash) == 0 {
		return
	}

	boot.mutReconcile.Lock()
	boot.pendingReconcile = &reconcileEvidence{
		nonce:          nonce,
		localHash:      localHash,
		competitorHash: competitorHash,
		// first evaluated after the arming round turns: a fired roll back starts round-aligned,
		// when no consensus commit can still be in flight
		lastEvaluatedRound: boot.roundHandler.Index(),
	}
	boot.mutReconcile.Unlock()

	log.Warn("equivocation proof observed at the final chain tip, reconcile evidence recorded",
		"nonce", nonce,
		"local hash", localHash,
		"competitor hash", competitorHash)
}

func pickCompetitorHash(localHash []byte, headerProof data.HeaderProofHandler, competingProofs []data.HeaderProofHandler) []byte {
	if !bytes.Equal(headerProof.GetHeaderHash(), localHash) {
		return headerProof.GetHeaderHash()
	}

	for _, competingProof := range competingProofs {
		if check.IfNil(competingProof) {
			continue
		}
		if !bytes.Equal(competingProof.GetHeaderHash(), localHash) {
			return competingProof.GetHeaderHash()
		}
	}

	return nil
}

// tryReconcileEquivocation overrides the final gate and forces the switch when the settlement
// authority settled the equivocation competitor and not the local block
func (boot *baseBootstrap) tryReconcileEquivocation(round int64) bool {
	evidence, shouldEvaluate := boot.reconcileEvidenceToEvaluate(round)
	if evidence == nil {
		return false
	}

	if !boot.reconcileEvidenceStillApplies(evidence) {
		boot.clearReconcileEvidence(evidence)
		return false
	}

	if !shouldEvaluate {
		return false
	}
	if evidence.selectedByAuthority {
		candidates := []notarizedHeaderCandidate{
			{hash: evidence.localHash, nonce: evidence.nonce},
			{hash: evidence.competitorHash, nonce: evidence.nonce},
		}
		selectedHash := boot.settlementChecker.resolveNotarizedHeader(evidence.nonce, candidates)
		if bytes.Equal(selectedHash, evidence.localHash) {
			boot.clearReconcileEvidence(evidence)
			return false
		}
		if !bytes.Equal(selectedHash, evidence.competitorHash) ||
			!boot.proofs.HasProof(boot.shardCoordinator.SelfId(), evidence.competitorHash) {
			_, _, nextCursor := boot.settlementChecker.prepareInclusionScan(evidence.scanCursor)
			boot.storeReconcileScanCursor(evidence, nextCursor)
			return false
		}

		return boot.applyReconcileSwitch(evidence)
	}

	scanFrom, scanTo, nextCursor := boot.settlementChecker.prepareInclusionScan(evidence.scanCursor)
	boot.storeReconcileScanCursor(evidence, nextCursor)

	competitorHash := evidence.competitorHash
	if !boot.proofs.HasProof(boot.shardCoordinator.SelfId(), competitorHash) {
		competitorHash = nil
	}
	localSettled, competitorSettled := boot.settlementChecker.settlementVerdict(
		evidence.nonce, evidence.localHash, competitorHash, scanFrom, scanTo)
	if localSettled {
		boot.clearReconcileEvidence(evidence)
		return false
	}

	if !competitorSettled {
		// the authority's verdict may still arrive; keep the evidence armed for the next round
		return false
	}

	return boot.applyReconcileSwitch(evidence)
}

func (boot *baseBootstrap) applyReconcileSwitch(evidence *reconcileEvidence) bool {
	boot.clearReconcileEvidence(evidence)
	var reconciled bool
	if evidence.selectedByAuthority {
		reconciled = boot.forkDetector.ReconcileFinalCheckpointFromAuthority(evidence.nonce, evidence.competitorHash)
	} else {
		reconciled = boot.forkDetector.ReconcileFinalCheckpoint(evidence.nonce)
	}
	if !reconciled {
		return false
	}

	log.Error("reconcile backstop: switching away from a finalized block on equivocation evidence",
		"nonce", evidence.nonce,
		"local hash", evidence.localHash,
		"competitor hash", evidence.competitorHash)
	boot.statusHandler.Increment(common.MetricNumReconcileSwitches)

	process.AddHeaderToBlackList(boot.blackListHandler, evidence.localHash)
	boot.forkDetector.SetRollBackNonce(evidence.nonce)

	return true
}

// the settlement checks walk the pools, so the authority is consulted at most once per round
func (boot *baseBootstrap) reconcileEvidenceToEvaluate(round int64) (*reconcileEvidence, bool) {
	boot.mutReconcile.Lock()
	defer boot.mutReconcile.Unlock()

	evidence := boot.pendingReconcile
	if evidence == nil {
		return nil, false
	}

	shouldEvaluate := evidence.lastEvaluatedRound != round
	if shouldEvaluate {
		evidence.lastEvaluatedRound = round
	}

	return evidence, shouldEvaluate
}

// invalidateNodeState forces the next iteration to recompute the fork info, so a roll back armed
// here is picked up without waiting for the round to change
func (boot *baseBootstrap) invalidateNodeState() {
	boot.mutNodeState.Lock()
	boot.isNodeStateCalculated = false
	boot.mutNodeState.Unlock()
}

func (boot *baseBootstrap) reconcileEvidenceStillApplies(evidence *reconcileEvidence) bool {
	if evidence.selectedByAuthority {
		authority, ok := boot.forkDetector.(notarizedHeaderAuthority)
		if !ok || !bytes.Equal(authority.getProcessedHeaderHash(evidence.nonce), evidence.localHash) {
			return false
		}
		currentHeader := boot.chainHandler.GetCurrentBlockHeader()
		return !check.IfNil(currentHeader) && currentHeader.GetNonce() >= evidence.nonce
	}

	currentHeader := boot.chainHandler.GetCurrentBlockHeader()
	currentHash := boot.chainHandler.GetCurrentBlockHeaderHash()

	return !check.IfNil(currentHeader) &&
		currentHeader.GetNonce() == evidence.nonce &&
		bytes.Equal(currentHash, evidence.localHash) &&
		evidence.nonce == boot.forkDetector.GetHighestFinalBlockNonce()
}

// tryReconcileDivergence rolls back the certainly-dead own suffix above the block referencing a
// dead cross-notarized meta; the per-block pointer pops make deeper divergences converge round by round
func (boot *baseBootstrap) tryReconcileDivergence(round int64) bool {
	if boot.divergenceEvaluatedRound == round {
		return false
	}
	boot.divergenceEvaluatedRound = round

	deadMeta, deadMetaHash, isDead := boot.settlementChecker.deadCrossNotarizedMeta()
	if !isDead {
		return false
	}

	headHash := boot.chainHandler.GetCurrentBlockHeaderHash()
	earliestDeadNonce, deadOwnHashes, collected := boot.collectOwnBlocksReferencing(deadMetaHash)
	if !collected {
		return false
	}

	// the chain should only move from this goroutine, re-checked out of caution
	if !bytes.Equal(boot.chainHandler.GetCurrentBlockHeaderHash(), headHash) {
		return false
	}

	if !boot.forkDetector.ReconcileFinalCheckpointBelow(earliestDeadNonce) {
		return false
	}

	log.Error("divergence backstop: rolling back own blocks referencing a dead meta block",
		"dead meta nonce", deadMeta.GetNonce(),
		"dead meta hash", deadMetaHash,
		"earliest dead own nonce", earliestDeadNonce,
		"num dead own blocks", len(deadOwnHashes))
	boot.statusHandler.Increment(common.MetricNumReconcileSwitches)

	boot.disarmDeadEpochStartIfNeeded(deadMeta, deadMetaHash)

	for _, deadOwnHash := range deadOwnHashes {
		process.AddHeaderToBlackList(boot.blackListHandler, deadOwnHash)
	}
	boot.forkDetector.SetRollBackNonce(earliestDeadNonce)

	return true
}

// collectOwnBlocksReferencing walks the own chain down to the block holding the cross-notarization
// pointer; blocks above it reference no meta block at all, so the whole suffix dies with it
func (boot *baseBootstrap) collectOwnBlocksReferencing(deadMetaHash []byte) (uint64, [][]byte, bool) {
	currHeader, err := boot.blockBootstrapper.getCurrHeader()
	if err != nil {
		return 0, nil, false
	}
	currHash := boot.chainHandler.GetCurrentBlockHeaderHash()
	settledNonce, _ := boot.forkDetector.GetHighestSettledBlockInfo()

	deadOwnHashes := make([][]byte, 0)
	for {
		if check.IfNil(currHeader) || currHeader.GetNonce() == 0 || currHeader.GetNonce() <= settledNonce {
			log.Warn("collectOwnBlocksReferencing: no block referencing the dead meta above the settled checkpoint",
				"dead meta hash", deadMetaHash)
			return 0, nil, false
		}

		deadOwnHashes = append(deadOwnHashes, currHash)

		numReferences, referencesDeadMeta := metaReferencesOfShardHeader(currHeader, deadMetaHash)
		if referencesDeadMeta {
			return currHeader.GetNonce(), deadOwnHashes, true
		}
		if numReferences > 0 {
			log.Warn("collectOwnBlocksReferencing: pointer block does not reference the dead meta",
				"nonce", currHeader.GetNonce(),
				"dead meta hash", deadMetaHash)
			return 0, nil, false
		}

		prevHash := currHeader.GetPrevHash()
		currHeader, err = boot.blockBootstrapper.getPrevHeader(currHeader, boot.headerStore)
		if err != nil {
			return 0, nil, false
		}
		currHash = prevHash
	}
}

func metaReferencesOfShardHeader(header data.HeaderHandler, metaHash []byte) (int, bool) {
	shardHeader, ok := header.(data.ShardHeaderHandler)
	if !ok {
		return 0, false
	}

	metaHashes := shardHeader.GetMetaBlockHashes()
	for _, hash := range metaHashes {
		if bytes.Equal(hash, metaHash) {
			return len(metaHashes), true
		}
	}

	return len(metaHashes), false
}

func (boot *baseBootstrap) disarmDeadEpochStartIfNeeded(deadMeta data.HeaderHandler, deadMetaHash []byte) {
	if !deadMeta.IsStartOfEpochBlock() {
		return
	}
	if boot.epochStartDisarmer == nil {
		log.Warn("dead epoch start meta block, no disarm capable trigger wired", "hash", deadMetaHash)
		return
	}

	disarmed := boot.epochStartDisarmer.DisarmDeadEpochStartActivation(deadMeta.GetEpoch(), deadMetaHash)
	log.Warn("dead epoch start meta block, trigger disarm attempted",
		"epoch", deadMeta.GetEpoch(),
		"hash", deadMetaHash,
		"disarmed", disarmed)
}

func (boot *baseBootstrap) storeReconcileScanCursor(evidence *reconcileEvidence, nextCursor uint64) {
	boot.mutReconcile.Lock()
	if boot.pendingReconcile == evidence {
		evidence.scanCursor = nextCursor
	}
	boot.mutReconcile.Unlock()
}

func (boot *baseBootstrap) clearReconcileEvidence(evidence *reconcileEvidence) {
	boot.mutReconcile.Lock()
	if boot.pendingReconcile == evidence {
		boot.pendingReconcile = nil
	}
	boot.mutReconcile.Unlock()
}

func (boot *baseBootstrap) isForcedRollBackOneBlock() bool {
	return boot.forkInfo.IsDetected &&
		boot.forkInfo.Nonce == math.MaxUint64 &&
		boot.forkInfo.Hash == nil
}

func (boot *baseBootstrap) isForcedRollBackToNonce() bool {
	return boot.forkInfo.IsDetected &&
		boot.forkInfo.Round == math.MaxUint64 &&
		boot.forkInfo.Hash == nil
}

func (boot *baseBootstrap) rollBackOneBlockForced() {
	rolledBackHeader := boot.getCurrentBlock()

	err := boot.rollBack(false)
	if err != nil {
		log.Debug("rollBackOneBlockForced", "error", err.Error())
	}

	boot.forkDetector.ResetFork()
	boot.removeHeadersHigherThanNonceFromPool(boot.getNonceForCurrentBlock())

	if err == nil && common.IsProofsFlagEnabledForHeader(boot.enableEpochsHandler, rolledBackHeader) {
		boot.blockBootstrapper.requestProofByNonce(rolledBackHeader.GetNonce())
	}
}

func (boot *baseBootstrap) rollBackToNonceForced() {
	err := boot.rollBack(true)
	if err != nil {
		log.Debug("rollBackToNonceForced", "error", err.Error())
	}

	boot.forkDetector.ResetProbableHighestNonce()
	boot.removeHeadersHigherThanNonceFromPool(boot.getNonceForCurrentBlock())
}

func (boot *baseBootstrap) restoreState(
	currHeaderHash []byte,
	currHeader data.HeaderHandler,
	currRootHash []byte,
) {
	log.Debug("revert state to header",
		"nonce", currHeader.GetNonce(),
		"hash", currHeaderHash,
		"current root hash", currRootHash)

	err := boot.chainHandler.SetCurrentBlockHeaderAndRootHash(currHeader, currRootHash)
	if err != nil {
		log.Debug("SetCurrentBlockHeader", "error", err.Error())
	}

	boot.chainHandler.SetCurrentBlockHeaderHash(currHeaderHash)

	// for legacy (non-V3) headers, keep last executed block header in sync with current block header
	if check.IfNil(currHeader) || !currHeader.IsHeaderV3() {
		boot.chainHandler.SetLastExecutedBlockHeaderAndRootHash(currHeader, currHeaderHash, currRootHash)
	}

	err = boot.scheduledTxsExecutionHandler.RollBackToBlock(currHeaderHash)
	if err != nil {
		scheduledInfo := &process.ScheduledInfo{
			RootHash:        currHeader.GetRootHash(),
			IntermediateTxs: make(map[block.Type][]data.TransactionHandler),
			GasAndFees:      process.GetZeroGasAndFees(),
			MiniBlocks:      make(block.MiniBlockSlice, 0),
		}
		boot.scheduledTxsExecutionHandler.SetScheduledInfo(scheduledInfo)
	}

	err = boot.blockProcessor.RevertStateToBlock(currHeader, boot.scheduledTxsExecutionHandler.GetScheduledRootHash())
	if err != nil {
		log.Debug("RevertState", "error", err.Error())
	}
}

func (boot *baseBootstrap) setCurrentBlockInfo(
	headerHash []byte,
	header data.HeaderHandler,
	rootHash []byte,
) error {

	err := boot.chainHandler.SetCurrentBlockHeaderAndRootHash(header, rootHash)
	if err != nil {
		return err
	}

	boot.chainHandler.SetCurrentBlockHeaderHash(headerHash)

	// for legacy (non-V3) headers, keep last executed block header in sync with current block header
	if check.IfNil(header) || !header.IsHeaderV3() {
		boot.chainHandler.SetLastExecutedBlockHeaderAndRootHash(header, headerHash, rootHash)
	}

	return nil
}

// setRequestedMiniBlocks method sets the body hash requested by the sync mechanism
func (boot *baseBootstrap) setRequestedMiniBlocks(hashes [][]byte) {
	boot.requestedHashes.SetHashes(hashes)
}

// receivedMiniblock method is a call back function which is called when a new body is added
// in the block bodies pool
func (boot *baseBootstrap) receivedMiniblock(hash []byte, _ interface{}) {
	boot.mutRcvMiniBlocks.Lock()
	if len(boot.requestedHashes.ExpectedData()) == 0 {
		boot.mutRcvMiniBlocks.Unlock()
		return
	}

	boot.requestedHashes.SetReceivedHash(hash)
	if boot.requestedHashes.ReceivedAll() {
		log.Debug("received all the requested mini blocks from network")
		boot.setRequestedMiniBlocks(nil)
		boot.mutRcvMiniBlocks.Unlock()
		boot.chRcvMiniBlocks <- true
	} else {
		boot.mutRcvMiniBlocks.Unlock()
	}
}

// requestMiniBlocksByHashes method requests a block body from network when it is not found in the pool
func (boot *baseBootstrap) requestMiniBlocksByHashes(hashes [][]byte) {
	boot.setRequestedMiniBlocks(hashes)
	log.Debug("requesting mini blocks from network",
		"num miniblocks", len(hashes),
	)
	boot.requestHandler.RequestMiniBlocks(boot.shardCoordinator.SelfId(), hashes)
}

// getMiniBlocksRequestingIfMissing method gets the body with given nonce from pool, if it exists there,
// and if not it will be requested from network
// the func returns interface{} as to match the next implementations for block body fetchers
// that will be added. The block executor should decide by parsing the header block body type value
// what kind of block body received.
func (boot *baseBootstrap) getMiniBlocksRequestingIfMissing(hashes [][]byte) (block.MiniBlockSlice, error) {
	miniBlocksAndHashes, missingMiniBlocksHashes := boot.miniBlocksProvider.GetMiniBlocksFromPool(hashes)
	if len(missingMiniBlocksHashes) == 0 {
		miniBlocks := make([]*block.MiniBlock, len(miniBlocksAndHashes))
		for index, miniBlockAndHash := range miniBlocksAndHashes {
			miniBlocks[index] = miniBlockAndHash.Miniblock
		}

		return miniBlocks, nil
	}

	_ = core.EmptyChannel(boot.chRcvMiniBlocks)
	boot.requestMiniBlocksByHashes(missingMiniBlocksHashes)
	err := boot.waitForMiniBlocks()
	if err != nil {
		return nil, err
	}

	receivedMiniBlocksAndHashes, unreceivedMiniBlocksHashes := boot.miniBlocksProvider.GetMiniBlocksFromPool(missingMiniBlocksHashes)
	if len(unreceivedMiniBlocksHashes) > 0 {
		return nil, process.ErrMissingBody
	}

	miniBlocksAndHashes = append(miniBlocksAndHashes, receivedMiniBlocksAndHashes...)

	return getOrderedMiniBlocks(hashes, miniBlocksAndHashes)
}

func (boot *baseBootstrap) getHeaderMiniBlocksRequestingIfMissing(
	header data.HeaderHandler,
) (block.MiniBlockSlice, error) {
	miniBlockHeaderHandlers := header.GetMiniBlockHeaderHandlers()

	hashes := make([][]byte, len(miniBlockHeaderHandlers))
	for i, miniBlockHeaderHandler := range miniBlockHeaderHandlers {
		hashes[i] = miniBlockHeaderHandler.GetHash()
	}

	boot.setRequestedMiniBlocks(nil)

	return boot.getMiniBlocksRequestingIfMissing(hashes)
}

func getOrderedMiniBlocks(
	hashes [][]byte,
	miniBlocksAndHashes []*block.MiniblockAndHash,
) (block.MiniBlockSlice, error) {

	mapHashMiniBlock := make(map[string]*block.MiniBlock, len(miniBlocksAndHashes))
	for _, miniBlockAndHash := range miniBlocksAndHashes {
		mapHashMiniBlock[string(miniBlockAndHash.Hash)] = miniBlockAndHash.Miniblock
	}

	orderedMiniBlocks := make(block.MiniBlockSlice, len(hashes))
	for index, hash := range hashes {
		miniBlock, ok := mapHashMiniBlock[string(hash)]
		if !ok {
			return nil, process.ErrMissingBody
		}

		orderedMiniBlocks[index] = miniBlock
	}

	return orderedMiniBlocks, nil
}

// waitForMiniBlocks method wait for body with the requested nonce to be received
func (boot *baseBootstrap) waitForMiniBlocks() error {
	select {
	case <-boot.chRcvMiniBlocks:
		return nil
	case <-time.After(boot.getWaitTime()):
		return process.ErrTimeIsOut
	}
}

func (boot *baseBootstrap) init() {
	boot.forkInfo = process.NewForkInfo()

	boot.chRcvHdrNonce = make(chan bool)
	boot.chRcvHdrHash = make(chan bool)
	boot.chRcvMiniBlocks = make(chan bool)
	boot.signalProcessCompletionChan = boot.executionManager.GetSignalProcessCompletionChan()

	boot.setRequestedHeaderNonce(nil)
	boot.setRequestedHeaderHash(nil)
	boot.setRequestedMiniBlocks(nil)

	boot.poolsHolder.MiniBlocks().RegisterHandler(boot.receivedMiniblock, core.UniqueIdentifier())
	boot.headers.RegisterHandler(boot.processReceivedHeader)
	boot.proofs.RegisterHandler(boot.processReceivedProof)
	boot.proofs.RegisterEquivocationHandler(boot.onEquivocationEvidence)

	boot.syncStateListeners = make([]func(bool), 0)
	boot.requestedHashes = process.RequiredDataPool{}
	boot.mapNonceSyncedWithErrors = make(map[uint64]uint32)
	boot.mapNonceRecoveryAttempts = make(map[uint64]*nonceRecoveryInfo)
	boot.executionResultsRecoveryCooldown = defaultExecutionResultsRecoveryCooldown
}

func (boot *baseBootstrap) requestHeaders(fromNonce uint64, toNonce uint64) {
	boot.mutRequestHeaders.Lock()
	defer boot.mutRequestHeaders.Unlock()

	for currentNonce := fromNonce; currentNonce <= toNonce; currentNonce++ {
		hdr, hash, err := boot.getHeaderFromPoolWithNonce(currentNonce)
		hasHeader := err == nil
		if hasHeader && boot.hasProof(hash, hdr) {
			continue
		}

		if hasHeader {
			boot.requestHandler.SetEpoch(hdr.GetEpoch())
		}

		needsProof := boot.checkNeedsProofByNonce(currentNonce, hdr, hash)
		if !hasHeader {
			boot.blockBootstrapper.requestHeaderByNonce(currentNonce)
		}

		if needsProof {
			boot.blockBootstrapper.requestProofByNonce(currentNonce)
		}
	}
}

// GetNodeState method returns the sync state of the node. If it returns 'NsNotSynchronized', this means that the node
// is not synchronized yet, and it has to continue the bootstrapping mechanism. If it returns 'NsSynchronized', this means
// that the node is already synced, and it can participate in the consensus. This method could also return 'NsNotCalculated'
// which means that the state of the node in the current round is not calculated yet. Note that when the node is not
// connected to the network, GetNodeState could return 'NsNotSynchronized' but the SyncBlock is not automatically called.
func (boot *baseBootstrap) GetNodeState() common.NodeState {
	if boot.isInImportMode {
		return common.NsNotSynchronized
	}
	currentSyncedEpoch := boot.getEpochOfCurrentBlock()
	if !boot.currentEpochProvider.EpochIsActiveInNetwork(currentSyncedEpoch) {
		return common.NsNotSynchronized
	}

	boot.mutNodeState.RLock()
	isNodeStateCalculatedInCurrentRound := boot.roundIndex == boot.roundHandler.Index() && boot.isNodeStateCalculated
	isNodeSynchronized := boot.isNodeSynchronized
	boot.mutNodeState.RUnlock()

	if !isNodeStateCalculatedInCurrentRound {
		return common.NsNotCalculated
	}

	if isNodeSynchronized {
		return common.NsSynchronized
	}

	return common.NsNotSynchronized
}

func (boot *baseBootstrap) handleAccountsTrieIteration() error {
	if boot.repopulateTokensSupplies {
		return boot.handleTokensSuppliesRepopulation()
	}

	// add more flags and trie iterators here
	return nil
}

func (boot *baseBootstrap) handleTokensSuppliesRepopulation() error {
	argsTrieAccountsIteratorProc := trieIterators.ArgsTrieAccountsIterator{
		Marshaller: boot.marshalizer,
		Accounts:   boot.accounts,
	}
	trieAccountsIteratorProc, err := trieIterators.NewTrieAccountsIterator(argsTrieAccountsIteratorProc)
	if err != nil {
		return err
	}

	argsTokensSuppliesProc := trieIterators.ArgsTokensSuppliesProcessor{
		StorageService: boot.store,
		Marshaller:     boot.marshalizer,
	}
	tokensSuppliesProc, err := trieIterators.NewTokensSuppliesProcessor(argsTokensSuppliesProc)
	if err != nil {
		return err
	}

	err = trieAccountsIteratorProc.Process(tokensSuppliesProc.HandleTrieAccountIteration)
	if err != nil {
		return err
	}

	return tokensSuppliesProc.SaveSupplies()
}

// Close will close the endless running go routine
func (boot *baseBootstrap) Close() error {
	if boot.cancelFunc != nil {
		boot.cancelFunc()
	}

	boot.cleanChannels()
	boot.closeRecovery()

	return nil
}

func (boot *baseBootstrap) cleanChannels() {
	nrReads := core.EmptyChannel(boot.chRcvHdrNonce)
	log.Debug("close baseSync: emptied channel", "chRcvHdrNonce nrReads", nrReads)

	nrReads = core.EmptyChannel(boot.chRcvHdrHash)
	log.Debug("close baseSync: emptied channel", "chRcvHdrHash nrReads", nrReads)

	nrReads = core.EmptyChannel(boot.chRcvMiniBlocks)
	log.Debug("close baseSync: emptied channel", "chRcvMiniBlocks nrReads", nrReads)

	if boot.signalProcessCompletionChan != nil {
		nrReads = common.EmptyUint64Channel(boot.signalProcessCompletionChan)
		log.Debug("close baseSync: emptied channel", "signalProcessCompletionChan nrReads", nrReads)
	}
}

func (boot *baseBootstrap) getHeaderMiniBlocks(
	header data.HeaderHandler,
) (block.MiniBlockSlice, error) {
	miniBlockHeaders := header.GetMiniBlockHeaderHandlers()

	hashes := make([][]byte, len(miniBlockHeaders))
	for i, miniBlockHeader := range miniBlockHeaders {
		hashes[i] = miniBlockHeader.GetHash()
	}

	miniBlocksAndHashes, missingMiniBlocksHashes := boot.miniBlocksProvider.GetMiniBlocks(hashes)
	if len(missingMiniBlocksHashes) > 0 {
		return nil, process.ErrMissingBody
	}

	miniBlocks := make([]*block.MiniBlock, len(miniBlocksAndHashes))
	for index, miniBlockAndHash := range miniBlocksAndHashes {
		miniBlocks[index] = miniBlockAndHash.Miniblock
	}

	return miniBlocks, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (boot *baseBootstrap) IsInterfaceNil() bool {
	return boot == nil
}

func (boot *baseBootstrap) createTxSyncer() error {
	var err error

	miniBlocksStorer, err := boot.store.GetStorer(dataRetriever.MiniBlockUnit)
	if err != nil {
		return err
	}

	syncMiniBlocksArgs := updateSync.ArgsNewPendingMiniBlocksSyncer{
		Storage:        miniBlocksStorer,
		Cache:          boot.dataPool.MiniBlocks(),
		Marshalizer:    boot.marshalizer,
		RequestHandler: boot.requestHandler,
	}
	boot.miniBlocksSyncer, err = updateSync.NewPendingMiniBlocksSyncer(syncMiniBlocksArgs)
	if err != nil {
		return err
	}

	syncTxsArgs := updateSync.ArgsNewTransactionsSyncer{
		DataPools:      boot.dataPool,
		Storages:       boot.store,
		Marshaller:     boot.marshalizer,
		RequestHandler: boot.requestHandler,
	}
	boot.txSyncer, err = updateSync.NewTransactionsSyncer(syncTxsArgs)
	if err != nil {
		return err
	}

	return nil
}
