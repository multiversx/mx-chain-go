package pendingMb

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/logger"
)

var log = logger.GetOrCreate("process/block/pendingMb")

type pendingMiniBlocksHeaders struct {
	mutPending            sync.RWMutex
	//TODO: Also this map should be saved in BootStorer
	mapMiniBlockHeaders   map[string]block.ShardMiniBlockHeader
	mapShardNumMiniBlocks map[uint32]uint32
}

// NewPendingMiniBlocks will create a new pendingMiniBlocksHeaders object
func NewPendingMiniBlocks() (*pendingMiniBlocksHeaders, error) {
	return &pendingMiniBlocksHeaders{
		mapMiniBlockHeaders:   make(map[string]block.ShardMiniBlockHeader),
		mapShardNumMiniBlocks: make(map[uint32]uint32),
	}, nil
}

func (p *pendingMiniBlocksHeaders) getAllCrossShardMiniBlocks(metaBlock *block.MetaBlock) map[string]block.ShardMiniBlockHeader {
	crossShardMiniBlocks := make(map[string]block.ShardMiniBlockHeader)

	for _, mbHeader := range metaBlock.MiniBlockHeaders {
		if mbHeader.SenderShardID != core.MetachainShardId && mbHeader.ReceiverShardID == core.MetachainShardId {
			continue
		}
		//TODO: Activate the following code when EN-5464-peer-rating-process-in-shard will be merged in dev
		//if mbHeader.SenderShardID == core.MetachainShardId && mbHeader.ReceiverShardID == core.AllShardId {
		//	continue
		//}

		shardMiniBlockHeader := block.ShardMiniBlockHeader{
			Hash:            mbHeader.Hash,
			ReceiverShardID: mbHeader.ReceiverShardID,
			SenderShardID:   mbHeader.SenderShardID,
			TxCount:         mbHeader.TxCount,
		}
		crossShardMiniBlocks[string(mbHeader.Hash)] = shardMiniBlockHeader
	}

	for _, shardData := range metaBlock.ShardInfo {
		for _, mbHeader := range shardData.ShardMiniBlockHeaders {
			if mbHeader.SenderShardID == mbHeader.ReceiverShardID {
				continue
			}
			if mbHeader.SenderShardID != core.MetachainShardId && mbHeader.ReceiverShardID == core.MetachainShardId {
				continue
			}

			crossShardMiniBlocks[string(mbHeader.Hash)] = mbHeader
		}
	}

	return crossShardMiniBlocks
}

// AddProcessedHeader will add all miniblocks headers in a map
func (p *pendingMiniBlocksHeaders) AddProcessedHeader(headerHandler data.HeaderHandler) error {
	if check.IfNil(headerHandler) {
		return epochStart.ErrNilHeaderHandler
	}

	metaBlock, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return epochStart.ErrWrongTypeAssertion
	}

	crossShardMiniBlocks := p.getAllCrossShardMiniBlocks(metaBlock)

	p.mutPending.Lock()
	for mbHash, mbHeader := range crossShardMiniBlocks {
		if _, ok = p.mapMiniBlockHeaders[mbHash]; !ok {
			p.mapMiniBlockHeaders[mbHash] = mbHeader
			p.mapShardNumMiniBlocks[mbHeader.ReceiverShardID]++
			continue
		}

		delete(p.mapMiniBlockHeaders, mbHash)
		if p.mapShardNumMiniBlocks[mbHeader.ReceiverShardID] == 0 {
			log.Error("AddProcessedHeader: trying to remove an unpending miniblock",
				"epoch", metaBlock.Epoch,
				"round", metaBlock.Round,
				"nonce", metaBlock.Nonce,
				"mb hash", mbHash,
				"mb sender shard", mbHeader.SenderShardID,
				"mb receiver shard", mbHeader.ReceiverShardID)
			continue
		}
		p.mapShardNumMiniBlocks[mbHeader.ReceiverShardID]--
	}

	for shardID, numPendingMiniBlocks := range p.mapShardNumMiniBlocks {
		log.Trace("pending miniblocks", "shard", shardID, "num", numPendingMiniBlocks)
	}

	p.mutPending.Unlock()

	return nil
}

// RevertHeader will remove all miniblocks headers that are in metablock from pending
func (p *pendingMiniBlocksHeaders) RevertHeader(headerHandler data.HeaderHandler) error {
	if check.IfNil(headerHandler) {
		return epochStart.ErrNilHeaderHandler
	}

	metaBlock, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return epochStart.ErrWrongTypeAssertion
	}

	crossShardMiniBlocks := p.getAllCrossShardMiniBlocks(metaBlock)

	p.mutPending.Lock()
	for mbHash, mbHeader := range crossShardMiniBlocks {
		if _, ok = p.mapMiniBlockHeaders[mbHash]; ok {
			delete(p.mapMiniBlockHeaders, mbHash)
			if p.mapShardNumMiniBlocks[mbHeader.ReceiverShardID] == 0 {
				log.Error("RevertHeader: trying to remove an unpending miniblock",
					"epoch", metaBlock.Epoch,
					"round", metaBlock.Round,
					"nonce", metaBlock.Nonce,
					"mb hash", mbHash,
					"mb sender shard", mbHeader.SenderShardID,
					"mb receiver shard", mbHeader.ReceiverShardID)
				continue
			}

			p.mapShardNumMiniBlocks[mbHeader.ReceiverShardID]--
			continue
		}

		p.mapMiniBlockHeaders[mbHash] = mbHeader
		p.mapShardNumMiniBlocks[mbHeader.ReceiverShardID]++
	}
	p.mutPending.Unlock()

	return nil
}

// GetNumPendingMiniBlocks will return the number of pending miniblocks for a given shard
func (p *pendingMiniBlocksHeaders) GetNumPendingMiniBlocks(shardID uint32) uint32 {
	p.mutPending.RLock()
	numPendingMiniBlocks := p.mapShardNumMiniBlocks[shardID]
	p.mutPending.RUnlock()

	return numPendingMiniBlocks
}

// SetNumPendingMiniBlocks will set the number of pending miniblocks for a given shard
func (p *pendingMiniBlocksHeaders) SetNumPendingMiniBlocks(shardID uint32, numPendingMiniBlocks uint32) {
	p.mutPending.Lock()
	p.mapShardNumMiniBlocks[shardID] = numPendingMiniBlocks
	p.mutPending.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *pendingMiniBlocksHeaders) IsInterfaceNil() bool {
	return p == nil
}
