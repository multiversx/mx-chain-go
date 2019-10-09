package sync

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.DefaultLogger()

// sleepTime defines the time in milliseconds between each iteration made in syncBlocks method
const sleepTime = 5 * time.Millisecond

// maxRoundsToWait defines the maximum rounds to wait, when bootstrapping, after which the node will add an empty
// block through recovery mechanism, if its block request is not resolved and no new block header is received meantime
const maxRoundsToWait = 5

// maxHeadersToRequestInAdvance defines the maximum number of headers which will be requested in advance if they are missing
const maxHeadersToRequestInAdvance = 10

type notarizedInfo struct {
	lastNotarized           map[uint32]uint64
	finalNotarized          map[uint32]uint64
	blockWithLastNotarized  map[uint32]uint64
	blockWithFinalNotarized map[uint32]uint64
	startNonce              uint64
}

func (ni *notarizedInfo) reset() {
	ni.lastNotarized = make(map[uint32]uint64, 0)
	ni.finalNotarized = make(map[uint32]uint64, 0)
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

	isForkDetected bool
	forkNonce      uint64
	forkHash       []byte

	mutRcvHdrInfo         sync.RWMutex
	syncStateListeners    []func(bool)
	mutSyncStateListeners sync.RWMutex
	uint64Converter       typeConverters.Uint64ByteSliceConverter
	bootstrapRoundIndex   uint64
	requestsWithTimeout   uint32
}

func (boot *baseBootstrap) loadBlocks(
	blockFinality uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	var err error
	var validNonce uint64

	highestNonceInStorer := boot.computeHighestNonce(hdrNonceHashDataUnit)

	log.Info(fmt.Sprintf("the highest header nonce committed in storer is %d\n", highestNonceInStorer))

	var finalNotarized map[uint32]uint64
	var lastNotarized map[uint32]uint64

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
				log.Info(fmt.Sprintf("apply block with nonce %d: %s\n", i, err.Error()))
				break
			}
		}

		if err == nil {
			err = boot.accounts.RecreateTrie(boot.blkc.GetCurrentBlockHeader().GetRootHash())
			if err != nil {
				log.Info(fmt.Sprintf("recreate trie for block with nonce %d in shard %d: %s\n",
					boot.blkc.GetCurrentBlockHeader().GetNonce(),
					boot.blkc.GetCurrentBlockHeader().GetShardID(),
					err.Error()))
				currentNonce--
				continue
			}

			break
		}

		currentNonce--
	}

	defer func() {
		if err != nil {
			lastNotarized = make(map[uint32]uint64, 0)
			finalNotarized = make(map[uint32]uint64, 0)
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
		boot.storageBootstrapper.addHeaderToForkDetector(shardId, i, lastNotarized[sharding.MetachainShardId])
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

	log.Info(fmt.Sprintf("apply block with nonce %d and round %d\n", header.GetNonce(), header.GetRound()))

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
		log.Info(fmt.Sprintf("remove block body with nonce %d: %s\n", nonce, errNotCritical.Error()))
	}

	errNotCritical = boot.removeBlockHeader(nonce, blockUnit, hdrNonceHashDataUnit)
	if errNotCritical != nil {
		log.Info(fmt.Sprintf("remove block header with nonce %d: %s\n", nonce, errNotCritical.Error()))
	}
}

func (boot *baseBootstrap) removeBlockHeader(
	nonce uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	headerStore := boot.store.GetStorer(blockUnit)
	if headerStore == nil {
		return process.ErrNilHeadersStorage
	}

	headerNonceHashStore := boot.store.GetStorer(hdrNonceHashDataUnit)
	if headerNonceHashStore == nil {
		return process.ErrNilHeadersNonceHashStorage
	}

	nonceToByteSlice := boot.uint64Converter.ToByteSlice(nonce)
	headerHash, err := boot.store.Get(hdrNonceHashDataUnit, nonceToByteSlice)
	if err != nil {
		return err
	}

	err = headerStore.Remove(headerHash)
	if err != nil {
		return err
	}

	err = headerNonceHashStore.Remove(nonceToByteSlice)
	if err != nil {
		return err
	}

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
	log.Debug(fmt.Sprintf("received header with hash %s and nonce %d from network\n",
		core.ToB64(headerHash),
		headerHandler.GetNonce()))

	err := boot.forkDetector.AddHeader(headerHandler, headerHash, process.BHReceived, nil, nil)
	if err != nil {
		log.Debug(err.Error())
	}

	hash := boot.requestedHeaderHash()
	if hash == nil {
		return
	}

	if bytes.Equal(hash, headerHash) {
		log.Info(fmt.Sprintf("received requested header with hash %s and nonce %d from network\n",
			core.ToB64(hash),
			headerHandler.GetNonce()))
		boot.setRequestedHeaderHash(nil)
		boot.chRcvHdrHash <- true
	}
}

