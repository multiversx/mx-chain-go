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
const sleepTime = time.Duration(5 * time.Millisecond)

// maxRoundsToWait defines the maximum rounds to wait, when bootstrapping, after which the node will add an empty
// block through recovery mechanism, if its block request is not resolved and no new block header is received meantime
const maxRoundsToWait = 5

type baseBootstrap struct {
	headers       storage.Cacher
	headersNonces dataRetriever.Uint64Cacher

	blkc        data.ChainHandler
	blkExecutor process.BlockProcessor
	store       dataRetriever.StorageService

	rounder          consensus.Rounder
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	forkDetector     process.ForkDetector
	shardCoordinator sharding.Coordinator
	accounts         state.AccountsAdapter

	mutHeader   sync.RWMutex
	headerNonce *uint64
	chRcvHdr    chan bool

	requestedHashes process.RequiredDataPool

	chStopSync chan bool
	waitTime   time.Duration

	isNodeSynchronized bool
	hasLastBlock       bool
	roundIndex         int32

	isForkDetected bool
	forkNonce      uint64

	mutRcvHdrInfo         sync.RWMutex
	syncStateListeners    []func(bool)
	mutSyncStateListeners sync.RWMutex
	uint64Converter       typeConverters.Uint64ByteSliceConverter
	bootstrapRoundIndex   uint32
}

func (boot *baseBootstrap) loadBlocks(
	blockFinality uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
	getHeader func(uint32, uint64) (data.HeaderHandler, []byte, error),
	getBlockBody func(data.HeaderHandler) (data.BodyHandler, error),
	removeBlockBody func(uint64, dataRetriever.UnitType, dataRetriever.UnitType) error,
	getNonceWithLastNotarized func(uint64, uint64) (uint64, map[uint32]uint64, map[uint32]uint64),
	applyNotarizedBlocks func(map[uint32]uint64, map[uint32]uint64) error,
) error {
	var err error
	var currentNonce uint64

	highestNonceInStorer := boot.computeHighestNonce(hdrNonceHashDataUnit)

	log.Info(fmt.Sprintf("the highest header nonce committed in storer is %d\n", highestNonceInStorer))

	var finalNotarized map[uint32]uint64
	var lastNotarized map[uint32]uint64

	for currentNonce = highestNonceInStorer; currentNonce > blockFinality; currentNonce-- {
		currentNonce, finalNotarized, lastNotarized = getNonceWithLastNotarized(currentNonce, blockFinality)
		if currentNonce <= blockFinality {
			return process.ErrNotEnoughValidBlocksInStorage
		}

		for i := currentNonce - blockFinality; i <= currentNonce; i++ {
			err = boot.applyBlock(boot.shardCoordinator.SelfId(), i, getHeader, getBlockBody)
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
				continue
			}

			break
		}
	}

	if currentNonce <= blockFinality {
		return process.ErrNotEnoughValidBlocksInStorage
	}

	for i := currentNonce + 1; i <= highestNonceInStorer; i++ {
		boot.cleanupStorage(removeBlockBody, i, blockUnit, hdrNonceHashDataUnit)
	}

	err = applyNotarizedBlocks(finalNotarized, lastNotarized)
	if err != nil {
		return err
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

func (boot *baseBootstrap) applyBlock(
	shardId uint32,
	nonce uint64,
	getHeader func(uint32, uint64) (data.HeaderHandler, []byte, error),
	getBlockBody func(data.HeaderHandler) (data.BodyHandler, error),
) error {
	header, headerHash, err := getHeader(shardId, nonce)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("apply block with nonce %d and round %d\n", header.GetNonce(), header.GetRound()))

	errNotCritical := boot.forkDetector.AddHeader(header, headerHash, process.BHProcessed)
	if errNotCritical != nil {
		log.Info(errNotCritical.Error())
	}

	blockBody, err := getBlockBody(header)
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
	removeBlockBody func(uint64, dataRetriever.UnitType, dataRetriever.UnitType) error,
	nonce uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType) {

	errNotCritical := removeBlockBody(nonce, blockUnit, hdrNonceHashDataUnit)
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

func (boot *baseBootstrap) getShardHeaderFromStorage(shardId uint32, nonce uint64) (data.HeaderHandler, []byte, error) {
	nonceToByteSlice := boot.uint64Converter.ToByteSlice(nonce)
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardId)
	headerHash, err := boot.store.Get(hdrNonceHashDataUnit, nonceToByteSlice)
	if err != nil {
		return nil, nil, err
	}

	header, err := process.GetShardHeaderFromStorage(headerHash, boot.marshalizer, boot.store)

	return header, headerHash, err
}

