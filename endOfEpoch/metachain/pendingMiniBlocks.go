package metachain

import (
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
	Marshalizer marshal.Marshalizer
	Storage     storage.Storer
}

type pendingMiniBlockHeaders struct {
	marshalizer marshal.Marshalizer
	storage     storage.Storer

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

	return &pendingMiniBlockHeaders{
		marshalizer:         args.Marshalizer,
		storage:             args.Storage,
		mapMiniBlockHeaders: make(map[string]block.ShardMiniBlockHeader),
	}, nil
}

//PendingMiniBlockHeaders will return a sorted list of ShardMiniBlockHeaders
func (p *pendingMiniBlockHeaders) PendingMiniBlockHeaders() []block.ShardMiniBlockHeader {
	shardMiniBlockHeaders := make([]block.ShardMiniBlockHeader, 0)

	p.mutPending.Lock()
	defer p.mutPending.Unlock()

	for _, shMbHdr := range p.mapMiniBlockHeaders {
		shardMiniBlockHeaders = append(shardMiniBlockHeaders, shMbHdr)
	}

	sort.Slice(shardMiniBlockHeaders, func(i, j int) bool {
		return shardMiniBlockHeaders[i].TxCount < shardMiniBlockHeaders[j].TxCount
	})

	return shardMiniBlockHeaders
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

	// TODO: uncomment after merge with system vm
	//mapMetaMiniBlockHdrs := make(map[string]*block.MiniBlockHeader)
	//for _, miniBlockHeader := range metaHdr.MiniBlockHeaders {
	//  if miniBlockHeader.ReceiverShardId != sharding.MetachainShardId {
	//  	continue
	//  }
	//
	//	mapMetaMiniBlockHdrs[string(miniBlockHeader.Hash)] = &miniBlockHeader
	//}

	var err error
	p.mutPending.Lock()
	defer func() {
		p.mutPending.Unlock()
		if err != nil {
			_ = p.RevertHeader(handler)
		}
	}()

	for _, shardData := range metaHdr.ShardInfo {
		for _, mbHeader := range shardData.ShardMiniBlockHeaders {
			if mbHeader.SenderShardId == mbHeader.ReceiverShardId {
				continue
			}

			// TODO: uncomment after merge with system vm
			//if _, ok := mapMetaMiniBlockHdrs[string(mbHeader.Hash)]; ok {
			//	continue
			//}

			if _, ok = p.mapMiniBlockHeaders[string(mbHeader.Hash)]; !ok {
				p.mapMiniBlockHeaders[string(mbHeader.Hash)] = mbHeader
				continue
			}

			delete(p.mapMiniBlockHeaders, string(mbHeader.Hash))

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
