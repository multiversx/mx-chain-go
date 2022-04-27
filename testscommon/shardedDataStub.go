package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/core/counting"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ShardedDataStub -
type ShardedDataStub struct {
	RegisterOnAddedCalled                  func(func(key []byte, value interface{}))
	ShardDataStoreCalled                   func(cacheID string) storage.Cacher
	AddDataCalled                          func(key []byte, data interface{}, sizeInBytes int, cacheID string)
	SearchFirstDataCalled                  func(key []byte) (value interface{}, ok bool)
	RemoveDataCalled                       func(key []byte, cacheID string)
	RemoveDataFromAllShardsCalled          func(key []byte)
	MergeShardStoresCalled                 func(sourceCacheID, destCacheID string)
	MoveDataCalled                         func(sourceCacheID, destCacheID string, key [][]byte)
	ClearCalled                            func()
	ClearShardStoreCalled                  func(cacheID string)
	RemoveSetOfDataFromPoolCalled          func(keys [][]byte, destCacheID string)
	ImmunizeSetOfDataAgainstEvictionCalled func(keys [][]byte, cacheID string)
	CreateShardStoreCalled                 func(destCacheID string)
	GetCountsCalled                        func() counting.CountsWithSize
	KeysCalled                             func() [][]byte
}

// NewShardedDataStub -
func NewShardedDataStub() *ShardedDataStub {
	return &ShardedDataStub{}
}

// RegisterOnAdded -
func (sd *ShardedDataStub) RegisterOnAdded(handler func(key []byte, value interface{})) {
	if sd.RegisterOnAddedCalled != nil {
		sd.RegisterOnAddedCalled(handler)
	}
}

// ShardDataStore -
func (sd *ShardedDataStub) ShardDataStore(cacheID string) storage.Cacher {
	return sd.ShardDataStoreCalled(cacheID)
}

// AddData -
func (sd *ShardedDataStub) AddData(key []byte, data interface{}, sizeInBytes int, cacheID string) {
	if sd.AddDataCalled != nil {
		sd.AddDataCalled(key, data, sizeInBytes, cacheID)
	}
}

// SearchFirstData -
func (sd *ShardedDataStub) SearchFirstData(key []byte) (value interface{}, ok bool) {
	return sd.SearchFirstDataCalled(key)
}

// RemoveData -
func (sd *ShardedDataStub) RemoveData(key []byte, cacheID string) {
	sd.RemoveDataCalled(key, cacheID)
}

// RemoveDataFromAllShards -
func (sd *ShardedDataStub) RemoveDataFromAllShards(key []byte) {
	sd.RemoveDataFromAllShardsCalled(key)
}

// MergeShardStores -
func (sd *ShardedDataStub) MergeShardStores(sourceCacheID, destCacheID string) {
	sd.MergeShardStoresCalled(sourceCacheID, destCacheID)
}

// Clear -
func (sd *ShardedDataStub) Clear() {
	sd.ClearCalled()
}

// ClearShardStore -
func (sd *ShardedDataStub) ClearShardStore(cacheID string) {
	sd.ClearShardStoreCalled(cacheID)
}

// RemoveSetOfDataFromPool -
func (sd *ShardedDataStub) RemoveSetOfDataFromPool(keys [][]byte, cacheID string) {
	sd.RemoveSetOfDataFromPoolCalled(keys, cacheID)
}

// ImmunizeSetOfDataAgainstEviction -
func (sd *ShardedDataStub) ImmunizeSetOfDataAgainstEviction(keys [][]byte, cacheID string) {
	if sd.ImmunizeSetOfDataAgainstEvictionCalled != nil {
		sd.ImmunizeSetOfDataAgainstEvictionCalled(keys, cacheID)
	}
}

// GetCounts -
func (sd *ShardedDataStub) GetCounts() counting.CountsWithSize {
	if sd.GetCountsCalled != nil {
		return sd.GetCountsCalled()
	}

	return &counting.NullCounts{}
}

// Keys -
func (sd *ShardedDataStub) Keys() [][]byte {
	if sd.KeysCalled != nil {
		return sd.KeysCalled()
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sd *ShardedDataStub) IsInterfaceNil() bool {
	return sd == nil
}
