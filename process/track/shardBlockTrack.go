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
)

type shardBlockTrack struct {
	*baseBlockTrack
	store dataRetriever.StorageService
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
		hasher:            hasher,
		marshalizer:       marshalizer,
		rounder:           rounder,
		shardCoordinator:  shardCoordinator,
		shardHeadersPool:  poolsHolder.Headers(),
		metaBlocksPool:    poolsHolder.MetaBlocks(),
		headersNoncesPool: poolsHolder.HeadersNonces(),
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
		baseBlockTrack: bbt,
		store:          store,
	}

	sbt.headers = make(map[uint32]map[uint64][]*headerInfo)
	sbt.shardHeadersPool.RegisterHandler(sbt.receivedShardHeader)
	sbt.metaBlocksPool.RegisterHandler(sbt.receivedMetaBlock)

	sbt.selfNotarizedHeadersHandlers = make([]func(headers []data.HeaderHandler, headersHashes [][]byte), 0)
	sbt.crossNotarizedHeadersHandlers = make([]func(headers []data.HeaderHandler, headersHashes [][]byte), 0)

	sbt.blockFinality = process.BlockFinality

	sbt.blockTracker = sbt

	return sbt, nil
}

func (sbt *shardBlockTrack) getSelfHeaders(headerHandler data.HeaderHandler) []*headerInfo {
	selfHeadersInfo := make([]*headerInfo, 0)

	metaBlock, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		log.Debug("getSelfHeaders", process.ErrWrongTypeAssertion)
		return selfHeadersInfo
	}

	for _, shardInfo := range metaBlock.ShardInfo {
		if shardInfo.ShardID != sbt.shardCoordinator.SelfId() {
			continue
		}

		header, err := process.GetShardHeader(shardInfo.HeaderHash, sbt.shardHeadersPool, sbt.marshalizer, sbt.store)
		if err != nil {
			log.Debug("GetShardHeader", err.Error())
			continue
		}

		selfHeadersInfo = append(selfHeadersInfo, &headerInfo{hash: shardInfo.HeaderHash, header: header})
	}

	return selfHeadersInfo
}

func (sbt *shardBlockTrack) computeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	lastSelfNotarizedHeaderInfo, err := sbt.getLastSelfNotarizedHeader(sharding.MetachainShardId)
	if err != nil {
		return nil, nil, nil, nil
	}

	headers, hashes := sbt.ComputeLongestChain(sbt.shardCoordinator.SelfId(), lastSelfNotarizedHeaderInfo.header)
	return lastSelfNotarizedHeaderInfo.header, lastSelfNotarizedHeaderInfo.hash, headers, hashes
}
