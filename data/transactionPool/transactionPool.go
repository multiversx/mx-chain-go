package transactionPool

import (
	"sync"

	"errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/config"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

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

	marsher marshal.Marshalizer
	hasher  hashing.Hasher
}

type miniPool struct {
	ShardID uint32
	TxStore storage.Cacher
}

// NewTransactionPool is responsible for creating an empty pool of transactions
func NewTransactionPool(marsher marshal.Marshalizer, hasher hashing.Hasher) *TransactionPool {
	return &TransactionPool{
		miniPoolsStore: make(map[uint32]*miniPool),
		marsher:        marsher,
		hasher:         hasher,
	}
}

// NewMiniPool is responsible for creating an empty mini pool
func NewMiniPool(destShardID uint32) *miniPool {
	cacher, err := storage.CreateCacheFromConf(config.TestnetBlockchainConfig.TxPoolStorage)
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
	tp.miniPoolsStore[destShardID] = NewMiniPool(destShardID)
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

// ReceivedTransaction handle the received transactions and add them to the specific pools
func (tp *TransactionPool) ReceivedTransaction(txHash []byte, tx *transaction.Transaction, destShardId uint32) error {
	if tp.marsher == nil {
		return errors.New("marsher not defined")
	}

	buff, err := tp.marsher.Marshal(tx)

	if err != nil {
		return errors.New("unable to marshalize")
	}

	if txHash == nil {
		if tp.hasher == nil {
			return errors.New("hasher not defined")
		}

		txHash = tp.hasher.Compute(string(buff))
	}

	tp.AddTransaction(txHash, buff, destShardId)

	return nil
}

// AddTransaction will add a transaction to the coresponding pool
func (tp *TransactionPool) AddTransaction(txHash []byte, tx []byte, destShardID uint32) {
	if tp.MiniPool(destShardID) == nil {
		tp.NewMiniPool(destShardID)
	}
	mp := tp.MiniPoolTxStore(destShardID)
	mp.HasOrAdd(txHash, tx)
}

// RemoveTransaction will remove a transaction hash from the coresponding pool
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
			tp.AddTransaction(txHash, tx, destShardID)
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
			tp.AddTransaction(txHash, tx, destShardID)
			tp.RemoveTransaction(txHash, sourceShardID)
		}
	}
}

// Clear will delete all minipools and associated transactions
func (tp *TransactionPool) Clear() {
	tp.lock.Lock()
	for m := range tp.miniPoolsStore {
		delete(tp.miniPoolsStore, m)
	}
	tp.lock.Unlock()
}

// ClearMiniPool will delete all transactions associated with a given destination shardID
func (tp *TransactionPool) ClearMiniPool(shardID uint32) {
	mp := tp.MiniPoolTxStore(shardID)
	if mp == nil {
		return
	}
	mp.Clear()
}
