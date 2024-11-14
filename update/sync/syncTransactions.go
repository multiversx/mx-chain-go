package sync

import (
	"context"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	"github.com/multiversx/mx-chain-go/update"
)

var _ update.TransactionsSyncHandler = (*transactionsSync)(nil)

type transactionsSync struct {
	mutPendingTx            sync.Mutex
	mapTransactions         map[string]data.TransactionHandler
	mapTxsToMiniBlocks      map[string]*block.MiniBlock
	mapValidatorsInfo       map[string]*state.ShardValidatorInfo
	txPools                 map[block.Type]dataRetriever.ShardedDataCacherNotifier
	storage                 map[block.Type]update.HistoryStorer
	chReceivedAll           chan bool
	requestHandler          process.RequestHandler
	marshaller              marshal.Marshalizer
	epochToSync             uint32
	stopSync                bool
	syncedAll               bool
	waitTimeBetweenRequests time.Duration
}

// ArgsNewTransactionsSyncer defines the arguments needed for a new transactions syncer
type ArgsNewTransactionsSyncer struct {
	DataPools      dataRetriever.PoolsHolder
	Storages       dataRetriever.StorageService
	Marshaller     marshal.Marshalizer
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
	if check.IfNil(args.Marshaller) {
		return nil, dataRetriever.ErrNilMarshalizer
	}
	if check.IfNil(args.RequestHandler) {
		return nil, process.ErrNilRequestHandler
	}

	ts := &transactionsSync{
		mutPendingTx:            sync.Mutex{},
		mapTransactions:         make(map[string]data.TransactionHandler),
		mapTxsToMiniBlocks:      make(map[string]*block.MiniBlock),
		mapValidatorsInfo:       make(map[string]*state.ShardValidatorInfo),
		chReceivedAll:           make(chan bool),
		requestHandler:          args.RequestHandler,
		marshaller:              args.Marshaller,
		stopSync:                true,
		syncedAll:               true,
		waitTimeBetweenRequests: args.RequestHandler.RequestInterval(),
	}

	ts.txPools = make(map[block.Type]dataRetriever.ShardedDataCacherNotifier)
	ts.txPools[block.TxBlock] = args.DataPools.Transactions()
	ts.txPools[block.SmartContractResultBlock] = args.DataPools.UnsignedTransactions()
	ts.txPools[block.RewardsBlock] = args.DataPools.RewardTransactions()
	ts.txPools[block.PeerBlock] = args.DataPools.ValidatorsInfo()

	var err error
	ts.storage = make(map[block.Type]update.HistoryStorer)
	ts.storage[block.TxBlock], err = args.Storages.GetStorer(dataRetriever.TransactionUnit)
	if err != nil {
		return nil, err
	}

	ts.storage[block.SmartContractResultBlock], err = args.Storages.GetStorer(dataRetriever.UnsignedTransactionUnit)
	if err != nil {
		return nil, err
	}

	ts.storage[block.RewardsBlock], err = args.Storages.GetStorer(dataRetriever.RewardTransactionUnit)
	if err != nil {
		return nil, err
	}

	ts.storage[block.PeerBlock], err = args.Storages.GetStorer(dataRetriever.UnsignedTransactionUnit)
	if err != nil {
		return nil, err
	}

	for poolType, pool := range ts.txPools {
		if poolType == block.PeerBlock {
			pool.RegisterOnAdded(ts.receivedValidatorInfo)
			continue
		}

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

		numRequestedTxs := 0
		for _, miniBlock := range miniBlocks {
			for _, txHash := range miniBlock.TxHashes {
				ts.mapTxsToMiniBlocks[string(txHash)] = miniBlock
				log.Debug("transactionsSync.SyncTransactionsFor", "mb type", miniBlock.Type, "mb sender", miniBlock.SenderShardID, "mb receiver", miniBlock.ReceiverShardID, "tx hash needed", txHash)
			}
			numRequestedTxs += ts.requestTransactionsFor(miniBlock)
		}
		ts.mutPendingTx.Unlock()

		if numRequestedTxs == 0 {
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
			ts.mutPendingTx.Lock()
			log.Debug("transactionsSync.SyncTransactionsFor", "num txs needed", len(ts.mapTxsToMiniBlocks), "num txs got", len(ts.mapTransactions))
			ts.mutPendingTx.Unlock()
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
	if miniBlock.Type == block.PeerBlock {
		return ts.requestTransactionsForPeerMiniBlock(miniBlock)
	}

	return ts.requestTransactionsForNonPeerMiniBlock(miniBlock)

}

func (ts *transactionsSync) requestTransactionsForNonPeerMiniBlock(miniBlock *block.MiniBlock) int {
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
		ts.mapTxsToMiniBlocks[string(txHash)] = miniBlock
		log.Debug("transactionsSync.requestTransactionsForNonPeerMiniBlock", "mb type", miniBlock.Type, "mb sender", miniBlock.SenderShardID, "mb receiver", miniBlock.ReceiverShardID, "tx hash missing", txHash)
	}

	mbType := miniBlock.Type
	if mbType == block.InvalidBlock {
		mbType = block.TxBlock
	}

	switch mbType {
	case block.TxBlock:
		ts.requestHandler.RequestTransaction(miniBlock.SenderShardID, missingTxs)
		ts.requestHandler.RequestTransaction(miniBlock.ReceiverShardID, missingTxs)
	case block.SmartContractResultBlock:
		ts.requestHandler.RequestUnsignedTransactions(miniBlock.SenderShardID, missingTxs)
		ts.requestHandler.RequestUnsignedTransactions(miniBlock.ReceiverShardID, missingTxs)
	case block.RewardsBlock:
		ts.requestHandler.RequestRewardTransactions(miniBlock.SenderShardID, missingTxs)
		ts.requestHandler.RequestRewardTransactions(miniBlock.ReceiverShardID, missingTxs)
	}

	return len(missingTxs)
}

func (ts *transactionsSync) requestTransactionsForPeerMiniBlock(miniBlock *block.MiniBlock) int {
	missingValidatorsInfo := make([][]byte, 0)
	for _, txHash := range miniBlock.TxHashes {
		_, isValidatorInfoFound := ts.mapValidatorsInfo[string(txHash)]
		if isValidatorInfoFound {
			continue
		}

		validatorInfo, ok := ts.getValidatorInfoFromPoolOrStorage(txHash)
		if ok {
			ts.mapValidatorsInfo[string(txHash)] = validatorInfo
			continue
		}

		missingValidatorsInfo = append(missingValidatorsInfo, txHash)
	}

	for _, txHash := range missingValidatorsInfo {
		ts.mapTxsToMiniBlocks[string(txHash)] = miniBlock
		log.Debug("transactionsSync.requestTransactionsForPeerMiniBlock", "mb type", miniBlock.Type, "mb sender", miniBlock.SenderShardID, "mb receiver", miniBlock.ReceiverShardID, "tx hash missing", txHash)
	}

	go ts.requestHandler.RequestValidatorsInfo(missingValidatorsInfo)

	return len(missingValidatorsInfo)
}

func (ts *transactionsSync) receivedTransaction(txHash []byte, val interface{}) {
	ts.mutPendingTx.Lock()
	if ts.stopSync {
		ts.mutPendingTx.Unlock()
		return
	}

	miniBlock, foundInMap := ts.mapTxsToMiniBlocks[string(txHash)]
	if !foundInMap {
		ts.mutPendingTx.Unlock()
		return
	}
	_, foundInMap = ts.mapTransactions[string(txHash)]
	if foundInMap {
		ts.mutPendingTx.Unlock()
		return
	}

	var tx data.TransactionHandler
	var wrappedTx *txcache.WrappedTransaction
	var ok bool

	tx, ok = val.(data.TransactionHandler)
	if !ok {
		wrappedTx, ok = val.(*txcache.WrappedTransaction)
		if !ok {
			ts.mutPendingTx.Unlock()
			log.Error("transactionsSync.receivedTransaction", "tx hash", txHash, "error", update.ErrWrongTypeAssertion)
			return
		}

		tx = wrappedTx.Tx
	}

	log.Debug("transactionsSync.receivedTransaction", "mb type", miniBlock.Type, "mb sender", miniBlock.SenderShardID, "mb receiver", miniBlock.ReceiverShardID, "tx hash got", txHash)

	ts.mapTransactions[string(txHash)] = tx
	receivedAllMissing := len(ts.mapTxsToMiniBlocks) == len(ts.mapTransactions)+len(ts.mapValidatorsInfo)
	ts.mutPendingTx.Unlock()

	if receivedAllMissing {
		ts.chReceivedAll <- true
	}
}

func (ts *transactionsSync) receivedValidatorInfo(txHash []byte, val interface{}) {
	ts.mutPendingTx.Lock()
	if ts.stopSync {
		ts.mutPendingTx.Unlock()
		return
	}

	miniBlock, ok := ts.mapTxsToMiniBlocks[string(txHash)]
	if !ok {
		ts.mutPendingTx.Unlock()
		return
	}
	_, ok = ts.mapValidatorsInfo[string(txHash)]
	if ok {
		ts.mutPendingTx.Unlock()
		return
	}

	validatorInfo, ok := val.(*state.ShardValidatorInfo)
	if !ok {
		ts.mutPendingTx.Unlock()
		log.Error("transactionsSync.receivedValidatorInfo", "tx hash", txHash, "error", update.ErrWrongTypeAssertion)
		return
	}

	log.Debug("transactionsSync.receivedValidatorInfo", "mb type", miniBlock.Type, "mb sender", miniBlock.SenderShardID, "mb receiver", miniBlock.ReceiverShardID, "tx hash got", txHash)

	ts.mapValidatorsInfo[string(txHash)] = validatorInfo
	receivedAllMissing := len(ts.mapTxsToMiniBlocks) == len(ts.mapValidatorsInfo)+len(ts.mapTransactions)
	ts.mutPendingTx.Unlock()

	if receivedAllMissing {
		ts.chReceivedAll <- true
	}
}

func (ts *transactionsSync) getTransactionFromPool(txHash []byte) (data.TransactionHandler, bool) {
	mb, ok := ts.mapTxsToMiniBlocks[string(txHash)]
	if !ok {
		return nil, false
	}

	mbType := mb.Type
	if mbType == block.InvalidBlock {
		mbType = block.TxBlock
	}

	if _, ok = ts.txPools[mbType]; !ok {
		log.Debug("transactionsSync.getTransactionFromPool: missing mini block type from sharded data cacher notifier map",
			"tx hash", txHash,
			"original mb type", mb.Type,
			"mb type", mbType,
			"mb sender shard", mb.SenderShardID,
			"mb receiver shard", mb.ReceiverShardID,
			"mb num txs", len(mb.TxHashes))
		return nil, false
	}

	storeId := process.ShardCacherIdentifier(mb.SenderShardID, mb.ReceiverShardID)
	shardTxStore := ts.txPools[mbType].ShardDataStore(storeId)
	if check.IfNil(shardTxStore) {
		return nil, false
	}

	val, ok := shardTxStore.Peek(txHash)
	if !ok {
		return nil, false
	}

	tx, ok := val.(data.TransactionHandler)
	if !ok {
		log.Error("transactionsSync.getTransactionFromPool", "tx hash", txHash, "error", update.ErrWrongTypeAssertion)
		return nil, false
	}

	return tx, true
}

func (ts *transactionsSync) getValidatorInfoFromPool(txHash []byte) (*state.ShardValidatorInfo, bool) {
	mb, ok := ts.mapTxsToMiniBlocks[string(txHash)]
	if !ok {
		return nil, false
	}

	if _, ok = ts.txPools[block.PeerBlock]; !ok {
		log.Debug("transactionsSync.getValidatorInfoFromPool: missing mini block type from sharded data cacher notifier map",
			"tx hash", txHash,
			"original mb type", mb.Type,
			"mb type", block.PeerBlock,
			"mb sender shard", mb.SenderShardID,
			"mb receiver shard", mb.ReceiverShardID,
			"mb num txs", len(mb.TxHashes))
		return nil, false
	}

	storeId := process.ShardCacherIdentifier(mb.SenderShardID, mb.ReceiverShardID)
	shardTxStore := ts.txPools[block.PeerBlock].ShardDataStore(storeId)
	if check.IfNil(shardTxStore) {
		return nil, false
	}

	val, ok := shardTxStore.Peek(txHash)
	if !ok {
		return nil, false
	}

	validatorInfo, ok := val.(*state.ShardValidatorInfo)
	if !ok {
		log.Error("transactionsSync.getValidatorInfoFromPool", "tx hash", txHash, "error", update.ErrWrongTypeAssertion)
		return nil, false
	}

	return validatorInfo, true
}

func (ts *transactionsSync) getTransactionFromPoolWithSearchFirst(
	txHash []byte,
	cacher dataRetriever.ShardedDataCacherNotifier,
) (data.TransactionHandler, bool) {
	val, ok := cacher.SearchFirstData(txHash)
	if !ok {
		return nil, false
	}

	tx, ok := val.(data.TransactionHandler)
	if !ok {
		return nil, false
	}

	return tx, true
}

func (ts *transactionsSync) getValidatorInfoFromPoolWithSearchFirst(
	txHash []byte,
	cacher dataRetriever.ShardedDataCacherNotifier,
) (*state.ShardValidatorInfo, bool) {
	val, ok := cacher.SearchFirstData(txHash)
	if !ok {
		return nil, false
	}

	validatorInfo, ok := val.(*state.ShardValidatorInfo)
	if !ok {
		return nil, false
	}

	return validatorInfo, true
}

func (ts *transactionsSync) getTransactionFromPoolOrStorage(hash []byte) (data.TransactionHandler, bool) {
	txFromPool, ok := ts.getTransactionFromPool(hash)
	if ok {
		return txFromPool, true
	}

	miniBlock, ok := ts.mapTxsToMiniBlocks[string(hash)]
	if !ok {
		return nil, false
	}

	mbType := miniBlock.Type
	if mbType == block.InvalidBlock {
		mbType = block.TxBlock
	}

	txFromPoolWithSearchFirst, ok := ts.getTransactionFromPoolWithSearchFirst(hash, ts.txPools[mbType])
	if ok {
		log.Debug("transactionsSync.getTransactionFromPoolWithSearchFirst: found transaction using search first", "mb type", miniBlock.Type, "mb sender", miniBlock.SenderShardID, "mb receiver", miniBlock.ReceiverShardID, "tx hash", hash)
		return txFromPoolWithSearchFirst, true
	}

	txData, err := GetDataFromStorage(hash, ts.storage[mbType])
	if err != nil {
		return nil, false
	}

	var tx data.TransactionHandler
	switch mbType {
	case block.TxBlock:
		tx = &transaction.Transaction{}
	case block.SmartContractResultBlock:
		tx = &smartContractResult.SmartContractResult{}
	case block.RewardsBlock:
		tx = &rewardTx.RewardTx{}
	}

	err = ts.marshaller.Unmarshal(tx, txData)
	if err != nil {
		return nil, false
	}

	return tx, true
}

func (ts *transactionsSync) getValidatorInfoFromPoolOrStorage(hash []byte) (*state.ShardValidatorInfo, bool) {
	validatorInfoFromPool, ok := ts.getValidatorInfoFromPool(hash)
	if ok {
		return validatorInfoFromPool, true
	}

	miniBlock, ok := ts.mapTxsToMiniBlocks[string(hash)]
	if !ok {
		return nil, false
	}

	validatorInfoFromPoolWithSearchFirst, ok := ts.getValidatorInfoFromPoolWithSearchFirst(hash, ts.txPools[block.PeerBlock])
	if ok {
		log.Debug("transactionsSync.getValidatorInfoFromPoolOrStorage: found transaction using search first", "mb type", miniBlock.Type, "mb sender", miniBlock.SenderShardID, "mb receiver", miniBlock.ReceiverShardID, "tx hash", hash)
		return validatorInfoFromPoolWithSearchFirst, true
	}

	validatorInfoData, err := GetDataFromStorage(hash, ts.storage[block.PeerBlock])
	if err != nil {
		return nil, false
	}

	validatorInfo := &state.ShardValidatorInfo{}

	err = ts.marshaller.Unmarshal(validatorInfo, validatorInfoData)
	if err != nil {
		return nil, false
	}

	return validatorInfo, true
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

// GetValidatorsInfo returns the synced validators info
func (ts *transactionsSync) GetValidatorsInfo() (map[string]*state.ShardValidatorInfo, error) {
	ts.mutPendingTx.Lock()
	defer ts.mutPendingTx.Unlock()
	if !ts.syncedAll {
		return nil, update.ErrNotSynced
	}

	return ts.mapValidatorsInfo, nil
}

// ClearFields will clear all the maps
func (ts *transactionsSync) ClearFields() {
	ts.mutPendingTx.Lock()
	ts.mapTransactions = make(map[string]data.TransactionHandler)
	ts.mapTxsToMiniBlocks = make(map[string]*block.MiniBlock)
	ts.mapValidatorsInfo = make(map[string]*state.ShardValidatorInfo)
	ts.mutPendingTx.Unlock()
}

// IsInterfaceNil returns true if underlying object is nil
func (ts *transactionsSync) IsInterfaceNil() bool {
	return ts == nil
}
