package sync

import (
	"bytes"
	"math"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
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

	rounder             consensus.Rounder
	hasher              hashing.Hasher
	marshalizer         marshal.Marshalizer
	forkDetector        process.ForkDetector
	shardCoordinator    sharding.Coordinator
	accounts            state.AccountsAdapter
	storageBootstrapper storageBootstrapper
	blockBootstrapper   blockBootstrapper
	blackListHandler    process.BlackListHandler

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
	bootstrapRoundIndex   uint64
	requestsWithTimeout   uint32

	requestMiniBlocks func(uint32, uint64)

	networkWatcher    process.NetworkConnectionWatcher
	getHeaderFromPool func([]byte) (data.HeaderHandler, error)

	headerStore          storage.Storer
	headerNonceHashStore storage.Storer
	hdrRes               dataRetriever.HeaderResolver
	syncStarter          syncStarter
}

func (boot *baseBootstrap) loadBlocks(
	blockFinality uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	var err error
	var validNonce uint64

	highestNonceInStorer := boot.computeHighestNonce(hdrNonceHashDataUnit)

	log.Debug("the highest header nonce committed in storer",
		"nonce", highestNonceInStorer,
	)

	var finalNotarized map[uint32]*HdrInfo
	var lastNotarized map[uint32]*HdrInfo

	shardId := boot.shardCoordinator.SelfId()

	currentNonce := highestNonceInStorer
	for currentNonce > blockFinality {
		validNonce, finalNotarized, lastNotarized = boot.storageBootstrapper.getNonceWithLastNotarized(currentNonce)
		if validNonce <= blockFinality {
			break
		}

		if validNonce < currentNonce {
			currentNonce = validNonce
		}

		for i := validNonce - blockFinality; i <= validNonce; i++ {
			err = boot.applyBlock(shardId, i)
			if err != nil {
				log.Debug("apply block",
					"nonce", i,
					"error", err.Error(),
				)
				break
			}
		}

		if err == nil {
			err = boot.blkExecutor.RevertStateToBlock(boot.blkc.GetCurrentBlockHeader())
			if err != nil {
				log.Debug("recreate trie for block",
					"nonce", boot.blkc.GetCurrentBlockHeader().GetNonce(),
					"shard", boot.blkc.GetCurrentBlockHeader().GetShardID(),
					"error", err.Error(),
				)
				currentNonce--
				continue
			}

			break
		}

		currentNonce--
	}

	defer func() {
		if err != nil {
			lastNotarized = make(map[uint32]*HdrInfo, 0)
			finalNotarized = make(map[uint32]*HdrInfo, 0)
			validNonce = 0
		}

		for i := validNonce + 1; i <= highestNonceInStorer; i++ {
			boot.cleanupStorage(i, blockUnit, hdrNonceHashDataUnit)
		}

		boot.storageBootstrapper.cleanupNotarizedStorage(lastNotarized)
	}()

	if currentNonce <= blockFinality || validNonce <= blockFinality {
		err = process.ErrNotEnoughValidBlocksInStorage
		return err
	}

	err = boot.storageBootstrapper.applyNotarizedBlocks(finalNotarized, lastNotarized)
	if err != nil {
		return err
	}

	for i := validNonce - blockFinality; i <= validNonce; i++ {
		withFinalHeaders := i == validNonce-blockFinality
		boot.addHeaderToForkDetector(shardId, i, withFinalHeaders)
	}

	return nil
}

func (boot *baseBootstrap) computeHighestNonce(hdrNonceHashDataUnit dataRetriever.UnitType) uint64 {
	highestNonceInStorer := uint64(0)

	for {
		highestNonceInStorer++
		nonceToByteSlice := boot.uint64Converter.ToByteSlice(highestNonceInStorer)
		err := boot.store.Has(hdrNonceHashDataUnit, nonceToByteSlice)
		if err != nil {
			highestNonceInStorer--
			break
		}
	}

	return highestNonceInStorer
}

func (boot *baseBootstrap) applyBlock(shardId uint32, nonce uint64) error {
	header, headerHash, err := boot.storageBootstrapper.getHeader(shardId, nonce)
	if err != nil {
		return err
	}

	log.Debug("apply block",
		"nonce", header.GetNonce(),
		"round", header.GetRound(),
	)

	blockBody, err := boot.storageBootstrapper.getBlockBody(header)
	if err != nil {
		return err
	}

	err = boot.blkc.SetCurrentBlockBody(blockBody)
	if err != nil {
		return err
	}

	err = boot.blkc.SetCurrentBlockHeader(header)
	if err != nil {
		return err
	}

	boot.blkc.SetCurrentBlockHeaderHash(headerHash)

	return nil
}

func (boot *baseBootstrap) cleanupStorage(
	nonce uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
) {
	errNotCritical := boot.storageBootstrapper.removeBlockBody(nonce, blockUnit, hdrNonceHashDataUnit)
	if errNotCritical != nil {
		log.Debug("remove block body",
			"nonce", nonce,
			"error", errNotCritical.Error(),
		)
	}

	errNotCritical = boot.removeBlockHeader(nonce, blockUnit, hdrNonceHashDataUnit)
	if errNotCritical != nil {
		log.Debug("remove block header",
			"nonce", nonce,
			"error", errNotCritical.Error(),
		)
	}
}

