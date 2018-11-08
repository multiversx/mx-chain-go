package transaction_pool

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/config"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

type TransactionPool struct {
	lock 			sync.Mutex
	// MiniPools is a key value store
	// Each key represents a destination shard id and the value will contain all
	//  transaction hashes that have that shard as destination
	MiniPoolsStore 	map[uint32]*MiniPool
}

type MiniPool struct {
	ShardID       uint32
	TxHashStore *storage.Cacher
}

// NewTransactionPool is responsible for creating an empty pool of transactions
func NewTransactionPool() (*TransactionPool) {
	return &TransactionPool{
		MiniPoolsStore: make(map[uint32]*MiniPool),
	}
}

// NewMiniPool is responsible for creating an empty mini pool
func NewMiniPool(destShardID uint32) (*MiniPool) {
	cacher, _ := storage.CreateCacheFromConf(config.TestnetBlockchainConfig.TxPoolStorage)
	return &MiniPool{
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

// AddTransaction will add a transaction hash to the coresponding pool
// Returns true if the transaction was added, false if it already existed
func (tp *TransactionPool) AddTransaction(txHash []byte, destShardID uint32) (bool) {
	if tp.MiniPoolsStore[destShardID] == nil {
		tp.NewMiniPool(destShardID)
	}
	mp := *tp.GetMiniPoolTxStore(destShardID)
	if !mp.Has(txHash) {
		mp.Put(txHash, nil)
		return true
	}
	return false
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

// GetMiniPool returns a minipool of transactions associated with a given destination shardID
func (tp *TransactionPool) GetMiniPool(shardID uint32) (*MiniPool) {
	return tp.MiniPoolsStore[shardID]
}

// GetMiniPoolTxStore returns the minipool transaction store containing transaction hashes
//  associated with a given destination shardID
func (tp *TransactionPool) GetMiniPoolTxStore(shardID uint32) (c *storage.Cacher) {
	mp := *tp.GetMiniPool(shardID)
	return mp.TxHashStore
}