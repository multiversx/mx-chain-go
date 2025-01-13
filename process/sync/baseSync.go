package sync

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/closing"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/typeConverters"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	logger "github.com/multiversx/mx-chain-logger-go"

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

var _ closing.Closer = (*baseBootstrap)(nil)

// sleepTime defines the time in milliseconds between each iteration made in syncBlocks method
const sleepTime = 5 * time.Millisecond
const minimumProcessWaitTime = time.Millisecond * 100

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

type baseBootstrap struct {
	log         logger.Logger
	historyRepo dblookupext.HistoryRepository
	headers     dataRetriever.HeadersPool
	proofs      dataRetriever.ProofsPool

	chainHandler   data.ChainHandler
	blockProcessor process.BlockProcessor
	store          dataRetriever.StorageService

	roundHandler        consensus.RoundHandler
	hasher              hashing.Hasher
	marshalizer         marshal.Marshalizer
	epochHandler        dataRetriever.EpochHandler
	forkDetector        process.ForkDetector
	requestHandler      process.RequestHandler
	shardCoordinator    sharding.Coordinator
	accounts            state.AccountsAdapter
	blockBootstrapper   blockBootstrapper
	blackListHandler    process.TimeCacher
	enableEpochsHandler common.EnableEpochsHandler

	mutHeader     sync.RWMutex
	headerNonce   *uint64
	headerhash    []byte
	chRcvHdrNonce chan bool
	chRcvHdrHash  chan bool

	requestedHashes process.RequiredDataPool

	statusHandler core.AppStatusHandler

	chStopSync chan bool
	waitTime   time.Duration

	mutNodeState          sync.RWMutex
	isNodeSynchronized    bool
	isNodeStateCalculated bool
	hasLastBlock          bool
	roundIndex            int64

	forkInfo *process.ForkInfo

	mutRcvHdrNonce           sync.RWMutex
	mutRcvHdrHash            sync.RWMutex
	syncStateListeners       []func(bool)
	mutSyncStateListeners    sync.RWMutex
	uint64Converter          typeConverters.Uint64ByteSliceConverter
	mapNonceSyncedWithErrors map[uint64]uint32
	mutNonceSyncedWithErrors sync.RWMutex

	requestMiniBlocks func(headerHandler data.HeaderHandler)

	networkWatcher    process.NetworkConnectionWatcher
	getHeaderFromPool func([]byte) (data.HeaderHandler, error)

	headerStore          storage.Storer
	headerNonceHashStore storage.Storer
	syncStarter          syncStarter
	bootStorer           process.BootStorer
	storageBootstrapper  process.BootstrapperFromStorage
	currentEpochProvider process.CurrentNetworkEpochProviderHandler

	outportHandler   outport.OutportHandler
	accountsDBSyncer process.AccountsDBSyncer

	chRcvMiniBlocks              chan bool
	mutRcvMiniBlocks             sync.Mutex
	miniBlocksProvider           process.MiniBlockProvider
	poolsHolder                  dataRetriever.PoolsHolder
	mutRequestHeaders            sync.Mutex
	cancelFunc                   func()
	isInImportMode               bool
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	processWaitTime              time.Duration

	repopulateTokensSupplies bool
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
	boot.forkDetector.ReceivedProof(headerProof)
}