// receivedHeaderNonce method is a call back function which is called when a new header is added
// in the block headers pool
func (boot *baseBootstrap) receivedHeaderNonce(nonce uint64, shardId uint32, hash []byte) {
	log.Debug(fmt.Sprintf("received header with nonce %d and hash %s from network\n",
		nonce,
		core.ToB64(hash)))

	n := boot.requestedHeaderNonce()
	if n == nil {
		return
	}

	if *n == nonce {
		log.Info(fmt.Sprintf("received requested header with nonce %d and hash %s from network\n",
			nonce,
			core.ToB64(hash)))
		boot.setRequestedHeaderNonce(nil)
		boot.chRcvHdrNonce <- true
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

// ShouldSync method returns the synch state of the node. If it returns 'true', this means that the node
// is not synchronized yet and it has to continue the bootstrapping mechanism, otherwise the node is already
// synched and it can participate to the consensus, if it is in the jobDone group of this rounder
func (boot *baseBootstrap) ShouldSync() bool {
	boot.mutNodeSynched.Lock()
	defer boot.mutNodeSynched.Unlock()

	isNodeSynchronizedInCurrentRound := boot.roundIndex == boot.rounder.Index() && boot.isNodeSynchronized
	if isNodeSynchronizedInCurrentRound {
		return false
	}

	boot.isForkDetected, boot.forkNonce, boot.forkHash = boot.forkDetector.CheckFork()

	if boot.blkc.GetCurrentBlockHeader() == nil {
		boot.hasLastBlock = boot.forkDetector.ProbableHighestNonce() <= 0
	} else {
		boot.hasLastBlock = boot.forkDetector.ProbableHighestNonce() <= boot.blkc.GetCurrentBlockHeader().GetNonce()
	}

	isNodeSynchronized := !boot.isForkDetected && boot.hasLastBlock
	if isNodeSynchronized != boot.isNodeSynchronized {
		log.Info(fmt.Sprintf("node has changed its synchronized state to %v\n", isNodeSynchronized))
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
		log.Info(err.Error())
		return nil
	}

	//TODO: boot.headers.Remove(hash) should not be called, just to have a restore point if it is needed later
	return hash
}

func (boot *baseBootstrap) cleanCachesAndStorageOnRollback(
	header data.HeaderHandler,
	headerStore storage.Storer,
	headerNonceHashStore storage.Storer) {

	hash := boot.removeHeaderFromPools(header)
	boot.forkDetector.RemoveHeaders(header.GetNonce(), hash)
	_ = headerStore.Remove(hash)
	nonceToByteSlice := boot.uint64Converter.ToByteSlice(header.GetNonce())
	_ = headerNonceHashStore.Remove(nonceToByteSlice)
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
) error {
	if blkc == nil || blkc.IsInterfaceNil() {
		return process.ErrNilBlockChain
	}
	if rounder == nil || rounder.IsInterfaceNil() {
		return process.ErrNilRounder
	}
	if blkExecutor == nil || blkExecutor.IsInterfaceNil() {
		return process.ErrNilBlockExecutor
	}
	if hasher == nil || hasher.IsInterfaceNil() {
		return process.ErrNilHasher
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return process.ErrNilMarshalizer
	}
	if forkDetector == nil || forkDetector.IsInterfaceNil() {
		return process.ErrNilForkDetector
	}
	if resolversFinder == nil || resolversFinder.IsInterfaceNil() {
		return process.ErrNilResolverContainer
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return process.ErrNilShardCoordinator
	}
	if accounts == nil || accounts.IsInterfaceNil() {
		return process.ErrNilAccountsAdapter
	}
	if store == nil || store.IsInterfaceNil() {
		return process.ErrNilStore
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

func (boot *baseBootstrap) requestHeadersFromNonceIfMissing(nonce uint64, hdrRes dataRetriever.HeaderResolver) {
	var err error
	nbRequestedHdrs := 0
	maxNonce := core.MinUint64(nonce+maxHeadersToRequestInAdvance-1, boot.forkDetector.ProbableHighestNonce())
	for currentNonce := nonce; currentNonce <= maxNonce; currentNonce++ {
		if boot.shardCoordinator.SelfId() == sharding.MetachainShardId {
			_, _, err = process.GetMetaHeaderFromPoolWithNonce(
				currentNonce,
				boot.headers,
				boot.headersNonces)
		} else {
			_, _, err = process.GetShardHeaderFromPoolWithNonce(
				currentNonce,
				boot.shardCoordinator.SelfId(),
				boot.headers,
				boot.headersNonces)
		}

		if err != nil {
			err = hdrRes.RequestDataFromNonce(currentNonce)
			if err != nil {
				log.Error(err.Error())
				continue
			}

			nbRequestedHdrs++
		}
	}

	if nbRequestedHdrs > 0 {
		log.Info(fmt.Sprintf("requested in advance %d headers from nonce %d to nonce %d and probable highest nonce is %d\n",
			nbRequestedHdrs,
			nonce,
			maxNonce,
			boot.forkDetector.ProbableHighestNonce()))
	}
}
