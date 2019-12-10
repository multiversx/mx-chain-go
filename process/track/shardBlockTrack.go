package track

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type shardBlockTrack struct {
	*baseBlockTrack
	headersPool    storage.Cacher
	metaBlocksPool storage.Cacher
}

// NewShardBlockTrack creates an object for tracking the received shard blocks
func NewShardBlockTrack(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	poolsHolder dataRetriever.PoolsHolder,
	rounder consensus.Rounder,
	shardCoordinator sharding.Coordinator,
	startHeaders map[uint32]data.HeaderHandler,
) (*shardBlockTrack, error) {

	err := checkTrackerNilParameters(hasher, marshalizer, rounder, shardCoordinator)
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

	bbt := &baseBlockTrack{
		hasher:           hasher,
		marshalizer:      marshalizer,
		rounder:          rounder,
		shardCoordinator: shardCoordinator,
	}

	err = bbt.setNotarizedHeaders(startHeaders)
	if err != nil {
		return nil, err
	}

	err = bbt.setFinalizedHeaders(startHeaders)
	if err != nil {
		return nil, err
	}

	sbt := &shardBlockTrack{
		baseBlockTrack: bbt,
		headersPool:    poolsHolder.Headers(),
		metaBlocksPool: poolsHolder.MetaBlocks(),
	}

	sbt.headers = make(map[uint32]map[uint64][]*headerInfo)
	sbt.longestChainHeadersIndexes = make([]int, 0)
	sbt.headersPool.RegisterHandler(sbt.receivedHeader)
	sbt.metaBlocksPool.RegisterHandler(sbt.receivedMetaBlock)

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

	sbt.AddHeader(header, headerHash)
	sbt.doJobOnReceivedBlock(header.GetShardID())
}

func (sbt *shardBlockTrack) receivedMetaBlock(metaBlockHash []byte) {
	metaBlock, err := process.GetMetaHeaderFromPool(metaBlockHash, sbt.metaBlocksPool)
	if err != nil {
		log.Trace("GetMetaBlockFromPool", "error", err.Error())
		return
	}

	log.Debug("received meta block from network in block tracker",
		"shard", metaBlock.GetShardID(),
		"round", metaBlock.GetRound(),
		"nonce", metaBlock.GetNonce(),
		"hash", metaBlockHash,
	)

	sbt.AddHeader(metaBlock, metaBlockHash)
	sbt.doJobOnReceivedCrossNotarizedBlock(metaBlock.GetShardID())
}

func (sbt *shardBlockTrack) computeFinalizedHeaders(headers []data.HeaderHandler) []data.HeaderHandler {
	return nil
}
