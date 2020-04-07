package poolsCleaner

import (
	"bytes"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
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
	addressConverter         state.AddressConverter
	blockTransactionsPool    dataRetriever.ShardedDataCacherNotifier
	rewardTransactionsPool   dataRetriever.ShardedDataCacherNotifier
	unsignedTransactionsPool dataRetriever.ShardedDataCacherNotifier
	rounder                  process.Rounder
	shardCoordinator         sharding.Coordinator

	mutMapCrossTxsRounds sync.RWMutex
	mapCrossTxsRounds    map[string]*txInfo
	emptyAddress         []byte
}

// NewCrossTxsPoolsCleaner will return a new cross txs pools cleaner
func NewCrossTxsPoolsCleaner(
	addressConverter state.AddressConverter,
	dataPool dataRetriever.PoolsHolder,
	rounder process.Rounder,
	shardCoordinator sharding.Coordinator,
) (*crossTxsPoolsCleaner, error) {

	if check.IfNil(addressConverter) {
		return nil, process.ErrNilAddressConverter
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
		addressConverter:         addressConverter,
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

	ctpc.emptyAddress = make([]byte, ctpc.addressConverter.AddressLen())

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

	wrappedTx, ok := value.(*txcache.WrappedTransaction)
	if !ok {
		log.Warn("crossTxsPoolsCleaner.receivedBlockTx", "error", process.ErrWrongTypeAssertion)
		return
	}

	ctpc.processReceivedTx(key, wrappedTx.SenderShardID, wrappedTx.ReceiverShardID, blockTx)
}

func (ctpc *crossTxsPoolsCleaner) receivedRewardTx(key []byte, value interface{}) {
	if key == nil {
		return
	}

	log.Trace("crossTxsPoolsCleaner.receivedRewardTx", "hash", key)

	tx, ok := value.(data.TransactionHandler)
	if !ok {
		log.Warn("crossTxsPoolsCleaner.receivedRewardTx", "error", process.ErrWrongTypeAssertion)
		return
	}

	senderShardID, receiverShardID, err := ctpc.computeSenderAndReceiverShards(tx)
	if err != nil {
		log.Debug("crossTxsPoolsCleaner.receivedRewardTx", "error", err.Error())
		return
	}

	ctpc.processReceivedTx(key, senderShardID, receiverShardID, rewardTx)
}

func (ctpc *crossTxsPoolsCleaner) receivedUnsignedTx(key []byte, value interface{}) {
	if key == nil {
		return
	}

	log.Trace("crossTxsPoolsCleaner.receivedUnsignedTx", "hash", key)

	tx, ok := value.(data.TransactionHandler)
	if !ok {
		log.Warn("crossTxsPoolsCleaner.receivedUnsignedTx", "error", process.ErrWrongTypeAssertion)
		return
	}

	senderShardID, receiverShardID, err := ctpc.computeSenderAndReceiverShards(tx)
	if err != nil {
		log.Debug("crossTxsPoolsCleaner.receivedUnsignedTx", "error", err.Error())
		return
	}

	ctpc.processReceivedTx(key, senderShardID, receiverShardID, unsignedTx)
}

func (ctpc *crossTxsPoolsCleaner) processReceivedTx(
	key []byte,
	senderShardID uint32,
	receiverShardID uint32,
	txType int8,
) {
	if senderShardID == ctpc.shardCoordinator.SelfId() {
		return
	}

	ctpc.mutMapCrossTxsRounds.Lock()
	defer ctpc.mutMapCrossTxsRounds.Unlock()

	if _, ok := ctpc.mapCrossTxsRounds[string(key)]; !ok {
		transactionPool := ctpc.getTransactionPool(txType)
		if transactionPool == nil {
			return
		}

		strCache := process.ShardCacherIdentifier(senderShardID, receiverShardID)
		txStore := transactionPool.ShardDataStore(strCache)
		if txStore == nil {
			return
		}

		crossTxInfo := &txInfo{
			round:           ctpc.rounder.Index(),
			senderShardID:   senderShardID,
			receiverShardID: receiverShardID,
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

	numTxsCleaned := 0

	for hash, crossTxInfo := range ctpc.mapCrossTxsRounds {
		_, ok := crossTxInfo.txStore.Get([]byte(hash))
		if !ok {
			log.Trace("transaction not found in pool",
				"hash", []byte(hash),
				"round", crossTxInfo.round,
				"sender", crossTxInfo.senderShardID,
				"receiver", crossTxInfo.receiverShardID,
				"type", getTxTypeName(crossTxInfo.txType))
			delete(ctpc.mapCrossTxsRounds, hash)
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

func (ctpc *crossTxsPoolsCleaner) computeSenderAndReceiverShards(tx data.TransactionHandler) (uint32, uint32, error) {
	senderShardID, err := ctpc.getShardFromAddress(tx.GetSndAddr())
	if err != nil {
		return 0, 0, err
	}

	receiverShardID, err := ctpc.getShardFromAddress(tx.GetRcvAddr())
	if err != nil {
		return 0, 0, err
	}

	return senderShardID, receiverShardID, nil
}

func (ctpc *crossTxsPoolsCleaner) getShardFromAddress(address []byte) (uint32, error) {
	isEmptyAddress := bytes.Equal(address, ctpc.emptyAddress)
	if isEmptyAddress {
		return ctpc.shardCoordinator.SelfId(), nil
	}

	addressContainer, err := ctpc.addressConverter.CreateAddressFromPublicKeyBytes(address)
	if err != nil {
		return 0, err
	}

	return ctpc.shardCoordinator.ComputeId(addressContainer), nil
}
