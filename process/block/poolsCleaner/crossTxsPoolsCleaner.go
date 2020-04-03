package poolsCleaner

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/txcache"
)

// sleepTime defines the time between each iteration made in clean...Pools methods
const sleepTime = time.Minute

const (
	blockTx = iota
	rewardTx
	unsignedTx
)

type txInfo struct {
	round           int64
	senderShardID   uint32
	receiverShardID uint32
	txType          int8
	txStore         storage.Cacher
}

// crossTxsPoolsCleaner represents a pools cleaner that checks and cleans cross txs which should not be in pool anymore
type crossTxsPoolsCleaner struct {
	blockTracker             BlockTracker
	blockTransactionsPool    dataRetriever.ShardedDataCacherNotifier
	rewardTransactionsPool   dataRetriever.ShardedDataCacherNotifier
	unsignedTransactionsPool dataRetriever.ShardedDataCacherNotifier
	rounder                  process.Rounder
	shardCoordinator         sharding.Coordinator

	mutMapCrossTxsRounds sync.RWMutex
	mapCrossTxsRounds    map[string]*txInfo
}

// NewCrossTxsPoolsCleaner will return a new cross txs pools cleaner
func NewCrossTxsPoolsCleaner(
	blockTracker BlockTracker,
	dataPool dataRetriever.PoolsHolder,
	rounder process.Rounder,
	shardCoordinator sharding.Coordinator,
) (*crossTxsPoolsCleaner, error) {

	if check.IfNil(blockTracker) {
		return nil, process.ErrNilBlockTracker
	}
	if check.IfNil(dataPool) {
		return nil, process.ErrNilPoolsHolder
	}
	if check.IfNil(dataPool.Transactions()) {
		return nil, process.ErrNilTransactionPool
	}
	if check.IfNil(dataPool.RewardTransactions()) {
		return nil, process.ErrNilRewardTxDataPool
	}
	if check.IfNil(dataPool.UnsignedTransactions()) {
		return nil, process.ErrNilUnsignedTxDataPool
	}
	if check.IfNil(rounder) {
		return nil, process.ErrNilRounder
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	ctpc := crossTxsPoolsCleaner{
		blockTracker:             blockTracker,
		blockTransactionsPool:    dataPool.Transactions(),
		rewardTransactionsPool:   dataPool.RewardTransactions(),
		unsignedTransactionsPool: dataPool.UnsignedTransactions(),
		rounder:                  rounder,
		shardCoordinator:         shardCoordinator,
	}

	ctpc.mapCrossTxsRounds = make(map[string]*txInfo)

	ctpc.blockTransactionsPool.RegisterHandler(ctpc.receivedBlockTx)
	ctpc.rewardTransactionsPool.RegisterHandler(ctpc.receivedRewardTx)
	ctpc.unsignedTransactionsPool.RegisterHandler(ctpc.receivedUnsignedTx)

	go ctpc.cleanCrossTxsPools()

	return &ctpc, nil
}

func (ctpc *crossTxsPoolsCleaner) cleanCrossTxsPools() {
	for {
		time.Sleep(sleepTime)
		numCrossTxsInMap := ctpc.cleanCrossTxsPoolsIfNeeded()
		log.Debug("crossTxsPoolsCleaner.cleanCrossTxsPools", "num cross txs in map", numCrossTxsInMap)
	}
}

func (ctpc *crossTxsPoolsCleaner) receivedBlockTx(key []byte, value interface{}) {
	if key == nil {
		return
	}

	log.Trace("crossTxsPoolsCleaner.receivedBlockTx", "hash", key)
	ctpc.processReceivedTx(key, value, blockTx)
}

func (ctpc *crossTxsPoolsCleaner) receivedRewardTx(key []byte, value interface{}) {
	if key == nil {
		return
	}

	log.Trace("crossTxsPoolsCleaner.receivedRewardTx", "hash", key)
	ctpc.processReceivedTx(key, value, rewardTx)
}

func (ctpc *crossTxsPoolsCleaner) receivedUnsignedTx(key []byte, value interface{}) {
	if key == nil {
		return
	}

	log.Trace("crossTxsPoolsCleaner.receivedUnsignedTx", "hash", key)
	ctpc.processReceivedTx(key, value, unsignedTx)
}

func (ctpc *crossTxsPoolsCleaner) processReceivedTx(key []byte, value interface{}, txType int8) {
	wrappedTx, ok := value.(*txcache.WrappedTransaction)
	if !ok {
		log.Warn("crossTxsPoolsCleaner.processReceivedTx", "error", process.ErrWrongTypeAssertion)
		return
	}

	if wrappedTx.SenderShardID == ctpc.shardCoordinator.SelfId() {
		return
	}

	ctpc.mutMapCrossTxsRounds.Lock()
	defer ctpc.mutMapCrossTxsRounds.Unlock()

	if _, ok := ctpc.mapCrossTxsRounds[string(key)]; !ok {
		transactionPool := ctpc.getTransactionPool(txType)
		if transactionPool == nil {
			return
		}

		strCache := process.ShardCacherIdentifier(wrappedTx.SenderShardID, wrappedTx.ReceiverShardID)
		txStore := transactionPool.ShardDataStore(strCache)
		if txStore == nil {
			return
		}

		crossTxInfo := &txInfo{
			round:           ctpc.rounder.Index(),
			senderShardID:   wrappedTx.SenderShardID,
			receiverShardID: wrappedTx.ReceiverShardID,
			txType:          txType,
			txStore:         txStore,
		}

		ctpc.mapCrossTxsRounds[string(key)] = crossTxInfo

		log.Trace("transaction has been added",
			"hash", key,
			"round", crossTxInfo.round,
			"sender", crossTxInfo.senderShardID,
			"receiver", crossTxInfo.receiverShardID,
			"type", getTxTypeName(crossTxInfo.txType))
	}
}

func (ctpc *crossTxsPoolsCleaner) cleanCrossTxsPoolsIfNeeded() int {
	ctpc.mutMapCrossTxsRounds.Lock()
	defer ctpc.mutMapCrossTxsRounds.Unlock()

	selfShardID := ctpc.shardCoordinator.SelfId()
	numPendingMiniBlocks := ctpc.blockTracker.GetNumPendingMiniBlocks(selfShardID)
	numTxsCleaned := 0

	for hash, crossTxInfo := range ctpc.mapCrossTxsRounds {
		percentUsed := float64(crossTxInfo.txStore.Len()) / float64(crossTxInfo.txStore.MaxSize())
		if numPendingMiniBlocks > 0 && percentUsed < percentAllowed {
			log.Trace("cleaning cross transaction not yet allowed",
				"hash", []byte(hash),
				"round", crossTxInfo.round,
				"sender", crossTxInfo.senderShardID,
				"receiver", crossTxInfo.receiverShardID,
				"type", getTxTypeName(crossTxInfo.txType),
				"num pending miniblocks", numPendingMiniBlocks,
				"transactions pool percent used", percentUsed)

			continue
		}

		roundDif := ctpc.rounder.Index() - crossTxInfo.round
		if roundDif <= process.MaxRoundsToKeepUnprocessedTransactions {
			log.Trace("cleaning cross transaction not yet allowed",
				"hash", []byte(hash),
				"round", crossTxInfo.round,
				"sender", crossTxInfo.senderShardID,
				"receiver", crossTxInfo.receiverShardID,
				"type", getTxTypeName(crossTxInfo.txType),
				"round dif", roundDif)

			continue
		}

		crossTxInfo.txStore.Remove([]byte(hash))
		delete(ctpc.mapCrossTxsRounds, hash)
		numTxsCleaned++

		log.Trace("transaction has been cleaned",
			"hash", []byte(hash),
			"round", crossTxInfo.round,
			"sender", crossTxInfo.senderShardID,
			"receiver", crossTxInfo.receiverShardID,
			"type", getTxTypeName(crossTxInfo.txType))
	}

	if numTxsCleaned > 0 {
		log.Debug("crossTxsPoolsCleaner.cleanCrossTxsPoolsIfNeeded", "num txs cleaned", numTxsCleaned)
	}

	return len(ctpc.mapCrossTxsRounds)
}

func (ctpc *crossTxsPoolsCleaner) getTransactionPool(txType int8) dataRetriever.ShardedDataCacherNotifier {
	switch txType {
	case blockTx:
		return ctpc.blockTransactionsPool
	case rewardTx:
		return ctpc.rewardTransactionsPool
	case unsignedTx:
		return ctpc.unsignedTransactionsPool
	}

	return nil
}

func getTxTypeName(txType int8) string {
	switch txType {
	case blockTx:
		return "blockTx"
	case rewardTx:
		return "rewardTx"
	case unsignedTx:
		return "unsignedTx"
	}

	return "unknownTx"
}
