package pendingMb

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-logger-go"
)

var _ process.PendingMiniBlocksHandler = (*pendingMiniBlocks)(nil)

var log = logger.GetOrCreate("process/block/pendingMb")

type pendingMiniBlocks struct {
	mutPendingMbShard          sync.RWMutex
	mapPendingMbShard          map[string]uint32
	beforeRevertPendingMbShard map[string]uint32
}

// NewPendingMiniBlocks will create a new pendingMiniBlocks object
func NewPendingMiniBlocks() (*pendingMiniBlocks, error) {
	return &pendingMiniBlocks{
		mapPendingMbShard:          make(map[string]uint32),
		beforeRevertPendingMbShard: make(map[string]uint32),
	}, nil
}

func (p *pendingMiniBlocks) getAllCrossShardMiniBlocksHashes(metaBlock data.MetaHeaderHandler) map[string]uint32 {
	crossShardMiniBlocks := make(map[string]uint32)

	for _, mbHeader := range metaBlock.GetMiniBlockHeaderHandlers() {
		if !shouldConsiderCrossShardMiniBlock(mbHeader.GetSenderShardID(), mbHeader.GetReceiverShardID()) {
			continue
		}

		crossShardMiniBlocks[string(mbHeader.GetHash())] = mbHeader.GetReceiverShardID()
	}

	shardInfoHandlers := metaBlock.GetShardInfoHandlers()
	for _, shardData := range shardInfoHandlers {
		miniblockHandlers := shardData.GetShardMiniBlockHeaderHandlers()
		for _, mbHeader := range miniblockHandlers {
			if !shouldConsiderCrossShardMiniBlock(mbHeader.GetSenderShardID(), mbHeader.GetReceiverShardID()) {
				continue
			}

			crossShardMiniBlocks[string(mbHeader.GetHash())] = mbHeader.GetReceiverShardID()
		}
	}

	return crossShardMiniBlocks
}

func shouldConsiderCrossShardMiniBlock(senderShardID uint32, receiverShardID uint32) bool {
	if senderShardID == receiverShardID {
		return false
	}
	if senderShardID != core.MetachainShardId && receiverShardID == core.MetachainShardId {
		return false
	}
	if senderShardID == core.MetachainShardId && receiverShardID == core.AllShardId {
		return false
	}

	return true
}

// AddProcessedHeader will add in pending list all miniblocks hashes from a given metablock
func (p *pendingMiniBlocks) AddProcessedHeader(headerHandler data.HeaderHandler) error {
	if check.IfNil(headerHandler) {
		return process.ErrNilHeaderHandler
	}
	metaHandler, ok := headerHandler.(data.MetaHeaderHandler)
	if !ok {
		return fmt.Errorf("%w in pendingMiniBlocks.AddProcessedHeader", process.ErrWrongTypeAssertion)
	}

	log.Trace("AddProcessedHeader",
		"shard", metaHandler.GetShardID(),
		"epoch", metaHandler.GetEpoch(),
		"round", metaHandler.GetRound(),
		"nonce", metaHandler.GetNonce(),
		"is start of epoch", metaHandler.IsStartOfEpochBlock(),
	)

	if metaHandler.IsStartOfEpochBlock() {
		return p.processStartOfEpochMeta(metaHandler)
	}

	return p.processHeader(metaHandler)
}

// RevertHeader will remove from pending list all miniblocks hashes from a given metablock
func (p *pendingMiniBlocks) RevertHeader(headerHandler data.HeaderHandler) error {
	if check.IfNil(headerHandler) {
		return process.ErrNilHeaderHandler
	}
	metaHandler, ok := headerHandler.(data.MetaHeaderHandler)
	if !ok {
		return fmt.Errorf("%w in pendingMiniBlocks.RevertHeader", process.ErrWrongTypeAssertion)
	}

	log.Trace("RevertHeader",
		"shard", metaHandler.GetShardID(),
		"epoch", metaHandler.GetEpoch(),
		"round", metaHandler.GetRound(),
		"nonce", metaHandler.GetNonce(),
		"is start of epoch", metaHandler.IsStartOfEpochBlock(),
	)

	if metaHandler.IsStartOfEpochBlock() {
		p.revertStartOfEpochMeta()
		return nil
	}

	return p.processHeader(metaHandler)
}

