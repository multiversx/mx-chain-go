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

const percentAllowed = 0.8

// miniBlocksPoolsCleaner represents a pools cleaner that checks and cleans miniblocks which should not be in pool anymore
type miniBlocksPoolsCleaner struct {
	blockTracker     BlockTracker
	miniblocksPool   storage.Cacher
	rounder          process.Rounder
	shardCoordinator sharding.Coordinator

	mutMapMiniBlocksRounds sync.RWMutex
	mapMiniBlocksRounds    map[string]int64
}

// NewMiniBlocksPoolsCleaner will return a new miniblocks pools cleaner
func NewMiniBlocksPoolsCleaner(
	blockTracker BlockTracker,
	miniblocksPool storage.Cacher,
	rounder process.Rounder,
	shardCoordinator sharding.Coordinator,
) (*miniBlocksPoolsCleaner, error) {

	if check.IfNil(blockTracker) {
		return nil, process.ErrNilBlockTracker
	}
	if check.IfNil(miniblocksPool) {
		return nil, process.ErrNilMiniBlockPool
	}
	if check.IfNil(rounder) {
		return nil, process.ErrNilRounder
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	mbpc := miniBlocksPoolsCleaner{
		blockTracker:     blockTracker,
		miniblocksPool:   miniblocksPool,
		rounder:          rounder,
		shardCoordinator: shardCoordinator,
	}

	mbpc.mapMiniBlocksRounds = make(map[string]int64)
	mbpc.miniblocksPool.RegisterHandler(mbpc.receivedMiniBlock)

	return &mbpc, nil
}

func (mbpc *miniBlocksPoolsCleaner) receivedMiniBlock(key []byte) {
	if key == nil {
		return
	}

	log.Trace("miniBlocksPoolsCleaner.receivedMiniBlock", "hash", key)

	mbpc.mutMapMiniBlocksRounds.Lock()
	defer mbpc.mutMapMiniBlocksRounds.Unlock()

	if _, ok := mbpc.mapMiniBlocksRounds[string(key)]; !ok {
		mbpc.mapMiniBlocksRounds[string(key)] = mbpc.rounder.Index()

		log.Trace("miniblock has been added",
			"hash", key,
			"round", mbpc.rounder.Index())
	}

	mbpc.cleanMiniblocksPoolsIfNeeded()
}

func (mbpc *miniBlocksPoolsCleaner) cleanMiniblocksPoolsIfNeeded() {
	selfShardID := mbpc.shardCoordinator.SelfId()
	numPendingMiniBlocks := mbpc.blockTracker.GetNumPendingMiniBlocks(selfShardID)
	percentUsed := float64(mbpc.miniblocksPool.Len()) / float64(mbpc.miniblocksPool.MaxSize())
	numMbsCleaned := 0

	for hash, round := range mbpc.mapMiniBlocksRounds {
		value, ok := mbpc.miniblocksPool.Get([]byte(hash))
		if !ok {
			log.Trace("miniblock not found in pool",
				"hash", []byte(hash),
				"round", round)
			delete(mbpc.mapMiniBlocksRounds, hash)
			continue
		}

		miniBlock, ok := value.(*block.MiniBlock)
		if !ok {
			log.Debug("cleanMiniblocksPoolsIfNeeded", "error", process.ErrWrongTypeAssertion,
				"hash", []byte(hash),
				"round", round)
			continue
		}

		if miniBlock.SenderShardID != selfShardID {
			if numPendingMiniBlocks > 0 && percentUsed < percentAllowed {
				log.Trace("cleaning cross miniblock not yet allowed",
					"hash", []byte(hash),
					"round", round,
					"num pending miniblocks", numPendingMiniBlocks,
					"miniblocks pool percent used", percentUsed,
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
				"hash", []byte(hash),
				"round", round,
				"round dif", roundDif,
				"type", miniBlock.Type,
				"sender", miniBlock.SenderShardID,
				"receiver", miniBlock.ReceiverShardID,
				"num txs", len(miniBlock.TxHashes))
			continue
		}

		mbpc.miniblocksPool.Remove([]byte(hash))
		delete(mbpc.mapMiniBlocksRounds, hash)
		numMbsCleaned++

		log.Trace("miniblock has been cleaned",
			"hash", []byte(hash),
			"round", round,
			"type", miniBlock.Type,
			"sender", miniBlock.SenderShardID,
			"receiver", miniBlock.ReceiverShardID,
			"num txs", len(miniBlock.TxHashes))
	}

	if numMbsCleaned > 0 {
		log.Debug("miniBlocksPoolsCleaner.cleanMiniblocksPoolsIfNeeded",
			"num mbs cleaned", numMbsCleaned)
	}
}
