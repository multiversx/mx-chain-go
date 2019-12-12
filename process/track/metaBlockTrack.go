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

type metaBlockTrack struct {
	*baseBlockTrack
	store dataRetriever.StorageService
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
		hasher:            hasher,
		marshalizer:       marshalizer,
		rounder:           rounder,
		shardCoordinator:  shardCoordinator,
		metaBlocksPool:    poolsHolder.MetaBlocks(),
		shardHeadersPool:  poolsHolder.ShardHeaders(),
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

	mbt := &metaBlockTrack{
		baseBlockTrack: bbt,
		store:          store,
	}

	mbt.headers = make(map[uint32]map[uint64][]*headerInfo)
	mbt.longestChainHeadersIndexes = make([]int, 0)
	mbt.metaBlocksPool.RegisterHandler(mbt.receivedMetaBlock)
	mbt.shardHeadersPool.RegisterHandler(mbt.receivedShardHeader)

	mbt.selfNotarizedHeadersHandlers = make([]func(headers []data.HeaderHandler, headersHashes [][]byte), 0)

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
