package track

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type shardBlockTrack struct {
	*baseBlockTrack
}

// NewShardBlockTrack creates an object for tracking the received shard blocks
func NewShardBlockTrack(arguments ArgShardTracker) (*shardBlockTrack, error) {
	err := checkTrackerNilParameters(arguments.ArgBaseTracker)
	if err != nil {
		return nil, err
	}

	if check.IfNil(arguments.PoolsHolder) {
		return nil, process.ErrNilPoolsHolder
	}
	if check.IfNil(arguments.PoolsHolder.Headers()) {
		return nil, process.ErrNilHeadersDataPool
	}
	if check.IfNil(arguments.PoolsHolder.MetaBlocks()) {
		return nil, process.ErrNilMetaBlocksPool
	}
	if check.IfNil(arguments.PoolsHolder.HeadersNonces()) {
		return nil, process.ErrNilHeadersNoncesDataPool
	}

	bbt := &baseBlockTrack{
		hasher:            arguments.Hasher,
		headerValidator:   arguments.HeaderValidator,
		marshalizer:       arguments.Marshalizer,
		rounder:           arguments.Rounder,
		shardCoordinator:  arguments.ShardCoordinator,
		shardHeadersPool:  arguments.PoolsHolder.Headers(),
		metaBlocksPool:    arguments.PoolsHolder.MetaBlocks(),
		headersNoncesPool: arguments.PoolsHolder.HeadersNonces(),
		store:             arguments.Store,
	}

	err = bbt.initCrossNotarizedHeaders(arguments.StartHeaders)
	if err != nil {
		return nil, err
	}

	err = bbt.initSelfNotarizedHeaders(arguments.StartHeaders)
	if err != nil {
		return nil, err
	}

	sbt := &shardBlockTrack{
		baseBlockTrack: bbt,
	}

	sbt.headers = make(map[uint32]map[uint64][]*headerInfo)
	sbt.shardHeadersPool.RegisterHandler(sbt.receivedShardHeader)
	sbt.metaBlocksPool.RegisterHandler(sbt.receivedMetaBlock)

	sbt.selfNotarizedHeadersHandlers = make([]func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte), 0)
	sbt.crossNotarizedHeadersHandlers = make([]func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte), 0)

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
