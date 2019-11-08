package sync

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// MetaBootstrap implements the bootstrap mechanism
type MetaBootstrap struct {
	*baseBootstrap

	resolversFinder dataRetriever.ResolversFinder
}

// NewMetaBootstrap creates a new Bootstrap object
func NewMetaBootstrap(
	poolsHolder dataRetriever.MetaPoolsHolder,
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
	networkWatcher process.NetworkConnectionWatcher,
) (*MetaBootstrap, error) {

	if poolsHolder == nil || poolsHolder.IsInterfaceNil() {
		return nil, process.ErrNilPoolsHolder
	}
	if poolsHolder.HeadersNonces() == nil || poolsHolder.HeadersNonces().IsInterfaceNil() {
		return nil, process.ErrNilHeadersNoncesDataPool
	}
	if poolsHolder.MetaBlocks() == nil || poolsHolder.MetaBlocks().IsInterfaceNil() {
		return nil, process.ErrNilMetaBlockPool
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
		networkWatcher,
	)
	if err != nil {
		return nil, err
	}

	base := &baseBootstrap{
		blkc:                blkc,
		blkExecutor:         blkExecutor,
		store:               store,
		headers:             poolsHolder.MetaBlocks(),
		headersNonces:       poolsHolder.HeadersNonces(),
		rounder:             rounder,
		waitTime:            waitTime,
		hasher:              hasher,
		marshalizer:         marshalizer,
		forkDetector:        forkDetector,
		shardCoordinator:    shardCoordinator,
		accounts:            accounts,
		bootstrapRoundIndex: bootstrapRoundIndex,
		networkWatcher:      networkWatcher,
	}

	boot := MetaBootstrap{
		baseBootstrap: base,
	}

	base.storageBootstrapper = &boot
	base.blockBootstrapper = &boot
	base.getHeaderFromPool = boot.getMetaHeaderFromPool
	base.syncStarter = &boot

	//there is one header topic so it is ok to save it
	hdrResolver, err := resolversFinder.MetaChainResolver(factory.MetachainBlocksTopic)
	if err != nil {
		return nil, err
	}

	//placed in struct fields for performance reasons
	base.headerStore = boot.store.GetStorer(dataRetriever.MetaBlockUnit)
	base.headerNonceHashStore = boot.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)

	hdrRes, ok := hdrResolver.(dataRetriever.HeaderResolver)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	base.hdrRes = hdrRes

	boot.chRcvHdrNonce = make(chan bool)
	boot.chRcvHdrHash = make(chan bool)

	boot.setRequestedHeaderNonce(nil)
	boot.setRequestedHeaderHash(nil)
	boot.headersNonces.RegisterHandler(boot.receivedHeaderNonce)
	boot.headers.RegisterHandler(boot.receivedHeader)

	boot.chStopSync = make(chan bool)

	boot.statusHandler = statusHandler.NewNilStatusHandler()

	boot.syncStateListeners = make([]func(bool), 0)
	boot.requestedHashes = process.RequiredDataPool{}

	//TODO: This should be injected when BlockProcessor will be refactored
	boot.uint64Converter = uint64ByteSlice.NewBigEndianConverter()

	return &boot, nil
}

func (boot *MetaBootstrap) syncFromStorer(
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

func (boot *MetaBootstrap) addHeaderToForkDetector(shardId uint32, nonce uint64, lastNotarizedMeta uint64) {
	header, headerHash, errNotCritical := boot.storageBootstrapper.getHeader(shardId, nonce)
	if errNotCritical != nil {
		log.Debug("storageBootstrapper.getHeader", "error", errNotCritical.Error())
		return
	}

	if shardId == sharding.MetachainShardId {
		errNotCritical = boot.forkDetector.AddHeader(header, headerHash, process.BHProcessed, nil, nil)
		if errNotCritical != nil {
			log.Debug("forkDetector.AddHeader", "error", errNotCritical.Error())
		}

		return
	}
}

func (boot *MetaBootstrap) getHeader(shardId uint32, nonce uint64) (data.HeaderHandler, []byte, error) {
	return boot.getMetaHeaderFromStorage(shardId, nonce)
}

func (boot *MetaBootstrap) getBlockBody(headerHandler data.HeaderHandler) (data.BodyHandler, error) {
	return &block.MetaBlockBody{}, nil
}

func (boot *MetaBootstrap) removeBlockBody(
	nonce uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	return nil
}

func (boot *MetaBootstrap) getNonceWithLastNotarized(nonce uint64) (uint64, map[uint32]uint64, map[uint32]uint64) {
	ni := notarizedInfo{}
	ni.reset()
	for currentNonce := nonce; currentNonce > 0; currentNonce-- {
		metaBlock, ok := boot.isMetaBlockValid(currentNonce)
		if !ok {
			ni.reset()
			continue
		}

		if ni.startNonce == 0 {
			ni.startNonce = currentNonce
		}

		if len(metaBlock.ShardInfo) == 0 {
			continue
		}

		maxNonce, err := boot.getMaxNotarizedHeadersNoncesInMetaBlock(metaBlock, &ni)
		if err != nil {
			log.Debug("getMaxNotarizedHeadersNoncesInMetaBlock", "error", err.Error())
			ni.reset()
			continue
		}

		if boot.areNotarizedShardHeadersFound(&ni, maxNonce, currentNonce) {
			break
		}
	}

	log.Debug("bootstrap from meta block",
		"nonce", ni.startNonce,
	)

	for i := uint32(0); i < boot.shardCoordinator.NumberOfShards(); i++ {
		if nonce > ni.blockWithLastNotarized[i] {
			ni.finalNotarized[i] = ni.lastNotarized[i]
		}

		log.Debug("last notarized block",
			"shard", i,
			"nonce", ni.lastNotarized[i],
		)
		log.Debug("final notarized block",
			"nonce", ni.finalNotarized[i],
		)
	}

	return ni.startNonce, ni.finalNotarized, ni.lastNotarized
}

func (boot *MetaBootstrap) isMetaBlockValid(nonce uint64) (*block.MetaBlock, bool) {
	headerHandler, _, err := boot.getHeader(sharding.MetachainShardId, nonce)
	if err != nil {
		log.Debug("getHeader", "error", err.Error())
		return nil, false
	}

	metaBlock, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		log.Debug("headerHandler is not of type metaBlock", "error", process.ErrWrongTypeAssertion.Error())
		return nil, false
	}

	if metaBlock.Round > boot.bootstrapRoundIndex {
		log.Debug("higher round in metablock",
			"round", metaBlock.Round,
			"bootstrapRoundIndex", boot.bootstrapRoundIndex,
			"error", ErrHigherRoundInBlock.Error())
		return nil, false
	}

	return metaBlock, true
}

