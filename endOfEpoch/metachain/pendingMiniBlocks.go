package metachain

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/endOfEpoch"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"sort"
	"sync"
)

type ArgsPendingMiniBlocks struct {
	Marshalizer marshal.Marshalizer
	Storage     storage.Storer
}

type pendingMiniBlockHeaders struct {
	marshalizer marshal.Marshalizer
	storage     storage.Storer

	mutPending          sync.Mutex
	mapMiniBlockHeaders map[string]block.ShardMiniBlockHeader
}

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

	return &pendingMiniBlockHeaders{}, nil
}

func (p *pendingMiniBlockHeaders) PendingMiniBlockHeaders() []block.ShardMiniBlockHeader {
	shardMiniBlokcHeaders := make([]block.ShardMiniBlockHeader, 0)

	p.mutPending.Lock()
	defer p.mutPending.Unlock()

	for _, shMbHdr := range p.mapMiniBlockHeaders {
		shardMiniBlokcHeaders = append(shardMiniBlokcHeaders, shMbHdr)
	}

	sort.Slice(shardMiniBlokcHeaders, func(i, j int) bool {
		return shardMiniBlokcHeaders[i].TxCount < shardMiniBlokcHeaders[j].TxCount
	})

	return shardMiniBlokcHeaders
}

func (p *pendingMiniBlockHeaders) AddCommittedHeader(handler data.HeaderHandler) error {
	metaHdr, ok := handler.(*block.MetaBlock)
	if !ok {
		return endOfEpoch.ErrWongTypeAssertion
	}

	// TODO: uncomment after merge with system vm
	//mapMetaMiniBlockHdrs := make(map[string]*block.MiniBlockHeader)
	//for _, miniBlockHeader := range metaHdr.MiniBlockHeaders {
	//  if miniBlockHeader.ReceiverShardId != sharding.MetachainShardId {
	//  	continue
	//  }
	//
	//	mapMetaMiniBlockHdrs[string(miniBlockHeader.Hash)] = &miniBlockHeader
	//}

	p.mutPending.Lock()
	defer p.mutPending.Unlock()

	for _, shardData := range metaHdr.ShardInfo {
		for _, mbHeader := range shardData.ShardMiniBlockHeaders {
			if mbHeader.SenderShardId == mbHeader.ReceiverShardId {
				continue
			}

			// TODO: uncomment after merge with system vm
			//if _, ok := mapMetaMiniBlockHdrs[string(mbHeader.Hash)]; ok {
			//	continue
			//}

			if _, ok := p.mapMiniBlockHeaders[string(mbHeader.Hash)]; !ok {
				p.mapMiniBlockHeaders[string(mbHeader.Hash)] = mbHeader
				continue
			}

			delete(p.mapMiniBlockHeaders, string(mbHeader.Hash))

			buff, err := p.marshalizer.Marshal(mbHeader)
			if err != nil {
				return err
			}

			err = p.storage.Put(mbHeader.Hash, buff)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *pendingMiniBlockHeaders) RevertHeader(handler data.HeaderHandler) error {
	metaHdr, ok := handler.(*block.MetaBlock)
	if !ok {
		return endOfEpoch.ErrWongTypeAssertion
	}

	// TODO: uncomment after merge with system vm
	//mapMetaMiniBlockHdrs := make(map[string]*block.MiniBlockHeader)
	//for _, miniBlockHeader := range metaHdr.MiniBlockHeaders {
	//  if miniBlockHeader.ReceiverShardId != sharding.MetachainShardId {
	//  	continue
	//  }
	//
	//	mapMetaMiniBlockHdrs[string(miniBlockHeader.Hash)] = &miniBlockHeader
	//}

	p.mutPending.Lock()
	defer p.mutPending.Unlock()

	for _, shardData := range metaHdr.ShardInfo {
		for _, mbHeader := range shardData.ShardMiniBlockHeaders {
			if mbHeader.SenderShardId == mbHeader.ReceiverShardId {
				continue
			}

			// TODO: uncomment after merge with system vm
			//if _, ok := mapMetaMiniBlockHdrs[string(mbHeader.Hash)]; ok {
			//	continue
			//}

			if _, ok := p.mapMiniBlockHeaders[string(mbHeader.Hash)]; ok {
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

func (p *pendingMiniBlockHeaders) IsInterfaceNil() bool {
	return p == nil
}
