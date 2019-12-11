package track

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type metaBlockTrack struct {
	*baseBlockTrack
	metaBlocksPool    storage.Cacher
	shardHeadersPool  storage.Cacher
	headersNoncesPool dataRetriever.Uint64SyncMapCacher
	store             dataRetriever.StorageService
}

// NewMetaBlockTrack creates an object for tracking the received meta blocks
func NewMetaBlockTrack(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	poolsHolder dataRetriever.MetaPoolsHolder,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	store dataRetriever.StorageService,
	startHeaders map[uint32]data.HeaderHandler,
) (*metaBlockTrack, error) {

	err := checkTrackerNilParameters(hasher, marshalizer, rounder, shardCoordinator, store)
	if err != nil {
		return nil, err
	}

	if check.IfNil(poolsHolder) {
		return nil, process.ErrNilPoolsHolder
	}
	if check.IfNil(poolsHolder.MetaBlocks()) {
		return nil, process.ErrNilMetaBlocksPool
	}
	if check.IfNil(poolsHolder.ShardHeaders()) {
		return nil, process.ErrNilShardBlockPool
	}
	if check.IfNil(poolsHolder.HeadersNonces()) {
		return nil, process.ErrNilHeadersNoncesDataPool
	}

	bbt := &baseBlockTrack{
		hasher:           hasher,
		marshalizer:      marshalizer,
		rounder:          rounder,
		shardCoordinator: shardCoordinator,
	}

	err = bbt.setCrossNotarizedHeaders(startHeaders)
	if err != nil {
		return nil, err
	}

	err = bbt.setSelfNotarizedHeaders(startHeaders)
	if err != nil {
		return nil, err
	}

	mbt := &metaBlockTrack{
		baseBlockTrack:    bbt,
		metaBlocksPool:    poolsHolder.MetaBlocks(),
		shardHeadersPool:  poolsHolder.ShardHeaders(),
		headersNoncesPool: poolsHolder.HeadersNonces(),
		store:             store,
	}

	mbt.headers = make(map[uint32]map[uint64][]*headerInfo)
	mbt.longestChainHeadersIndexes = make([]int, 0)
	mbt.metaBlocksPool.RegisterHandler(mbt.receivedMetaBlock)
	mbt.shardHeadersPool.RegisterHandler(mbt.receivedShardHeader)

	mbt.selfNotarizedHandlers = make([]func(headers []data.HeaderHandler), 0)

	mbt.blockFinality = process.ShardBlockFinality

	mbt.blockTracker = mbt

	return mbt, nil
}

func (mbt *metaBlockTrack) receivedMetaBlock(metaBlockHash []byte) {
	metaBlock, err := process.GetMetaHeaderFromPool(metaBlockHash, mbt.metaBlocksPool)
	if err != nil {
		log.Trace("GetMetaHeaderFromPool", "error", err.Error())
		return
	}

	log.Debug("received meta block from network in block tracker",
		"shard", metaBlock.GetShardID(),
		"round", metaBlock.GetRound(),
		"nonce", metaBlock.GetNonce(),
		"hash", metaBlockHash,
	)

	lastSelfNotarizedHeaderNonce := mbt.getLastSelfNotarizedHeaderNonce(metaBlock.GetShardID())
	if mbt.isHeaderOutOfRange(metaBlock.GetNonce(), lastSelfNotarizedHeaderNonce) {
		log.Debug("meta block is out of range",
			"received nonce", metaBlock.GetNonce(),
			"last self notarized nonce", lastSelfNotarizedHeaderNonce,
		)
		return
	}

	mbt.addMissingMetaBlocks(
		lastSelfNotarizedHeaderNonce+2,
		metaBlock.GetNonce()-1,
		mbt.metaBlocksPool,
		mbt.headersNoncesPool)

	mbt.addHeader(metaBlock, metaBlockHash)
	mbt.doJobOnReceivedBlock(metaBlock.GetShardID())
}

func (mbt *metaBlockTrack) receivedShardHeader(headerHash []byte) {
	header, err := process.GetShardHeaderFromPool(headerHash, mbt.shardHeadersPool)
	if err != nil {
		log.Trace("GetShardHeaderFromPool", "error", err.Error())
		return
	}

	log.Debug("received shard header from network in block tracker",
		"shard", header.GetShardID(),
		"round", header.GetRound(),
		"nonce", header.GetNonce(),
		"hash", headerHash,
	)

	lastCrossNotarizedHeaderNonce := mbt.getLastCrossNotarizedHeaderNonce(header.GetShardID())
	if mbt.isHeaderOutOfRange(header.GetNonce(), lastCrossNotarizedHeaderNonce) {
		log.Debug("shard header is out of range",
			"received nonce", header.GetNonce(),
			"last cross notarized nonce", lastCrossNotarizedHeaderNonce,
		)
		return
	}

	mbt.addMissingShardHeaders(
		header.GetShardID(),
		lastCrossNotarizedHeaderNonce+2,
		header.GetNonce()-1,
		mbt.shardHeadersPool,
		mbt.headersNoncesPool)

	mbt.addHeader(header, headerHash)
	mbt.doJobOnReceivedCrossNotarizedBlock(header.GetShardID())
}

func (mbt *metaBlockTrack) computeSelfNotarizedHeaders(headers []data.HeaderHandler) []data.HeaderHandler {
	selfNotarizedHeaders := make([]data.HeaderHandler, 0)

	for _, header := range headers {
		header, ok := header.(*block.Header)
		if !ok {
			log.Debug("computeSelfNotarizedHeaders", process.ErrWrongTypeAssertion)
			continue
		}

		selfMetaBlocks := mbt.getSelfMetaBlocks(header)
		if len(selfMetaBlocks) > 0 {
			selfNotarizedHeaders = append(selfNotarizedHeaders, selfMetaBlocks...)
		}
	}

	process.SortHeadersByNonce(selfNotarizedHeaders)

	return selfNotarizedHeaders
}

func (mbt *metaBlockTrack) getSelfMetaBlocks(header *block.Header) []data.HeaderHandler {
	selfMetaBlocks := make([]data.HeaderHandler, 0)

	for _, metaBlockHash := range header.MetaBlockHashes {
		metaBlock, err := process.GetMetaHeader(metaBlockHash, mbt.metaBlocksPool, mbt.marshalizer, mbt.store)
		if err != nil {
			log.Debug("GetMetaHeader", err.Error())
			continue
		}

		selfMetaBlocks = append(selfMetaBlocks, metaBlock)
	}

	return selfMetaBlocks
}

func (mbt *metaBlockTrack) cleanupHeadersForSelfShard() {
	//TODO: Should be analyzed if this nonce should be calculated differently
	nonce := mbt.getLastSelfNotarizedHeaderNonce(mbt.shardCoordinator.SelfId())
	mbt.cleanupHeadersForShardBehindNonce(mbt.shardCoordinator.SelfId(), nonce)
}
