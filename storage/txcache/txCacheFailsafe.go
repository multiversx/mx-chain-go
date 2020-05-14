package txcache

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ storage.Cacher = (*txCacheFailsafeDecorator)(nil)
var _ txCache = (*txCacheFailsafeDecorator)(nil)

// txCacheFailsafeDecorator is a decorator over a cache, a decorator which wraps some operations in try / catch constructs
type txCacheFailsafeDecorator struct {
	backingCache txCache
}

func newTxCacheFailsafe(cache txCache) txCache {
	return &txCacheFailsafeDecorator{
		backingCache: cache,
	}
}

// AddTx delegates to backing cache, with try / catch
func (decorator *txCacheFailsafeDecorator) AddTx(tx *WrappedTransaction) (ok bool, added bool) {
	core.TryCatch(func() {
		ok, added = decorator.backingCache.AddTx(tx)
	}, decorator.catchAndLog, "AddTx")

	return
}

// GetByTxHash delegates to backing cache
func (decorator *txCacheFailsafeDecorator) GetByTxHash(txHash []byte) (*WrappedTransaction, bool) {
	return decorator.backingCache.GetByTxHash(txHash)
}

// SelectTransactions deledates to backing cache, with try / cache
func (decorator *txCacheFailsafeDecorator) SelectTransactions(numRequested int, batchSizePerSender int) []*WrappedTransaction {
	selection := make([]*WrappedTransaction, 0)

	core.TryCatch(func() {
		selection = decorator.backingCache.SelectTransactions(numRequested, batchSizePerSender)
	}, decorator.catchAndLog, "SelectTransactions")

	return selection
}

// RemoveTxByHash delegates to backing cache, with try / cache
func (decorator *txCacheFailsafeDecorator) RemoveTxByHash(txHash []byte) error {
	var err error

	core.TryCatch(func() {
		err = decorator.backingCache.RemoveTxByHash(txHash)
	}, decorator.catchAndLog, "RemoveTxByHash")

	return err
}

// CountTx delegates to backing cache
func (decorator *txCacheFailsafeDecorator) CountTx() int64 {
	return decorator.backingCache.CountTx()
}

// Len delegates to backing cache
func (decorator *txCacheFailsafeDecorator) Len() int {
	return decorator.backingCache.Len()
}

// ForEachTransaction delegates to backing cache
func (decorator *txCacheFailsafeDecorator) ForEachTransaction(function ForEachTransaction) {
	decorator.backingCache.ForEachTransaction(function)
}

// Clear delegates to backing cache
func (decorator *txCacheFailsafeDecorator) Clear() {
	decorator.backingCache.Clear()
}

// Put delegates to backing cache
func (decorator *txCacheFailsafeDecorator) Put(key []byte, value interface{}) (evicted bool) {
	return decorator.backingCache.Put(key, value)
}

// Get delegates to backing cache
func (decorator *txCacheFailsafeDecorator) Get(key []byte) (value interface{}, ok bool) {
	return decorator.backingCache.Get(key)
}

// Has delegates to backing cache
func (decorator *txCacheFailsafeDecorator) Has(key []byte) bool {
	return decorator.backingCache.Has(key)
}

// Peek delegates to backing cache
func (decorator *txCacheFailsafeDecorator) Peek(key []byte) (value interface{}, ok bool) {
	return decorator.backingCache.Peek(key)
}

// HasOrAdd delegates to backing cache
func (decorator *txCacheFailsafeDecorator) HasOrAdd(key []byte, value interface{}) (ok, evicted bool) {
	return decorator.backingCache.HasOrAdd(key, value)
}

// Remove delegates to backing cache
func (decorator *txCacheFailsafeDecorator) Remove(key []byte) {
	decorator.RemoveTxByHash(key)
}

// RemoveOldest delegates to backing cache
func (decorator *txCacheFailsafeDecorator) RemoveOldest() {
	decorator.backingCache.RemoveOldest()
}

// Keys delegates to backing cache
func (decorator *txCacheFailsafeDecorator) Keys() txHashes {
	return decorator.backingCache.Keys()
}

// MaxSize delegates to backing cache
func (decorator *txCacheFailsafeDecorator) MaxSize() int {
	return decorator.backingCache.MaxSize()
}

// RegisterHandler delegates to backing cache
func (decorator *txCacheFailsafeDecorator) RegisterHandler(function func(key []byte, value interface{})) {
	decorator.backingCache.RegisterHandler(function)
}

// IsInterfaceNil returns true if there is no value under the interface
func (decorator *txCacheFailsafeDecorator) IsInterfaceNil() bool {
	return decorator == nil
}

func (decorator *txCacheFailsafeDecorator) catchAndLog(err error) {
	log.Error("txCacheFailsafe.catchAndLog()", "err", err)
}
