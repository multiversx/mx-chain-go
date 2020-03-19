package pendingMb

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/process"
)

var log = logger.GetOrCreate("process/block/pendingMb")

type pendingMiniBlocks struct {
	mutPendingMbShard sync.RWMutex
	mapPendingMbShard map[string]uint32
}

// NewPendingMiniBlocks will create a new pendingMiniBlocks object
func NewPendingMiniBlocks() (*pendingMiniBlocks, error) {
	return &pendingMiniBlocks{
		mapPendingMbShard: make(map[string]uint32),
	}, nil
}

func (p *pendingMiniBlocks) getAllCrossShardMiniBlocksHashes(metaBlock *block.MetaBlock) map[string]uint32 {
	crossShardMiniBlocks := make(map[string]uint32)

	for _, mbHeader := range metaBlock.MiniBlockHeaders {
		if mbHeader.SenderShardID != core.MetachainShardId && mbHeader.ReceiverShardID == core.MetachainShardId {
			continue
		}
		if mbHeader.SenderShardID == core.MetachainShardId && mbHeader.ReceiverShardID == core.AllShardId {
			continue
		}

		crossShardMiniBlocks[string(mbHeader.Hash)] = mbHeader.ReceiverShardID
	}

	for _, shardData := range metaBlock.ShardInfo {
		for _, mbHeader := range shardData.ShardMiniBlockHeaders {
			if mbHeader.SenderShardID == mbHeader.ReceiverShardID {
				continue
			}
			if mbHeader.SenderShardID != core.MetachainShardId && mbHeader.ReceiverShardID == core.MetachainShardId {
				continue
			}

			crossShardMiniBlocks[string(mbHeader.Hash)] = mbHeader.ReceiverShardID
		}
	}

	return crossShardMiniBlocks
}

// AddProcessedHeader will add in pending list all miniblocks hashes from a given metablock
func (p *pendingMiniBlocks) AddProcessedHeader(headerHandler data.HeaderHandler) error {
	if check.IfNil(headerHandler) {
		return process.ErrNilHeaderHandler
	}

	log.Trace("AddProcessedHeader",
		"shard", headerHandler.GetShardID(),
		"epoch", headerHandler.GetEpoch(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce())

	return p.processHeader(headerHandler)
}

// RevertHeader will remove from pending list all miniblocks hashes from a given metablock
func (p *pendingMiniBlocks) RevertHeader(headerHandler data.HeaderHandler) error {
	if check.IfNil(headerHandler) {
		return process.ErrNilHeaderHandler
	}

	log.Trace("RevertHeader",
		"shard", headerHandler.GetShardID(),
		"epoch", headerHandler.GetEpoch(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce())

	return p.processHeader(headerHandler)
}

func (p *pendingMiniBlocks) processHeader(headerHandler data.HeaderHandler) error {
	metaBlock, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	crossShardMiniBlocksHashes := p.getAllCrossShardMiniBlocksHashes(metaBlock)

	p.mutPendingMbShard.Lock()
	defer p.mutPendingMbShard.Unlock()

	for mbHash, shardID := range crossShardMiniBlocksHashes {
		if _, ok = p.mapPendingMbShard[mbHash]; !ok {
			p.mapPendingMbShard[mbHash] = shardID
			continue
		}

		delete(p.mapPendingMbShard, mbHash)
	}

	for shardID, mbHash := range p.mapPendingMbShard {
		log.Debug("pending miniblocks", "shard", shardID, "hash", mbHash)
	}

	return nil
}

// GetPendingMiniBlocks will return the pending miniblocks hashes for a given shard
func (p *pendingMiniBlocks) GetPendingMiniBlocks(shardID uint32) [][]byte {
	p.mutPendingMbShard.RLock()
	defer p.mutPendingMbShard.RUnlock()

	pendingMiniBlocks := make([][]byte, 0)
	for mbHash, mbShardID := range p.mapPendingMbShard {
		if mbShardID != shardID {
			continue
		}

		pendingMiniBlocks = append(pendingMiniBlocks, []byte(mbHash))
	}

	return pendingMiniBlocks
}

// SetPendingMiniBlocks will set the pending miniblocks hashes for a given shard
func (p *pendingMiniBlocks) SetPendingMiniBlocks(shardID uint32, mbHashes [][]byte) {
	p.mutPendingMbShard.Lock()
	defer p.mutPendingMbShard.Unlock()

	for _, mbHash := range mbHashes {
		p.mapPendingMbShard[string(mbHash)] = shardID
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (p *pendingMiniBlocks) IsInterfaceNil() bool {
	return p == nil
}
