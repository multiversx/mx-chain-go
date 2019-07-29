package track

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.DefaultLogger()

type headerInfo struct {
	header           data.HeaderHandler
	broadcastInRound int32
}

// shardBlockTracker implements NotarisedBlocksTracker interface which tracks notarised blocks
type shardBlockTracker struct {
	dataPool         dataRetriever.PoolsHolder
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator
	store            dataRetriever.StorageService

	mutUnnotarisedHeaders sync.RWMutex
	unnotarisedHeaders    map[uint64]*headerInfo
}

// NewShardBlockTracker creates a new shardBlockTracker object
func NewShardBlockTracker(
	dataPool dataRetriever.PoolsHolder,
	marshalizer marshal.Marshalizer,
	shardCoordinator sharding.Coordinator,
	store dataRetriever.StorageService,
) (*shardBlockTracker, error) {
	err := checkTrackerNilParameters(
		dataPool,
		marshalizer,
		shardCoordinator,
		store)
	if err != nil {
		return nil, err
	}

	sbt := shardBlockTracker{
		dataPool:         dataPool,
		marshalizer:      marshalizer,
		shardCoordinator: shardCoordinator,
		store:            store,
	}

	sbt.unnotarisedHeaders = make(map[uint64]*headerInfo)

	return &sbt, nil
}

// checkTrackerNilParameters will check the imput parameters for nil values
func checkTrackerNilParameters(
	dataPool dataRetriever.PoolsHolder,
	marshalizer marshal.Marshalizer,
	shardCoordinator sharding.Coordinator,
	store dataRetriever.StorageService,
) error {
	if dataPool == nil {
		return process.ErrNilDataPoolHolder
	}
	if marshalizer == nil {
		return process.ErrNilMarshalizer
	}
	if shardCoordinator == nil {
		return process.ErrNilShardCoordinator
	}
	if store == nil {
		return process.ErrNilStorage
	}

	return nil
}

// AddBlock adds new block to be tracked
func (sbt *shardBlockTracker) AddBlock(headerHandler data.HeaderHandler) {
	sbt.mutUnnotarisedHeaders.Lock()
	sbt.unnotarisedHeaders[headerHandler.GetNonce()] = &headerInfo{header: headerHandler, broadcastInRound: 0}
	sbt.mutUnnotarisedHeaders.Unlock()
}

// RemoveNotarisedBlocks removes all the blocks which already have been notarised
func (sbt *shardBlockTracker) RemoveNotarisedBlocks(headerHandler data.HeaderHandler) error {
	metaBlock, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	for _, shardData := range metaBlock.ShardInfo {
		if shardData.ShardId != sbt.shardCoordinator.SelfId() {
			continue
		}

		//TODO: Shoud be called process.GetShardHeader?
		header, err := process.GetShardHeaderFromPool(
			shardData.HeaderHash,
			sbt.dataPool.Headers())
		if err != nil {
			continue
		}

		sbt.mutUnnotarisedHeaders.Lock()
		delete(sbt.unnotarisedHeaders, header.Nonce)
		sbt.mutUnnotarisedHeaders.Unlock()

		log.Debug(fmt.Sprintf("shardBlock with nonce %d and hash %s has been notarised by metachain\n",
			header.GetNonce(),
			core.ToB64(shardData.HeaderHash)))
	}

	return nil
}

// UnnotarisedBlocks gets all the blocks which are not notarised yet
func (sbt *shardBlockTracker) UnnotarisedBlocks() []data.HeaderHandler {
	sbt.mutUnnotarisedHeaders.RLock()

	hdrs := make([]data.HeaderHandler, 0)
	for _, hInfo := range sbt.unnotarisedHeaders {
		hdrs = append(hdrs, hInfo.header)
	}

	sbt.mutUnnotarisedHeaders.RUnlock()

	return hdrs
}

// SetBlockBroadcastRound sets the round in which the block with the given nonce has been broadcast
func (sbt *shardBlockTracker) SetBlockBroadcastRound(nonce uint64, round int32) {
	sbt.mutUnnotarisedHeaders.Lock()

	hInfo := sbt.unnotarisedHeaders[nonce]
	if hInfo != nil {
		hInfo.broadcastInRound = round
		sbt.unnotarisedHeaders[nonce] = hInfo
	}

	sbt.mutUnnotarisedHeaders.Unlock()
}

// BlockBroadcastRound gets the round in which the block with given nonce has been broadcast
func (sbt *shardBlockTracker) BlockBroadcastRound(nonce uint64) int32 {
	sbt.mutUnnotarisedHeaders.RLock()
	hInfo := sbt.unnotarisedHeaders[nonce]
	sbt.mutUnnotarisedHeaders.RUnlock()

	if hInfo == nil {
		return 0
	}

	return hInfo.broadcastInRound
}
