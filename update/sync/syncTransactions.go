package sync

import (
	"context"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/update"
)

var _ update.TransactionsSyncHandler = (*transactionsSync)(nil)

type transactionsSync struct {
	mutPendingTx            sync.Mutex
	mapTransactions         map[string]data.TransactionHandler
	mapHashes               map[string]*block.MiniBlock
	txPools                 map[block.Type]dataRetriever.ShardedDataCacherNotifier
	storage                 map[block.Type]update.HistoryStorer
	chReceivedAll           chan bool
	requestHandler          process.RequestHandler
	marshalizer             marshal.Marshalizer
	epochToSync             uint32
	stopSync                bool
	syncedAll               bool
	waitTimeBetweenRequests time.Duration
}

// ArgsNewTransactionsSyncer defines the arguments needed for a new transactions syncer
type ArgsNewTransactionsSyncer struct {
	DataPools      dataRetriever.PoolsHolder
	Storages       dataRetriever.StorageService
	Marshalizer    marshal.Marshalizer
	RequestHandler process.RequestHandler
}

// NewTransactionsSyncer creates a new transactions syncer
func NewTransactionsSyncer(args ArgsNewTransactionsSyncer) (*transactionsSync, error) {
	if check.IfNil(args.Storages) {
		return nil, dataRetriever.ErrNilHeadersStorage
	}
	if check.IfNil(args.DataPools) {
		return nil, dataRetriever.ErrNilDataPoolHolder
	}
	if check.IfNil(args.Marshalizer) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(args.RequestHandler) {
		return nil, process.ErrNilRequestHandler
	}

	ts := &transactionsSync{
		mutPendingTx:            sync.Mutex{},
		mapTransactions:         make(map[string]data.TransactionHandler),
		mapHashes:               make(map[string]*block.MiniBlock),
		chReceivedAll:           make(chan bool),
		requestHandler:          args.RequestHandler,
		marshalizer:             args.Marshalizer,
		stopSync:                true,
		syncedAll:               true,
		waitTimeBetweenRequests: args.RequestHandler.RequestInterval(),
	}

	ts.txPools = make(map[block.Type]dataRetriever.ShardedDataCacherNotifier)
	ts.txPools[block.TxBlock] = args.DataPools.Transactions()
	ts.txPools[block.SmartContractResultBlock] = args.DataPools.UnsignedTransactions()
	ts.txPools[block.RewardsBlock] = args.DataPools.RewardTransactions()

	ts.storage = make(map[block.Type]update.HistoryStorer)
	ts.storage[block.TxBlock] = args.Storages.GetStorer(dataRetriever.TransactionUnit)
	ts.storage[block.SmartContractResultBlock] = args.Storages.GetStorer(dataRetriever.UnsignedTransactionUnit)
	ts.storage[block.RewardsBlock] = args.Storages.GetStorer(dataRetriever.RewardTransactionUnit)

	for _, pool := range ts.txPools {
		pool.RegisterOnAdded(ts.receivedTransaction)
	}

	return ts, nil
}

// SyncTransactionsFor syncs transactions for a list of miniblocks
func (ts *transactionsSync) SyncTransactionsFor(miniBlocks map[string]*block.MiniBlock, epoch uint32, ctx context.Context) error {
	_ = core.EmptyChannel(ts.chReceivedAll)

	for {
		ts.mutPendingTx.Lock()
		ts.epochToSync = epoch
		ts.syncedAll = false
		ts.stopSync = false

		requestedTxs := 0
		for _, miniBlock := range miniBlocks {
			for _, txHash := range miniBlock.TxHashes {
				ts.mapHashes[string(txHash)] = miniBlock
			}
			requestedTxs += ts.requestTransactionsFor(miniBlock)
		}
		ts.mutPendingTx.Unlock()

		if requestedTxs == 0 {
			ts.mutPendingTx.Lock()
			ts.stopSync = true
			ts.syncedAll = true
			ts.mutPendingTx.Unlock()
			return nil
		}

		select {
		case <-ts.chReceivedAll:
			ts.mutPendingTx.Lock()
			ts.stopSync = true
			ts.syncedAll = true
			ts.mutPendingTx.Unlock()
			return nil
		case <-time.After(ts.waitTimeBetweenRequests):
			continue
		case <-ctx.Done():
			ts.mutPendingTx.Lock()
			ts.stopSync = true
			ts.mutPendingTx.Unlock()
			return update.ErrTimeIsOut
		}
	}
}