func (boot *baseBootstrap) removeBlockHeader(
	nonce uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	return nil
}

func (boot *baseBootstrap) getShardHeaderFromStorage(
	shardId uint32,
	nonce uint64,
) (data.HeaderHandler, []byte, error) {

	nonceToByteSlice := boot.uint64Converter.ToByteSlice(nonce)
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardId)
	headerHash, err := boot.store.Get(hdrNonceHashDataUnit, nonceToByteSlice)
	if err != nil {
		return nil, nil, err
	}

	header, err := process.GetShardHeaderFromStorage(headerHash, boot.marshalizer, boot.store)

	return header, headerHash, err
}

func (boot *baseBootstrap) getMetaHeaderFromStorage(
	shardId uint32,
	nonce uint64,
) (data.HeaderHandler, []byte, error) {

	nonceToByteSlice := boot.uint64Converter.ToByteSlice(nonce)
	headerHash, err := boot.store.Get(dataRetriever.MetaHdrNonceHashDataUnit, nonceToByteSlice)
	if err != nil {
		return nil, nil, err
	}

	header, err := process.GetMetaHeaderFromStorage(headerHash, boot.marshalizer, boot.store)

	return header, headerHash, err
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
	log.Trace("received header from network",
		"nonce", nonce,
		"hash", hash,
	)

	err := boot.addReceivedHeaderToForkDetector(hash)
	if err != nil {
		log.Debug("addReceivedHeaderToForkDetector", "error", err.Error())
	}

	if boot.requestMiniBlocks != nil {
		go boot.requestMiniBlocks(shardId, nonce)
	}

	boot.mutRcvHdrNonce.Lock()
	n := boot.requestedHeaderNonce()
	if n == nil {
		boot.mutRcvHdrNonce.Unlock()
		return
	}

	if *n == nonce {
		log.Debug("received requested header from network",
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
	if !boot.networkWatcher.IsConnectedToTheNetwork() {
		return true
	}

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

	isNodeSynchronized := !boot.forkInfo.IsDetected && boot.hasLastBlock
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

	nbRequestedHdrs := 0
	maxNonce := core.MinUint64(nonce+process.MaxHeadersToRequestInAdvance-1, boot.forkDetector.ProbableHighestNonce())
	for currentNonce := nonce; currentNonce <= maxNonce; currentNonce++ {
		haveHeader := haveHeaderInPoolWithNonce(nonce)
		if !haveHeader {
			err := hdrRes.RequestDataFromNonce(currentNonce)
			if err != nil {
				log.Debug("RequestDataFromNonce", "error", err.Error())
				continue
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
	if err != nil {
		return err
	}
	elapsedTime := time.Now().Sub(startTime).Seconds()
	log.Debug("elapsed time to process block",
		"time [s]", elapsedTime,
	)

	startTime = time.Now()
	err = boot.blkExecutor.CommitBlock(boot.blkc, hdr, blockBody)
	if err != nil {
		return err
	}
	elapsedTime = time.Now().Sub(startTime).Seconds()
	log.Debug("elapsed time to commit block",
		"time [s]", elapsedTime,
	)
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
		currBlockBody, err := boot.blockBootstrapper.getBlockBody(currHeader)
		if err != nil {
			return err
		}
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
			prevHeader,
			prevBlockBody)

		if err != nil {
			return err
		}

		boot.statusHandler.Decrement(core.MetricCountConsensusAcceptedBlocks)

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
	prevHeader data.HeaderHandler,
	prevBlockBody data.BodyHandler,
) error {

	var err error

	defer func() {
		if err != nil {
			boot.restoreState(currHeaderHash, currHeader, currBlockBody)
		}
	}()

	var prevHeaderHash []byte

	if currHeader.GetNonce() > 1 {
		prevHeaderHash = currHeader.GetPrevHash()
	}

	err = boot.blkc.SetCurrentBlockHeader(prevHeader)
	if err != nil {
		return err
	}

	err = boot.blkc.SetCurrentBlockBody(prevBlockBody)
	if err != nil {
		return err
	}

	boot.blkc.SetCurrentBlockHeaderHash(prevHeaderHash)

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

func (boot *baseBootstrap) addHeaderToForkDetector(
	shardId uint32,
	nonce uint64,
	withFinalHeaders bool,
) {

	header, headerHash, errNotCritical := boot.storageBootstrapper.getHeader(shardId, nonce)
	if errNotCritical != nil {
		log.Debug("getHeader", "error", errNotCritical.Error())
		return
	}

	var finalHeaders []data.HeaderHandler
	var finalHeadersHashes [][]byte

	if withFinalHeaders {
		finalHeaders = append(finalHeaders, header)
		finalHeadersHashes = append(finalHeadersHashes, headerHash)
	}

	errNotCritical = boot.forkDetector.AddHeader(header, headerHash, process.BHProcessed, finalHeaders, finalHeadersHashes, false)
	if errNotCritical != nil {
		log.Debug(errNotCritical.Error())
	}

	return
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
