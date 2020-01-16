package sync

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
)

var log = logger.GetOrCreate("update/genesis")

type pendingMiniBlocks struct {
	mutPendingMb  sync.Mutex
	mapMiniBlocks map[string]*block.MiniBlock
	mapHashes     map[string]struct{}
	pool          storage.Cacher
	storage       update.HistoryStorer
	chReceivedAll chan bool
}

type pendingTransactions struct {
	mutPendingTx    sync.Mutex
	mapTransactions map[string]data.TransactionHandler
	mapHashes       map[string]*block.MiniBlock
	txPools         map[block.Type]dataRetriever.ShardedDataCacherNotifier
	storage         map[block.Type]update.HistoryStorer
	chReceivedAll   chan bool
}

type syncState struct {
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator
	trieSyncers      update.TrieSyncContainer
	epochHandler     update.EpochStartVerifier
	requestHandler   process.RequestHandler
	headerValidator  process.HeaderConstructionValidator

	tries        map[string]data.Trie
	syncingEpoch uint32

	activeAccountsAdapters update.AccountsHandlerContainer

	headers update.HeaderSyncHandler

	miniBlocks   pendingMiniBlocks
	transactions pendingTransactions
}

// Arguments for the NewSync
type ArgsNewSyncState struct {
	Hasher           hashing.Hasher
	Marshalizer      marshal.Marshalizer
	ShardCoordinator sharding.Coordinator
	TrieSyncers      update.TrieSyncContainer
	EpochHandler     update.EpochStartVerifier
	Storages         dataRetriever.StorageService
	DataPools        dataRetriever.PoolsHolder
	RequestHandler   process.RequestHandler
	HeaderValidator  process.HeaderConstructionValidator
	AccountHandlers  update.AccountsHandlerContainer
}

// NewSyncState creates a complete syncer which saves the state of the blockchain with pending values as well
func NewSyncState(args ArgsNewSyncState) (*syncState, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, data.ErrNilShardCoordinator
	}
	if check.IfNil(args.TrieSyncers) {
		return nil, dataRetriever.ErrNilResolverContainer
	}
	if check.IfNil(args.Storages) {
		return nil, epochStart.ErrNilStorageService
	}
	if check.IfNil(args.EpochHandler) {
		return nil, dataRetriever.ErrNilEpochHandler
	}
	if check.IfNil(args.DataPools) {
		return nil, dataRetriever.ErrNilDataPoolHolder
	}
	if check.IfNil(args.DataPools.MetaBlocks()) {
		return nil, dataRetriever.ErrNilMetaBlockPool
	}
	if check.IfNil(args.DataPools.Transactions()) {
		return nil, dataRetriever.ErrNilTxDataPool
	}
	if check.IfNil(args.DataPools.RewardTransactions()) {
		return nil, dataRetriever.ErrNilRewardTransactionPool
	}
	if check.IfNil(args.DataPools.UnsignedTransactions()) {
		return nil, dataRetriever.ErrNilUnsignedTransactionPool
	}
	if check.IfNil(args.RequestHandler) {
		return nil, epochStart.ErrNilRequestHandler
	}
	if check.IfNil(args.HeaderValidator) {
		return nil, process.ErrNilHeaderValidator
	}

	ss := &syncState{
		hasher:                 args.Hasher,
		marshalizer:            args.Marshalizer,
		shardCoordinator:       args.ShardCoordinator,
		trieSyncers:            args.TrieSyncers,
		epochHandler:           args.EpochHandler,
		requestHandler:         args.RequestHandler,
		headerValidator:        args.HeaderValidator,
		tries:                  make(map[string]data.Trie),
		syncingEpoch:           0,
		activeAccountsAdapters: args.AccountHandlers,
	}

	ss.miniBlocks = pendingMiniBlocks{
		mutPendingMb:  sync.Mutex{},
		mapMiniBlocks: make(map[string]*block.MiniBlock),
		mapHashes:     make(map[string]struct{}),
		pool:          args.DataPools.MiniBlocks(),
		storage:       args.Storages.GetStorer(dataRetriever.MiniBlockUnit),
		chReceivedAll: make(chan bool),
	}
	if check.IfNil(ss.miniBlocks.storage) {
		return nil, update.ErrNilMiniBlocksStorage
	}

	ss.transactions = pendingTransactions{
		mutPendingTx:    sync.Mutex{},
		mapTransactions: make(map[string]data.TransactionHandler),
		mapHashes:       make(map[string]*block.MiniBlock),
		chReceivedAll:   make(chan bool),
	}

	ss.transactions.txPools = make(map[block.Type]dataRetriever.ShardedDataCacherNotifier)
	ss.transactions.txPools[block.TxBlock] = args.DataPools.Transactions()
	ss.transactions.txPools[block.SmartContractResultBlock] = args.DataPools.UnsignedTransactions()
	ss.transactions.txPools[block.RewardsBlock] = args.DataPools.RewardTransactions()

	ss.transactions.storage = make(map[block.Type]update.HistoryStorer)
	ss.transactions.storage[block.TxBlock] = args.Storages.GetStorer(dataRetriever.TransactionUnit)
	ss.transactions.storage[block.SmartContractResultBlock] = args.Storages.GetStorer(dataRetriever.UnsignedTransactionUnit)
	ss.transactions.storage[block.RewardsBlock] = args.Storages.GetStorer(dataRetriever.RewardTransactionUnit)

	for _, pool := range ss.transactions.txPools {
		pool.RegisterHandler(ss.receivedTransaction)
	}

	for _, store := range ss.transactions.storage {
		if check.IfNil(store) {
			return nil, epochStart.ErrNilStorage
		}
	}

	ss.transactions.chReceivedAll = make(chan bool)
	ss.miniBlocks.chReceivedAll = make(chan bool)

	ss.miniBlocks.pool.RegisterHandler(ss.receivedMiniBlock)

	return ss, nil
}