func (boot *baseBootstrap) getMetaHeaderFromStorage(shardId uint32, nonce uint64) (data.HeaderHandler, []byte, error) {
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

// requestedHeaderNonce method gets the header nonce requested by the sync mechanism
func (boot *baseBootstrap) requestedHeaderNonce() *uint64 {
	boot.mutHeader.RLock()
	defer boot.mutHeader.RUnlock()
	return boot.headerNonce
}

func (boot *baseBootstrap) processReceivedHeader(headerHandler data.HeaderHandler, headerHash []byte) {
	log.Debug(fmt.Sprintf("receivedHeaders: received header with nonce %d and hash %s from network\n",
		headerHandler.GetNonce(),
		core.ToB64(headerHash)))

	err := boot.forkDetector.AddHeader(headerHandler, headerHash, process.BHReceived)
	if err != nil {
		log.Info(err.Error())
	}
}

// receivedHeaderNonce method is a call back function which is called when a new header is added
// in the block headers pool
func (boot *baseBootstrap) receivedHeaderNonce(nonce uint64) {
	headerHash, _ := boot.headersNonces.Get(nonce)
	byteHeaderHash, ok := headerHash.([]byte)

	if ok {
		log.Debug(fmt.Sprintf("receivedHeaderNonce: received header with nonce %d and hash %s from network\n",
			nonce,
			core.ToB64(byteHeaderHash)))
	}

	n := boot.requestedHeaderNonce()
	if n == nil {
		return
	}

	if *n == nonce {
		log.Info(fmt.Sprintf("received requested header with nonce %d from network and probable highest nonce is %d\n",
			nonce,
			boot.forkDetector.ProbableHighestNonce()))
		boot.setRequestedHeaderNonce(nil)
		boot.chRcvHdr <- true
	}
}

// AddSyncStateListener adds a syncStateListener that get notified each time the sync status of the node changes
func (boot *baseBootstrap) AddSyncStateListener(syncStateListener func(bool)) {
	boot.mutSyncStateListeners.Lock()
	boot.syncStateListeners = append(boot.syncStateListeners, syncStateListener)
	boot.mutSyncStateListeners.Unlock()
}

func (boot *baseBootstrap) notifySyncStateListeners() {
	boot.mutSyncStateListeners.RLock()
	for i := 0; i < len(boot.syncStateListeners); i++ {
		go boot.syncStateListeners[i](boot.isNodeSynchronized)
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
	case <-boot.chRcvHdr:
		return nil
	case <-time.After(boot.waitTime):
		return process.ErrTimeIsOut
	}
}

// ShouldSync method returns the synch state of the node. If it returns 'true', this means that the node
// is not synchronized yet and it has to continue the bootstrapping mechanism, otherwise the node is already
// synched and it can participate to the consensus, if it is in the jobDone group of this rounder
func (boot *baseBootstrap) ShouldSync() bool {
	isNodeSynchronizedInCurrentRound := boot.roundIndex == boot.rounder.Index() && boot.isNodeSynchronized
	if isNodeSynchronizedInCurrentRound {
		return false
	}

	boot.isForkDetected, boot.forkNonce = boot.forkDetector.CheckFork()

	if boot.blkc.GetCurrentBlockHeader() == nil {
		boot.hasLastBlock = boot.forkDetector.ProbableHighestNonce() <= 0
	} else {
		boot.hasLastBlock = boot.forkDetector.ProbableHighestNonce() <= boot.blkc.GetCurrentBlockHeader().GetNonce()
	}

	isNodeSynchronized := !boot.isForkDetected && boot.hasLastBlock
	if isNodeSynchronized != boot.isNodeSynchronized {
		log.Info(fmt.Sprintf("node has changed its synchronized state to %v\n", isNodeSynchronized))
		boot.isNodeSynchronized = isNodeSynchronized
		boot.notifySyncStateListeners()
	}

	boot.roundIndex = boot.rounder.Index()

	return !isNodeSynchronized
}

func (boot *baseBootstrap) removeHeaderFromPools(header data.HeaderHandler) []byte {
	value, _ := boot.headersNonces.Get(header.GetNonce())
	boot.headersNonces.Remove(header.GetNonce())

	hash, ok := value.([]byte)
	if ok {
		boot.headers.Remove(hash)
	}

	return hash
}

func (boot *baseBootstrap) cleanCachesOnRollback(
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
	if blkc == nil {
		return process.ErrNilBlockChain
	}
	if rounder == nil {
		return process.ErrNilRounder
	}
	if blkExecutor == nil {
		return process.ErrNilBlockExecutor
	}
	if hasher == nil {
		return process.ErrNilHasher
	}
	if marshalizer == nil {
		return process.ErrNilMarshalizer
	}
	if forkDetector == nil {
		return process.ErrNilForkDetector
	}
	if resolversFinder == nil {
		return process.ErrNilResolverContainer
	}
	if shardCoordinator == nil {
		return process.ErrNilShardCoordinator
	}
	if accounts == nil {
		return process.ErrNilAccountsAdapter
	}
	if store == nil {
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
