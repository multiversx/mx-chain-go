package sync

import (
	"math"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ShardBootstrap implements the bootstrap mechanism
type ShardBootstrap struct {
	*baseBootstrap

	miniBlocks storage.Cacher

	chRcvMiniBlocks  chan bool
	mutRcvMiniBlocks sync.Mutex

	resolversFinder    dataRetriever.ResolversFinder
	miniBlocksResolver dataRetriever.MiniBlocksResolver
}

// NewShardBootstrap creates a new Bootstrap object
func NewShardBootstrap(
	poolsHolder dataRetriever.PoolsHolder,
	store dataRetriever.StorageService,
	blkc data.ChainHandler,
	rounder consensus.Rounder,
	blkExecutor process.BlockProcessor,
	waitTime time.Duration,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	forkDetector process.ForkDetector,
	resolversFinder dataRetriever.ResolversFinder,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
	bootstrapRoundIndex uint64,
	blackListHandler process.BlackListHandler,
	networkWatcher process.NetworkConnectionWatcher,
) (*ShardBootstrap, error) {

	if check.IfNil(poolsHolder) {
		return nil, process.ErrNilPoolsHolder
	}
	if check.IfNil(poolsHolder.Headers()) {
		return nil, process.ErrNilHeadersDataPool
	}
	if check.IfNil(poolsHolder.HeadersNonces()) {
		return nil, process.ErrNilHeadersNoncesDataPool
	}
	if check.IfNil(poolsHolder.MiniBlocks()) {
		return nil, process.ErrNilTxBlockBody
	}

	err := checkBootstrapNilParameters(
		blkc,
		rounder,
		blkExecutor,
		hasher,
		marshalizer,
		forkDetector,
		resolversFinder,
		shardCoordinator,
		accounts,
		store,
		blackListHandler,
		networkWatcher,
	)
	if err != nil {
		return nil, err
	}

	base := &baseBootstrap{
		blkc:                blkc,
		blkExecutor:         blkExecutor,
		store:               store,
		headers:             poolsHolder.Headers(),
		headersNonces:       poolsHolder.HeadersNonces(),
		rounder:             rounder,
		waitTime:            waitTime,
		hasher:              hasher,
		marshalizer:         marshalizer,
		forkDetector:        forkDetector,
		shardCoordinator:    shardCoordinator,
		accounts:            accounts,
		bootstrapRoundIndex: bootstrapRoundIndex,
		blackListHandler:    blackListHandler,
		networkWatcher:      networkWatcher,
	}

	boot := ShardBootstrap{
		baseBootstrap: base,
		miniBlocks:    poolsHolder.MiniBlocks(),
	}

	base.storageBootstrapper = &boot
	base.blockBootstrapper = &boot
	base.getHeaderFromPool = boot.getShardHeaderFromPool
	base.syncStarter = &boot
	base.requestMiniBlocks = boot.requestMiniBlocksFromHeaderWithNonceIfMissing

	//there is one header topic so it is ok to save it
	hdrResolver, err := resolversFinder.IntraShardResolver(factory.HeadersTopic)
	if err != nil {
		return nil, err
	}

	//sync should request the missing block body on the intrashard topic
	miniBlocksResolver, err := resolversFinder.IntraShardResolver(factory.MiniBlocksTopic)
	if err != nil {
		return nil, err
	}

	//placed in struct fields for performance reasons
	base.headerStore = boot.store.GetStorer(dataRetriever.BlockHeaderUnit)
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(boot.shardCoordinator.SelfId())
	base.headerNonceHashStore = boot.store.GetStorer(hdrNonceHashDataUnit)

	hdrRes, ok := hdrResolver.(dataRetriever.HeaderResolver)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	base.hdrRes = hdrRes
	base.forkInfo = process.NewForkInfo()

	miniBlocksRes, ok := miniBlocksResolver.(dataRetriever.MiniBlocksResolver)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	boot.miniBlocksResolver = miniBlocksRes

	boot.chRcvHdrNonce = make(chan bool)
	boot.chRcvHdrHash = make(chan bool)
	boot.chRcvMiniBlocks = make(chan bool)

	boot.setRequestedHeaderNonce(nil)
	boot.setRequestedHeaderHash(nil)
	boot.setRequestedMiniBlocks(nil)

	boot.headersNonces.RegisterHandler(boot.receivedHeaderNonce)
	boot.miniBlocks.RegisterHandler(boot.receivedBodyHash)
	boot.headers.RegisterHandler(boot.receivedHeaders)

	boot.chStopSync = make(chan bool)

	boot.statusHandler = statusHandler.NewNilStatusHandler()

	boot.syncStateListeners = make([]func(bool), 0)
	boot.requestedHashes = process.RequiredDataPool{}

	//TODO: This should be injected when BlockProcessor will be refactored
	boot.uint64Converter = uint64ByteSlice.NewBigEndianConverter()

	return &boot, nil
}

func (boot *ShardBootstrap) syncFromStorer(
	blockFinality uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
	notarizedBlockFinality uint64,
) error {
	err := boot.loadBlocks(
		blockFinality,
		blockUnit,
		hdrNonceHashDataUnit)
	if err != nil {
		return err
	}

	return nil
}

func (boot *ShardBootstrap) getHeader(shardId uint32, nonce uint64) (data.HeaderHandler, []byte, error) {
	return boot.getShardHeaderFromStorage(shardId, nonce)
}

func (boot *ShardBootstrap) getBlockBody(headerHandler data.HeaderHandler) (data.BodyHandler, error) {
	header, ok := headerHandler.(*block.Header)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	hashes := make([][]byte, len(header.MiniBlockHeaders))
	for i := 0; i < len(header.MiniBlockHeaders); i++ {
		hashes[i] = header.MiniBlockHeaders[i].Hash
	}

	miniBlocks, missingMiniBlocksHashes := boot.miniBlocksResolver.GetMiniBlocks(hashes)
	if len(missingMiniBlocksHashes) > 0 {
		return nil, process.ErrMissingBody
	}

	return block.Body(miniBlocks), nil
}

func (boot *ShardBootstrap) removeBlockBody(
	nonce uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	return nil
}

func (boot *ShardBootstrap) getNonceWithLastNotarized(nonce uint64) (uint64, map[uint32]*HdrInfo, map[uint32]*HdrInfo) {
	ni := notarizedInfo{}
	ni.reset()
	shardId := sharding.MetachainShardId
	for currentNonce := nonce; currentNonce > 0; currentNonce-- {
		header, ok := boot.isHeaderValid(currentNonce)
		if !ok {
			ni.reset()
			continue
		}

		if len(header.MetaBlockHashes) == 0 {
			continue
		}

		minNonce, err := boot.getMinNotarizedMetaBlockNonceInHeader(header, &ni)
		if err != nil {
			log.Debug("getMinNotarizedMetaBlockNonceInHeader", "error", err.Error())
			ni.reset()
			continue
		}

		if boot.areNotarizedMetaBlocksFound(&ni, minNonce) {
			break
		}
	}

	ni.startNonce = ni.blockWithLastNotarized[shardId]

	if ni.lastNotarized[shardId] != nil && ni.finalNotarized[shardId] != nil {
		log.Debug("bootstrap from shard block",
			"nonce", ni.startNonce,
			"last notarized nonce", ni.lastNotarized[shardId].Nonce,
			"final meta nonce", ni.finalNotarized[shardId].Nonce,
		)
	}

	return ni.startNonce, ni.finalNotarized, ni.lastNotarized
}

func (boot *ShardBootstrap) isHeaderValid(nonce uint64) (*block.Header, bool) {
	headerHandler, _, err := boot.getHeader(boot.shardCoordinator.SelfId(), nonce)
	if err != nil {
		log.Debug("getHeader", "error", err.Error())
		return nil, false
	}

	header, ok := headerHandler.(*block.Header)
	if !ok {
		log.Debug("headerHandler is not type of header", "error", process.ErrWrongTypeAssertion.Error())
		return nil, false
	}

	if header.Round > boot.bootstrapRoundIndex {
		log.Debug("higher round in block",
			"round", header.Round,
			"bootstrapRoundIndex", boot.bootstrapRoundIndex,
			"error", ErrHigherRoundInBlock.Error())
		return nil, false
	}

	return header, true
}

func (boot *ShardBootstrap) getMinNotarizedMetaBlockNonceInHeader(
	header *block.Header,
	ni *notarizedInfo,
) (uint64, error) {

	minNonce := uint64(math.MaxUint64)
	shardId := sharding.MetachainShardId
	for _, metaBlockHash := range header.MetaBlockHashes {
		metaBlock, err := process.GetMetaHeaderFromStorage(metaBlockHash, boot.marshalizer, boot.store)
		if err != nil {
			return minNonce, err
		}

		foundShardHeaderWithLastNotarizedMeta := ni.blockWithLastNotarized[shardId] == 0 &&
			ni.lastNotarized[shardId] != nil && ni.lastNotarized[shardId].Nonce == metaBlock.Nonce

		if foundShardHeaderWithLastNotarizedMeta {
			ni.lastNotarized[shardId].Hash = metaBlockHash
			ni.blockWithLastNotarized[shardId] = header.Nonce
		}

		foundShardHeaderWithFinalNotarizedMeta := ni.blockWithFinalNotarized[shardId] == 0 &&
			ni.finalNotarized[shardId] != nil && ni.finalNotarized[shardId].Nonce == metaBlock.Nonce

		if foundShardHeaderWithFinalNotarizedMeta {
			ni.finalNotarized[shardId].Hash = metaBlockHash
			ni.blockWithFinalNotarized[shardId] = header.Nonce
		}

		if metaBlock.Nonce < minNonce {
			minNonce = metaBlock.Nonce
		}
	}

	return minNonce, nil
}

func (boot *ShardBootstrap) areNotarizedMetaBlocksFound(ni *notarizedInfo, notarizedNonce uint64) bool {
	shardId := sharding.MetachainShardId
	if notarizedNonce == 0 {
		return false
	}

	if ni.lastNotarized[shardId] == nil {
		ni.lastNotarized[shardId] = &HdrInfo{Nonce: notarizedNonce - 1}
		return false
	}

	if ni.blockWithLastNotarized[shardId] == 0 {
		return false
	}

	if ni.finalNotarized[shardId] == nil {
		ni.finalNotarized[shardId] = &HdrInfo{Nonce: notarizedNonce - 1}
		return false
	}

	if ni.blockWithFinalNotarized[shardId] == 0 {
		return false
	}

	return true
}

func (boot *ShardBootstrap) applyNotarizedBlocks(
	finalNotarized map[uint32]*HdrInfo,
	lastNotarized map[uint32]*HdrInfo,
) error {
	if finalNotarized[sharding.MetachainShardId] == nil || lastNotarized[sharding.MetachainShardId] == nil {
		return ErrNilNotarizedHeader
	}
	if finalNotarized[sharding.MetachainShardId].Hash == nil || lastNotarized[sharding.MetachainShardId].Hash == nil {
		return ErrNilHash
	}

	metaBlock, err := process.GetMetaHeaderFromStorage(finalNotarized[sharding.MetachainShardId].Hash, boot.marshalizer, boot.store)
	if err != nil {
		return err
	}

	boot.blkExecutor.AddLastNotarizedHdr(sharding.MetachainShardId, metaBlock)

	metaBlock, err = process.GetMetaHeaderFromStorage(lastNotarized[sharding.MetachainShardId].Hash, boot.marshalizer, boot.store)
	if err != nil {
		return err
	}

	boot.blkExecutor.AddLastNotarizedHdr(sharding.MetachainShardId, metaBlock)

	return nil
}

func (boot *ShardBootstrap) cleanupNotarizedStorage(lastNotarized map[uint32]*HdrInfo) {
	highestNonceInStorer := boot.computeHighestNonce(dataRetriever.MetaHdrNonceHashDataUnit)

	if lastNotarized[sharding.MetachainShardId] == nil {
		return
	}

	for nonce := lastNotarized[sharding.MetachainShardId].Nonce + 1; nonce <= highestNonceInStorer; nonce++ {
		errNotCritical := boot.removeBlockHeader(nonce, dataRetriever.MetaBlockUnit, dataRetriever.MetaHdrNonceHashDataUnit)
		if errNotCritical != nil {
			log.Debug("remove notarized block header",
				"nonce", nonce,
				"error", errNotCritical.Error(),
			)
		}
	}
}

func (boot *ShardBootstrap) receivedHeaders(headerHash []byte) {
	header, err := process.GetShardHeaderFromPool(headerHash, boot.headers)
	if err != nil {
		log.Trace("GetShardHeaderFromPool", "error", err.Error())
		return
	}

	boot.processReceivedHeader(header, headerHash)
}

// setRequestedMiniBlocks method sets the body hash requested by the sync mechanism
func (boot *ShardBootstrap) setRequestedMiniBlocks(hashes [][]byte) {
	boot.requestedHashes.SetHashes(hashes)
}

// receivedBody method is a call back function which is called when a new body is added
// in the block bodies pool
func (boot *ShardBootstrap) receivedBodyHash(hash []byte) {
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

// StartSync method will start SyncBlocks as a go routine
func (boot *ShardBootstrap) StartSync() {
	// when a node starts it first tries to bootstrap from storage, if there already exist a database saved
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(boot.shardCoordinator.SelfId())
	errNotCritical := boot.syncFromStorer(process.ShardBlockFinality,
		dataRetriever.BlockHeaderUnit,
		hdrNonceHashDataUnit,
		process.MetaBlockFinality)
	if errNotCritical != nil {
		log.Debug("boot.syncFromStorer",
			"error", errNotCritical.Error(),
		)
	}

	go boot.syncBlocks()
}

// SyncBlock method actually does the synchronization. It requests the next block header from the pool
// and if it is not found there it will be requested from the network. After the header is received,
// it requests the block body in the same way(pool and than, if it is not found in the pool, from network).
// If either header and body are received the ProcessBlock and CommitBlock method will be called successively.
// These methods will execute the block and its transactions. Finally if everything works, the block will be committed
// in the blockchain, and all this mechanism will be reiterated for the next block.
func (boot *ShardBootstrap) SyncBlock() error {
	return boot.syncBlock()
}

// requestHeaderWithNonce method requests a block header from network when it is not found in the pool
func (boot *ShardBootstrap) requestHeaderWithNonce(nonce uint64) {
	boot.setRequestedHeaderNonce(&nonce)
	err := boot.hdrRes.RequestDataFromNonce(nonce)

	log.Debug("requested header from network",
		"nonce", nonce,
		"probable highest nonce", boot.forkDetector.ProbableHighestNonce(),
	)

	if err != nil {
		log.Debug("RequestDataFromNonce", "error", err.Error())
	}

	log.Debug("requested header from network",
		"nonce", nonce,
	)
	log.Debug("probable highest nonce",
		"nonce", boot.forkDetector.ProbableHighestNonce(),
	)
}

// requestHeaderWithHash method requests a block header from network when it is not found in the pool
func (boot *ShardBootstrap) requestHeaderWithHash(hash []byte) {
	boot.setRequestedHeaderHash(hash)
	err := boot.hdrRes.RequestDataFromHash(hash)
	if err != nil {
		log.Debug("RequestDataFromHash", "error", err.Error())
	}

	log.Debug("requested header from network",
		"hash", hash,
	)
}

// getHeaderWithNonceRequestingIfMissing method gets the header with a given nonce from pool. If it is not found there, it will
// be requested from network
func (boot *ShardBootstrap) getHeaderWithNonceRequestingIfMissing(nonce uint64) (data.HeaderHandler, error) {
	hdr, _, err := process.GetShardHeaderFromPoolWithNonce(
		nonce,
		boot.shardCoordinator.SelfId(),
		boot.headers,
		boot.headersNonces)
	if err != nil {
		_ = process.EmptyChannel(boot.chRcvHdrNonce)
		boot.requestHeaderWithNonce(nonce)
		err := boot.waitForHeaderNonce()
		if err != nil {
			return nil, err
		}

		hdr, _, err = process.GetShardHeaderFromPoolWithNonce(
			nonce,
			boot.shardCoordinator.SelfId(),
			boot.headers,
			boot.headersNonces)
		if err != nil {
			return nil, err
		}
	}

	return hdr, nil
}

// getHeaderWithHashRequestingIfMissing method gets the header with a given hash from pool. If it is not found there,
// it will be requested from network
func (boot *ShardBootstrap) getHeaderWithHashRequestingIfMissing(hash []byte) (data.HeaderHandler, error) {
	hdr, err := process.GetShardHeader(hash, boot.headers, boot.marshalizer, boot.store)
	if err != nil {
		_ = process.EmptyChannel(boot.chRcvHdrHash)
		boot.requestHeaderWithHash(hash)
		err := boot.waitForHeaderHash()
		if err != nil {
			return nil, err
		}

		hdr, err = process.GetShardHeaderFromPool(hash, boot.headers)
		if err != nil {
			return nil, err
		}
	}

	return hdr, nil
}

// requestMiniBlocks method requests a block body from network when it is not found in the pool
func (boot *ShardBootstrap) requestMiniBlocks(hashes [][]byte) {
	boot.setRequestedMiniBlocks(hashes)
	err := boot.miniBlocksResolver.RequestDataFromHashArray(hashes)
	if err != nil {
		log.Debug("RequestDataFromHashArray", "error", err.Error())
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
func (boot *ShardBootstrap) getMiniBlocksRequestingIfMissing(hashes [][]byte) (block.MiniBlockSlice, error) {
	miniBlocks, missingMiniBlocksHashes := boot.miniBlocksResolver.GetMiniBlocksFromPool(hashes)
	if len(missingMiniBlocksHashes) > 0 {
		_ = process.EmptyChannel(boot.chRcvMiniBlocks)
		boot.requestMiniBlocks(missingMiniBlocksHashes)
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
func (boot *ShardBootstrap) waitForMiniBlocks() error {
	select {
	case <-boot.chRcvMiniBlocks:
		return nil
	case <-time.After(boot.waitTime):
		return process.ErrTimeIsOut
	}
}

func (boot *ShardBootstrap) getPrevHeader(
	header data.HeaderHandler,
	headerStore storage.Storer,
) (data.HeaderHandler, error) {

	prevHash := header.GetPrevHash()
	buffHeader, err := headerStore.Get(prevHash)
	if err != nil {
		return nil, err
	}

	prevHeader := &block.Header{}
	err = boot.marshalizer.Unmarshal(prevHeader, buffHeader)
	if err != nil {
		return nil, err
	}

	return prevHeader, nil
}

func (boot *ShardBootstrap) getCurrHeader() (data.HeaderHandler, error) {
	blockHeader := boot.blkc.GetCurrentBlockHeader()
	if blockHeader == nil {
		return nil, process.ErrNilBlockHeader
	}

	header, ok := blockHeader.(*block.Header)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return header, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (boot *ShardBootstrap) IsInterfaceNil() bool {
	if boot == nil {
		return true
	}
	return false
}

func (boot *ShardBootstrap) haveHeaderInPoolWithNonce(nonce uint64) bool {
	_, _, err := process.GetShardHeaderFromPoolWithNonce(
		nonce,
		boot.shardCoordinator.SelfId(),
		boot.headers,
		boot.headersNonces)

	return err == nil
}

func (boot *ShardBootstrap) getShardHeaderFromPool(headerHash []byte) (data.HeaderHandler, error) {
	return process.GetShardHeaderFromPool(headerHash, boot.headers)
}

func (boot *ShardBootstrap) requestMiniBlocksFromHeaderWithNonceIfMissing(shardId uint32, nonce uint64) {
	nextBlockNonce := boot.getNonceForNextBlock()
	maxNonce := core.MinUint64(nextBlockNonce+process.MaxHeadersToRequestInAdvance-1, boot.forkDetector.ProbableHighestNonce())
	if nonce < nextBlockNonce || nonce > maxNonce {
		return
	}

	header, _, err := process.GetShardHeaderFromPoolWithNonce(
		nonce,
		shardId,
		boot.headers,
		boot.headersNonces)

	if err != nil {
		log.Trace("GetShardHeaderFromPoolWithNonce", "error", err.Error())
		return
	}

	hashes := make([][]byte, len(header.MiniBlockHeaders))
	for i := 0; i < len(header.MiniBlockHeaders); i++ {
		hashes[i] = header.MiniBlockHeaders[i].Hash
	}

	_, missingMiniBlocksHashes := boot.miniBlocksResolver.GetMiniBlocksFromPool(hashes)
	if len(missingMiniBlocksHashes) > 0 {
		err := boot.miniBlocksResolver.RequestDataFromHashArray(missingMiniBlocksHashes)
		if err != nil {
			log.Debug("RequestDataFromHashArray", "error", err.Error())
			return
		}

		log.Trace("requested in advance mini blocks",
			"num miniblocks", len(missingMiniBlocksHashes),
			"header nonce", header.Nonce,
		)
	}
}

func (boot *ShardBootstrap) getBlockBodyRequestingIfMissing(headerHandler data.HeaderHandler) (data.BodyHandler, error) {
	header, ok := headerHandler.(*block.Header)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	hashes := make([][]byte, len(header.MiniBlockHeaders))
	for i := 0; i < len(header.MiniBlockHeaders); i++ {
		hashes[i] = header.MiniBlockHeaders[i].Hash
	}

	boot.setRequestedMiniBlocks(nil)

	miniBlockSlice, err := boot.getMiniBlocksRequestingIfMissing(hashes)
	if err != nil {
		return nil, err
	}

	blockBody := block.Body(miniBlockSlice)

	return blockBody, nil
}

func (boot *ShardBootstrap) isForkTriggeredByMeta() bool {
	return boot.forkInfo.IsDetected &&
		boot.forkInfo.Nonce != math.MaxUint64 &&
		boot.forkInfo.Round == process.MinForkRound &&
		boot.forkInfo.Hash != nil
}