// SyncAllState gets an epoch number and will sync the complete data for that epoch start metablock
func (ss *syncState) SyncAllState(epoch uint32) error {
	if epoch == ss.syncingEpoch {
		return nil
	}
	ss.syncingEpoch = epoch

	meta, err := ss.headers.SyncEpochStartMetaHeader(epoch, time.Minute)
	if err != nil {
		return err
	}

	ss.syncingEpoch = meta.GetEpoch()

	wg := sync.WaitGroup{}
	wg.Add(len(meta.EpochStart.LastFinalizedHeaders) + 1)

	var errFound error
	mutErr := sync.Mutex{}

	go func() {
		errMeta := ss.syncMeta(meta, &wg)
		if errMeta != nil {
			mutErr.Lock()
			errFound = errMeta
			mutErr.Unlock()
		}
	}()

	for _, shData := range meta.EpochStart.LastFinalizedHeaders {
		go func(shardData block.EpochStartShardData) {
			err := ss.syncShard(shardData, &wg)
			if err != nil {
				mutErr.Lock()
				errFound = err
				mutErr.Unlock()
			}
		}(shData)
	}

	wg.Wait()

	if errFound != nil {
		return errFound
	}

	return nil
}

func (ss *syncState) syncShard(shardData block.EpochStartShardData, wg *sync.WaitGroup) error {
	defer wg.Done()
	err := ss.syncTrieOfType(factory.UserAccount, shardData.ShardId, shardData.RootHash)
	if err != nil {
		return err
	}

	_ = process.EmptyChannel(ss.transactions.chReceivedAll)
	_ = process.EmptyChannel(ss.miniBlocks.chReceivedAll)

	ss.miniBlocks.mutPendingMb.Lock()
	ss.transactions.mutPendingTx.Lock()

	requestedMBs := 0
	requestedTxs := 0
	for _, mbHeader := range shardData.PendingMiniBlockHeaders {
		ss.miniBlocks.mapHashes[string(mbHeader.Hash)] = struct{}{}
		miniBlock, ok := ss.getMiniBlockFromPoolOrStorage(mbHeader.Hash)
		if ok {
			ss.miniBlocks.mapMiniBlocks[string(mbHeader.Hash)] = miniBlock
			for _, txHash := range miniBlock.TxHashes {
				ss.transactions.mapHashes[string(txHash)] = miniBlock
			}
			requestedTxs += ss.requestTransactionsForMB(miniBlock)
			continue
		}

		requestedMBs++
		ss.requestHandler.RequestMiniBlock(mbHeader.SenderShardID, mbHeader.Hash)
	}

	ss.transactions.mutPendingTx.Unlock()
	ss.miniBlocks.mutPendingMb.Unlock()

	if requestedMBs > 0 {
		err := WaitFor(ss.miniBlocks.chReceivedAll, time.Hour)
		log.Warn("could not finish syncing", "error", err)
		if err != nil {
			return err
		}
	}
	if requestedTxs > 0 {
		err := WaitFor(ss.transactions.chReceivedAll, time.Hour)
		log.Warn("could not finish syncing", "error", err)
		if err != nil {
			return err
		}
	}

	return nil
}

// receivedMiniBlock is a callback function when a new miniblock was received
// it will further ask for missing transactions
func (ss *syncState) receivedMiniBlock(miniBlockHash []byte) {
	ss.miniBlocks.mutPendingMb.Lock()

	if _, ok := ss.miniBlocks.mapHashes[string(miniBlockHash)]; ok {
		ss.miniBlocks.mutPendingMb.Unlock()
		return
	}

	if _, ok := ss.miniBlocks.mapMiniBlocks[string(miniBlockHash)]; ok {
		ss.miniBlocks.mutPendingMb.Unlock()
		return
	}

	miniBlock, ok := ss.getMiniBlockFromPool(miniBlockHash)
	if !ok {
		ss.miniBlocks.mutPendingMb.Unlock()
		return
	}

	ss.miniBlocks.mapMiniBlocks[string(miniBlockHash)] = miniBlock
	_ = ss.requestTransactionsForMB(miniBlock)

	receivedAll := len(ss.miniBlocks.mapHashes) == len(ss.miniBlocks.mapMiniBlocks)
	ss.miniBlocks.mutPendingMb.Unlock()
	if receivedAll {
		ss.miniBlocks.chReceivedAll <- true
	}
}

