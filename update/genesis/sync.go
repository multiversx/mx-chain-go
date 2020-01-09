package genesis

import (
	"bytes"
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
	storage         update.HistoryStorer
	chReceivedAll   chan bool
}

type syncState struct {
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	shardCoordinator sharding.Coordinator
	trieSyncers      update.TrieSyncContainer
	metaBlockCache   storage.Cacher
	metaBlockStorage storage.Storer
	epochHandler     update.EpochHandler
	pruningStorer    update.HistoryStorer
	requestHandler   process.RequestHandler

	tries           map[string]data.Trie
	lastSyncedEpoch uint32

	activeAccountsAdapters update.AccountsHandlerContainer

	miniBlocks   pendingMiniBlocks
	transactions pendingTransactions
}

// Arguments for the NewSync
type ArgsNewSyncState struct {
	ShardCoordinator sharding.Coordinator
	TrieSyncers      update.TrieSyncContainer
	EpochHandler     update.EpochHandler
	Storages         dataRetriever.StorageService
}

// NewSyncState creates a complete syncer which saves the state of the blockchain with pending values as well
func NewSyncState(args ArgsNewSyncState) (*syncState, error) {
	if check.IfNil(args.ShardCoordinator) {
		return nil, data.ErrNilShardCoordinator
	}
	if check.IfNil(args.TrieSyncers) {
		return nil, dataRetriever.ErrNilResolverContainer
	}

	ss := &syncState{
		shardCoordinator: args.ShardCoordinator,
		trieSyncers:      args.TrieSyncers,
		lastSyncedEpoch:  0,
	}

	for _, pool := range ss.transactions.txPools {
		pool.RegisterHandler(ss.receivedTransaction)
	}

	ss.transactions.chReceivedAll = make(chan bool)
	ss.miniBlocks.chReceivedAll = make(chan bool)

	ss.miniBlocks.pool.RegisterHandler(ss.receivedMiniBlock)

	return ss, nil
}

// SyncAllState gets an epoch number and will sync the complete data for that epoch start metablock
func (ss *syncState) SyncAllState(epoch uint32) error {
	if epoch == ss.lastSyncedEpoch {
		return nil
	}

	meta, err := ss.getEpochStartMetaHeader(epoch)
	if err != nil {
		return err
	}

	ss.lastSyncedEpoch = meta.GetEpoch()

	wg := sync.WaitGroup{}
	wg.Add(len(meta.EpochStart.LastFinalizedHeaders) + 1)

	go ss.syncMeta(meta, &wg)

	for _, shardData := range meta.EpochStart.LastFinalizedHeaders {
		go ss.syncShard(shardData, &wg)
	}

	wg.Wait()

	return nil
}

func (ss *syncState) syncShard(shardData block.EpochStartShardData, wg *sync.WaitGroup) {
	ss.syncTrieOfType(factory.UserAccount, shardData.ShardId, shardData.RootHash)

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

	if requestedMBs > 0 || requestedTxs > 0 {
		err := ss.waitForMBsAndTxs(time.Hour)
		log.Warn("could not finish syncing", "error", err)
	}

	wg.Done()
}

func (ss *syncState) waitForMBsAndTxs(waitTime time.Duration) error {
	select {
	case <-ss.miniBlocks.chReceivedAll:
		break
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}

	select {
	case <-ss.transactions.chReceivedAll:
		break
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}

	return nil
}

func (ss *syncState) syncTrieOfType(accountType factory.Type, shardId uint32, rootHash []byte) {
	accAdapterIdentifier := update.CreateTrieIdentifier(shardId, accountType)

	success := ss.tryRecreateTrie(accAdapterIdentifier, rootHash)
	if success {
		return
	}

	trieSyncer, err := ss.trieSyncers.Get(accAdapterIdentifier)
	if err != nil {
		// critical error - should not happen - maybe recreate trie syncer here
		return
	}

	err = trieSyncer.StartSyncing(rootHash)
	if err != nil {
		// critical error - should not happen - maybe recreate trie syncer here
		return
	}

	ss.tries[accAdapterIdentifier] = trieSyncer.Trie()
}

func (ss *syncState) tryRecreateTrie(id string, rootHash []byte) bool {
	savedTrie, ok := ss.tries[id]
	if ok {
		currHash, err := savedTrie.Root()
		if err == nil && bytes.Equal(currHash, rootHash) {
			return true
		}
	}

	accounts, err := ss.activeAccountsAdapters.Get(id)
	if err != nil {
		return false
	}

	trie, err := accounts.CopyRecreateTrie(rootHash)
	if err != nil {
		return false
	}

	err = trie.Commit()
	if err != nil {
		return false
	}

	ss.tries[id] = trie
	return true
}

func (ss *syncState) syncMeta(meta *block.MetaBlock, wg *sync.WaitGroup) {
	ss.syncTrieOfType(factory.UserAccount, sharding.MetachainShardId, meta.RootHash)
	ss.syncTrieOfType(factory.ValidatorAccount, sharding.MetachainShardId, meta.ValidatorStatsRootHash)

	wg.Done()
}

func (ss *syncState) getEpochStartMetaHeader(epoch uint32) (*block.MetaBlock, error) {
	return nil, nil
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

	mbData, err := ss.getDataFromStorage(hash, ss.miniBlocks.storage)
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

func (ss *syncState) getDataFromStorage(hash []byte, storer update.HistoryStorer) ([]byte, error) {
	var err error
	currData := make([]byte, 0)
	currData, err = storer.Get(hash)
	if err != nil {
		currData, err = storer.GetFromEpoch(hash, ss.lastSyncedEpoch)
		if err != nil {
			currData, err = storer.GetFromEpoch(hash, ss.lastSyncedEpoch-1)
		}
	}

	return currData, err
}

func (ss *syncState) getTransactionFromPoolOrStorage(hash []byte) (data.TransactionHandler, bool) {
	txFromPool, ok := ss.getTransactionFromPool(hash)
	if ok {
		return txFromPool, true
	}

	txData, err := ss.getDataFromStorage(hash, ss.transactions.storage)
	if err != nil {
		return nil, false
	}

	miniBlock, ok := ss.transactions.mapHashes[string(hash)]
	if !ok {
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
