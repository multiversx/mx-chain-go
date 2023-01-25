package testscommon

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/counting"
	"github.com/multiversx/mx-chain-go/storage"
)

// ShardedDataCacheNotifierMock -
type ShardedDataCacheNotifierMock struct {
	mutCaches sync.RWMutex
	caches    map[string]storage.Cacher
}

// NewShardedDataCacheNotifierMock -
func NewShardedDataCacheNotifierMock() *ShardedDataCacheNotifierMock {
	return &ShardedDataCacheNotifierMock{
		caches: make(map[string]storage.Cacher),
	}
}

// RegisterOnAdded -
func (mock *ShardedDataCacheNotifierMock) RegisterOnAdded(_ func(key []byte, value interface{})) {
}

// ShardDataStore -
func (mock *ShardedDataCacheNotifierMock) ShardDataStore(cacheId string) (c storage.Cacher) {
	mock.mutCaches.Lock()
	defer mock.mutCaches.Unlock()

	cache, found := mock.caches[cacheId]
	if !found {
		cache = NewCacherMock()
		mock.caches[cacheId] = cache
	}

	return cache
}

// AddData -
func (mock *ShardedDataCacheNotifierMock) AddData(key []byte, data interface{}, sizeInBytes int, cacheId string) {
	cache := mock.ShardDataStore(cacheId)
	cache.HasOrAdd(key, data, sizeInBytes)
}

// SearchFirstData -
func (mock *ShardedDataCacheNotifierMock) SearchFirstData(key []byte) (interface{}, bool) {
	mock.mutCaches.RLock()
	defer mock.mutCaches.RUnlock()

	for _, cache := range mock.caches {
		buff, ok := cache.Get(key)
		if ok {
			return buff, true
		}
	}

	return nil, false
}

// RemoveData -
func (mock *ShardedDataCacheNotifierMock) RemoveData(key []byte, cacheId string) {
	cache := mock.ShardDataStore(cacheId)
	cache.Remove(key)
}

// RemoveSetOfDataFromPool -
func (mock *ShardedDataCacheNotifierMock) RemoveSetOfDataFromPool(keys [][]byte, cacheId string) {
	cache := mock.ShardDataStore(cacheId)
	for _, key := range keys {
		cache.Remove(key)
	}
}

// ImmunizeSetOfDataAgainstEviction -
func (mock *ShardedDataCacheNotifierMock) ImmunizeSetOfDataAgainstEviction(_ [][]byte, _ string) {
}

// RemoveDataFromAllShards -
func (mock *ShardedDataCacheNotifierMock) RemoveDataFromAllShards(key []byte) {
	mock.mutCaches.RLock()
	defer mock.mutCaches.RUnlock()

	for _, cache := range mock.caches {
		cache.Remove(key)
	}
}

// MergeShardStores -
func (mock *ShardedDataCacheNotifierMock) MergeShardStores(_, _ string) {
}

// Clear -
func (mock *ShardedDataCacheNotifierMock) Clear() {
	mock.mutCaches.Lock()
	defer mock.mutCaches.Unlock()

	mock.caches = make(map[string]storage.Cacher)
}

// ClearShardStore -
func (mock *ShardedDataCacheNotifierMock) ClearShardStore(cacheId string) {
	cache := mock.ShardDataStore(cacheId)
	cache.Clear()
}

// GetCounts -
func (mock *ShardedDataCacheNotifierMock) GetCounts() counting.CountsWithSize {
	return nil
}

// Keys -
func (mock *ShardedDataCacheNotifierMock) Keys() [][]byte {
	mock.mutCaches.Lock()
	defer mock.mutCaches.Unlock()

	keys := make([][]byte, 0)
	for _, cache := range mock.caches {
		keys = append(keys, cache.Keys()...)
	}

	return keys
}

// IsInterfaceNil -
func (mock *ShardedDataCacheNotifierMock) IsInterfaceNil() bool {
	return mock == nil
}
