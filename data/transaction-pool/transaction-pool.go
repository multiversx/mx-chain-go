package transaction_pool

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/config"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// TransactionPool holds the list of transactions organised by destination shard
//
// Through the MiniPools store it maps a a cacher, containing transaction
//  hashes, to a corresponding shard id. It is able to add or remove
//  transactions given the shard id it is associated with. It can
//  also merge and split pools when required
type TransactionPool struct {
	lock 			sync.Mutex
	// MiniPools is a key value store
	// Each key represents a destination shard id and the value will contain all
	//  transaction hashes that have that shard as destination
	MiniPoolsStore 	map[uint32]*miniPool
}

type miniPool struct {
	ShardID       uint32
	TxHashStore *storage.Cacher
}

// NewTransactionPool is responsible for creating an empty pool of transactions
func NewTransactionPool() (*TransactionPool) {
	return &TransactionPool{
		MiniPoolsStore: make(map[uint32]*miniPool),
	}
}

// NewMiniPool is responsible for creating an empty mini pool
func NewMiniPool(destShardID uint32) (*miniPool) {
	cacher, err := storage.CreateCacheFromConf(config.TestnetBlockchainConfig.TxPoolStorage)
	if err != nil {
		// TODO: This should be replaced with the correct log panic
		panic("Could not create cache storage for pools")
	}
	return &miniPool{
		destShardID,
		&cacher,
	}
}

// NewMiniPool is a TransactionPool method that is responsible for creating
//  a new mini pool at the destShardID index in the MiniPoolsStore map
func (tp *TransactionPool) NewMiniPool(destShardID uint32) {
	tp.lock.Lock()
	tp.MiniPoolsStore[destShardID] = NewMiniPool(destShardID)
	tp.lock.Unlock()
}

// GetMiniPool returns a minipool of transactions associated with a given destination shardID
func (tp *TransactionPool) GetMiniPool(shardID uint32) (*miniPool) {
	return tp.MiniPoolsStore[shardID]
}

// GetMiniPoolTxStore returns the minipool transaction store containing transaction hashes
//  associated with a given destination shardID
func (tp *TransactionPool) GetMiniPoolTxStore(shardID uint32) (c *storage.Cacher) {
	mp := tp.GetMiniPool(shardID)
	if mp == nil {
		return nil
	}
	return (*mp).TxHashStore
}

// AddTransaction will add a transaction hash to the coresponding pool
// Returns true if the transaction was added, false if it already existed
func (tp *TransactionPool) AddTransaction(txHash []byte, destShardID uint32) {
	if tp.MiniPoolsStore[destShardID] == nil {
		tp.NewMiniPool(destShardID)
	}
	mp := *tp.GetMiniPoolTxStore(destShardID)
	mp.HasOrAdd(txHash, nil)
}

// RemoveTransaction will remove a transaction hash from the coresponding pool
func (tp *TransactionPool) RemoveTransaction(txHash []byte, destShardID uint32) {
	if tp.MiniPoolsStore[destShardID] != nil {
		(*tp.GetMiniPoolTxStore(destShardID)).Remove(txHash)
	}
}

// FindAndRemoveTransaction will remove a transaction hash from the pool given only
//  the transaction hash. It will itearate over all mini pools and remove it everywhere
func (tp *TransactionPool) FindAndRemoveTransaction(txHash []byte) {
	for k := range tp.MiniPoolsStore {
		m := *tp.GetMiniPoolTxStore(k)
		if m.Has(txHash) {
			m.Remove(txHash)
		}
	}
}

// MergeMiniPools will take all transactions associated with the mergedShardId and move them
//  to the destShardID. It will then remove the mergedShardID key from the store map
func (tp *TransactionPool) MergeMiniPools(destShardID, mergedShardID uint32) {
	mergedStore := *tp.GetMiniPoolTxStore(mergedShardID)
	mergedStoreTxs := mergedStore.Keys()
	for _, tx := range mergedStoreTxs {
		tp.AddTransaction(tx, destShardID)
	}
	delete(tp.MiniPoolsStore, mergedShardID)
}

// SplitMiniPool will move all given transactions to a new minipool and remove them from the old one
func (tp *TransactionPool) SplitMiniPool(shardID, newShardID uint32, txs[][]byte) {
	for _, tx := range txs {
		tp.RemoveTransaction(tx, shardID)
		tp.AddTransaction(tx, newShardID)
	}
}

// Clear will delete all minipools and associated transactions
func (tp *TransactionPool) Clear() {
	for m := range tp.MiniPoolsStore {
		delete(tp.MiniPoolsStore, m)
	}
}

// ClearMiniPool will delete all transactions associated with a given destination shardID
func (tp *TransactionPool) ClearMiniPool(shardID uint32) {
	mp := tp.GetMiniPoolTxStore(shardID)
	if mp == nil {
		return
	}
	(*mp).Clear()
}