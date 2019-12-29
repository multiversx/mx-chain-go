package sync

import (
	"bytes"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("process/sync")

// sleepTime defines the time in milliseconds between each iteration made in syncBlocks method
const sleepTime = 5 * time.Millisecond

// HdrInfo hold the data related to a header
type HdrInfo struct {
	Nonce uint64
	Hash  []byte
}

type notarizedInfo struct {
	lastNotarized           map[uint32]*HdrInfo
	finalNotarized          map[uint32]*HdrInfo
	blockWithLastNotarized  map[uint32]uint64
	blockWithFinalNotarized map[uint32]uint64
	startNonce              uint64
}

func (ni *notarizedInfo) reset() {
	ni.lastNotarized = make(map[uint32]*HdrInfo, 0)
	ni.finalNotarized = make(map[uint32]*HdrInfo, 0)
	ni.blockWithLastNotarized = make(map[uint32]uint64, 0)
	ni.blockWithFinalNotarized = make(map[uint32]uint64, 0)
	ni.startNonce = uint64(0)
}

type baseBootstrap struct {
	headers       storage.Cacher
	headersNonces dataRetriever.Uint64SyncMapCacher

	blkc        data.ChainHandler
	blkExecutor process.BlockProcessor
	store       dataRetriever.StorageService

	rounder           consensus.Rounder
	hasher            hashing.Hasher
	marshalizer       marshal.Marshalizer
	forkDetector      process.ForkDetector
	shardCoordinator  sharding.Coordinator
	accounts          state.AccountsAdapter
	blockBootstrapper blockBootstrapper
	blackListHandler  process.BlackListHandler

	mutHeader     sync.RWMutex
	headerNonce   *uint64
	headerhash    []byte
	chRcvHdrNonce chan bool
	chRcvHdrHash  chan bool

	requestedHashes process.RequiredDataPool

	statusHandler core.AppStatusHandler

	chStopSync chan bool
	waitTime   time.Duration

	mutNodeSynched     sync.RWMutex
	isNodeSynchronized bool
	hasLastBlock       bool
	roundIndex         int64

	forkInfo *process.ForkInfo

	mutRcvHdrNonce        sync.RWMutex
	mutRcvHdrHash         sync.RWMutex
	syncStateListeners    []func(bool)
	mutSyncStateListeners sync.RWMutex
	uint64Converter       typeConverters.Uint64ByteSliceConverter
	requestsWithTimeout   uint32

	requestMiniBlocks func(uint32, uint64)

	networkWatcher    process.NetworkConnectionWatcher
	getHeaderFromPool func([]byte) (data.HeaderHandler, error)

	headerStore           storage.Storer
	headerNonceHashStore  storage.Storer
	hdrRes                dataRetriever.HeaderResolver
	syncStarter           syncStarter
	bootStorer            process.BootStorer
	storageBootstrapper   process.BootstrapperFromStorage
	requestedItemsHandler dataRetriever.RequestedItemsHandler

	miniBlocks         storage.Cacher
	chRcvMiniBlocks    chan bool
	mutRcvMiniBlocks   sync.Mutex
	miniBlocksResolver dataRetriever.MiniBlocksResolver
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

func (boot *baseBootstrap) processReceivedHeader(headerHandler data.HeaderHandler, headerHash []byte) {
	log.Trace("received header from network",
		"shard", headerHandler.GetShardID(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
		"hash", headerHash,
	)

	err := boot.forkDetector.AddHeader(headerHandler, headerHash, process.BHReceived, nil, nil, false)
	if err != nil {
		log.Debug("forkDetector.AddHeader", "error", err.Error())
	}

	boot.mutRcvHdrHash.Lock()
	hash := boot.requestedHeaderHash()
	if hash == nil {
		boot.mutRcvHdrHash.Unlock()
		return
	}

	if bytes.Equal(hash, headerHash) {
		log.Debug("received requested header from network",
			"shard", headerHandler.GetShardID(),
			"round", headerHandler.GetRound(),
			"nonce", headerHandler.GetNonce(),
			"hash", hash,
		)
		boot.setRequestedHeaderHash(nil)
		boot.mutRcvHdrHash.Unlock()
		boot.chRcvHdrHash <- true
	} else {
		boot.mutRcvHdrHash.Unlock()
	}
}

// receivedHeaderNonce method is a call back function which is called when a new header is added
// in the block headers pool
func (boot *baseBootstrap) receivedHeaderNonce(nonce uint64, shardId uint32, hash []byte) {
	if boot.shardCoordinator.SelfId() != shardId {
		return
	}

	log.Trace("received header from network",
		"shard", shardId,
		"nonce", nonce,
		"hash", hash,
	)

	err := boot.addReceivedHeaderToForkDetector(hash)
	if err != nil {
		log.Debug("addReceivedHeaderToForkDetector", "error", err.Error())
	}

	go boot.requestMiniBlocks(shardId, nonce)

	boot.mutRcvHdrNonce.Lock()
	n := boot.requestedHeaderNonce()
	if n == nil {
		boot.mutRcvHdrNonce.Unlock()
		return
	}

	if *n == nonce {
		log.Debug("received requested header from network",
			"shard", shardId,
			"nonce", nonce,
			"hash", hash,
		)
		boot.setRequestedHeaderNonce(nil)
		boot.mutRcvHdrNonce.Unlock()
		boot.chRcvHdrNonce <- true
	} else {
		boot.mutRcvHdrNonce.Unlock()
	}
}

// AddSyncStateListener adds a syncStateListener that get notified each time the sync status of the node changes
func (boot *baseBootstrap) AddSyncStateListener(syncStateListener func(isSyncing bool)) {
	boot.mutSyncStateListeners.Lock()
	boot.syncStateListeners = append(boot.syncStateListeners, syncStateListener)
	boot.mutSyncStateListeners.Unlock()
}

// SetStatusHandler will set the instance of the AppStatusHandler
func (boot *baseBootstrap) SetStatusHandler(handler core.AppStatusHandler) error {
	if handler == nil || handler.IsInterfaceNil() {
		return process.ErrNilAppStatusHandler
	}
	boot.statusHandler = handler

	return nil
}

func (boot *baseBootstrap) notifySyncStateListeners(isNodeSynchronized bool) {
	boot.mutSyncStateListeners.RLock()
	for i := 0; i < len(boot.syncStateListeners); i++ {
		go boot.syncStateListeners[i](isNodeSynchronized)
	}
	boot.mutSyncStateListeners.RUnlock()
}

// getNonceForNextBlock will get the nonce for the next block we should request
func (boot *baseBootstrap) getNonceForNextBlock() uint64 {
	nonce := uint64(1) // first block nonce after genesis block
	if boot.blkc != nil && boot.blkc.GetCurrentBlockHeader() != nil {
		nonce = boot.blkc.GetCurrentBlockHeader().GetNonce() + 1
	}
	return nonce
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

// ShouldSync method returns the sync state of the node. If it returns 'true', this means that the node
// is not synchronized yet and it has to continue the bootstrapping mechanism, otherwise the node is already
// synced and it can participate to the consensus, if it is in the jobDone group of this rounder.
// Note that when the node is not connected to the network, ShouldSync returns true but the SyncBlock
// is not automatically called
func (boot *baseBootstrap) ShouldSync() bool {
	boot.mutNodeSynched.Lock()
	defer boot.mutNodeSynched.Unlock()

	isNodeSynchronizedInCurrentRound := boot.roundIndex == boot.rounder.Index() && boot.isNodeSynchronized
	if isNodeSynchronizedInCurrentRound {
		return false
	}

	boot.forkInfo = boot.forkDetector.CheckFork()

	if boot.blkc.GetCurrentBlockHeader() == nil {
		boot.hasLastBlock = boot.forkDetector.ProbableHighestNonce() <= 0
	} else {
		boot.hasLastBlock = boot.forkDetector.ProbableHighestNonce() <= boot.blkc.GetCurrentBlockHeader().GetNonce()
	}

	isNodeConnectedToTheNetwork := boot.networkWatcher.IsConnectedToTheNetwork()

	isNodeSynchronized := !boot.forkInfo.IsDetected && boot.hasLastBlock && isNodeConnectedToTheNetwork
	if isNodeSynchronized != boot.isNodeSynchronized {
		log.Debug("node has changed its synchronized state",
			"state", isNodeSynchronized,
		)
		boot.isNodeSynchronized = isNodeSynchronized
		boot.notifySyncStateListeners(isNodeSynchronized)
	}

	boot.roundIndex = boot.rounder.Index()

	var result uint64
	if isNodeSynchronized {
		result = uint64(0)
	} else {
		result = uint64(1)
	}
	boot.statusHandler.SetUInt64Value(core.MetricIsSyncing, result)

	return !isNodeSynchronized
}

func (boot *baseBootstrap) removeHeaderFromPools(header data.HeaderHandler) []byte {
	boot.headersNonces.Remove(header.GetNonce(), header.GetShardID())

	hash, err := core.CalculateHash(boot.marshalizer, boot.hasher, header)
	if err != nil {
		log.Debug("CalculateHash", "error", err.Error())
		return nil
	}

	return hash
}

func (boot *baseBootstrap) cleanCachesAndStorageOnRollback(header data.HeaderHandler) {
	hash := boot.removeHeaderFromPools(header)
	boot.forkDetector.RemoveHeaders(header.GetNonce(), hash)
	nonceToByteSlice := boot.uint64Converter.ToByteSlice(header.GetNonce())
	_ = boot.headerNonceHashStore.Remove(nonceToByteSlice)
}

// checkBootstrapNilParameters will check the imput parameters for nil values
func checkBootstrapNilParameters(
	blkc data.ChainHandler,
	rounder consensus.Rounder,
	blkExecutor process.BlockProcessor,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	forkDetector process.ForkDetector,
	resolversFinder dataRetriever.ResolversContainer,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
	store dataRetriever.StorageService,
	blackListHandler process.BlackListHandler,
	watcher process.NetworkConnectionWatcher,
	requestedItemsHandler dataRetriever.RequestedItemsHandler,
) error {
	if check.IfNil(blkc) {
		return process.ErrNilBlockChain
	}
	if check.IfNil(rounder) {
		return process.ErrNilRounder
	}
	if check.IfNil(blkExecutor) {
		return process.ErrNilBlockExecutor
	}
	if check.IfNil(hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(forkDetector) {
		return process.ErrNilForkDetector
	}
	if check.IfNil(resolversFinder) {
		return process.ErrNilResolverContainer
	}
	if check.IfNil(shardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(accounts) {
		return process.ErrNilAccountsAdapter
	}
	if check.IfNil(store) {
		return process.ErrNilStore
	}
	if check.IfNil(blackListHandler) {
		return process.ErrNilBlackListHandler
	}
	if check.IfNil(watcher) {
		return process.ErrNilNetworkWatcher
	}
	if check.IfNil(requestedItemsHandler) {
		return dataRetriever.ErrNilRequestedItemsHandler
	}

	return nil
}

// isSigned verifies if a block is signed
func isSigned(header data.HeaderHandler) bool {
	// TODO: Later, here it should be done a more complex verification (signature for this round matches with the bitmap,
	// and validators which signed here, were in this round consensus group)
	bitmap := header.GetPubKeysBitmap()
	isBitmapEmpty := bytes.Equal(bitmap, make([]byte, len(bitmap)))

	return !isBitmapEmpty
}

// isRandomSeedValid verifies if the random seed is valid (equal with a signed previous rand seed)
func isRandomSeedValid(header data.HeaderHandler) bool {
	// TODO: Later, here should be done a more complex verification (random seed should be equal with the previous rand
	// seed signed by the proposer of this round)
	prevRandSeed := header.GetPrevRandSeed()
	randSeed := header.GetRandSeed()
	isPrevRandSeedNilOrEmpty := len(prevRandSeed) == 0
	isRandSeedNilOrEmpty := len(randSeed) == 0

	return !isPrevRandSeedNilOrEmpty && !isRandSeedNilOrEmpty
}

func (boot *baseBootstrap) requestHeadersFromNonceIfMissing(
	nonce uint64,
	haveHeaderInPoolWithNonce func(uint64) bool,
	hdrRes dataRetriever.HeaderResolver) {

	boot.requestedItemsHandler.Sweep()

	nbRequestedHdrs := 0
	maxNonce := core.MinUint64(nonce+process.MaxHeadersToRequestInAdvance-1, boot.forkDetector.ProbableHighestNonce())
	for currentNonce := nonce; currentNonce <= maxNonce; currentNonce++ {
		key := fmt.Sprintf("%d-%d", boot.shardCoordinator.SelfId(), currentNonce)
		if boot.requestedItemsHandler.Has(key) {
			continue
		}

		haveHeader := haveHeaderInPoolWithNonce(nonce)
		if !haveHeader {
			err := hdrRes.RequestDataFromNonce(currentNonce)
			if err != nil {
				log.Debug("RequestDataFromNonce", "error", err.Error())
				continue
			}

			err = boot.requestedItemsHandler.Add(key)
			if err != nil {
				log.Trace("add requested item with error", "error", err.Error())
			}

			nbRequestedHdrs++
		}
	}

	if nbRequestedHdrs > 0 {
		log.Debug("requested in advance headers",
			"num headers", nbRequestedHdrs,
			"from nonce", nonce,
			"to", maxNonce,
		)
		log.Debug("probable highest nonce",
			"nonce", boot.forkDetector.ProbableHighestNonce(),
		)
	}
}

// StopSync method will stop SyncBlocks
func (boot *baseBootstrap) StopSync() {
	boot.chStopSync <- true
}

// syncBlocks method calls repeatedly synchronization method SyncBlock
func (boot *baseBootstrap) syncBlocks() {
	for {
		time.Sleep(sleepTime)

		if !boot.networkWatcher.IsConnectedToTheNetwork() {
			continue
		}

		select {
		case <-boot.chStopSync:
			return
		default:
			err := boot.syncStarter.SyncBlock()
			if err != nil {
				log.Debug("SyncBlock", "error", err.Error())
			}
		}
	}
}

func (boot *baseBootstrap) doJobOnSyncBlockFail(headerHandler data.HeaderHandler, err error) {
	if err == process.ErrTimeIsOut {
		boot.requestsWithTimeout++
	}

	allowedRequestsWithTimeOutHaveReached := boot.requestsWithTimeout >= process.MaxRequestsWithTimeoutAllowed
	isInProperRound := process.IsInProperRound(boot.rounder.Index())

	shouldRollBack := err != process.ErrTimeIsOut || (allowedRequestsWithTimeOutHaveReached && isInProperRound)
	if shouldRollBack {
		boot.requestsWithTimeout = 0

		if headerHandler != nil {
			hash := boot.removeHeaderFromPools(headerHandler)
			boot.forkDetector.RemoveHeaders(headerHandler.GetNonce(), hash)
		}

		errNotCritical := boot.rollBack(false)
		if errNotCritical != nil {
			log.Debug("rollBack", "error", errNotCritical.Error())
		}

		if allowedRequestsWithTimeOutHaveReached && isInProperRound {
			boot.forkDetector.ResetProbableHighestNonce()
		}
	}
}

// syncBlock method actually does the synchronization. It requests the next block header from the pool
// and if it is not found there it will be requested from the network. After the header is received,
// it requests the block body in the same way(pool and than, if it is not found in the pool, from network).
// If either header and body are received the ProcessBlock and CommitBlock method will be called successively.
// These methods will execute the block and its transactions. Finally if everything works, the block will be committed
// in the blockchain, and all this mechanism will be reiterated for the next block.
func (boot *baseBootstrap) syncBlock() error {
	if !boot.ShouldSync() {
		return nil
	}

	if boot.forkInfo.IsDetected {
		boot.statusHandler.Increment(core.MetricNumTimesInForkChoice)

		if boot.isForcedFork() {
			log.Debug("fork has been forced")
			boot.rollBackOnForcedFork()
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

	var hdr data.HeaderHandler
	var err error

	defer func() {
		if err != nil {
			boot.doJobOnSyncBlockFail(hdr, err)
		}
	}()

	hdr, err = boot.getNextHeaderRequestingIfMissing()
	if err != nil {
		return err
	}

	go boot.requestHeadersFromNonceIfMissing(hdr.GetNonce()+1, boot.blockBootstrapper.haveHeaderInPoolWithNonce, boot.hdrRes)

	blockBody, err := boot.blockBootstrapper.getBlockBodyRequestingIfMissing(hdr)
	if err != nil {
		return err
	}

	haveTime := func() time.Duration {
		return boot.rounder.TimeDuration()
	}

	startTime := time.Now()
	err = boot.blkExecutor.ProcessBlock(boot.blkc, hdr, blockBody, haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to process block",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	startTime = time.Now()
	err = boot.blkExecutor.CommitBlock(boot.blkc, hdr, blockBody)
	elapsedTime = time.Since(startTime)
	log.Debug("elapsed time to commit block",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	log.Debug("block has been synced successfully",
		"nonce", hdr.GetNonce(),
	)
	boot.requestsWithTimeout = 0

	return nil
}

// rollBack decides if rollBackOneBlock must be called
func (boot *baseBootstrap) rollBack(revertUsingForkNonce bool) error {
	if boot.headerStore == nil {
		return process.ErrNilHeadersStorage
	}
	if boot.headerNonceHashStore == nil {
		return process.ErrNilHeadersNonceHashStorage
	}

	log.Debug("starting roll back")
	for {
		currHeaderHash := boot.blkc.GetCurrentBlockHeaderHash()
		currHeader, err := boot.blockBootstrapper.getCurrHeader()
		if err != nil {
			return err
		}
		if !revertUsingForkNonce && currHeader.GetNonce() <= boot.forkDetector.GetHighestFinalBlockNonce() {
			return ErrRollBackBehindFinalHeader
		}

		shouldEndRollBack := revertUsingForkNonce && currHeader.GetNonce() < boot.forkInfo.Nonce
		if shouldEndRollBack {
			return ErrRollBackBehindForkNonce
		}

		currBlockBody, err := boot.blockBootstrapper.getBlockBody(currHeader)
		if err != nil {
			return err
		}
		prevHeaderHash := currHeader.GetPrevHash()
		prevHeader, err := boot.blockBootstrapper.getPrevHeader(currHeader, boot.headerStore)
		if err != nil {
			return err
		}
		prevBlockBody, err := boot.blockBootstrapper.getBlockBody(prevHeader)
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

		err = boot.rollBackOneBlock(
			currHeaderHash,
			currHeader,
			currBlockBody,
			prevHeaderHash,
			prevHeader,
			prevBlockBody)

		if err != nil {
			return err
		}

		_, _ = updateMetricsFromStorage(boot.store, boot.uint64Converter, boot.marshalizer, boot.statusHandler, prevHeader.GetNonce())

		err = boot.bootStorer.SaveLastRound(int64(prevHeader.GetRound()))
		if err != nil {
			log.Info(fmt.Sprintf("cannot save last round in storage %s", err.Error()))
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

func (boot *baseBootstrap) rollBackOneBlock(
	currHeaderHash []byte,
	currHeader data.HeaderHandler,
	currBlockBody data.BodyHandler,
	prevHeaderHash []byte,
	prevHeader data.HeaderHandler,
	prevBlockBody data.BodyHandler,
) error {

	var err error

	defer func() {
		if err != nil {
			boot.restoreState(currHeaderHash, currHeader, currBlockBody)
		}
	}()

	if currHeader.GetNonce() > 1 {
		err = boot.setCurrentBlockInfo(prevHeaderHash, prevHeader, prevBlockBody)
		if err != nil {
			return err
		}
	} else {
		err = boot.setCurrentBlockInfo(nil, nil, nil)
		if err != nil {
			return err
		}
	}

	err = boot.blkExecutor.RevertStateToBlock(prevHeader)
	if err != nil {
		return err
	}

	err = boot.blkExecutor.RestoreBlockIntoPools(currHeader, currBlockBody)
	if err != nil {
		return err
	}

	boot.cleanCachesAndStorageOnRollback(currHeader)

	return nil
}

func (boot *baseBootstrap) getNextHeaderRequestingIfMissing() (data.HeaderHandler, error) {
	nonce := boot.getNonceForNextBlock()

	boot.setRequestedHeaderHash(nil)
	boot.setRequestedHeaderNonce(nil)

	hash := boot.forkDetector.GetNotarizedHeaderHash(nonce)
	if boot.forkInfo.IsDetected {
		hash = boot.forkInfo.Hash
	}

	if hash != nil {
		return boot.blockBootstrapper.getHeaderWithHashRequestingIfMissing(hash)
	}

	return boot.blockBootstrapper.getHeaderWithNonceRequestingIfMissing(nonce)
}

func (boot *baseBootstrap) addReceivedHeaderToForkDetector(hash []byte) error {
	header, err := boot.getHeaderFromPool(hash)
	if err != nil {
		return err
	}

	err = boot.forkDetector.AddHeader(header, hash, process.BHReceived, nil, nil, false)
	if err != nil {
		return err
	}

	return nil
}

func (boot *baseBootstrap) isForcedFork() bool {
	return boot.forkInfo.IsDetected &&
		boot.forkInfo.Nonce == math.MaxUint64 &&
		boot.forkInfo.Hash == nil
}

func (boot *baseBootstrap) rollBackOnForcedFork() {
	err := boot.rollBack(false)
	if err != nil {
		log.Debug("rollBack", "error", err.Error())
	}

	boot.forkDetector.ResetProbableHighestNonce()
	boot.forkDetector.ResetFork()
}

func (boot *baseBootstrap) restoreState(
	currHeaderHash []byte,
	currHeader data.HeaderHandler,
	currBlockBody data.BodyHandler,
) {
	log.Debug("revert state to header",
		"nonce", currHeader.GetNonce(),
		"hash", currHeaderHash)

	err := boot.blkc.SetCurrentBlockHeader(currHeader)
	if err != nil {
		log.Debug("SetCurrentBlockHeader", "error", err.Error())
	}

	err = boot.blkc.SetCurrentBlockBody(currBlockBody)
	if err != nil {
		log.Debug("SetCurrentBlockBody", "error", err.Error())
	}

	boot.blkc.SetCurrentBlockHeaderHash(currHeaderHash)

	err = boot.blkExecutor.RevertStateToBlock(currHeader)
	if err != nil {
		log.Debug("RevertStateToBlock", "error", err.Error())
	}
}

func (boot *baseBootstrap) setCurrentBlockInfo(
	headerHash []byte,
	header data.HeaderHandler,
	body data.BodyHandler,
) error {

	err := boot.blkc.SetCurrentBlockHeader(header)
	if err != nil {
		return err
	}

	err = boot.blkc.SetCurrentBlockBody(body)
	if err != nil {
		return err
	}

	boot.blkc.SetCurrentBlockHeaderHash(headerHash)

	return nil
}

// setRequestedMiniBlocks method sets the body hash requested by the sync mechanism
func (boot *baseBootstrap) setRequestedMiniBlocks(hashes [][]byte) {
	boot.requestedHashes.SetHashes(hashes)
}

// receivedBodyHash method is a call back function which is called when a new body is added
// in the block bodies pool
func (boot *baseBootstrap) receivedBodyHash(hash []byte) {
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
	err := boot.miniBlocksResolver.RequestDataFromHashArray(hashes)
	if err != nil {
		log.Debug("RequestDataFromHashArray", "error", err.Error())
		return
	}

	boot.requestedItemsHandler.Sweep()

	for _, hash := range hashes {
		err = boot.requestedItemsHandler.Add(string(hash))
		if err != nil {
			log.Trace("add requested item with error", "error", err.Error())
		}
	}

	log.Debug("requested mini blocks from network",
		"num miniblocks", len(hashes),
	)
}

// getMiniBlocksRequestingIfMissing method gets the body with given nonce from pool, if it exist there,
// and if not it will be requested from network
// the func returns interface{} as to match the next implementations for block body fetchers
// that will be added. The block executor should decide by parsing the header block body type value
// what kind of block body received.
func (boot *baseBootstrap) getMiniBlocksRequestingIfMissing(hashes [][]byte) (block.MiniBlockSlice, error) {
	miniBlocks, missingMiniBlocksHashes := boot.miniBlocksResolver.GetMiniBlocksFromPool(hashes)
	if len(missingMiniBlocksHashes) > 0 {
		_ = process.EmptyChannel(boot.chRcvMiniBlocks)
		boot.requestMiniBlocksByHashes(missingMiniBlocksHashes)
		err := boot.waitForMiniBlocks()
		if err != nil {
			return nil, err
		}

		receivedMiniBlocks, unreceivedMiniBlocksHashes := boot.miniBlocksResolver.GetMiniBlocksFromPool(missingMiniBlocksHashes)
		if len(unreceivedMiniBlocksHashes) > 0 {
			return nil, process.ErrMissingBody
		}

		miniBlocks = append(miniBlocks, receivedMiniBlocks...)
	}

	return miniBlocks, nil
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