func (boot *MetaBootstrap) getMaxNotarizedHeadersNoncesInMetaBlock(
	metaBlock *block.MetaBlock,
	ni *notarizedInfo,
) (map[uint32]uint64, error) {

	maxNonce := make(map[uint32]uint64, 0)
	for _, shardData := range metaBlock.ShardInfo {
		header, err := process.GetShardHeaderFromStorage(shardData.HeaderHash, boot.marshalizer, boot.store)
		if err != nil {
			return maxNonce, err
		}

		if header.Nonce > maxNonce[shardData.ShardId] {
			maxNonce[shardData.ShardId] = header.Nonce
		}
	}

	return maxNonce, nil
}

func (boot *MetaBootstrap) areNotarizedShardHeadersFound(
	ni *notarizedInfo,
	notarizedNonce map[uint32]uint64,
	nonce uint64,
) bool {

	for i := uint32(0); i < boot.shardCoordinator.NumberOfShards(); i++ {
		if ni.lastNotarized[i] == 0 {
			ni.lastNotarized[i] = notarizedNonce[i]
			ni.blockWithLastNotarized[i] = nonce
			continue
		}

		if ni.finalNotarized[i] == 0 {
			ni.finalNotarized[i] = notarizedNonce[i]
			ni.blockWithFinalNotarized[i] = nonce
			continue
		}
	}

	foundAllNotarizedShardHeaders := true
	for i := uint32(0); i < boot.shardCoordinator.NumberOfShards(); i++ {
		if ni.lastNotarized[i] == 0 || ni.finalNotarized[i] == 0 {
			foundAllNotarizedShardHeaders = false
			break
		}
	}

	return foundAllNotarizedShardHeaders
}

func (boot *MetaBootstrap) applyNotarizedBlocks(
	finalNotarized map[uint32]uint64,
	lastNotarized map[uint32]uint64,
) error {
	for i := uint32(0); i < boot.shardCoordinator.NumberOfShards(); i++ {
		nonce := finalNotarized[i]
		if nonce > 0 {
			headerHandler, _, err := boot.getShardHeaderFromStorage(i, nonce)
			if err != nil {
				return err
			}

			boot.blkExecutor.AddLastNotarizedHdr(i, headerHandler)
		}

		nonce = lastNotarized[i]
		if nonce > 0 {
			headerHandler, _, err := boot.getShardHeaderFromStorage(i, nonce)
			if err != nil {
				return err
			}

			boot.blkExecutor.AddLastNotarizedHdr(i, headerHandler)
		}
	}

	return nil
}

func (boot *MetaBootstrap) cleanupNotarizedStorage(lastNotarized map[uint32]uint64) {
	for i := uint32(0); i < boot.shardCoordinator.NumberOfShards(); i++ {
		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(i)
		highestNonceInStorer := boot.computeHighestNonce(hdrNonceHashDataUnit)

		for i := lastNotarized[i] + 1; i <= highestNonceInStorer; i++ {
			errNotCritical := boot.removeBlockHeader(i, dataRetriever.BlockHeaderUnit, hdrNonceHashDataUnit)
			if errNotCritical != nil {
				log.Debug("remove notarized block header",
					"nonce", i,
					"error", errNotCritical.Error(),
				)
			}
		}
	}
}