func (ts *transactionsSync) requestTransactionsFor(miniBlock *block.MiniBlock) int {
	missingTxs := make([][]byte, 0)
	for _, txHash := range miniBlock.TxHashes {
		if _, ok := ts.mapTransactions[string(txHash)]; ok {
			continue
		}

		tx, ok := ts.getTransactionFromPoolOrStorage(txHash)
		if ok {
			ts.mapTransactions[string(txHash)] = tx
			continue
		}

		missingTxs = append(missingTxs, txHash)
	}

	for _, txHash := range missingTxs {
		ts.mapHashes[string(txHash)] = miniBlock
	}

	switch miniBlock.Type {
	case block.TxBlock:
		ts.requestHandler.RequestTransaction(miniBlock.SenderShardID, missingTxs)
	case block.SmartContractResultBlock:
		ts.requestHandler.RequestUnsignedTransactions(miniBlock.SenderShardID, missingTxs)
	case block.RewardsBlock:
		ts.requestHandler.RequestRewardTransactions(miniBlock.SenderShardID, missingTxs)
	}

	return len(missingTxs)
}

func (ts *transactionsSync) receivedTransaction(txHash []byte, val interface{}) {
	ts.mutPendingTx.Lock()
	if ts.stopSync {
		ts.mutPendingTx.Unlock()
		return
	}

	if _, ok := ts.mapHashes[string(txHash)]; !ok {
		ts.mutPendingTx.Unlock()
		return
	}
	if _, ok := ts.mapTransactions[string(txHash)]; ok {
		ts.mutPendingTx.Unlock()
		return
	}

	tx, ok := val.(data.TransactionHandler)
	if !ok {
		ts.mutPendingTx.Unlock()
		return
	}

	ts.mapTransactions[string(txHash)] = tx
	receivedAllMissing := len(ts.mapHashes) == len(ts.mapTransactions)
	ts.mutPendingTx.Unlock()

	if receivedAllMissing {
		ts.chReceivedAll <- true
	}
}

func (ts *transactionsSync) getTransactionFromPool(txHash []byte) (data.TransactionHandler, bool) {
	mb := ts.mapHashes[string(txHash)]
	storeId := process.ShardCacherIdentifier(mb.SenderShardID, mb.ReceiverShardID)
	shardTxStore := ts.txPools[mb.Type].ShardDataStore(storeId)
	if check.IfNil(shardTxStore) {
		return nil, false
	}

	val, ok := shardTxStore.Peek(txHash)
	if !ok {
		return nil, false
	}

	tx, ok := val.(data.TransactionHandler)
	if !ok {
		return nil, false
	}

	return tx, true
}

func (ts *transactionsSync) getTransactionFromPoolOrStorage(hash []byte) (data.TransactionHandler, bool) {
	txFromPool, ok := ts.getTransactionFromPool(hash)
	if ok {
		return txFromPool, true
	}

	miniBlock, ok := ts.mapHashes[string(hash)]
	if !ok {
		return nil, false
	}

	txData, err := GetDataFromStorage(hash, ts.storage[miniBlock.Type])
	if err != nil {
		return nil, false
	}

	var tx data.TransactionHandler
	switch miniBlock.Type {
	case block.TxBlock:
		tx = &transaction.Transaction{}
	case block.SmartContractResultBlock:
		tx = &smartContractResult.SmartContractResult{}
	case block.RewardsBlock:
		tx = &rewardTx.RewardTx{}
	}

	err = ts.marshalizer.Unmarshal(tx, txData)
	if err != nil {
		return nil, false
	}

	return tx, true
}

// GetTransactions returns the synced transactions
func (ts *transactionsSync) GetTransactions() (map[string]data.TransactionHandler, error) {
	ts.mutPendingTx.Lock()
	defer ts.mutPendingTx.Unlock()
	if !ts.syncedAll {
		return nil, update.ErrNotSynced
	}

	return ts.mapTransactions, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (ts *transactionsSync) IsInterfaceNil() bool {
	return ts == nil
}
