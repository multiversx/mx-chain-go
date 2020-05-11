package txcache

import (
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ storage.Cacher = (*DisabledCache)(nil)

// DisabledCache represents a disabled cache
type DisabledCache struct {
}

// NewDisabledCache creates a new disabled cache
func NewDisabledCache() *DisabledCache {
	return &DisabledCache{}
}

// AddTx -
func (cache *DisabledCache) AddTx(tx *WrappedTransaction) (ok bool, added bool) {
	log.Error("DisabledCache.AddTx()")
	return false, false
}

// GetByTxHash -
func (cache *DisabledCache) GetByTxHash(txHash []byte) (*WrappedTransaction, bool) {
	log.Error("DisabledCache.GetByTxHash()")
	return nil, false
}

// SelectTransactions -
func (cache *DisabledCache) SelectTransactions(numRequested int, batchSizePerSender int) []*WrappedTransaction {
	log.Error("DisabledCache.SelectTransactions()")
	return make([]*WrappedTransaction, 0)
}

// RemoveTxByHash -
func (cache *DisabledCache) RemoveTxByHash(txHash []byte) error {
	log.Error("DisabledCache.RemoveTxByHash()")
	return nil
}

// CountTx -
func (cache *DisabledCache) CountTx() int64 {
	log.Error("DisabledCache.CountTx()")
	return 0
}

// Len -
func (cache *DisabledCache) Len() int {
	log.Error("DisabledCache.Len()")
	return 0
}

// ForEachTransaction -
func (cache *DisabledCache) ForEachTransaction(function ForEachTransaction) {
	log.Error("DisabledCache.ForEachTransaction()")
}

// Clear -
func (cache *DisabledCache) Clear() {
	log.Error("DisabledCache.Clear()")
}

// Put -
func (cache *DisabledCache) Put(key []byte, value interface{}) (evicted bool) {
	log.Error("DisabledCache.Put()")
	return false
}

// Get -
func (cache *DisabledCache) Get(key []byte) (value interface{}, ok bool) {
	tx, ok := cache.GetByTxHash(key)
	if ok {
		return tx.Tx, true
	}
	return nil, false
}

// Has -
func (cache *DisabledCache) Has(key []byte) bool {
	log.Error("DisabledCache.Has is not implemented")
	return false
}

// Peek -
func (cache *DisabledCache) Peek(key []byte) (value interface{}, ok bool) {
	log.Error("DisabledCache.DisabledCache()")
	return nil, false
}

// HasOrAdd -
func (cache *DisabledCache) HasOrAdd(key []byte, value interface{}) (ok, evicted bool) {
	log.Error("DisabledCache.HasOrAdd()")
	return false, false
}

// Remove -
func (cache *DisabledCache) Remove(key []byte) {
	log.Error("DisabledCache.Remove()")
}

// RemoveOldest -
func (cache *DisabledCache) RemoveOldest() {
	log.Error("DisabledCache.RemoveOldest()")
}

// Keys -
func (cache *DisabledCache) Keys() txHashes {
	log.Error("DisabledCache.Keys()")
	return make([][]byte, 0)
}

// MaxSize -
func (cache *DisabledCache) MaxSize() int {
	log.Error("DisabledCache.MaxSize()")
	return 0
}

// RegisterHandler -
func (cache *DisabledCache) RegisterHandler(func(key []byte, value interface{})) {
	log.Error("DisabledCache.RegisterHandler()")
}

// IsInterfaceNil returns true if there is no value under the interface
func (cache *DisabledCache) IsInterfaceNil() bool {
	return cache == nil
}
