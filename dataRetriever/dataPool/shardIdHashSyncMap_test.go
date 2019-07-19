package dataPool_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/stretchr/testify/assert"
)

func TestShardIdHashSyncMap_StoreLoadShouldWork(t *testing.T) {
	t.Parallel()

	sihsm := dataPool.ShardIdHashSyncMap{}

	shardIds := []uint32{45, 38, 56}
	hashes := [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")}

	for idx := 0; idx < len(shardIds); idx++ {
		sihsm.Store(shardIds[idx], hashes[idx])

		retrievedHash, ok := sihsm.Load(shardIds[idx])

		assert.True(t, ok)
		assert.Equal(t, hashes[idx], retrievedHash)
	}
}

func TestShardIdHashSyncMap_LoadingNilHash(t *testing.T) {
	t.Parallel()

	sihsm := dataPool.ShardIdHashSyncMap{}

	shardId := uint32(67)
	sihsm.Store(shardId, nil)

	retrievedHash, ok := sihsm.Load(shardId)

	assert.True(t, ok)
	assert.Nil(t, retrievedHash)
}

func TestShardIdHashSyncMap_LoadingNotExistingElement(t *testing.T) {
	t.Parallel()

	sihsm := dataPool.ShardIdHashSyncMap{}

	shardId := uint32(67)
	retrievedHash, ok := sihsm.Load(shardId)

	assert.False(t, ok)
	assert.Nil(t, retrievedHash)
}

func TestShardIdHashSyncMap_RangeWithNilHandlerShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not paniced: %v", r))
		}
	}()

	sihsm := dataPool.ShardIdHashSyncMap{}

	shardId := uint32(67)
	hash := []byte("hash")
	sihsm.Store(shardId, hash)

	sihsm.Range(nil)
}

func TestShardIdHashSyncMap_RangeShouldWork(t *testing.T) {
	t.Parallel()

	sihsm := dataPool.ShardIdHashSyncMap{}

	shardIds := []uint32{45, 38, 56}
	hashes := [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")}

	mutVisitedMap := sync.Mutex{}
	visitedMap := make(map[uint32]bool)
	for i := 0; i < len(shardIds); i++ {
		visitedMap[shardIds[i]] = false
		sihsm.Store(shardIds[i], hashes[i])
	}

	sihsm.Range(func(shardId uint32, hash []byte) bool {
		mutVisitedMap.Lock()
		defer mutVisitedMap.Unlock()

		visitedMap[shardId] = true
		return true
	})

	mutVisitedMap.Lock()
	for _, wasVisited := range visitedMap {
		assert.True(t, wasVisited)
	}
	mutVisitedMap.Unlock()

}

func TestShardIdHashSyncMap_Delete(t *testing.T) {
	t.Parallel()

	sihsm := dataPool.ShardIdHashSyncMap{}

	shardIds := []uint32{45, 38, 56}
	hashes := [][]byte{[]byte("hash1"), []byte("hash2"), []byte("hash3")}

	mutVisitedMap := sync.Mutex{}
	visitedMap := make(map[uint32]bool)
	for i := 0; i < len(shardIds); i++ {
		visitedMap[shardIds[i]] = false
		sihsm.Store(shardIds[i], hashes[i])
	}

	deletedIndex := 1
	sihsm.Delete(shardIds[deletedIndex])

	sihsm.Range(func(shardId uint32, hash []byte) bool {
		mutVisitedMap.Lock()
		defer mutVisitedMap.Unlock()

		visitedMap[shardId] = true
		return true
	})

	mutVisitedMap.Lock()
	for shardId, wasVisited := range visitedMap {
		shouldBeVisited := shardId != shardIds[deletedIndex]
		assert.Equal(t, shouldBeVisited, wasVisited)
	}
	mutVisitedMap.Unlock()
}
