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

type shardBlockTrack struct {
	*baseBlockTrack
	headersPool       storage.Cacher
	metaBlocksPool    storage.Cacher
	headersNoncesPool dataRetriever.Uint64SyncMapCacher
	store             dataRetriever.StorageService
}

// NewShardBlockTrack creates an object for tracking the received shard blocks
func NewShardBlockTrack(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	poolsHolder dataRetriever.PoolsHolder,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	store dataRetriever.StorageService,
	startHeaders map[uint32]data.HeaderHandler,
) (*shardBlockTrack, error) {

	err := checkTrackerNilParameters(hasher, marshalizer, rounder, shardCoordinator, store)
	if err != nil {
		return nil, err
	}

	if check.IfNil(poolsHolder) {
		return nil, process.ErrNilPoolsHolder
	}
	if check.IfNil(poolsHolder.Headers()) {
		return nil, process.ErrNilHeadersDataPool
	}
	if check.IfNil(poolsHolder.MetaBlocks()) {
		return nil, process.ErrNilMetaBlocksPool
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

	sbt := &shardBlockTrack{
		baseBlockTrack:    bbt,
		headersPool:       poolsHolder.Headers(),
		metaBlocksPool:    poolsHolder.MetaBlocks(),
		headersNoncesPool: poolsHolder.HeadersNonces(),
		store:             store,
	}

	sbt.headers = make(map[uint32]map[uint64][]*headerInfo)
	sbt.longestChainHeadersIndexes = make([]int, 0)
	sbt.headersPool.RegisterHandler(sbt.receivedHeader)
	sbt.metaBlocksPool.RegisterHandler(sbt.receivedMetaBlock)

	sbt.selfNotarizedHandlers = make([]func(headers []data.HeaderHandler), 0)

	sbt.blockFinality = process.MetaBlockFinality

	sbt.blockTracker = sbt

	return sbt, nil
}

func (sbt *shardBlockTrack) receivedHeader(headerHash []byte) {
	header, err := process.GetShardHeaderFromPool(headerHash, sbt.headersPool)
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

	lastSelfNotarizedHeaderNonce := sbt.getLastSelfNotarizedHeaderNonce(header.GetShardID())
	if sbt.isHeaderOutOfRange(header.GetNonce(), lastSelfNotarizedHeaderNonce) {
		log.Debug("shard header is out of range",
			"received nonce", header.GetNonce(),
			"last self notarized nonce", lastSelfNotarizedHeaderNonce,
		)
		return
	}

	sbt.addMissingShardHeaders(
		header.GetShardID(),
		lastSelfNotarizedHeaderNonce+2,
		header.GetNonce()-1,
		sbt.headersPool,
		sbt.headersNoncesPool)

	sbt.addHeader(header, headerHash)
	sbt.doJobOnReceivedBlock(header.GetShardID())
}

func (sbt *shardBlockTrack) receivedMetaBlock(metaBlockHash []byte) {
	metaBlock, err := process.GetMetaHeaderFromPool(metaBlockHash, sbt.metaBlocksPool)
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

	lastCrossNotarizedHeaderNonce := sbt.getLastCrossNotarizedHeaderNonce(metaBlock.GetShardID())
	if sbt.isHeaderOutOfRange(metaBlock.GetNonce(), lastCrossNotarizedHeaderNonce) {
		log.Debug("meta block is out of range",
			"received nonce", metaBlock.GetNonce(),
			"last cross notarized nonce", lastCrossNotarizedHeaderNonce,
		)
		return
	}

	sbt.addMissingMetaBlocks(
		lastCrossNotarizedHeaderNonce+2,
		metaBlock.GetNonce()-1,
		sbt.metaBlocksPool,
		sbt.headersNoncesPool)

	sbt.addHeader(metaBlock, metaBlockHash)
	sbt.doJobOnReceivedCrossNotarizedBlock(metaBlock.GetShardID())
}

func (sbt *shardBlockTrack) computeSelfNotarizedHeaders(headers []data.HeaderHandler) []data.HeaderHandler {
	selfNotarizedHeaders := make([]data.HeaderHandler, 0)

	for _, header := range headers {
		metaBlock, ok := header.(*block.MetaBlock)
		if !ok {
			log.Debug("computeSelfNotarizedHeaders", process.ErrWrongTypeAssertion)
			continue
		}

		selfHeaders := sbt.getSelfHeaders(metaBlock)
		if len(selfHeaders) > 0 {
			selfNotarizedHeaders = append(selfNotarizedHeaders, selfHeaders...)
		}
	}

	process.SortHeadersByNonce(selfNotarizedHeaders)

	return selfNotarizedHeaders
}

func (sbt *shardBlockTrack) getSelfHeaders(metaBlock *block.MetaBlock) []data.HeaderHandler {
	selfHeaders := make([]data.HeaderHandler, 0)

	for _, shardInfo := range metaBlock.ShardInfo {
		if shardInfo.ShardID != sbt.shardCoordinator.SelfId() {
			continue
		}

		header, err := process.GetShardHeader(shardInfo.HeaderHash, sbt.headersPool, sbt.marshalizer, sbt.store)
		if err != nil {
			log.Debug("GetShardHeader", err.Error())
			continue
		}

		selfHeaders = append(selfHeaders, header)
	}

	return selfHeaders
}

func (sbt *shardBlockTrack) cleanupHeadersForSelfShard() {
	//TODO: Should be analyzed if this nonce should be calculated differently
	nonce := sbt.getLastSelfNotarizedHeaderNonce(sharding.MetachainShardId)
	sbt.cleanupHeadersForShardBehindNonce(sbt.shardCoordinator.SelfId(), nonce)
}
