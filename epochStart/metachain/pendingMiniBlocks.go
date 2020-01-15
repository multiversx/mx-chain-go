package metachain

import (
	"bytes"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgsPendingMiniBlocks is a structure that contains components that are used to create a new pendingMiniBlockHeaders object
type ArgsPendingMiniBlocks struct {
	Marshalizer      marshal.Marshalizer
	Storage          storage.Storer
	MetaBlockStorage storage.Storer
	MetaBlockPool    dataRetriever.HeadersPool
}

type pendingMiniBlockHeaders struct {
	marshalizer              marshal.Marshalizer
	metaBlockStorage         storage.Storer
	metaBlockPool            dataRetriever.HeadersPool
	storage                  storage.Storer
	mutPending               sync.RWMutex
	mapMiniBlockHeaders      map[string]block.ShardMiniBlockHeader
	mapShardMiniBlockHeaders map[uint32]uint32
}

// NewPendingMiniBlocks will create a new pendingMiniBlockHeaders object
func NewPendingMiniBlocks(args *ArgsPendingMiniBlocks) (*pendingMiniBlockHeaders, error) {
	if args == nil {
		return nil, epochStart.ErrNilArgsPendingMiniblocks
	}
	if check.IfNil(args.Marshalizer) {
		return nil, epochStart.ErrNilMarshalizer
	}
	if check.IfNil(args.Storage) {
		return nil, epochStart.ErrNilStorage
	}
	if check.IfNil(args.MetaBlockStorage) {
		return nil, epochStart.ErrNilMetaBlockStorage
	}
	if check.IfNil(args.MetaBlockPool) {
		return nil, epochStart.ErrNilMetaBlocksPool
	}

	return &pendingMiniBlockHeaders{
		marshalizer:              args.Marshalizer,
		storage:                  args.Storage,
		mapMiniBlockHeaders:      make(map[string]block.ShardMiniBlockHeader),
		mapShardMiniBlockHeaders: make(map[uint32]uint32),
		metaBlockPool:            args.MetaBlockPool,
		metaBlockStorage:         args.MetaBlockStorage,
	}, nil
}

// PendingMiniBlockHeaders will return a sorted list of ShardMiniBlockHeaders
func (p *pendingMiniBlockHeaders) PendingMiniBlockHeaders(
	lastNotarizedHeaders []data.HeaderHandler,
) ([]block.ShardMiniBlockHeader, error) {
	shardMiniBlockHeaders := make([]block.ShardMiniBlockHeader, 0)

	mapLastUsedMetaBlocks, err := p.getLastUsedMetaBlockFromShardHeaders(lastNotarizedHeaders)
	if err != nil {
		return nil, err
	}

	// make a list map of shardminiblock headers which are in these metablocks
	mapShardMiniBlockHeaders := make(map[string]block.ShardMiniBlockHeader)
	for _, lastMetaHdr := range mapLastUsedMetaBlocks {
		crossShard := p.getAllCrossShardMiniBlocks(lastMetaHdr)
		for key, shardMBHeader := range crossShard {
			mapShardMiniBlockHeaders[key] = shardMBHeader
		}
	}

	// pending miniblocks are only those which are still pending and are from the aforementioned list
	p.mutPending.RLock()
	defer p.mutPending.RUnlock()

	for key, shMbHdr := range p.mapMiniBlockHeaders {
		if _, ok := mapShardMiniBlockHeaders[key]; !ok {
			continue
		}
		shardMiniBlockHeaders = append(shardMiniBlockHeaders, shMbHdr)
	}

	sort.Slice(shardMiniBlockHeaders, func(i, j int) bool {
		return bytes.Compare(shardMiniBlockHeaders[i].Hash, shardMiniBlockHeaders[j].Hash) < 0
	})

	return shardMiniBlockHeaders, nil
}

func (p *pendingMiniBlockHeaders) getAllCrossShardMiniBlocks(metaHdr *block.MetaBlock) map[string]block.ShardMiniBlockHeader {
	crossShard := make(map[string]block.ShardMiniBlockHeader)

	for _, miniBlockHeader := range metaHdr.MiniBlockHeaders {
		if miniBlockHeader.ReceiverShardID != sharding.MetachainShardId {
			continue
		}

		shardMiniBlockHeader := block.ShardMiniBlockHeader{
			Hash:            miniBlockHeader.Hash,
			ReceiverShardID: miniBlockHeader.ReceiverShardID,
			SenderShardID:   miniBlockHeader.SenderShardID,
			TxCount:         miniBlockHeader.TxCount,
		}
		crossShard[string(miniBlockHeader.Hash)] = shardMiniBlockHeader
	}

	for _, shardData := range metaHdr.ShardInfo {
		for _, mbHeader := range shardData.ShardMiniBlockHeaders {
			if mbHeader.SenderShardID == mbHeader.ReceiverShardID {
				continue
			}
			if mbHeader.ReceiverShardID == sharding.MetachainShardId {
				continue
			}

			crossShard[string(mbHeader.Hash)] = mbHeader
		}
	}

	return crossShard
}

func (p *pendingMiniBlockHeaders) getLastUsedMetaBlockFromShardHeaders(
	lastNotarizedHeaders []data.HeaderHandler,
) (map[string]*block.MetaBlock, error) {
	mapLastUsedMetaBlocks := make(map[string]*block.MetaBlock)
	for _, header := range lastNotarizedHeaders {
		shardHdr, ok := header.(*block.Header)
		if !ok {
			return nil, epochStart.ErrWrongTypeAssertion
		}

		numMetas := len(shardHdr.MetaBlockHashes)
		if numMetas == 0 {
			continue
		}

		lastMetaBlockHash := shardHdr.MetaBlockHashes[numMetas-1]
		if _, ok := mapLastUsedMetaBlocks[string(lastMetaBlockHash)]; ok {
			continue
		}

		lastMetaHdr, err := p.getMetaBlockByHash(lastMetaBlockHash)
		if err != nil {
			return nil, err
		}

		mapLastUsedMetaBlocks[string(lastMetaBlockHash)] = lastMetaHdr
	}

	return mapLastUsedMetaBlocks, nil
}

func (p *pendingMiniBlockHeaders) getMetaBlockByHash(metaHash []byte) (*block.MetaBlock, error) {
	peekedData, _ := p.metaBlockPool.GetHeaderByHash(metaHash)
	metaHdr, ok := peekedData.(*block.MetaBlock)
	if ok {
		return metaHdr, nil
	}

	buff, err := p.metaBlockStorage.Get(metaHash)
	if err != nil {
		return nil, err
	}

	var metaHeader block.MetaBlock
	err = p.marshalizer.Unmarshal(&metaHeader, buff)
	if err != nil {
		return nil, err
	}

	return &metaHeader, nil
}

// AddProcessedHeader will add all miniblocks headers in a map
func (p *pendingMiniBlockHeaders) AddProcessedHeader(handler data.HeaderHandler) error {
	if check.IfNil(handler) {
		return epochStart.ErrNilHeaderHandler
	}

	metaHdr, ok := handler.(*block.MetaBlock)
	if !ok {
		return epochStart.ErrWrongTypeAssertion
	}

	crossShard := p.getAllCrossShardMiniBlocks(metaHdr)

	var err error
	p.mutPending.Lock()
	defer func() {
		p.mutPending.Unlock()
		if err != nil {
			_ = p.RevertHeader(handler)
		}
	}()

	for key, mbHeader := range crossShard {
		if _, ok = p.mapMiniBlockHeaders[key]; !ok {
			p.mapMiniBlockHeaders[key] = mbHeader
			p.mapShardMiniBlockHeaders[mbHeader.ReceiverShardID]++
			continue
		}

		delete(p.mapMiniBlockHeaders, key)
		if p.mapShardMiniBlockHeaders[mbHeader.ReceiverShardID] > 0 { // this condition should be always true
			p.mapShardMiniBlockHeaders[mbHeader.ReceiverShardID]--
		}

		var buff []byte
		buff, err = p.marshalizer.Marshal(mbHeader)
		if err != nil {
			return err
		}

		err = p.storage.Put(mbHeader.Hash, buff)
		if err != nil {
			return err
		}
	}

	return nil
}

// RevertHeader will remove all miniblocks headers that are in metablock from pending
func (p *pendingMiniBlockHeaders) RevertHeader(handler data.HeaderHandler) error {
	if check.IfNil(handler) {
		return epochStart.ErrNilHeaderHandler
	}

	metaHdr, ok := handler.(*block.MetaBlock)
	if !ok {
		return epochStart.ErrWrongTypeAssertion
	}

	crossShard := p.getAllCrossShardMiniBlocks(metaHdr)

	p.mutPending.Lock()
	for mbHash, mbHeader := range crossShard {
		if _, ok = p.mapMiniBlockHeaders[mbHash]; ok {
			delete(p.mapMiniBlockHeaders, mbHash)
			if p.mapShardMiniBlockHeaders[mbHeader.ReceiverShardID] > 0 { // this condition should be always true
				p.mapShardMiniBlockHeaders[mbHeader.ReceiverShardID]--
			}
			continue
		}

		_ = p.storage.Remove([]byte(mbHash))
		p.mapMiniBlockHeaders[mbHash] = mbHeader
		p.mapShardMiniBlockHeaders[mbHeader.ReceiverShardID]++
	}
	p.mutPending.Unlock()

	return nil
}

// PendingMiniBlockHeadersForShard will return the number of pending miniblock headers for a given shard
func (p *pendingMiniBlockHeaders) PendingMiniBlockHeadersForShard(shardID uint32) uint32 {
	p.mutPending.RLock()
	nbPendingMiniBlockHeaders := p.mapShardMiniBlockHeaders[shardID]
	p.mutPending.RUnlock()

	return nbPendingMiniBlockHeaders
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *pendingMiniBlockHeaders) IsInterfaceNil() bool {
	return p == nil
}
