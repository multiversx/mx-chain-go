package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

type Uint64SyncMapCacherStub struct {
	ClearCalled           func()
	GetCalled             func(nonce uint64) (dataRetriever.ShardIdHashMap, bool)
	MergeCalled           func(nonce uint64, src dataRetriever.ShardIdHashMap)
	RemoveNonceCalled     func(nonce uint64)
	RemoveShardIdCalled   func(nonce uint64, shardId uint32)
	RegisterHandlerCalled func(handler func(nonce uint64, shardId uint32, value []byte))
	HasCalled             func(nonce uint64, shardId uint32) bool
}

func (usmcs *Uint64SyncMapCacherStub) Clear() {
	usmcs.ClearCalled()
}

func (usmcs *Uint64SyncMapCacherStub) Get(nonce uint64) (dataRetriever.ShardIdHashMap, bool) {
	return usmcs.GetCalled(nonce)
}

func (usmcs *Uint64SyncMapCacherStub) Merge(nonce uint64, src dataRetriever.ShardIdHashMap) {
	usmcs.MergeCalled(nonce, src)
}

func (usmcs *Uint64SyncMapCacherStub) RemoveNonce(nonce uint64) {
	usmcs.RemoveNonceCalled(nonce)
}

func (usmcs *Uint64SyncMapCacherStub) RegisterHandler(handler func(nonce uint64, shardId uint32, value []byte)) {
	usmcs.RegisterHandlerCalled(handler)
}

func (usmcs *Uint64SyncMapCacherStub) Has(nonce uint64, shardId uint32) bool {
	return usmcs.HasCalled(nonce, shardId)
}

func (usmcs *Uint64SyncMapCacherStub) RemoveShardId(nonce uint64, shardId uint32) {
	usmcs.RemoveShardIdCalled(nonce, shardId)
}
