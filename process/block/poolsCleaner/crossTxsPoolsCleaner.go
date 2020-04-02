package poolsCleaner

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const (
	blockTx = iota
	rewardTx
	unsignedTx
)

type crossTxInfo struct {
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
	mapCrossTxsRounds    map[string]*crossTxInfo
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

	ctpc.mapCrossTxsRounds = make(map[string]*crossTxInfo)

	ctpc.blockTransactionsPool.RegisterHandler(ctpc.receivedBlockTx)
	ctpc.rewardTransactionsPool.RegisterHandler(ctpc.receivedRewardTx)
	ctpc.unsignedTransactionsPool.RegisterHandler(ctpc.receivedUnsignedTx)

	return &ctpc, nil
}

func (ctpc *crossTxsPoolsCleaner) receivedBlockTx(key []byte) {
	if key == nil {
		return
	}

	log.Trace("crossTxsPoolsCleaner.receivedBlockTx", "hash", key)
	ctpc.processReceivedTx(key, blockTx)
}

func (ctpc *crossTxsPoolsCleaner) receivedRewardTx(key []byte) {
	if key == nil {
		return
	}

	log.Trace("crossTxsPoolsCleaner.receivedRewardTx", "hash", key)
	ctpc.processReceivedTx(key, rewardTx)
}

func (ctpc *crossTxsPoolsCleaner) receivedUnsignedTx(key []byte) {
	if key == nil {
		return
	}

	log.Trace("crossTxsPoolsCleaner.receivedUnsignedTx", "hash", key)
	ctpc.processReceivedTx(key, unsignedTx)
}

func (ctpc *crossTxsPoolsCleaner) processReceivedTx(key []byte, txType int8) {
	isCrossTx := false
	var crossTxInfo *crossTxInfo
	for shardID := uint32(0); shardID < ctpc.shardCoordinator.NumberOfShards(); shardID++ {
		isCrossTx, crossTxInfo = ctpc.isCrossTxFromShard(key, shardID, txType)
		if isCrossTx {
			break
		}
	}

	if !isCrossTx {
		isCrossTx, crossTxInfo = ctpc.isCrossTxFromShard(key, core.MetachainShardId, txType)
		if !isCrossTx {
			return
		}
	}

	ctpc.mutMapCrossTxsRounds.Lock()
	defer ctpc.mutMapCrossTxsRounds.Unlock()

	if _, ok := ctpc.mapCrossTxsRounds[string(key)]; !ok {
		crossTxInfo.round = ctpc.rounder.Index()
		crossTxInfo.txType = txType
		ctpc.mapCrossTxsRounds[string(key)] = crossTxInfo

		log.Trace("transaction has been added",
			"hash", key,
			"round", crossTxInfo.round,
			"sender shard", crossTxInfo.senderShardID,
			"receiver shard", crossTxInfo.receiverShardID,
			"type", getTxTypeName(crossTxInfo.txType))
	}

	ctpc.cleanCrossTxsPoolsIfNeeded()
}

func (ctpc *crossTxsPoolsCleaner) isCrossTxFromShard(txHash []byte, shardID uint32, txType int8) (bool, *crossTxInfo) {
	transactionPool := ctpc.getTransactionPool(txType)
	if transactionPool == nil {
		return false, nil
	}

	selfShardID := ctpc.shardCoordinator.SelfId()
	if shardID == selfShardID {
		return false, nil
	}

	strCache := process.ShardCacherIdentifier(shardID, selfShardID)
	txStore := transactionPool.ShardDataStore(strCache)
	if txStore == nil {
		return false, nil
	}

	_, ok := txStore.Peek(txHash)
	if !ok {
		return false, nil
	}

	return true, &crossTxInfo{txStore: txStore, senderShardID: shardID, receiverShardID: selfShardID}
}

func (ctpc *crossTxsPoolsCleaner) cleanCrossTxsPoolsIfNeeded() {
	selfShardID := ctpc.shardCoordinator.SelfId()
	numPendingMiniBlocks := ctpc.blockTracker.GetNumPendingMiniBlocks(selfShardID)
	numTxsCleaned := 0

	for hash, crossTxInfo := range ctpc.mapCrossTxsRounds {
		value, ok := crossTxInfo.txStore.Get([]byte(hash))
		if !ok {
			log.Trace("transaction not found in pool",
				"hash", []byte(hash),
				"round", crossTxInfo.round,
				"sender shard", crossTxInfo.senderShardID,
				"receiver shard", crossTxInfo.receiverShardID,
				"type", getTxTypeName(crossTxInfo.txType))
			delete(ctpc.mapCrossTxsRounds, hash)
			continue
		}

		txHandler, ok := value.(data.TransactionHandler)
		if !ok {
			log.Debug("cleanCrossTxsPoolsIfNeeded", "error", process.ErrWrongTypeAssertion,
				"hash", []byte(hash),
				"round", crossTxInfo.round,
				"sender shard", crossTxInfo.senderShardID,
				"receiver shard", crossTxInfo.receiverShardID,
				"type", getTxTypeName(crossTxInfo.txType))
			continue
		}

		percentUsed := float64(crossTxInfo.txStore.Len()) / float64(crossTxInfo.txStore.MaxSize())
		if numPendingMiniBlocks > 0 && percentUsed < percentAllowed {
			log.Trace("cleaning cross transaction not yet allowed",
				"hash", []byte(hash),
				"round", crossTxInfo.round,
				"sender shard", crossTxInfo.senderShardID,
				"receiver shard", crossTxInfo.receiverShardID,
				"type", getTxTypeName(crossTxInfo.txType),
				"num pending miniblocks", numPendingMiniBlocks,
				"transactions pool percent used", percentUsed,
				"nonce", txHandler.GetNonce(),
				"sender address", txHandler.GetSndAddr(),
				"receiver address", txHandler.GetRcvAddr(),
				"value", txHandler.GetValue())
			continue
		}

		roundDif := ctpc.rounder.Index() - crossTxInfo.round
		if roundDif <= process.MaxRoundsToKeepUnprocessedTransactions {
			log.Trace("cleaning cross transaction not yet allowed",
				"hash", []byte(hash),
				"round", crossTxInfo.round,
				"sender shard", crossTxInfo.senderShardID,
				"receiver shard", crossTxInfo.receiverShardID,
				"type", getTxTypeName(crossTxInfo.txType),
				"round dif", roundDif,
				"nonce", txHandler.GetNonce(),
				"sender address", txHandler.GetSndAddr(),
				"receiver address", txHandler.GetRcvAddr(),
				"value", txHandler.GetValue())
			continue
		}

		crossTxInfo.txStore.Remove([]byte(hash))
		delete(ctpc.mapCrossTxsRounds, hash)
		numTxsCleaned++

		log.Trace("transaction has been cleaned",
			"hash", []byte(hash),
			"round", crossTxInfo.round,
			"sender shard", crossTxInfo.senderShardID,
			"receiver shard", crossTxInfo.receiverShardID,
			"type", getTxTypeName(crossTxInfo.txType),
			"nonce", txHandler.GetNonce(),
			"sender address", txHandler.GetSndAddr(),
			"receiver address", txHandler.GetRcvAddr(),
			"value", txHandler.GetValue())
	}

	if numTxsCleaned > 0 {
		log.Debug("crossTxsPoolsCleaner.cleanCrossTxsPoolsIfNeeded", "num txs cleaned", numTxsCleaned)
	}
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