func (boot *baseBootstrap) processReceivedHeader(headerHandler data.HeaderHandler, headerHash []byte) {
	if boot.shardCoordinator.SelfId() != headerHandler.GetShardID() {
		return
	}

	boot.log.Trace("received header from network",
		"shard", headerHandler.GetShardID(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
		"hash", headerHash,
	)

	err := boot.forkDetector.AddHeader(headerHandler, headerHash, process.BHReceived, nil, nil)
	if err != nil {
		boot.log.Debug("forkDetector.AddHeader", "error", err.Error())
	}

	go boot.requestMiniBlocks(headerHandler)

	boot.confirmHeaderReceivedByNonce(headerHandler, headerHash)
	boot.confirmHeaderReceivedByHash(headerHandler, headerHash)
}

func (boot *baseBootstrap) confirmHeaderReceivedByNonce(headerHandler data.HeaderHandler, hdrHash []byte) {
	boot.mutRcvHdrNonce.Lock()
	n := boot.requestedHeaderNonce()
	if n != nil && *n == headerHandler.GetNonce() {
		boot.log.Debug("received requested header from network",
			"shard", headerHandler.GetShardID(),
			"round", headerHandler.GetRound(),
			"nonce", headerHandler.GetNonce(),
			"hash", hdrHash,
		)
		boot.setRequestedHeaderNonce(nil)
		boot.mutRcvHdrNonce.Unlock()
		boot.chRcvHdrNonce <- true
		return
	}

	boot.mutRcvHdrNonce.Unlock()
}

func (boot *baseBootstrap) confirmHeaderReceivedByHash(headerHandler data.HeaderHandler, hdrHash []byte) {
	boot.mutRcvHdrHash.Lock()
	hash := boot.requestedHeaderHash()
	if hash != nil && bytes.Equal(hash, hdrHash) {
		boot.log.Debug("received requested header from network",
			"shard", headerHandler.GetShardID(),
			"round", headerHandler.GetRound(),
			"nonce", headerHandler.GetNonce(),
			"hash", hash,
		)
		boot.setRequestedHeaderHash(nil)
		boot.mutRcvHdrHash.Unlock()
		boot.chRcvHdrHash <- true

		return
	}
	boot.mutRcvHdrHash.Unlock()
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

// waitForHeaderNonce method wait for header with the requested nonce to be received
func (boot *baseBootstrap) waitForHeaderNonce() error {
	select {
	case <-boot.chRcvHdrNonce:
		return nil
	case <-time.After(boot.waitTime):
		return process.ErrTimeIsOut
	}
}

// waitForHeaderHash method wait for header with the requested hash to be received
func (boot *baseBootstrap) waitForHeaderHash() error {
	select {
	case <-boot.chRcvHdrHash:
		return nil
	case <-time.After(boot.waitTime):
		return process.ErrTimeIsOut
	}
}

func (boot *baseBootstrap) computeNodeState() {
	boot.mutNodeState.Lock()
	defer boot.mutNodeState.Unlock()

	isNodeStateCalculatedInCurrentRound := boot.roundIndex == boot.roundHandler.Index() && boot.isNodeStateCalculated
	if isNodeStateCalculatedInCurrentRound {
		return
	}

	boot.forkInfo = boot.forkDetector.CheckFork()

	genesisNonce := boot.chainHandler.GetGenesisHeader().GetNonce()
	currentHeader := boot.chainHandler.GetCurrentBlockHeader()
	if check.IfNil(currentHeader) {
		boot.hasLastBlock = boot.forkDetector.ProbableHighestNonce() == genesisNonce
		boot.log.Debug("computeNodeState",
			"probableHighestNonce", boot.forkDetector.ProbableHighestNonce(),
			"currentBlockNonce", nil,
			"boot.hasLastBlock", boot.hasLastBlock)
	} else {
		boot.hasLastBlock = boot.forkDetector.ProbableHighestNonce() <= boot.chainHandler.GetCurrentBlockHeader().GetNonce()
		boot.log.Debug("computeNodeState",
			"probableHighestNonce", boot.forkDetector.ProbableHighestNonce(),
			"currentBlockNonce", boot.chainHandler.GetCurrentBlockHeader().GetNonce(),
			"boot.hasLastBlock", boot.hasLastBlock)
	}

	isNodeConnectedToTheNetwork := boot.networkWatcher.IsConnectedToTheNetwork()
	isNodeSynchronized := !boot.forkInfo.IsDetected && boot.hasLastBlock && isNodeConnectedToTheNetwork
	if isNodeSynchronized != boot.isNodeSynchronized {
		boot.log.Debug("node has changed its synchronized state",
			"state", isNodeSynchronized,
		)
	}

	boot.isNodeSynchronized = isNodeSynchronized
	boot.isNodeStateCalculated = true
	boot.roundIndex = boot.roundHandler.Index()
	boot.notifySyncStateListeners(isNodeSynchronized)

	result := uint64(1)
	if isNodeSynchronized {
		result = uint64(0)
	}

	boot.statusHandler.SetUInt64Value(common.MetricIsSyncing, result)
	boot.log.Debug("computeNodeState",
		"isNodeStateCalculated", boot.isNodeStateCalculated,
		"isNodeSynchronized", boot.isNodeSynchronized)

	if boot.shouldTryToRequestHeaders() {
		go boot.requestHeadersIfSyncIsStuck()
	}
}

func (boot *baseBootstrap) shouldTryToRequestHeaders() bool {
	if boot.roundHandler.BeforeGenesis() {
		return false
	}
	if boot.isForcedRollBackOneBlock() {
		return false
	}
	if boot.isForcedRollBackToNonce() {
		return false
	}
	if !boot.isNodeSynchronized {
		return true
	}

	return boot.roundHandler.Index()%process.RoundModulusTriggerWhenSyncIsStuck == 0
}

func (boot *baseBootstrap) requestHeadersIfSyncIsStuck() {
	lastSyncedRound := boot.chainHandler.GetGenesisHeader().GetRound()
	currHeader := boot.chainHandler.GetCurrentBlockHeader()
	if !check.IfNil(currHeader) {
		lastSyncedRound = currHeader.GetRound()
	}

	roundDiff := uint64(boot.roundHandler.Index()) - lastSyncedRound
	if roundDiff <= process.MaxRoundsWithoutNewBlockReceived {
		return
	}

	fromNonce := boot.getNonceForNextBlock()
	numHeadersToRequest := core.MinUint64(process.MaxHeadersToRequestInAdvance, roundDiff-1)
	toNonce := fromNonce + numHeadersToRequest - 1

	if fromNonce > toNonce {
		return
	}

	boot.log.Debug("requestHeadersIfSyncIsStuck",
		"from nonce", fromNonce,
		"to nonce", toNonce,
		"probable highest nonce", boot.forkDetector.ProbableHighestNonce())

	boot.requestHeaders(fromNonce, toNonce)
}

func (boot *baseBootstrap) removeHeaderFromPools(header data.HeaderHandler) []byte {
	hash, err := core.CalculateHash(boot.marshalizer, boot.hasher, header)
	if err != nil {
		boot.log.Debug("CalculateHash", "error", err.Error())
		return nil
	}

	boot.log.Debug("removeHeaderFromPools",
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
	boot.log.Debug("removeHeadersHigherThanNonceFromPool",
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
	if check.IfNil(arguments.RoundHandler) {
		return process.ErrNilRoundHandler
	}
	if check.IfNil(arguments.BlockProcessor) {
		return process.ErrNilBlockProcessor
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
	if check.IfNil(arguments.EnableEpochsHandler) {
		return process.ErrNilEnableEpochsHandler
	}

	return nil
}

func (boot *baseBootstrap) requestHeadersFromNonceIfMissing(fromNonce uint64) {
	toNonce := core.MinUint64(fromNonce+process.MaxHeadersToRequestInAdvance-1, boot.forkDetector.ProbableHighestNonce())

	if fromNonce > toNonce {
		return
	}

	boot.log.Debug("requestHeadersFromNonceIfMissing",
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
			boot.log.Debug("bootstrap's go routine is stopping...")
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
				boot.log.Debug("SyncBlock finished, bootstrap's go routine is stopping...")
				return
			}

			boot.log.Debug("SyncBlock", "error", err.Error())
		}
	}
}

func (boot *baseBootstrap) doJobOnSyncBlockFail(bodyHandler data.BodyHandler, headerHandler data.HeaderHandler, err error) {
	processBlockStarted := !check.IfNil(bodyHandler) && !check.IfNil(headerHandler)
	isProcessWithError := processBlockStarted && err != process.ErrTimeIsOut

	numSyncedWithErrors := boot.incrementSyncedWithErrorsForNonce(boot.getNonceForNextBlock())
	allowedSyncWithErrorsLimitReached := numSyncedWithErrors >= process.MaxSyncWithErrorsAllowed
	isInProperRound := process.IsInProperRound(boot.roundHandler.Index())
	isSyncWithErrorsLimitReachedInProperRound := allowedSyncWithErrorsLimitReached && isInProperRound

	shouldRollBack := isProcessWithError || isSyncWithErrorsLimitReachedInProperRound
	if shouldRollBack {
		if !check.IfNil(headerHandler) {
			hash := boot.removeHeaderFromPools(headerHandler)
			boot.forkDetector.RemoveHeader(headerHandler.GetNonce(), hash)
		}

		errNotCritical := boot.rollBack(false)
		if errNotCritical != nil {
			boot.log.Debug("rollBack", "error", errNotCritical.Error())
		}

		if isSyncWithErrorsLimitReachedInProperRound {
			boot.forkDetector.ResetProbableHighestNonce()
			boot.removeHeadersHigherThanNonceFromPool(boot.getNonceForCurrentBlock())
		}
	}
}

func (boot *baseBootstrap) incrementSyncedWithErrorsForNonce(nonce uint64) uint32 {
	boot.mutNonceSyncedWithErrors.Lock()
	boot.mapNonceSyncedWithErrors[nonce]++
	numSyncedWithErrors := boot.mapNonceSyncedWithErrors[nonce]
	boot.mutNonceSyncedWithErrors.Unlock()

	return numSyncedWithErrors
}

// syncBlock method actually does the synchronization. It requests the next block header from the pool
// and if it is not found there it will be requested from the network. After the header is received,
// it requests the block body in the same way(pool and then, if it is not found in the pool, from network).
// If either header and body are received the ProcessBlock and CommitBlock method will be called successively.
// These methods will execute the block and its transactions. Finally, if everything works, the block will be committed
// in the blockchain, and all this mechanism will be reiterated for the next block.
func (boot *baseBootstrap) syncBlock() error {
	boot.computeNodeState()
	nodeState := boot.GetNodeState()
	if nodeState != common.NsNotSynchronized {
		return nil
	}

	defer func() {
		boot.mutNodeState.Lock()
		boot.isNodeStateCalculated = false
		boot.mutNodeState.Unlock()
	}()

	if boot.forkInfo.IsDetected {
		boot.statusHandler.Increment(common.MetricNumTimesInForkChoice)

		if boot.isForcedRollBackOneBlock() {
			boot.log.Debug("roll back one block has been forced")
			boot.rollBackOneBlockForced()
			return nil
		}

		if boot.isForcedRollBackToNonce() {
			boot.log.Debug("roll back to nonce has been forced", "nonce", boot.forkInfo.Nonce)
			boot.rollBackToNonceForced()
			return nil
		}

		boot.log.Debug("fork detected",
			"nonce", boot.forkInfo.Nonce,
			"hash", boot.forkInfo.Hash,
		)
		err := boot.rollBack(true)
		if err != nil {
			return err
		}
	}

	var body data.BodyHandler
	var header data.HeaderHandler
	var err error

	defer func() {
		if err != nil {
			boot.doJobOnSyncBlockFail(body, header, err)
		}
	}()

	header, headerHash, err := boot.getNextHeaderRequestingIfMissing()
	if err != nil {
		return err
	}

	go boot.requestHeadersFromNonceIfMissing(header.GetNonce() + 1)

	err = boot.handleEquivalentProof(header, headerHash)
	if err != nil {
		return err
	}

	body, err = boot.blockBootstrapper.getBlockBodyRequestingIfMissing(header)
	if err != nil {
		return err
	}

	startTime := time.Now()
	waitTime := boot.processWaitTime
	haveTime := func() time.Duration {
		return waitTime - time.Since(startTime)
	}

	startProcessBlockTime := time.Now()
	err = boot.blockProcessor.ProcessBlock(header, body, haveTime)
	elapsedTime := time.Since(startProcessBlockTime)
	boot.log.Debug("elapsed time to process block",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	startProcessScheduledBlockTime := time.Now()
	err = boot.blockProcessor.ProcessScheduledBlock(header, body, haveTime)
	elapsedTime = time.Since(startProcessScheduledBlockTime)
	boot.log.Debug("elapsed time to process scheduled block",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	startCommitBlockTime := time.Now()
	err = boot.blockProcessor.CommitBlock(header, body)
	elapsedTime = time.Since(startCommitBlockTime)
	if elapsedTime >= common.CommitMaxTime {
		boot.log.Warn("syncBlock.CommitBlock", "elapsed time", elapsedTime)
	} else {
		boot.log.Debug("elapsed time to commit block",
			"time [s]", elapsedTime,
		)
	}
	if err != nil {
		return err
	}

	boot.log.Debug("block has been synced successfully",
		"nonce", header.GetNonce(),
	)

	boot.cleanNoncesSyncedWithErrorsBehindFinal()
	boot.cleanProofsBehindFinal(header)

	return nil
}

func (boot *baseBootstrap) handleEquivalentProof(
	header data.HeaderHandler,
	headerHash []byte,
) error {
	if !boot.enableEpochsHandler.IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, header.GetEpoch()) {
		return nil
	}

	// process block only if there is a proof for it
	hasProof := boot.proofs.HasProof(header.GetShardID(), headerHash)
	if hasProof {
		return nil
	}

	boot.log.Trace("baseBootstrap.handleEquivalentProof: did not have proof for header, will try again", "headerHash", headerHash)

	_, _, err := boot.blockBootstrapper.getHeaderWithNonceRequestingIfMissing(header.GetNonce() + 1)
	if err != nil {
		return err
	}

	hasProof = boot.proofs.HasProof(header.GetShardID(), headerHash)
	if !hasProof {
		return fmt.Errorf("baseBootstrap.handleEquivalentProof: did not have proof for header, headerHash %s", hex.EncodeToString(headerHash))
	}

	return nil
}

func (boot *baseBootstrap) handleTrieSyncError(err error, ctx context.Context) {
	shouldOutputLog := err != nil && !common.IsContextDone(ctx)
	if shouldOutputLog {
		boot.log.Debug("SyncBlock syncTrie", "error", err)
	}
}

func (boot *baseBootstrap) syncUserAccountsState(key []byte) error {
	boot.log.Warn("base sync: started syncUserAccountsState")
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
}

func (boot *baseBootstrap) cleanProofsBehindFinal(header data.HeaderHandler) {
	if !boot.enableEpochsHandler.IsFlagEnabledInEpoch(common.EquivalentMessagesFlag, header.GetEpoch()) {
		return
	}

	// TODO: analyse fork detection by proofs
	finalNonce := boot.forkDetector.GetHighestFinalBlockNonce()

	err := boot.proofs.CleanupProofsBehindNonce(header.GetShardID(), finalNonce)
	if err != nil {
		boot.log.Warn("failed to cleanup notarized proofs behind nonce",
			"nonce", finalNonce,
			"shardID", header.GetShardID(),
			"error", err)
	}

	boot.log.Trace("baseBootstrap.cleanProofsBehindFinal cleanup successfully", "finalNonce", finalNonce)
}

// rollBack decides if rollBackOneBlock must be called
func (boot *baseBootstrap) rollBack(revertUsingForkNonce bool) error {
	var roleBackOneBlockExecuted bool
	var err error
	var currHeaderHash []byte
	var currHeader data.HeaderHandler
	var prevHeader data.HeaderHandler
	var currBody data.BodyHandler

	defer func() {
		if !roleBackOneBlockExecuted {
			err = boot.scheduledTxsExecutionHandler.RollBackToBlock(currHeaderHash)
			if err != nil {
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

	boot.log.Debug("starting roll back")
	for {
		currHeaderHash = boot.chainHandler.GetCurrentBlockHeaderHash()
		currHeader, err = boot.blockBootstrapper.getCurrHeader()
		if err != nil {
			return err
		}

		allowRollBack := boot.shouldAllowRollback(currHeader, currHeaderHash)
		if !revertUsingForkNonce && !allowRollBack {
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

		boot.log.Debug("roll back to block",
			"nonce", currHeader.GetNonce()-1,
			"hash", currHeader.GetPrevHash(),
		)
		boot.log.Debug("highest final block nonce",
			"nonce", boot.forkDetector.GetHighestFinalBlockNonce(),
		)

		currBody, err = boot.rollBackOneBlock(
			currHeaderHash,
			currHeader,
			prevHeaderHash,
			prevHeader,
		)
		roleBackOneBlockExecuted = true
		if err != nil {
			return err
		}

		_, _ = metricsLoader.UpdateMetricsFromStorage(boot.store, boot.uint64Converter, boot.marshalizer, boot.statusHandler, prevHeader.GetNonce())

		err = boot.bootStorer.SaveLastRound(int64(prevHeader.GetRound()))
		if err != nil {
			boot.log.Debug("save last round in storage",
				"error", err.Error(),
				"round", prevHeader.GetRound(),
			)
		}

		err = boot.historyRepo.RevertBlock(currHeader, currBody)
		if err != nil {
			boot.log.Debug("boot.historyRepo.RevertBlock",
				"error", err.Error(),
			)

			return err
		}

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

		err = boot.outportHandler.RevertIndexedBlock(&outportcore.HeaderDataWithBody{
			Body:       currBody,
			HeaderHash: currHeaderHash,
			Header:     currHeader,
		})
		if err != nil {
			boot.log.Warn("baseBootstrap.outportHandler.RevertIndexedBlock cannot revert indexed block", "error", err)
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

	boot.log.Debug("ending roll back")
	return nil
}

func (boot *baseBootstrap) shouldAllowRollback(currHeader data.HeaderHandler, currHeaderHash []byte) bool {
	finalBlockNonce := boot.forkDetector.GetHighestFinalBlockNonce()
	finalBlockHash := boot.forkDetector.GetHighestFinalBlockHash()
	isRollBackBehindFinal := currHeader.GetNonce() <= finalBlockNonce
	isFinalBlockRollBack := currHeader.GetNonce() == finalBlockNonce
	canRollbackBlock := boot.canRollbackBlock(currHeader)

	headerWithScheduledMiniBlocks := currHeader.HasScheduledMiniBlocks()
	headerHashDoesNotMatchWithFinalBlockHash := !bytes.Equal(currHeaderHash, finalBlockHash)
	allowFinalBlockRollBack := (headerWithScheduledMiniBlocks || headerHashDoesNotMatchWithFinalBlockHash) && isFinalBlockRollBack && canRollbackBlock
	allowRollBack := !isRollBackBehindFinal || allowFinalBlockRollBack

	boot.log.Debug("baseBootstrap.shouldAllowRollback",
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

func (boot *baseBootstrap) canRollbackBlock(currHeader data.HeaderHandler) bool {
	firstCommittedNonce := boot.blockProcessor.NonceOfFirstCommittedBlock()

	return currHeader.GetNonce() >= firstCommittedNonce.Value && firstCommittedNonce.HasValue
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
		boot.log.Debug("rollBackOneBlock getBlockBody error", "error", errNotCritical)
	}

	err = boot.blockProcessor.RestoreBlockIntoPools(currHeader, currBlockBody)
	if err != nil {
		return nil, err
	}

	boot.cleanCachesAndStorageOnRollback(currHeader)

	return currBlockBody, nil
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

	hash := boot.forkDetector.GetNotarizedHeaderHash(nonce)
	if boot.forkInfo.IsDetected {
		hash = boot.forkInfo.Hash
	}

	if hash != nil {
		header, err := boot.blockBootstrapper.getHeaderWithHashRequestingIfMissing(hash)
		return header, hash, err
	}

	return boot.blockBootstrapper.getHeaderWithNonceRequestingIfMissing(nonce)
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
	err := boot.rollBack(false)
	if err != nil {
		boot.log.Debug("rollBackOneBlockForced", "error", err.Error())
	}

	boot.forkDetector.ResetFork()
	boot.removeHeadersHigherThanNonceFromPool(boot.getNonceForCurrentBlock())
}

func (boot *baseBootstrap) rollBackToNonceForced() {
	err := boot.rollBack(true)
	if err != nil {
		boot.log.Debug("rollBackToNonceForced", "error", err.Error())
	}

	boot.forkDetector.ResetProbableHighestNonce()
	boot.removeHeadersHigherThanNonceFromPool(boot.getNonceForCurrentBlock())
}

func (boot *baseBootstrap) restoreState(
	currHeaderHash []byte,
	currHeader data.HeaderHandler,
	currRootHash []byte,
) {
	boot.log.Debug("revert state to header",
		"nonce", currHeader.GetNonce(),
		"hash", currHeaderHash,
		"current root hash", currRootHash)

	err := boot.chainHandler.SetCurrentBlockHeaderAndRootHash(currHeader, currRootHash)
	if err != nil {
		boot.log.Debug("SetCurrentBlockHeader", "error", err.Error())
	}

	boot.chainHandler.SetCurrentBlockHeaderHash(currHeaderHash)

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
		boot.log.Debug("RevertState", "error", err.Error())
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
		boot.log.Debug("received all the requested mini blocks from network")
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
	boot.log.Debug("requesting mini blocks from network",
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
	case <-time.After(boot.waitTime):
		return process.ErrTimeIsOut
	}
}

func (boot *baseBootstrap) init() {
	boot.forkInfo = process.NewForkInfo()

	boot.chRcvHdrNonce = make(chan bool)
	boot.chRcvHdrHash = make(chan bool)
	boot.chRcvMiniBlocks = make(chan bool)

	boot.setRequestedHeaderNonce(nil)
	boot.setRequestedHeaderHash(nil)
	boot.setRequestedMiniBlocks(nil)

	boot.poolsHolder.MiniBlocks().RegisterHandler(boot.receivedMiniblock, core.UniqueIdentifier())
	boot.headers.RegisterHandler(boot.processReceivedHeader)
	boot.proofs.RegisterHandler(boot.processReceivedProof)

	boot.syncStateListeners = make([]func(bool), 0)
	boot.requestedHashes = process.RequiredDataPool{}
	boot.mapNonceSyncedWithErrors = make(map[uint64]uint32)
}

func (boot *baseBootstrap) requestHeaders(fromNonce uint64, toNonce uint64) {
	boot.mutRequestHeaders.Lock()
	defer boot.mutRequestHeaders.Unlock()

	for currentNonce := fromNonce; currentNonce <= toNonce; currentNonce++ {
		haveHeader := boot.blockBootstrapper.haveHeaderInPoolWithNonce(currentNonce)
		if haveHeader {
			continue
		}

		boot.blockBootstrapper.requestHeaderByNonce(currentNonce)
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

	return nil
}

func (boot *baseBootstrap) cleanChannels() {
	nrReads := core.EmptyChannel(boot.chRcvHdrNonce)
	boot.log.Debug("close baseSync: emptied channel", "chRcvHdrNonce nrReads", nrReads)

	nrReads = core.EmptyChannel(boot.chRcvHdrHash)
	boot.log.Debug("close baseSync: emptied channel", "chRcvHdrHash nrReads", nrReads)

	nrReads = core.EmptyChannel(boot.chRcvMiniBlocks)
	boot.log.Debug("close baseSync: emptied channel", "chRcvMiniBlocks nrReads", nrReads)
}

// IsInterfaceNil returns true if there is no value under the interface
func (boot *baseBootstrap) IsInterfaceNil() bool {
	return boot == nil
}
