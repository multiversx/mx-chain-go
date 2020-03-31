package poolsCleaner

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("process/block/poolsCleaner")

// miniBlocksPoolsCleaner represents a pools cleaner that check and clean miniblocks which should not be in pool anymore
type miniBlocksPoolsCleaner struct {
	blockTracker     BlockTracker
	miniBlocksPool   storage.Cacher
	rounder          process.Rounder
	shardCoordinator sharding.Coordinator

	mutMapMiniBlocksRounds sync.RWMutex
	mapMiniBlocksRounds    map[string]int64
}

// NewMiniBlocksPoolsCleaner will return a new miniblocks pools cleaner
func NewMiniBlocksPoolsCleaner(
	blockTracker BlockTracker,
	miniBlocksPool storage.Cacher,
	rounder process.Rounder,
	shardCoordinator sharding.Coordinator,
) (*miniBlocksPoolsCleaner, error) {

	if check.IfNil(blockTracker) {
		return nil, process.ErrNilBlockTracker
	}
	if check.IfNil(miniBlocksPool) {
		return nil, process.ErrNilCacher
	}
	if check.IfNil(rounder) {
		return nil, process.ErrNilRounder
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	mbpc := miniBlocksPoolsCleaner{
		blockTracker:     blockTracker,
		miniBlocksPool:   miniBlocksPool,
		rounder:          rounder,
		shardCoordinator: shardCoordinator,
	}

	mbpc.mapMiniBlocksRounds = make(map[string]int64)
	mbpc.miniBlocksPool.RegisterHandler(mbpc.receivedMiniBlock)

	return &mbpc, nil
}

func (mbpc *miniBlocksPoolsCleaner) receivedMiniBlock(key []byte) {
	if key == nil {
		return
	}

	log.Trace("miniBlocksPoolsCleaner.receivedMiniBlock", "hash", key)

	mbpc.mutMapMiniBlocksRounds.Lock()
	if _, ok := mbpc.mapMiniBlocksRounds[string(key)]; !ok {
		mbpc.mapMiniBlocksRounds[string(key)] = mbpc.rounder.Index()
	}
	mbpc.cleanPoolIfNeeded()
	mbpc.mutMapMiniBlocksRounds.Unlock()

}

func (mbpc *miniBlocksPoolsCleaner) cleanPoolIfNeeded() {
	selfShard := mbpc.shardCoordinator.SelfId()
	numPendingMiniBlocks := mbpc.blockTracker.GetNumPendingMiniBlocks(selfShard)

	for hash, round := range mbpc.mapMiniBlocksRounds {
		value, ok := mbpc.miniBlocksPool.Get([]byte(hash))
		if !ok {
			log.Trace("miniblock not found in pool",
				"hash", hash,
				"round", round)
			delete(mbpc.mapMiniBlocksRounds, hash)
			continue
		}

		miniBlock, ok := value.(*block.MiniBlock)
		if !ok {
			log.Debug("cleanPoolIfNeeded", "error", process.ErrWrongTypeAssertion,
				"hash", hash,
				"round", round)
			continue
		}

		if miniBlock.SenderShardID != selfShard {
			if numPendingMiniBlocks > 0 {
				log.Trace("cleaning cross miniblock not yet allowed",
					"hash", hash,
					"round", round,
					"type", miniBlock.Type,
					"sender", miniBlock.SenderShardID,
					"receiver", miniBlock.ReceiverShardID,
					"num txs", len(miniBlock.TxHashes))
				continue
			}
		}

		roundDif := mbpc.rounder.Index() - round
		if roundDif <= process.MaxRoundsToKeepUnprocessedMiniBlocks {
			log.Trace("cleaning miniblock not yet allowed",
				"hash", hash,
				"round", round,
				"round dif", roundDif,
				"type", miniBlock.Type,
				"sender", miniBlock.SenderShardID,
				"receiver", miniBlock.ReceiverShardID,
				"num txs", len(miniBlock.TxHashes))
			continue
		}

		mbpc.miniBlocksPool.Remove([]byte(hash))
		delete(mbpc.mapMiniBlocksRounds, hash)

		log.Debug("miniblock has been cleaned",
			"hash", hash,
			"round", round,
			"type", miniBlock.Type,
			"sender", miniBlock.SenderShardID,
			"receiver", miniBlock.ReceiverShardID,
			"num txs", len(miniBlock.TxHashes))
	}
}
