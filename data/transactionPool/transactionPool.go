package transactionPool

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var defaultCacherConfig = &storage.CacheConfig{
	Size: 1000,
	Type: storage.LRUCache,
}

// TransactionPool holds the list of transactions organised by destination shard
//
// The miniPools field maps a cacher, containing transaction
//  hashes, to a corresponding shard id. It is able to add or remove
//  transactions given the shard id it is associated with. It can
//  also merge and split pools when required
type TransactionPool struct {
	lock sync.RWMutex
	// MiniPoolsStore is a key value store
	// Each key represents a destination shard id and the value will contain all
	//  transaction hashes that have that shard as destination
	miniPoolsStore map[uint32]*miniPool
	cacherConfig   *storage.CacheConfig

	handlerLock            sync.RWMutex
	addTransactionHandlers []func(txHash []byte)
}

type miniPool struct {
	ShardID uint32
	TxStore storage.Cacher
}

// NewTransactionPool is responsible for creating an empty pool of transactions
func NewTransactionPool(cacherConfig *storage.CacheConfig) *TransactionPool {
	if cacherConfig == nil {
		cacherConfig = defaultCacherConfig
	}
	return &TransactionPool{
		cacherConfig:   cacherConfig,
		miniPoolsStore: make(map[uint32]*miniPool),
	}
}

// NewMiniPool is responsible for creating an empty mini pool
func NewMiniPool(destShardID uint32, cacherConfig *storage.CacheConfig) *miniPool {
	if cacherConfig == nil {
		cacherConfig = defaultCacherConfig
	}
	cacher, err := storage.NewCache(cacherConfig.Type, cacherConfig.Size)
	if err != nil {
		// TODO: This should be replaced with the correct log panic
		panic("Could not create cache storage for pools")
	}
	return &miniPool{
		destShardID,
		cacher,
	}
}

// NewMiniPool is a TransactionPool method that is responsible for creating
//  a new mini pool at the destShardID index in the MiniPoolsStore map
func (tp *TransactionPool) NewMiniPool(destShardID uint32) {
	tp.lock.Lock()
	tp.miniPoolsStore[destShardID] = NewMiniPool(destShardID, tp.cacherConfig)
	tp.lock.Unlock()
}

// MiniPool returns a minipool of transactions associated with a given destination shardID
func (tp *TransactionPool) MiniPool(shardID uint32) *miniPool {
	tp.lock.RLock()
	mp := tp.miniPoolsStore[shardID]
	tp.lock.RUnlock()
	return mp
}

// MiniPoolTxStore returns the minipool transaction store containing transaction hashes
//  associated with a given destination shardID
func (tp *TransactionPool) MiniPoolTxStore(shardID uint32) (c storage.Cacher) {
	mp := tp.MiniPool(shardID)
	if mp == nil {
		return nil
	}
	return mp.TxStore
}

// AddTransaction will add a transaction to the corresponding pool
func (tp *TransactionPool) AddTransaction(txHash []byte, tx *transaction.Transaction, destShardID uint32) {
	if tp.MiniPool(destShardID) == nil {
		tp.NewMiniPool(destShardID)
	}
	mp := tp.MiniPoolTxStore(destShardID)
	found, _ := mp.HasOrAdd(txHash, tx)

	if tp.addTransactionHandlers != nil && !found {
		for _, handler := range tp.addTransactionHandlers {
			go handler(txHash)
		}
	}
}

// RemoveTransactionsFromPool removes a list of transactions from the corresponding pool
func(tp *TransactionPool) RemoveTransactionsFromPool(txHashes [][]byte, destShardID uint32){
	for _, txHash:=range txHashes{
		tp.RemoveTransaction(txHash, destShardID)
	}
}

// RemoveTransaction will remove a transaction hash from the corresponding pool
func (tp *TransactionPool) RemoveTransaction(txHash []byte, destShardID uint32) {
	mptx := tp.MiniPoolTxStore(destShardID)
	if mptx != nil {
		mptx.Remove(txHash)
	}
}

// RemoveTransactionFromAllShards will remove a transaction hash from the pool given only
//  the transaction hash. It will itearate over all mini pools and remove it everywhere
func (tp *TransactionPool) RemoveTransactionFromAllShards(txHash []byte) {
	for k := range tp.miniPoolsStore {
		m := tp.MiniPoolTxStore(k)
		if m.Has(txHash) {
			m.Remove(txHash)
		}
	}
}

// MergeMiniPools will take all transactions associated with the sourceShardId and move them
// to the destShardID. It will then remove the sourceShardID key from the store map
func (tp *TransactionPool) MergeMiniPools(sourceShardID, destShardID uint32) {
	sourceStore := tp.MiniPoolTxStore(sourceShardID)

	if sourceStore != nil {
		for _, txHash := range sourceStore.Keys() {
			tx, _ := sourceStore.Get(txHash)
			tp.AddTransaction(txHash, tx.(*transaction.Transaction), destShardID)
		}
	}

	tp.lock.Lock()
	delete(tp.miniPoolsStore, sourceShardID)
	tp.lock.Unlock()
}

// MoveTransactions will move all given transactions associated with the sourceShardId to the destShardId
func (tp *TransactionPool) MoveTransactions(sourceShardID, destShardID uint32, txHashes [][]byte) {
	sourceStore := tp.MiniPoolTxStore(sourceShardID)

	if sourceStore != nil {
		for _, txHash := range txHashes {
			tx, _ := sourceStore.Get(txHash)
			tp.AddTransaction(txHash, tx.(*transaction.Transaction), destShardID)
			tp.RemoveTransaction(txHash, sourceShardID)
		}
	}
}

// Clear will delete all minipools and associated transactions
func (tp *TransactionPool) Clear() {
	for m := range tp.miniPoolsStore {
		tp.lock.Lock()
		delete(tp.miniPoolsStore, m)
		tp.lock.Unlock()
	}
}

// ClearMiniPool will delete all transactions associated with a given destination shardID
func (tp *TransactionPool) ClearMiniPool(shardID uint32) {
	mp := tp.MiniPoolTxStore(shardID)
	if mp == nil {
		return
	}
	mp.Clear()
}

// ClearMiniPool will delete all transactions associated with a given destination shardID
func (tp *TransactionPool) RegisterTransactionHandler(transactionHandler func(txHash []byte)) {
	tp.handlerLock.Lock()

	if tp.addTransactionHandlers == nil {
		tp.addTransactionHandlers = make([]func(txHash []byte), 0)
	}

	tp.addTransactionHandlers = append(tp.addTransactionHandlers, transactionHandler)

	tp.handlerLock.Unlock()
}
