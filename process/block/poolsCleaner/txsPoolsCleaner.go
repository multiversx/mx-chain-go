package poolsCleaner

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/closing"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/txcache"
)

var _ closing.Closer = (*txsPoolsCleaner)(nil)

// sleepTime defines the time between each iteration made in clean...Pools methods
const sleepTime = time.Minute

const (
	blockTx = iota
	rewardTx
	unsignedTx
)

// ArgTxsPoolsCleaner represents the argument structure used to create a new txsPoolsCleaner instance
type ArgTxsPoolsCleaner struct {
	ArgBasePoolsCleaner
	AddressPubkeyConverter core.PubkeyConverter
	DataPool               dataRetriever.PoolsHolder
}

type txInfo struct {
	round           int64
	senderShardID   uint32
	receiverShardID uint32
	txType          int8
	txStore         storage.Cacher
}

// txsPoolsCleaner represents a pools cleaner that checks and cleans txs which should not be in pool anymore
type txsPoolsCleaner struct {
	basePoolsCleaner
	addressPubkeyConverter   core.PubkeyConverter
	blockTransactionsPool    dataRetriever.ShardedDataCacherNotifier
	rewardTransactionsPool   dataRetriever.ShardedDataCacherNotifier
	unsignedTransactionsPool dataRetriever.ShardedDataCacherNotifier

	mutMapTxsRounds sync.RWMutex
	mapTxsRounds    map[string]*txInfo
	emptyAddress    []byte
}

// NewTxsPoolsCleaner will return a new txs pools cleaner
func NewTxsPoolsCleaner(args ArgTxsPoolsCleaner) (*txsPoolsCleaner, error) {
	err := checkArgTxsPoolsCleaner(args)
	if err != nil {
		return nil, err
	}

	tpc := txsPoolsCleaner{
		basePoolsCleaner:         newBasePoolsCleaner(args.ArgBasePoolsCleaner),
		addressPubkeyConverter:   args.AddressPubkeyConverter,
		blockTransactionsPool:    args.DataPool.Transactions(),
		rewardTransactionsPool:   args.DataPool.RewardTransactions(),
		unsignedTransactionsPool: args.DataPool.UnsignedTransactions(),
		mapTxsRounds:             make(map[string]*txInfo),
		emptyAddress:             make([]byte, args.AddressPubkeyConverter.Len()),
	}

	tpc.blockTransactionsPool.RegisterOnAdded(tpc.receivedBlockTx)
	tpc.rewardTransactionsPool.RegisterOnAdded(tpc.receivedRewardTx)
	tpc.unsignedTransactionsPool.RegisterOnAdded(tpc.receivedUnsignedTx)

	return &tpc, nil
}

func checkArgTxsPoolsCleaner(args ArgTxsPoolsCleaner) error {
	err := checkBaseArgs(args.ArgBasePoolsCleaner)
	if err != nil {
		return err
	}
	if check.IfNil(args.AddressPubkeyConverter) {
		return process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.DataPool) {
		return process.ErrNilPoolsHolder
	}
	if check.IfNil(args.DataPool.Transactions()) {
		return process.ErrNilTransactionPool
	}
	if check.IfNil(args.DataPool.RewardTransactions()) {
		return process.ErrNilRewardTxDataPool
	}
	if check.IfNil(args.DataPool.UnsignedTransactions()) {
		return process.ErrNilUnsignedTxDataPool
	}

	return nil
}

// StartCleaning actually starts the pools cleaning mechanism
func (tpc *txsPoolsCleaner) StartCleaning() {
	tpc.mut.Lock()
	defer tpc.mut.Unlock()

	if tpc.isCleaningRoutineRunning {
		log.Error("txsPoolsCleaner cleaning routine already started...")
		return
	}

	tpc.isCleaningRoutineRunning = true
	var ctx context.Context
	ctx, tpc.cancelFunc = context.WithCancel(context.Background())
	go tpc.cleanTxsPools(ctx)
}

func (tpc *txsPoolsCleaner) cleanTxsPools(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Debug("txsPoolsCleaner's go routine is stopping...")
			return
		case <-time.After(sleepTime):
		}

		startTime := time.Now()
		numTxsInMap := tpc.cleanTxsPoolsIfNeeded()
		elapsedTime := time.Since(startTime)

		log.Debug("txsPoolsCleaner.cleanTxsPools",
			"num txs in map", numTxsInMap,
			"elapsed time", elapsedTime)
	}
}

func (tpc *txsPoolsCleaner) receivedBlockTx(key []byte, value interface{}) {
	if key == nil {
		return
	}

	log.Trace("txsPoolsCleaner.receivedBlockTx", "hash", key)

	wrappedTx, ok := value.(*txcache.WrappedTransaction)
	if !ok {
		log.Warn("txsPoolsCleaner.receivedBlockTx",
			"error", process.ErrWrongTypeAssertion,
			"found type", fmt.Sprintf("%T", value),
		)
		return
	}

	tpc.processReceivedTx(key, wrappedTx.SenderShardID, wrappedTx.ReceiverShardID, blockTx)
}

func (tpc *txsPoolsCleaner) receivedRewardTx(key []byte, value interface{}) {
	if key == nil {
		return
	}

	log.Trace("txsPoolsCleaner.receivedRewardTx", "hash", key)

	tx, ok := value.(data.TransactionHandler)
	if !ok {
		log.Warn("txsPoolsCleaner.receivedRewardTx",
			"error", process.ErrWrongTypeAssertion,
			"found type", fmt.Sprintf("%T", value),
		)
		return
	}

	senderShardID := core.MetachainShardId
	receiverShardID, err := tpc.getShardFromAddress(tx.GetRcvAddr())
	if err != nil {
		log.Debug("txsPoolsCleaner.receivedRewardTx", "error", err.Error())
		return
	}

	tpc.processReceivedTx(key, senderShardID, receiverShardID, rewardTx)
}

