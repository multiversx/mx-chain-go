package track

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

type metaBlockTrack struct {
	*baseBlockTrack
}

// NewMetaBlockTrack creates an object for tracking the received meta blocks
func NewMetaBlockTrack(arguments ArgMetaTracker) (*metaBlockTrack, error) {
	err := checkTrackerNilParameters(arguments.ArgBaseTracker)
	if err != nil {
		return nil, err
	}

	if check.IfNil(arguments.PoolsHolder) {
		return nil, process.ErrNilPoolsHolder
	}
	if check.IfNil(arguments.PoolsHolder.MetaBlocks()) {
		return nil, process.ErrNilMetaBlocksPool
	}
	if check.IfNil(arguments.PoolsHolder.ShardHeaders()) {
		return nil, process.ErrNilShardBlockPool
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
		metaBlocksPool:    arguments.PoolsHolder.MetaBlocks(),
		shardHeadersPool:  arguments.PoolsHolder.ShardHeaders(),
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

	mbt := &metaBlockTrack{
		baseBlockTrack: bbt,
	}

	mbt.headers = make(map[uint32]map[uint64][]*headerInfo)
	mbt.metaBlocksPool.RegisterHandler(mbt.receivedMetaBlock)
	mbt.shardHeadersPool.RegisterHandler(mbt.receivedShardHeader)

	mbt.selfNotarizedHeadersHandlers = make([]func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte), 0)
	mbt.crossNotarizedHeadersHandlers = make([]func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte), 0)

	mbt.blockFinality = process.BlockFinality

	mbt.blockTracker = mbt

	return mbt, nil
}

func (mbt *metaBlockTrack) getSelfHeaders(headerHandler data.HeaderHandler) []*headerInfo {
	selfMetaBlocksInfo := make([]*headerInfo, 0)

	header, ok := headerHandler.(*block.Header)
	if !ok {
		log.Debug("getSelfHeaders", process.ErrWrongTypeAssertion)
		return selfMetaBlocksInfo
	}

	for _, metaBlockHash := range header.MetaBlockHashes {
		metaBlock, err := process.GetMetaHeader(metaBlockHash, mbt.metaBlocksPool, mbt.marshalizer, mbt.store)
		if err != nil {
			log.Debug("GetMetaHeader", err.Error())
			continue
		}

		selfMetaBlocksInfo = append(selfMetaBlocksInfo, &headerInfo{hash: metaBlockHash, header: metaBlock})
	}

	return selfMetaBlocksInfo
}

func (mbt *metaBlockTrack) computeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	lastSelfNotarizedHeaderInfo, err := mbt.getLastSelfNotarizedHeader(mbt.shardCoordinator.SelfId())
	if err != nil {
		return nil, nil, nil, nil
	}

	headers, hashes := mbt.ComputeLongestChain(mbt.shardCoordinator.SelfId(), lastSelfNotarizedHeaderInfo.header)
	return lastSelfNotarizedHeaderInfo.header, lastSelfNotarizedHeaderInfo.hash, headers, hashes
}
