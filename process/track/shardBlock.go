package track

import (
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

var log = logger.DefaultLogger()

type headerInfo struct {
	header           data.HeaderHandler
	broadcastInRound int32
}

// shardBlock implements NotarisedBlocksTracker interface which tracks notarised blocks
type shardBlock struct {
	dataPool         dataRetriever.PoolsHolder
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator
	store            dataRetriever.StorageService

	mutUnnotarisedHeaders sync.RWMutex
	unnotarisedHeaders    map[uint64]*headerInfo
}

// NewShardBlock creates a new shardBlock object
func NewShardBlock(
	dataPool dataRetriever.PoolsHolder,
	marshalizer marshal.Marshalizer,
	shardCoordinator sharding.Coordinator,
	store dataRetriever.StorageService,
) (*shardBlock, error) {

	err := checkTrackerNilParameters(
		dataPool,
		marshalizer,
		shardCoordinator,
		store)
	if err != nil {
		return nil, err
	}

	sb := shardBlock{
		dataPool:         dataPool,
		marshalizer:      marshalizer,
		shardCoordinator: shardCoordinator,
		store:            store,
	}

	sb.unnotarisedHeaders = make(map[uint64]*headerInfo)

	return &sb, nil
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
func (sb *shardBlock) AddBlock(headerHandler data.HeaderHandler) {
	sb.mutUnnotarisedHeaders.Lock()
	sb.unnotarisedHeaders[headerHandler.GetNonce()] = &headerInfo{header: headerHandler, broadcastInRound: 0}
	sb.mutUnnotarisedHeaders.Unlock()
}

// RemoveNotarisedBlocks removes all the blocks which already have been notarised
func (sb *shardBlock) RemoveNotarisedBlocks(headerHandler data.HeaderHandler) {
	metaBlock, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		log.Debug(fmt.Sprintf("wrong type assertion: expected MetaBlock\n"))
		return
	}

	for _, shardData := range metaBlock.ShardInfo {
		if shardData.ShardId != sb.shardCoordinator.SelfId() {
			continue
		}

		header := sb.getHeader(shardData.HeaderHash)
		if header == nil {
			continue
		}

		log.Info(fmt.Sprintf("shardBlock with nonce %d and hash %s has been notarised by metachain\n",
			header.GetNonce(),
			process.ToB64(shardData.HeaderHash)))

		sb.mutUnnotarisedHeaders.Lock()
		delete(sb.unnotarisedHeaders, header.Nonce)
		sb.mutUnnotarisedHeaders.Unlock()
	}
}

// UnnotarisedBlocks gets all the blocks which are not notarised yet
func (sb *shardBlock) UnnotarisedBlocks() []data.HeaderHandler {
	sb.mutUnnotarisedHeaders.RLock()

	hdrs := make([]data.HeaderHandler, 0)
	for _, hInfo := range sb.unnotarisedHeaders {
		hdrs = append(hdrs, hInfo.header)
	}

	sb.mutUnnotarisedHeaders.RUnlock()

	return hdrs
}

// SetBlockBroadcastRound sets the round in which the block with the given nonce has been broadcast
func (sb *shardBlock) SetBlockBroadcastRound(nonce uint64, round int32) {
	sb.mutUnnotarisedHeaders.Lock()

	hInfo := sb.unnotarisedHeaders[nonce]
	if hInfo != nil {
		sb.unnotarisedHeaders[nonce] = &headerInfo{header: hInfo.header, broadcastInRound: round}
	}

	sb.mutUnnotarisedHeaders.Unlock()
}

// BlockBroadcastRound gets the round in which the block with given nonce has been broadcast
func (sb *shardBlock) BlockBroadcastRound(nonce uint64) int32 {
	sb.mutUnnotarisedHeaders.RLock()
	hInfo := sb.unnotarisedHeaders[nonce]
	sb.mutUnnotarisedHeaders.RUnlock()

	if hInfo == nil {
		return 0
	}

	return hInfo.broadcastInRound
}

func (sb *shardBlock) getHeader(hash []byte) *block.Header {
	hdr := sb.getHeaderFromPool(hash)
	if hdr != nil {
		return hdr
	}

	return sb.getHeaderFromStorage(hash)
}

func (sb *shardBlock) getHeaderFromPool(hash []byte) *block.Header {
	hdr, ok := sb.dataPool.Headers().Peek(hash)
	if !ok {
		log.Debug(fmt.Sprintf("header with hash %s not found in headers cache\n", process.ToB64(hash)))
		return nil
	}

	header, ok := hdr.(*block.Header)
	if !ok {
		log.Debug(fmt.Sprintf("data with hash %s is not header\n", process.ToB64(hash)))
		return nil
	}

	return header
}

func (sb *shardBlock) getHeaderFromStorage(hash []byte) *block.Header {
	headerStore := sb.store.GetStorer(dataRetriever.BlockHeaderUnit)

	if headerStore == nil {
		log.Error(process.ErrNilHeadersStorage.Error())
		return nil
	}

	buffHeader, err := headerStore.Get(hash)
	if err != nil {
		log.Debug(err.Error())
		return nil
	}

	header := &block.Header{}
	err = sb.marshalizer.Unmarshal(header, buffHeader)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return header
}
