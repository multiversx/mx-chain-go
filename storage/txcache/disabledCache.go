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

// AddTx does nothing
func (cache *DisabledCache) AddTx(tx *WrappedTransaction) (ok bool, added bool) {
	return false, false
}

// GetByTxHash returns no transaction
func (cache *DisabledCache) GetByTxHash(txHash []byte) (*WrappedTransaction, bool) {
	return nil, false
}

// SelectTransactions returns an empty slice
func (cache *DisabledCache) SelectTransactions(numRequested int, batchSizePerSender int) []*WrappedTransaction {
	return make([]*WrappedTransaction, 0)
}

// RemoveTxByHash does nothing
func (cache *DisabledCache) RemoveTxByHash(txHash []byte) error {
	return nil
}

// CountTx returns zero
func (cache *DisabledCache) CountTx() int64 {
	return 0
}

// Len returns zero
func (cache *DisabledCache) Len() int {
	return 0
}

// ForEachTransaction does nothing
func (cache *DisabledCache) ForEachTransaction(function ForEachTransaction) {
}

// Clear does nothing
func (cache *DisabledCache) Clear() {
}

// Put does nothing
func (cache *DisabledCache) Put(key []byte, value interface{}) (evicted bool) {
	return false
}

// Get returns no transaction
func (cache *DisabledCache) Get(key []byte) (value interface{}, ok bool) {
	return nil, false
}

// Has returns false
func (cache *DisabledCache) Has(key []byte) bool {
	return false
}

// Peek returns no transaction
func (cache *DisabledCache) Peek(key []byte) (value interface{}, ok bool) {
	return nil, false
}

// HasOrAdd returns false, does nothing
func (cache *DisabledCache) HasOrAdd(key []byte, value interface{}) (ok, evicted bool) {
	return false, false
}

// Remove does nothing
func (cache *DisabledCache) Remove(key []byte) {
}

// RemoveOldest does nothing
func (cache *DisabledCache) RemoveOldest() {
}

// Keys returns an empty slice
func (cache *DisabledCache) Keys() txHashes {
	return make([][]byte, 0)
}

// MaxSize returns zero
func (cache *DisabledCache) MaxSize() int {
	return 0
}

// RegisterHandler does nothing
func (cache *DisabledCache) RegisterHandler(func(key []byte, value interface{})) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (cache *DisabledCache) IsInterfaceNil() bool {
	return cache == nil
}
