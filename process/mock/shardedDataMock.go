package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"sync"
)

type ShardedDataMock struct {
	mutShardedDataStore  sync.RWMutex
	shardedDataStore     map[string]*shardStore
	cacherConfig         storageUnit.CacheConfig
	mutAddedDataHandlers sync.RWMutex
	addedDataHandlers    []func(key []byte)
}

type shardStore struct {
	CacheID   string
	DataStore storage.Cacher
}

func NewStardedDataMock(cacherConfig storageUnit.CacheConfig) *ShardedDataMock {
	return &ShardedDataMock{
		mutShardedDataStore:  sync.RWMutex{},
		shardedDataStore:     make(map[string]*shardStore),
		mutAddedDataHandlers: sync.RWMutex{},
		addedDataHandlers:    make([]func(key []byte), 0),
		cacherConfig:         cacherConfig,
	}
}

func (sdm *ShardedDataMock) RegisterHandler(handler func(key []byte)) {
	if handler == nil {
		//log.Error("attempt to register a nil handler to a ShardedData object")
		return
	}

	sdm.mutAddedDataHandlers.Lock()
	sdm.addedDataHandlers = append(sdm.addedDataHandlers, handler)
	sdm.mutAddedDataHandlers.Unlock()
}

// ShardStore returns a shard store of data associated with a given destination cacheId
func (sdm *ShardedDataMock) ShardStore(cacheId string) *shardStore {
	sdm.mutShardedDataStore.RLock()
	mp := sdm.shardedDataStore[cacheId]
	sdm.mutShardedDataStore.RUnlock()
	return mp
}

func (sdm *ShardedDataMock) ShardDataStore(cacheId string) (c storage.Cacher) {
	mp := sdm.ShardStore(cacheId)
	if mp == nil {
		return nil
	}
	return mp.DataStore
}

func (sdm *ShardedDataMock) AddData(key []byte, data interface{}, cacheId string) {
	var mp *shardStore

	sdm.mutShardedDataStore.Lock()
	mp = sdm.shardedDataStore[cacheId]
	if mp == nil {
		mp = sdm.newShardStoreNoLock(cacheId)
	}
	sdm.mutShardedDataStore.Unlock()

	found, _ := mp.DataStore.HasOrAdd(key, data)

	if !found {
		sdm.mutAddedDataHandlers.RLock()
		for _, handler := range sdm.addedDataHandlers {
			go handler(key)
		}
		sdm.mutAddedDataHandlers.RUnlock()
	}
}

func (sd *ShardedDataMock) newShardStoreNoLock(cacheId string) *shardStore {
	shardStore, _ := newShardStore(cacheId, sd.cacherConfig)
	//log.LogIfError(err)

	sd.shardedDataStore[cacheId] = shardStore
	return shardStore
}

func newShardStore(cacheId string, cacherConfig storageUnit.CacheConfig) (*shardStore, error) {
	//cacher, err := storageUnit.NewCache(cacherConfig.Type, cacherConfig.Size, cacherConfig.Shards)
	//if err != nil {
	//	return nil, err
	//}
	return &shardStore{
		CacheID:   cacheId,
		DataStore: mock.NewCacherMock(),
	}, nil
}

func (sdm *ShardedDataMock) SearchFirstData(key []byte) (value interface{}, ok bool) {
	panic("implement me")
}

func (sdm *ShardedDataMock) RemoveData(key []byte, cacheId string) {
	panic("implement me")
}

func (sdm *ShardedDataMock) RemoveSetOfDataFromPool(keys [][]byte, cacheId string) {
	panic("implement me")
}

func (sdm *ShardedDataMock) RemoveDataFromAllShards(key []byte) {
	panic("implement me")
}

func (sdm *ShardedDataMock) MergeShardStores(sourceCacheID, destCacheID string) {
	panic("implement me")
}

func (sdm *ShardedDataMock) MoveData(sourceCacheID, destCacheID string, key [][]byte) {
	panic("implement me")
}

func (sdm *ShardedDataMock) Clear() {
	panic("implement me")
}

func (sdm *ShardedDataMock) ClearShardStore(cacheId string) {
	panic("implement me")
}

func (sdm *ShardedDataMock) CreateShardStore(cacheId string) {
	panic("implement me")
}
