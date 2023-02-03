package mock

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
)

// Uint64SyncMapCacherStub -
type Uint64SyncMapCacherStub struct {
	ClearCalled             func()
	GetCalled               func(nonce uint64) (dataRetriever.ShardIdHashMap, bool)
	MergeCalled             func(nonce uint64, src dataRetriever.ShardIdHashMap)
	RemoveCalled            func(nonce uint64, shardId uint32)
	RegisterHandlerCalled   func(handler func(nonce uint64, shardId uint32, value []byte))
	HasCalled               func(nonce uint64, shardId uint32) bool
	UnRegisterHandlerCalled func(func(key []byte, value interface{}))
}

// UnRegisterHandler -
func (usmcs *Uint64SyncMapCacherStub) UnRegisterHandler(handler func(key []byte, value interface{})) {
	usmcs.UnRegisterHandlerCalled(handler)
}

// Clear -
func (usmcs *Uint64SyncMapCacherStub) Clear() {
	usmcs.ClearCalled()
}

// Get -
func (usmcs *Uint64SyncMapCacherStub) Get(nonce uint64) (dataRetriever.ShardIdHashMap, bool) {
	return usmcs.GetCalled(nonce)
}

// Merge -
func (usmcs *Uint64SyncMapCacherStub) Merge(nonce uint64, src dataRetriever.ShardIdHashMap) {
	usmcs.MergeCalled(nonce, src)
}

// RegisterHandler -
func (usmcs *Uint64SyncMapCacherStub) RegisterHandler(handler func(nonce uint64, shardId uint32, value []byte)) {
	usmcs.RegisterHandlerCalled(handler)
}

// Has -
func (usmcs *Uint64SyncMapCacherStub) Has(nonce uint64, shardId uint32) bool {
	return usmcs.HasCalled(nonce, shardId)
}

// Remove -
func (usmcs *Uint64SyncMapCacherStub) Remove(nonce uint64, shardId uint32) {
	usmcs.RemoveCalled(nonce, shardId)
}

// IsInterfaceNil returns true if there is no value under the interface
func (usmcs *Uint64SyncMapCacherStub) IsInterfaceNil() bool {
	return usmcs == nil
}
