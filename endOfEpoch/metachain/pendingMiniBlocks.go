package metachain

import (
	"github.com/ElrondNetwork/elrond-go/sharding"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/endOfEpoch"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

//ArgsPendingMiniBlocks is structure that contain components that are used to create a new pendingMiniBlockHeaders object
type ArgsPendingMiniBlocks struct {
	Marshalizer      marshal.Marshalizer
	Storage          storage.Storer
	MetaBlockStorage storage.Storer
	MetaBlockPool    storage.Cacher
}

type pendingMiniBlockHeaders struct {
	marshalizer      marshal.Marshalizer
	metaBlockStorage storage.Storer
	metaBlockPool    storage.Cacher

	storage             storage.Storer
	mutPending          sync.Mutex
	mapMiniBlockHeaders map[string]block.ShardMiniBlockHeader
}

// NewPendingMiniBlocks will create a new pendingMiniBlockHeaders object
func NewPendingMiniBlocks(args *ArgsPendingMiniBlocks) (*pendingMiniBlockHeaders, error) {
	if args == nil {
		return nil, endOfEpoch.ErrNilArgsPendingMiniblocks
	}
	if check.IfNil(args.Marshalizer) {
		return nil, endOfEpoch.ErrNilMarshalizer
	}
	if check.IfNil(args.Storage) {
		return nil, endOfEpoch.ErrNilStorage
	}
	if check.IfNil(args.MetaBlockStorage) {
		return nil, endOfEpoch.ErrNilMetaBlockStorage
	}
	if check.IfNil(args.MetaBlockPool) {
		return nil, endOfEpoch.ErrNilMetaBlocksPool
	}

	return &pendingMiniBlockHeaders{
		marshalizer:         args.Marshalizer,
		storage:             args.Storage,
		mapMiniBlockHeaders: make(map[string]block.ShardMiniBlockHeader),
		metaBlockPool:       args.MetaBlockPool,
		metaBlockStorage:    args.MetaBlockStorage,
	}, nil
}

//PendingMiniBlockHeaders will return a sorted list of ShardMiniBlockHeaders
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

	// pending miniblocks are only those which are still pending and ar from the aforementioned list
	p.mutPending.Lock()
	defer p.mutPending.Unlock()

	for key, shMbHdr := range p.mapMiniBlockHeaders {
		if _, ok := mapShardMiniBlockHeaders[key]; !ok {
			continue
		}
		shardMiniBlockHeaders = append(shardMiniBlockHeaders, shMbHdr)
	}

	sort.Slice(shardMiniBlockHeaders, func(i, j int) bool {
		return shardMiniBlockHeaders[i].TxCount < shardMiniBlockHeaders[j].TxCount
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
			return nil, endOfEpoch.ErrWrongTypeAssertion
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
	peekedData, _ := p.metaBlockPool.Peek(metaHash)
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
		return endOfEpoch.ErrNilHeaderHandler
	}

	metaHdr, ok := handler.(*block.MetaBlock)
	if !ok {
		return endOfEpoch.ErrWrongTypeAssertion
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
			continue
		}

		delete(p.mapMiniBlockHeaders, key)

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

// RevertHeader will remove  all minibloks headers that are in metablock from pending
func (p *pendingMiniBlockHeaders) RevertHeader(handler data.HeaderHandler) error {
	if check.IfNil(handler) {
		return endOfEpoch.ErrNilHeaderHandler
	}

	metaHdr, ok := handler.(*block.MetaBlock)
	if !ok {
		return endOfEpoch.ErrWrongTypeAssertion
	}

	mapMetaMiniBlockHdrs := make(map[string]*block.MiniBlockHeader)
	for _, miniBlockHeader := range metaHdr.MiniBlockHeaders {
		if miniBlockHeader.ReceiverShardID != sharding.MetachainShardId {
			continue
		}

		mapMetaMiniBlockHdrs[string(miniBlockHeader.Hash)] = &miniBlockHeader
	}

	p.mutPending.Lock()
	defer p.mutPending.Unlock()

	for _, shardData := range metaHdr.ShardInfo {
		for _, mbHeader := range shardData.ShardMiniBlockHeaders {
			if mbHeader.SenderShardID == mbHeader.ReceiverShardID {
				continue
			}

			if _, ok := mapMetaMiniBlockHdrs[string(mbHeader.Hash)]; ok {
				continue
			}

			if _, ok = p.mapMiniBlockHeaders[string(mbHeader.Hash)]; ok {
				delete(p.mapMiniBlockHeaders, string(mbHeader.Hash))
				continue
			}

			err := p.storage.Remove(mbHeader.Hash)
			if err != nil {
				return err
			}

			p.mapMiniBlockHeaders[string(mbHeader.Hash)] = mbHeader
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *pendingMiniBlockHeaders) IsInterfaceNil() bool {
	return p == nil
}