func (ss *syncState) requestTransactionsForMB(miniBlock *block.MiniBlock) int {
	ss.transactions.mutPendingTx.Lock()
	defer ss.transactions.mutPendingTx.Unlock()

	missingTxs := make([][]byte, 0)
	for _, txHash := range miniBlock.TxHashes {
		if _, ok := ss.transactions.mapTransactions[string(txHash)]; ok {
			continue
		}

		tx, ok := ss.getTransactionFromPoolOrStorage(txHash)
		if ok {
			ss.transactions.mapTransactions[string(txHash)] = tx
			continue
		}

		missingTxs = append(missingTxs, txHash)
	}

	for _, txHash := range missingTxs {
		ss.transactions.mapHashes[string(txHash)] = miniBlock
	}

	switch miniBlock.Type {
	case block.TxBlock:
		ss.requestHandler.RequestTransaction(miniBlock.SenderShardID, missingTxs)
	case block.SmartContractResultBlock:
		ss.requestHandler.RequestUnsignedTransactions(miniBlock.SenderShardID, missingTxs)
	case block.RewardsBlock:
		ss.requestHandler.RequestRewardTransactions(miniBlock.SenderShardID, missingTxs)
	}

	return len(missingTxs)
}

// receivedMiniBlock is a callback function when a new transactions was received
func (ss *syncState) receivedTransaction(txHash []byte) {
	ss.transactions.mutPendingTx.Lock()

	if _, ok := ss.transactions.mapHashes[string(txHash)]; ok {
		ss.transactions.mutPendingTx.Unlock()
		return
	}
	if _, ok := ss.transactions.mapTransactions[string(txHash)]; ok {
		ss.transactions.mutPendingTx.Unlock()
		return
	}

	tx, ok := ss.getTransactionFromPool(txHash)
	if !ok {
		ss.transactions.mutPendingTx.Unlock()
		return
	}

	ss.transactions.mapTransactions[string(txHash)] = tx
	receivedAllMissing := len(ss.transactions.mapHashes) == len(ss.transactions.mapTransactions)
	ss.transactions.mutPendingTx.Unlock()

	if receivedAllMissing {
		ss.transactions.chReceivedAll <- true
	}
}

func (ss *syncState) getMiniBlockFromPool(hash []byte) (*block.MiniBlock, bool) {
	val, ok := ss.miniBlocks.pool.Peek(hash)
	if !ok {
		return nil, false
	}

	miniBlock, ok := val.(*block.MiniBlock)
	if !ok {
		return nil, false
	}

	return miniBlock, true
}

func (ss *syncState) getTransactionFromPool(txHash []byte) (data.TransactionHandler, bool) {
	mb := ss.transactions.mapHashes[string(txHash)]
	storeId := process.ShardCacherIdentifier(mb.SenderShardID, mb.ReceiverShardID)
	shardTxStore := ss.transactions.txPools[mb.Type].ShardDataStore(storeId)

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

func (ss *syncState) getMiniBlockFromPoolOrStorage(hash []byte) (*block.MiniBlock, bool) {
	miniBlock, ok := ss.getMiniBlockFromPool(hash)
	if ok {
		return miniBlock, true
	}

	mbData, err := GetDataFromStorage(hash, ss.miniBlocks.storage, ss.syncingEpoch)
	if err != nil {
		return nil, false
	}

	mb := &block.MiniBlock{}
	err = ss.marshalizer.Unmarshal(mb, mbData)
	if err != nil {
		return nil, false
	}

	return mb, true
}

func (ss *syncState) getTransactionFromPoolOrStorage(hash []byte) (data.TransactionHandler, bool) {
	txFromPool, ok := ss.getTransactionFromPool(hash)
	if ok {
		return txFromPool, true
	}

	miniBlock, ok := ss.transactions.mapHashes[string(hash)]
	if !ok {
		return nil, false
	}

	txData, err := GetDataFromStorage(hash, ss.transactions.storage[miniBlock.Type], ss.syncingEpoch)
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

	err = ss.marshalizer.Unmarshal(tx, txData)
	if err != nil {
		return nil, false
	}

	return tx, true
}

func (ss *syncState) GetMetaBlock() *block.MetaBlock {
	meta, err := ss.headers.GetMetaBlock()
	if err != nil {
		log.Debug("meta not synced yet")
		return nil
	}

	return meta
}

func (ss *syncState) GetAllTries() (map[string]data.Trie, error) {
	return ss.tries, nil
}

func (ss *syncState) GetAllTransactions() (map[string]data.TransactionHandler, error) {
	return ss.transactions.mapTransactions, nil
}

func (ss *syncState) GetAllMiniBlocks() (map[string]*block.MiniBlock, error) {
	return ss.miniBlocks.mapMiniBlocks, nil
}

// IsInterfaceNil returns if underlying objects in nil
func (ss *syncState) IsInterfaceNil() bool {
	return ss == nil
}
