package txcachemocks

import "github.com/multiversx/mx-chain-storage-go/txcache"

// TxCacheMock -
type TxCacheMock struct {
	ClearCalled             func()
	PutCalled               func(key []byte, value interface{}, sizeInBytes int) (evicted bool)
	GetCalled               func(key []byte) (value interface{}, ok bool)
	HasCalled               func(key []byte) bool
	PeekCalled              func(key []byte) (value interface{}, ok bool)
	HasOrAddCalled          func(key []byte, value interface{}, sizeInBytes int) (has, added bool)
	RemoveCalled            func(key []byte)
	RemoveOldestCalled      func()
	KeysCalled              func() [][]byte
	LenCalled               func() int
	MaxSizeCalled           func() int
	RegisterHandlerCalled   func(func(key []byte, value interface{}))
	UnRegisterHandlerCalled func(id string)
	CloseCalled             func() error

	AddTxCalled                        func(tx *txcache.WrappedTransaction) (ok bool, added bool)
	GetByTxHashCalled                  func(txHash []byte) (*txcache.WrappedTransaction, bool)
	RemoveTxByHashCalled               func(txHash []byte) bool
	ImmunizeTxsAgainstEvictionCalled   func(keys [][]byte)
	ForEachTransactionCalled           func(txcache.ForEachTransaction)
	NumBytesCalled                     func() int
	DiagnoseCalled                     func(deep bool)
	GetTransactionsPoolForSenderCalled func(sender string) []*txcache.WrappedTransaction
}

// NewTxCacheStub -
func NewTxCacheStub() *TxCacheMock {
	return &TxCacheMock{}
}

// Clear -
func (cache *TxCacheMock) Clear() {
	if cache.ClearCalled != nil {
		cache.ClearCalled()
	}
}

// Put -
func (cache *TxCacheMock) Put(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
	if cache.PutCalled != nil {
		return cache.PutCalled(key, value, sizeInBytes)
	}

	return false
}

// Get -
func (cache *TxCacheMock) Get(key []byte) (value interface{}, ok bool) {
	if cache.GetCalled != nil {
		return cache.GetCalled(key)
	}

	return nil, false
}

// Has -
func (cache *TxCacheMock) Has(key []byte) bool {
	if cache.HasCalled != nil {
		return cache.HasCalled(key)
	}

	return false
}

// Peek -
func (cache *TxCacheMock) Peek(key []byte) (value interface{}, ok bool) {
	if cache.PeekCalled != nil {
		return cache.PeekCalled(key)
	}

	return nil, false
}

// HasOrAdd -
func (cache *TxCacheMock) HasOrAdd(key []byte, value interface{}, sizeInBytes int) (has, added bool) {
	if cache.HasOrAddCalled != nil {
		return cache.HasOrAddCalled(key, value, sizeInBytes)
	}

	return false, false
}

// Remove -
func (cache *TxCacheMock) Remove(key []byte) {
	if cache.RemoveCalled != nil {
		cache.RemoveCalled(key)
	}
}

// Keys -
func (cache *TxCacheMock) Keys() [][]byte {
	if cache.KeysCalled != nil {
		return cache.KeysCalled()
	}

	return make([][]byte, 0)
}

// Len -
func (cache *TxCacheMock) Len() int {
	if cache.LenCalled != nil {
		return cache.LenCalled()
	}

	return 0
}

// SizeInBytesContained -
func (cache *TxCacheMock) SizeInBytesContained() uint64 {
	return 0
}

// MaxSize -
func (cache *TxCacheMock) MaxSize() int {
	if cache.MaxSizeCalled != nil {
		return cache.MaxSizeCalled()
	}

	return 0
}

// RegisterHandler -
func (cache *TxCacheMock) RegisterHandler(handler func(key []byte, value interface{}), _ string) {
	if cache.RegisterHandlerCalled != nil {
		cache.RegisterHandlerCalled(handler)
	}
}

// UnRegisterHandler -
func (cache *TxCacheMock) UnRegisterHandler(id string) {
	if cache.UnRegisterHandlerCalled != nil {
		cache.UnRegisterHandlerCalled(id)
	}
}

// Close -
func (cache *TxCacheMock) Close() error {
	if cache.CloseCalled != nil {
		return cache.CloseCalled()
	}

	return nil
}

// AddTx -
func (cache *TxCacheMock) AddTx(tx *txcache.WrappedTransaction) (ok bool, added bool) {
	if cache.AddTxCalled != nil {
		return cache.AddTxCalled(tx)
	}

	return false, false
}

// GetByTxHash -
func (cache *TxCacheMock) GetByTxHash(txHash []byte) (*txcache.WrappedTransaction, bool) {
	if cache.GetByTxHashCalled != nil {
		return cache.GetByTxHashCalled(txHash)
	}

	return nil, false
}

// RemoveTxByHash -
func (cache *TxCacheMock) RemoveTxByHash(txHash []byte) bool {
	if cache.RemoveTxByHashCalled != nil {
		return cache.RemoveTxByHashCalled(txHash)
	}

	return false
}

// ImmunizeTxsAgainstEviction -
func (cache *TxCacheMock) ImmunizeTxsAgainstEviction(keys [][]byte) {
	if cache.ImmunizeTxsAgainstEvictionCalled != nil {
		cache.ImmunizeTxsAgainstEvictionCalled(keys)
	}
}

// ForEachTransaction -
func (cache *TxCacheMock) ForEachTransaction(fn txcache.ForEachTransaction) {
	if cache.ForEachTransactionCalled != nil {
		cache.ForEachTransactionCalled(fn)
	}
}

// NumBytes -
func (cache *TxCacheMock) NumBytes() int {
	if cache.NumBytesCalled != nil {
		return cache.NumBytesCalled()
	}

	return 0
}

// Diagnose -
func (cache *TxCacheMock) Diagnose(deep bool) {
	if cache.DiagnoseCalled != nil {
		cache.DiagnoseCalled(deep)
	}
}

// GetTransactionsPoolForSender -
func (cache *TxCacheMock) GetTransactionsPoolForSender(sender string) []*txcache.WrappedTransaction {
	if cache.GetTransactionsPoolForSenderCalled != nil {
		return cache.GetTransactionsPoolForSenderCalled(sender)
	}

	return nil
}

// IsInterfaceNil -
func (cache *TxCacheMock) IsInterfaceNil() bool {
	return cache == nil
}