func (p *pendingMiniBlocks) processHeader(metaHandler data.MetaHeaderHandler) error {
	crossShardMiniBlocksHashes := p.getAllCrossShardMiniBlocksHashes(metaHandler)

	p.mutPendingMbShard.Lock()
	defer p.mutPendingMbShard.Unlock()

	for mbHash, shardID := range crossShardMiniBlocksHashes {
		_, ok := p.mapPendingMbShard[mbHash]
		if !ok {
			p.mapPendingMbShard[mbHash] = shardID
			continue
		}

		delete(p.mapPendingMbShard, mbHash)
	}

	p.displayPendingMb()

	return nil
}

func (p *pendingMiniBlocks) processStartOfEpochMeta(metaHandler data.MetaHeaderHandler) error {
	p.mutPendingMbShard.Lock()
	defer p.mutPendingMbShard.Unlock()

	p.performBackup()
	return p.recreateMap(metaHandler)
}

func (p *pendingMiniBlocks) performBackup() {
	p.beforeRevertPendingMbShard = make(map[string]uint32)
	for hash, shardID := range p.mapPendingMbShard {
		p.beforeRevertPendingMbShard[hash] = shardID
	}
}

func (p *pendingMiniBlocks) recreateMap(metaHandler data.MetaHeaderHandler) error {
	p.mapPendingMbShard = make(map[string]uint32)
	epochStartHandler := metaHandler.GetEpochStartHandler()
	lastFinalizedHeaders := epochStartHandler.GetLastFinalizedHeaderHandlers()
	for _, hdr := range lastFinalizedHeaders {
		pendingMiniBlocksHeaders := hdr.GetPendingMiniBlockHeaderHandlers()
		for _, mbh := range pendingMiniBlocksHeaders {
			if !shouldConsiderCrossShardMiniBlock(mbh.GetSenderShardID(), mbh.GetReceiverShardID()) {
				continue
			}

			p.mapPendingMbShard[string(mbh.GetHash())] = mbh.GetReceiverShardID()
		}
	}

	return nil
}

func (p *pendingMiniBlocks) revertStartOfEpochMeta() {
	p.mutPendingMbShard.Lock()
	defer p.mutPendingMbShard.Unlock()

	p.mapPendingMbShard = make(map[string]uint32)
	for hash, shardID := range p.beforeRevertPendingMbShard {
		p.mapPendingMbShard[hash] = shardID
	}
}

// GetPendingMiniBlocks will return the pending miniblocks hashes for a given shard
func (p *pendingMiniBlocks) GetPendingMiniBlocks(shardID uint32) [][]byte {
	p.mutPendingMbShard.RLock()
	defer p.mutPendingMbShard.RUnlock()

	pendingMiniBlocksToReturn := make([][]byte, 0)
	for mbHash, mbShardID := range p.mapPendingMbShard {
		if mbShardID != shardID {
			continue
		}

		pendingMiniBlocksToReturn = append(pendingMiniBlocksToReturn, []byte(mbHash))
	}

	if len(pendingMiniBlocksToReturn) == 0 {
		return nil
	}

	return pendingMiniBlocksToReturn
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

func (p *pendingMiniBlocks) displayPendingMb() {
	mapShardNumPendingMb := make(map[uint32]int)
	for mbHash, shardID := range p.mapPendingMbShard {
		mapShardNumPendingMb[shardID]++
		log.Trace("pending miniblocks", "shard", shardID, "hash", []byte(mbHash))
	}

	for shardID, num := range mapShardNumPendingMb {
		log.Debug("pending miniblocks", "shard", shardID, "num", num)
	}
}
