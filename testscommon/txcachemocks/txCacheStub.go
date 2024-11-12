package txcachemocks

import "github.com/multiversx/mx-chain-go/storage/txcache"

// TxCacheStub -
type TxCacheStub struct {
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
	NotifyAccountNonceCalled           func(accountKey []byte, nonce uint64)
	ForgetAllAccountNoncesCalled       func()
	GetByTxHashCalled                  func(txHash []byte) (*txcache.WrappedTransaction, bool)
	RemoveTxByHashCalled               func(txHash []byte) bool
	ImmunizeTxsAgainstEvictionCalled   func(keys [][]byte)
	ForEachTransactionCalled           func(txcache.ForEachTransaction)
	NumBytesCalled                     func() int
	DiagnoseCalled                     func(deep bool)
	GetTransactionsPoolForSenderCalled func(sender string) []*txcache.WrappedTransaction
}

// NewTxCacheStub -
func NewTxCacheStub() *TxCacheStub {
	return &TxCacheStub{}
}

// Clear -
func (cache *TxCacheStub) Clear() {
	if cache.ClearCalled != nil {
		cache.ClearCalled()
	}
}

// Put -
func (cache *TxCacheStub) Put(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
	if cache.PutCalled != nil {
		return cache.PutCalled(key, value, sizeInBytes)
	}

	return false
}

// Get -
func (cache *TxCacheStub) Get(key []byte) (value interface{}, ok bool) {
	if cache.GetCalled != nil {
		return cache.GetCalled(key)
	}

	return nil, false
}

// Has -
func (cache *TxCacheStub) Has(key []byte) bool {
	if cache.HasCalled != nil {
		return cache.HasCalled(key)
	}

	return false
}

// Peek -
func (cache *TxCacheStub) Peek(key []byte) (value interface{}, ok bool) {
	if cache.PeekCalled != nil {
		return cache.PeekCalled(key)
	}

	return nil, false
}

// HasOrAdd -
func (cache *TxCacheStub) HasOrAdd(key []byte, value interface{}, sizeInBytes int) (has, added bool) {
	if cache.HasOrAddCalled != nil {
		return cache.HasOrAddCalled(key, value, sizeInBytes)
	}

	return false, false
}

// Remove -
func (cache *TxCacheStub) Remove(key []byte) {
	if cache.RemoveCalled != nil {
		cache.RemoveCalled(key)
	}
}

// Keys -
func (cache *TxCacheStub) Keys() [][]byte {
	if cache.KeysCalled != nil {
		return cache.KeysCalled()
	}

	return make([][]byte, 0)
}

// Len -
func (cache *TxCacheStub) Len() int {
	if cache.LenCalled != nil {
		return cache.LenCalled()
	}

	return 0
}

// SizeInBytesContained -
func (cache *TxCacheStub) SizeInBytesContained() uint64 {
	return 0
}

// MaxSize -
func (cache *TxCacheStub) MaxSize() int {
	if cache.MaxSizeCalled != nil {
		return cache.MaxSizeCalled()
	}

	return 0
}

// RegisterHandler -
func (cache *TxCacheStub) RegisterHandler(handler func(key []byte, value interface{}), _ string) {
	if cache.RegisterHandlerCalled != nil {
		cache.RegisterHandlerCalled(handler)
	}
}

// UnRegisterHandler -
func (cache *TxCacheStub) UnRegisterHandler(id string) {
	if cache.UnRegisterHandlerCalled != nil {
		cache.UnRegisterHandlerCalled(id)
	}
}

// Close -
func (cache *TxCacheStub) Close() error {
	if cache.CloseCalled != nil {
		return cache.CloseCalled()
	}

	return nil
}

// AddTx -
func (cache *TxCacheStub) AddTx(tx *txcache.WrappedTransaction) (ok bool, added bool) {
	if cache.AddTxCalled != nil {
		return cache.AddTxCalled(tx)
	}

	return false, false
}

// NotifyAccountNonce -
func (cache *TxCacheStub) NotifyAccountNonce(accountKey []byte, nonce uint64) {
	if cache.NotifyAccountNonceCalled != nil {
		cache.NotifyAccountNonceCalled(accountKey, nonce)
	}
}

// ForgetAllAccountNonces -
func (cache *TxCacheStub) ForgetAllAccountNonces() {
	if cache.ForgetAllAccountNoncesCalled != nil {
		cache.ForgetAllAccountNoncesCalled()
	}
}

// GetByTxHash -
func (cache *TxCacheStub) GetByTxHash(txHash []byte) (*txcache.WrappedTransaction, bool) {
	if cache.GetByTxHashCalled != nil {
		return cache.GetByTxHashCalled(txHash)
	}

	return nil, false
}

// RemoveTxByHash -
func (cache *TxCacheStub) RemoveTxByHash(txHash []byte) bool {
	if cache.RemoveTxByHashCalled != nil {
		return cache.RemoveTxByHashCalled(txHash)
	}

	return false
}

// ImmunizeTxsAgainstEviction -
func (cache *TxCacheStub) ImmunizeTxsAgainstEviction(keys [][]byte) {
	if cache.ImmunizeTxsAgainstEvictionCalled != nil {
		cache.ImmunizeTxsAgainstEvictionCalled(keys)
	}
}

// ForEachTransaction -
func (cache *TxCacheStub) ForEachTransaction(fn txcache.ForEachTransaction) {
	if cache.ForEachTransactionCalled != nil {
		cache.ForEachTransactionCalled(fn)
	}
}

// NumBytes -
func (cache *TxCacheStub) NumBytes() int {
	if cache.NumBytesCalled != nil {
		return cache.NumBytesCalled()
	}

	return 0
}

// Diagnose -
func (cache *TxCacheStub) Diagnose(deep bool) {
	if cache.DiagnoseCalled != nil {
		cache.DiagnoseCalled(deep)
	}
}

// GetTransactionsPoolForSender -
func (cache *TxCacheStub) GetTransactionsPoolForSender(sender string) []*txcache.WrappedTransaction {
	if cache.GetTransactionsPoolForSenderCalled != nil {
		return cache.GetTransactionsPoolForSenderCalled(sender)
	}

	return nil
}

// IsInterfaceNil -
func (cache *TxCacheStub) IsInterfaceNil() bool {
	return cache == nil
}