func (tpc *txsPoolsCleaner) receivedUnsignedTx(key []byte, value interface{}) {
	if key == nil {
		return
	}

	log.Trace("txsPoolsCleaner.receivedUnsignedTx", "hash", key)

	tx, ok := value.(data.TransactionHandler)
	if !ok {
		log.Warn("txsPoolsCleaner.receivedUnsignedTx",
			"error", process.ErrWrongTypeAssertion,
			"found type", fmt.Sprintf("%T", value),
		)
		return
	}

	senderShardID, receiverShardID, err := tpc.computeSenderAndReceiverShards(tx)
	if err != nil {
		log.Debug("txsPoolsCleaner.receivedUnsignedTx", "error", err.Error())
		return
	}

	tpc.processReceivedTx(key, senderShardID, receiverShardID, unsignedTx)
}

func (tpc *txsPoolsCleaner) processReceivedTx(
	key []byte,
	senderShardID uint32,
	receiverShardID uint32,
	txType int8,
) {
	tpc.mutMapTxsRounds.RLock()
	_, ok := tpc.mapTxsRounds[string(key)]
	tpc.mutMapTxsRounds.RUnlock()

	if !ok {
		transactionPool := tpc.getTransactionPool(txType)
		if check.IfNil(transactionPool) {
			return
		}

		strCache := process.ShardCacherIdentifier(senderShardID, receiverShardID)
		txStore := transactionPool.ShardDataStore(strCache)
		if check.IfNil(txStore) {
			return
		}

		currTxInfo := &txInfo{
			round:           tpc.roundHandler.Index(),
			senderShardID:   senderShardID,
			receiverShardID: receiverShardID,
			txType:          txType,
			txStore:         txStore,
		}

		tpc.mutMapTxsRounds.Lock()
		tpc.mapTxsRounds[string(key)] = currTxInfo
		tpc.mutMapTxsRounds.Unlock()

		log.Trace("transaction has been added",
			"hash", key,
			"round", currTxInfo.round,
			"sender", currTxInfo.senderShardID,
			"receiver", currTxInfo.receiverShardID,
			"type", getTxTypeName(currTxInfo.txType))
	}
}

func (tpc *txsPoolsCleaner) cleanTxsPoolsIfNeeded() int {
	numTxsCleaned := 0
	hashesToRemove := make(map[string]storage.Cacher)

	tpc.mutMapTxsRounds.Lock()
	for hash, currTxInfo := range tpc.mapTxsRounds {
		_, ok := currTxInfo.txStore.Get([]byte(hash))
		if !ok {
			log.Trace("transaction not found in pool",
				"hash", []byte(hash),
				"round", currTxInfo.round,
				"sender", currTxInfo.senderShardID,
				"receiver", currTxInfo.receiverShardID,
				"type", getTxTypeName(currTxInfo.txType))
			delete(tpc.mapTxsRounds, hash)
			continue
		}

		roundDif := tpc.roundHandler.Index() - currTxInfo.round
		if roundDif <= tpc.maxRoundsToKeepUnprocessedData {
			log.Trace("cleaning transaction not yet allowed",
				"hash", []byte(hash),
				"round", currTxInfo.round,
				"sender", currTxInfo.senderShardID,
				"receiver", currTxInfo.receiverShardID,
				"type", getTxTypeName(currTxInfo.txType),
				"round dif", roundDif)

			continue
		}

		hashesToRemove[hash] = currTxInfo.txStore
		delete(tpc.mapTxsRounds, hash)
		numTxsCleaned++

		log.Trace("transaction has been cleaned",
			"hash", []byte(hash),
			"round", currTxInfo.round,
			"sender", currTxInfo.senderShardID,
			"receiver", currTxInfo.receiverShardID,
			"type", getTxTypeName(currTxInfo.txType))
	}

	numTxsRounds := len(tpc.mapTxsRounds)
	tpc.mutMapTxsRounds.Unlock()

	startTime := time.Now()
	for hash, txStore := range hashesToRemove {
		txStore.Remove([]byte(hash))
	}
	elapsedTime := time.Since(startTime)

	if numTxsCleaned > 0 {
		log.Debug("txsPoolsCleaner.cleanTxsPoolsIfNeeded",
			"num txs cleaned", numTxsCleaned,
			"elapsed time to remove txs from cacher", elapsedTime)
	}

	return numTxsRounds
}

func (tpc *txsPoolsCleaner) getTransactionPool(txType int8) dataRetriever.ShardedDataCacherNotifier {
	switch txType {
	case blockTx:
		return tpc.blockTransactionsPool
	case rewardTx:
		return tpc.rewardTransactionsPool
	case unsignedTx:
		return tpc.unsignedTransactionsPool
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

func (tpc *txsPoolsCleaner) computeSenderAndReceiverShards(tx data.TransactionHandler) (uint32, uint32, error) {
	senderShardID, err := tpc.getShardFromAddress(tx.GetSndAddr())
	if err != nil {
		return 0, 0, err
	}

	receiverShardID, err := tpc.getShardFromAddress(tx.GetRcvAddr())
	if err != nil {
		return 0, 0, err
	}

	return senderShardID, receiverShardID, nil
}

func (tpc *txsPoolsCleaner) getShardFromAddress(address []byte) (uint32, error) {
	isEmptyAddress := bytes.Equal(address, tpc.emptyAddress)
	if isEmptyAddress {
		return tpc.shardCoordinator.SelfId(), nil
	}

	return tpc.shardCoordinator.ComputeId(address), nil
}
