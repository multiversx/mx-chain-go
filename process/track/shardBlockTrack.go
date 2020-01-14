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

	crossNotarizer, err := NewBlockNotarizer(arguments.Hasher, arguments.Marshalizer)
	if err != nil {
		return nil, err
	}

	selfNotarizer, err := NewBlockNotarizer(arguments.Hasher, arguments.Marshalizer)
	if err != nil {
		return nil, err
	}

	crossNotarizedHeadersNotifier, err := NewBlockNotifier()
	if err != nil {
		return nil, err
	}

	selfNotarizedHeadersNotifier, err := NewBlockNotifier()
	if err != nil {
		return nil, err
	}

	bbt := &baseBlockTrack{
		hasher:                        arguments.Hasher,
		headerValidator:               arguments.HeaderValidator,
		marshalizer:                   arguments.Marshalizer,
		rounder:                       arguments.Rounder,
		shardCoordinator:              arguments.ShardCoordinator,
		headersPool:                   arguments.PoolsHolder.Headers(),
		store:                         arguments.Store,
		crossNotarizer:                crossNotarizer,
		selfNotarizer:                 selfNotarizer,
		crossNotarizedHeadersNotifier: crossNotarizedHeadersNotifier,
		selfNotarizedHeadersNotifier:  selfNotarizedHeadersNotifier,
	}

	err = bbt.initNotarizedHeaders(arguments.StartHeaders)
	if err != nil {
		return nil, err
	}

	sbt := shardBlockTrack{
		baseBlockTrack: bbt,
	}

	blockProcessor, err := NewBlockProcessor(
		arguments.HeaderValidator,
		arguments.ShardCoordinator,
		&sbt,
		crossNotarizer,
		crossNotarizedHeadersNotifier,
		selfNotarizedHeadersNotifier,
	)
	if err != nil {
		return nil, err
	}

	sbt.blockProcessor = blockProcessor

	sbt.headers = make(map[uint32]map[uint64][]*headerInfo)
	sbt.headersPool.RegisterHandler(sbt.receivedHeader)

	return &sbt, nil
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

		header, err := process.GetShardHeader(shardInfo.HeaderHash, sbt.headersPool, sbt.marshalizer, sbt.store)
		if err != nil {
			log.Debug("GetShardHeader", err.Error())
			continue
		}

		selfHeadersInfo = append(selfHeadersInfo, &headerInfo{hash: shardInfo.HeaderHash, header: header})
	}

	return selfHeadersInfo
}

func (sbt *shardBlockTrack) computeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash, err := sbt.selfNotarizer.getLastNotarizedHeader(sharding.MetachainShardId)
	if err != nil {
		log.Warn("computeLongestSelfChain.getLastNotarizedHeader", "error", err.Error())
		return nil, nil, nil, nil
	}

	headers, hashes := sbt.ComputeLongestChain(sbt.shardCoordinator.SelfId(), lastSelfNotarizedHeader)
	return lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash, headers, hashes
}
