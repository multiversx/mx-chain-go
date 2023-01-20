package poolsCleaner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/closing"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("process/block/poolsCleaner")

var _ closing.Closer = (*miniBlocksPoolsCleaner)(nil)

// ArgMiniBlocksPoolsCleaner represents the argument structure used to create a new miniBlocksPoolsCleaner instance
type ArgMiniBlocksPoolsCleaner struct {
	ArgBasePoolsCleaner
	MiniblocksPool storage.Cacher
}

type mbInfo struct {
	round           int64
	senderShardID   uint32
	receiverShardID uint32
	mbType          block.Type
}

// miniBlocksPoolsCleaner represents a pools cleaner that checks and cleans miniblocks which should not be in pool anymore
type miniBlocksPoolsCleaner struct {
	basePoolsCleaner
	miniblocksPool         storage.Cacher
	mutMapMiniBlocksRounds sync.RWMutex
	mapMiniBlocksRounds    map[string]*mbInfo
}

// NewMiniBlocksPoolsCleaner will return a new miniblocks pools cleaner
func NewMiniBlocksPoolsCleaner(args ArgMiniBlocksPoolsCleaner) (*miniBlocksPoolsCleaner, error) {
	err := checkArgMiniBlocksPoolsCleaner(args)
	if err != nil {
		return nil, err
	}

	mbpc := miniBlocksPoolsCleaner{
		basePoolsCleaner:    newBasePoolsCleaner(args.ArgBasePoolsCleaner),
		miniblocksPool:      args.MiniblocksPool,
		mapMiniBlocksRounds: make(map[string]*mbInfo),
	}

	mbpc.miniblocksPool.RegisterHandler(mbpc.receivedMiniBlock, core.UniqueIdentifier())

	return &mbpc, nil
}

func checkArgMiniBlocksPoolsCleaner(args ArgMiniBlocksPoolsCleaner) error {
	err := checkBaseArgs(args.ArgBasePoolsCleaner)
	if err != nil {
		return err
	}
	if check.IfNil(args.MiniblocksPool) {
		return process.ErrNilMiniBlockPool
	}

	return err
}

// StartCleaning actually starts the pools cleaning mechanism
func (mbpc *miniBlocksPoolsCleaner) StartCleaning() {
	mbpc.mut.Lock()
	defer mbpc.mut.Unlock()

	if mbpc.isCleaningRoutineRunning {
		log.Error("miniBlocksPoolsCleaner cleaning routine already started...")
		return
	}

	mbpc.isCleaningRoutineRunning = true
	var ctx context.Context
	ctx, mbpc.cancelFunc = context.WithCancel(context.Background())
	go mbpc.cleanMiniblocksPools(ctx)
}

func (mbpc *miniBlocksPoolsCleaner) cleanMiniblocksPools(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("miniBlocksPoolsCleaner's go routine is stopping...")
			return
		case <-time.After(sleepTime):
		}

		startTime := time.Now()
		numMiniblocksInMap := mbpc.cleanMiniblocksPoolsIfNeeded()
		elapsedTime := time.Since(startTime)

		log.Debug("miniBlocksPoolsCleaner.cleanMiniblocksPools",
			"num miniblocks in map", numMiniblocksInMap,
			"elapsed time", elapsedTime)
	}
}

func (mbpc *miniBlocksPoolsCleaner) receivedMiniBlock(key []byte, value interface{}) {
	if key == nil {
		return
	}

	miniBlock, ok := value.(*block.MiniBlock)
	if !ok {
		log.Warn("miniBlocksPoolsCleaner.receivedMiniBlock",
			"error", process.ErrWrongTypeAssertion,
			"found type", fmt.Sprintf("%T", value),
		)
		return
	}

	log.Trace("miniBlocksPoolsCleaner.receivedMiniBlock", "hash", key)

	mbpc.mutMapMiniBlocksRounds.Lock()
	defer mbpc.mutMapMiniBlocksRounds.Unlock()

	if _, found := mbpc.mapMiniBlocksRounds[string(key)]; !found {
		receivedMbInfo := &mbInfo{
			round:           mbpc.roundHandler.Index(),
			senderShardID:   miniBlock.SenderShardID,
			receiverShardID: miniBlock.ReceiverShardID,
			mbType:          miniBlock.Type,
		}

		mbpc.mapMiniBlocksRounds[string(key)] = receivedMbInfo

		log.Trace("miniblock has been added",
			"hash", key,
			"round", receivedMbInfo.round,
			"sender", receivedMbInfo.senderShardID,
			"receiver", receivedMbInfo.receiverShardID,
			"type", receivedMbInfo.mbType)
	}
}

func (mbpc *miniBlocksPoolsCleaner) cleanMiniblocksPoolsIfNeeded() int {
	numMbsCleaned := 0
	hashesToRemove := make(map[string]struct{})

	mbpc.mutMapMiniBlocksRounds.Lock()
	for hash, mbi := range mbpc.mapMiniBlocksRounds {
		_, ok := mbpc.miniblocksPool.Get([]byte(hash))
		if !ok {
			log.Trace("miniblock not found in pool",
				"hash", []byte(hash),
				"round", mbi.round,
				"sender", mbi.senderShardID,
				"receiver", mbi.receiverShardID,
				"type", mbi.mbType)
			delete(mbpc.mapMiniBlocksRounds, hash)
			continue
		}

		roundDif := mbpc.roundHandler.Index() - mbi.round
		if roundDif <= mbpc.maxRoundsToKeepUnprocessedData {
			log.Trace("cleaning miniblock not yet allowed",
				"hash", []byte(hash),
				"round", mbi.round,
				"round dif", roundDif,
				"sender", mbi.senderShardID,
				"receiver", mbi.receiverShardID,
				"type", mbi.mbType)

			continue
		}

		hashesToRemove[hash] = struct{}{}
		delete(mbpc.mapMiniBlocksRounds, hash)
		numMbsCleaned++

		log.Trace("miniblock has been cleaned",
			"hash", []byte(hash),
			"round", mbi.round,
			"sender", mbi.senderShardID,
			"receiver", mbi.receiverShardID,
			"type", mbi.mbType)
	}

	numMiniBlocksRounds := len(mbpc.mapMiniBlocksRounds)
	mbpc.mutMapMiniBlocksRounds.Unlock()

	startTime := time.Now()
	for hash := range hashesToRemove {
		mbpc.miniblocksPool.Remove([]byte(hash))
	}
	elapsedTime := time.Since(startTime)

	if numMbsCleaned > 0 {
		log.Debug("miniBlocksPoolsCleaner.cleanMiniblocksPoolsIfNeeded",
			"num mbs cleaned", numMbsCleaned,
			"elapsed time to remove mbs from cacher", elapsedTime)
	}

	return numMiniBlocksRounds
}