func (boot *MetaBootstrap) receivedHeader(headerHash []byte) {
	header, err := process.GetMetaHeaderFromPool(headerHash, boot.headers)
	if err != nil {
		log.Trace("GetMetaHeaderFromPool", "error", err.Error())
		return
	}

	boot.processReceivedHeader(header, headerHash)
}

// StartSync method will start SyncBlocks as a go routine
func (boot *MetaBootstrap) StartSync() {
	// when a node starts it first tries to bootstrap from storage, if there already exist a database saved
	errNotCritical := boot.syncFromStorer(process.MetaBlockFinality,
		dataRetriever.MetaBlockUnit,
		dataRetriever.MetaHdrNonceHashDataUnit,
		process.ShardBlockFinality)
	if errNotCritical != nil {
		log.Debug("syncFromStorer", "error", errNotCritical.Error())
	}

	go boot.syncBlocks()
}

// SyncBlock method actually does the synchronization. It requests the next block header from the pool
// and if it is not found there it will be requested from the network. After the header is received,
// it requests the block body in the same way(pool and than, if it is not found in the pool, from network).
// If either header and body are received the ProcessBlock and CommitBlock method will be called successively.
// These methods will execute the block and its transactions. Finally if everything works, the block will be committed
// in the blockchain, and all this mechanism will be reiterated for the next block.
func (boot *MetaBootstrap) SyncBlock() error {
	return boot.syncBlock()
}

// requestHeaderWithNonce method requests a block header from network when it is not found in the pool
func (boot *MetaBootstrap) requestHeaderWithNonce(nonce uint64) {
	boot.setRequestedHeaderNonce(&nonce)
	err := boot.hdrRes.RequestDataFromNonce(nonce)
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
func (boot *MetaBootstrap) requestHeaderWithHash(hash []byte) {
	boot.setRequestedHeaderHash(hash)
	err := boot.hdrRes.RequestDataFromHash(hash)
	if err != nil {
		log.Debug("RequestDataFromHash", "error", err.Error())
	}

	log.Debug("requested header from network",
		"hash", display.ConvertHash(hash),
	)
}

// getHeaderWithNonceRequestingIfMissing method gets the header with a given nonce from pool. If it is not found there, it will
// be requested from network
func (boot *MetaBootstrap) getHeaderWithNonceRequestingIfMissing(nonce uint64) (data.HeaderHandler, error) {
	hdr, _, err := process.GetMetaHeaderFromPoolWithNonce(
		nonce,
		boot.headers,
		boot.headersNonces)
	if err != nil {
		_ = process.EmptyChannel(boot.chRcvHdrNonce)
		boot.requestHeaderWithNonce(nonce)
		err := boot.waitForHeaderNonce()
		if err != nil {
			return nil, err
		}

		hdr, _, err = process.GetMetaHeaderFromPoolWithNonce(
			nonce,
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
func (boot *MetaBootstrap) getHeaderWithHashRequestingIfMissing(hash []byte) (data.HeaderHandler, error) {
	hdr, err := process.GetMetaHeader(hash, boot.headers, boot.marshalizer, boot.store)
	if err != nil {
		_ = process.EmptyChannel(boot.chRcvHdrHash)
		boot.requestHeaderWithHash(hash)
		err := boot.waitForHeaderHash()
		if err != nil {
			return nil, err
		}

		hdr, err = process.GetMetaHeaderFromPool(hash, boot.headers)
		if err != nil {
			return nil, err
		}
	}

	return hdr, nil
}

func (boot *MetaBootstrap) getPrevHeader(
	header data.HeaderHandler,
	headerStore storage.Storer,
) (data.HeaderHandler, error) {

	prevHash := header.GetPrevHash()
	buffHeader, err := headerStore.Get(prevHash)
	if err != nil {
		return nil, err
	}

	prevHeader := &block.MetaBlock{}
	err = boot.marshalizer.Unmarshal(prevHeader, buffHeader)
	if err != nil {
		return nil, err
	}

	return prevHeader, nil
}

func (boot *MetaBootstrap) getCurrHeader() (data.HeaderHandler, error) {
	blockHeader := boot.blkc.GetCurrentBlockHeader()
	if blockHeader == nil {
		return nil, process.ErrNilBlockHeader
	}

	header, ok := blockHeader.(*block.MetaBlock)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return header, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (boot *MetaBootstrap) IsInterfaceNil() bool {
	if boot == nil {
		return true
	}
	return false
}

func (boot *MetaBootstrap) haveHeaderInPoolWithNonce(nonce uint64) bool {
	_, _, err := process.GetMetaHeaderFromPoolWithNonce(
		nonce,
		boot.headers,
		boot.headersNonces)

	return err == nil
}

func (boot *MetaBootstrap) getMetaHeaderFromPool(headerHash []byte) (data.HeaderHandler, error) {
	return process.GetMetaHeaderFromPool(headerHash, boot.headers)
}

func (boot *MetaBootstrap) getBlockBodyRequestingIfMissing(headerHandler data.HeaderHandler) (data.BodyHandler, error) {
	return boot.getBlockBody(headerHandler)
}
