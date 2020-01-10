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
		metaBlocksPool:                arguments.PoolsHolder.MetaBlocks(),
		shardHeadersPool:              arguments.PoolsHolder.ShardHeaders(),
		headersNoncesPool:             arguments.PoolsHolder.HeadersNonces(),
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

	mbt := metaBlockTrack{
		baseBlockTrack: bbt,
	}

	blockProcessor, err := NewBlockProcessor(
		arguments.HeaderValidator,
		arguments.RequestHandler,
		arguments.ShardCoordinator,
		&mbt,
		crossNotarizer,
		crossNotarizedHeadersNotifier,
		selfNotarizedHeadersNotifier,
	)
	if err != nil {
		return nil, err
	}

	mbt.blockProcessor = blockProcessor

	mbt.headers = make(map[uint32]map[uint64][]*headerInfo)
	mbt.metaBlocksPool.RegisterHandler(mbt.receivedMetaBlock)
	mbt.shardHeadersPool.RegisterHandler(mbt.receivedShardHeader)

	return &mbt, nil
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
	lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash, err := mbt.selfNotarizer.getLastNotarizedHeader(mbt.shardCoordinator.SelfId())
	if err != nil {
		log.Warn("computeLongestSelfChain.getLastNotarizedHeader", "error", err.Error())
		return nil, nil, nil, nil
	}

	headers, hashes := mbt.ComputeLongestChain(mbt.shardCoordinator.SelfId(), lastSelfNotarizedHeader)
	return lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash, headers, hashes
}
