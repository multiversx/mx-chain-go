package track

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type metaBlockTrack struct {
	*baseBlockTrack
	metaBlocksPool   storage.Cacher
	shardHeadersPool storage.Cacher
}

// NewMetaBlockTrack creates an object for tracking the received meta blocks
func NewMetaBlockTrack(
	poolsHolder dataRetriever.MetaPoolsHolder,
	rounder consensus.Rounder,
) (*metaBlockTrack, error) {

	if check.IfNil(poolsHolder) {
		return nil, process.ErrNilPoolsHolder
	}
	if check.IfNil(poolsHolder.MetaBlocks()) {
		return nil, process.ErrNilMetaBlocksPool
	}
	if check.IfNil(poolsHolder.ShardHeaders()) {
		return nil, process.ErrNilShardBlockPool
	}
	if check.IfNil(rounder) {
		return nil, process.ErrNilRounder
	}

	bbt := &baseBlockTrack{
		rounder: rounder,
	}

	mbt := &metaBlockTrack{
		baseBlockTrack:   bbt,
		metaBlocksPool:   poolsHolder.MetaBlocks(),
		shardHeadersPool: poolsHolder.ShardHeaders(),
	}

	mbt.headers = make(map[uint32]map[uint64][]*headerInfo)
	mbt.metaBlocksPool.RegisterHandler(mbt.receivedMetaBlock)
	mbt.shardHeadersPool.RegisterHandler(mbt.receivedShardHeader)

	return mbt, nil
}

func (mbt *metaBlockTrack) receivedMetaBlock(metaBlockHash []byte) {
	metaBlock, err := process.GetMetaHeaderFromPool(metaBlockHash, mbt.metaBlocksPool)
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

	mbt.AddHeader(metaBlock, metaBlockHash)
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

	mbt.AddHeader(header, headerHash)
}
