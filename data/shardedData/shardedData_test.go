package shardedData_test

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/stretchr/testify/assert"
)

var timeoutWaitForWaitGroups = time.Second * 2

var defaultTestConfig = storage.CacheConfig{
	Size: 1000,
	Type: storage.LRUCache,
}

func TestNewShardedData_BadConfigShouldErr(t *testing.T) {
	cacheConfigBad := storage.CacheConfig{
		Size: 0,
		Type: storage.LRUCache,
	}

	sd, err := shardedData.NewShardedData(cacheConfigBad)
	assert.NotNil(t, err)
	assert.Nil(t, sd)
}

func TestNewShardedData_GoodConfigShouldWork(t *testing.T) {
	cacheConfigBad := storage.CacheConfig{
		Size: 10,
		Type: storage.LRUCache,
	}

	sd, err := shardedData.NewShardedData(cacheConfigBad)
	assert.Nil(t, err)
	assert.NotNil(t, sd)
}

func TestShardedData_AddData(t *testing.T) {
	t.Parallel()

	sd, _ := shardedData.NewShardedData(defaultTestConfig)

	keyTx1 := []byte("hash_tx1")
	shardID1 := uint32(1)

	keyTx2 := []byte("hash_tx2")
	shardID2 := uint32(2)

	sd.AddData(keyTx1, &transaction.Transaction{Nonce: 1}, shardID1)

	shardStore := sd.ShardDataStore(1)
	has := shardStore.Has(keyTx1)
	assert.True(t, has, "Key was not added to minipool")
	assert.True(t, shardStore.Len() == 1,
		"Transaction pool length is not 1 after one element was added")

	sd.AddData(keyTx2, &transaction.Transaction{Nonce: 2}, shardID2)

	assert.False(t, shardStore.Has(keyTx2))
	assert.True(t, sd.ShardDataStore(shardID2).Has(keyTx2))
}

func TestShardedData_StorageEvictsData(t *testing.T) {
	t.Parallel()

	sd, _ := shardedData.NewShardedData(defaultTestConfig)

	for i := 1; i < int(defaultTestConfig.Size+100); i++ {
		key := []byte(strconv.Itoa(i))
		sd.AddData(key, &transaction.Transaction{Nonce: uint64(i)}, 1)
	}

	assert.Equal(t, int(defaultTestConfig.Size), sd.ShardDataStore(1).Len(),
		"Transaction pool entries excedes the maximum configured number")
}

func TestShardedData_NoDuplicates(t *testing.T) {
	t.Parallel()

	sd, _ := shardedData.NewShardedData(defaultTestConfig)

	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 1)
	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 1)
	assert.Equal(t, 1, sd.ShardDataStore(1).Len(),
		"Transaction pool should not contain duplicates")
}

func TestShardedData_AddDataInParallel(t *testing.T) {
	t.Parallel()

	sd, _ := shardedData.NewShardedData(defaultTestConfig)

	wg := sync.WaitGroup{}

	vals := int(defaultTestConfig.Size)
	wg.Add(vals)

	for i := 0; i < vals; i++ {
		key := []byte(strconv.Itoa(i))
		go func(i int, wg *sync.WaitGroup) {
			sd.AddData(key, &transaction.Transaction{Nonce: uint64(i)}, 1)
			wg.Done()
		}(i, &wg)
	}

	wg.Wait()

	//checking
	for i := 0; i < vals; i++ {
		key := []byte(strconv.Itoa(i))
		assert.True(t, sd.ShardStore(1).DataStore.Has(key), fmt.Sprintf("for val %d", i))
	}
}

func TestShardedData_RemoveData(t *testing.T) {
	t.Parallel()

	sd, _ := shardedData.NewShardedData(defaultTestConfig)

	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 1)
	assert.Equal(t, 1, sd.ShardDataStore(1).Len(),
		"AddData failed, length should be 1")
	sd.RemoveData([]byte("tx_hash1"), 1)
	assert.Equal(t, 0, sd.ShardDataStore(1).Len(),
		"RemoveData failed, length should be 0")

	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 1)
	sd.AddData([]byte("tx_hash2"), &transaction.Transaction{Nonce: 2}, 2)
	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 2)
	assert.Equal(t, 1, sd.ShardDataStore(1).Len(),
		"AddData failed, length should be 1")
	assert.Equal(t, 2, sd.ShardDataStore(2).Len(),
		"AddData failed, length should be 2")

	sd.RemoveDataFromAllShards([]byte("tx_hash1"))
	assert.Equal(t, 0, sd.ShardDataStore(1).Len(),
		"FindAndRemoveData failed, length should be 0 in shard 1")
	assert.Equal(t, 1, sd.ShardDataStore(2).Len(),
		"FindAndRemoveData failed, length should be 1 in shard 2")
}

func TestShardedData_Clear(t *testing.T) {
	t.Parallel()

	sd, _ := shardedData.NewShardedData(defaultTestConfig)

	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 1)
	sd.AddData([]byte("tx_hash2"), &transaction.Transaction{Nonce: 2}, 2)
	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 2)

	sd.ClearShardStore(2)
	assert.Equal(t, 0, sd.ShardDataStore(2).Len(),
		"Mini pool for shard 2 should be empty after clear")
	assert.Equal(t, 1, sd.ShardDataStore(1).Len(),
		"Mini pool for shard 1 should still have one element")

	sd.Clear()
	assert.Nil(t, sd.ShardStore(1), "Shard 1 should not be in the store anymore")
	assert.Nil(t, sd.ShardStore(2), "Shard 2 should not be in the store anymore")
}

func TestShardedData_MergeShardStores(t *testing.T) {
	t.Parallel()

	sd, _ := shardedData.NewShardedData(defaultTestConfig)

	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 1)
	sd.AddData([]byte("tx_hash2"), &transaction.Transaction{Nonce: 2}, 2)
	sd.AddData([]byte("tx_hash3"), &transaction.Transaction{Nonce: 3}, 2)

	sd.MergeShardStores(1, 2)
	assert.Equal(t, 3, sd.ShardDataStore(2).Len(),
		"Mini pool for shard 1 should have 3 elements")
	assert.Nil(t, sd.ShardDataStore(1))
}

func TestShardedData_MoveData(t *testing.T) {
	t.Parallel()

	sd, _ := shardedData.NewShardedData(defaultTestConfig)

	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 1)
	sd.AddData([]byte("tx_hash2"), &transaction.Transaction{Nonce: 2}, 2)
	sd.AddData([]byte("tx_hash3"), &transaction.Transaction{Nonce: 3}, 2)
	sd.AddData([]byte("tx_hash4"), &transaction.Transaction{Nonce: 4}, 2)
	sd.AddData([]byte("tx_hash5"), &transaction.Transaction{Nonce: 5}, 2)
	sd.AddData([]byte("tx_hash6"), &transaction.Transaction{Nonce: 6}, 2)

	sd.MoveData(2, 3, [][]byte{[]byte("tx_hash5"), []byte("tx_hash6")})

	assert.Equal(t, 3, sd.ShardDataStore(2).Len(),
		"Mini pool for shard 2 should have 3 elements")
	assert.Equal(t, 2, sd.ShardDataStore(3).Len(),
		"Mini pool for shard 3 should have 2 elements")
}

func TestShardedData_RegisterAddedDataHandlerNilHandlerShouldIgnore(t *testing.T) {
	t.Parallel()

	sd, _ := shardedData.NewShardedData(defaultTestConfig)

	sd.RegisterHandler(nil)

	assert.Equal(t, 0, len(sd.AddedDataHandlers()))
}

func TestShardedData_RegisterAddedDataHandlerShouldWork(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	chDone := make(chan bool, 0)

	f := func(key []byte) {
		if !bytes.Equal([]byte("aaaa"), key) {
			return
		}

		wg.Done()
	}

	go func() {
		wg.Wait()
		chDone <- true
	}()

	sd, _ := shardedData.NewShardedData(defaultTestConfig)

	sd.RegisterHandler(f)
	sd.AddData([]byte("aaaa"), "bbbb", 0)

	select {
	case <-chDone:
	case <-time.After(timeoutWaitForWaitGroups):
		assert.Fail(t, "should have been called")
		return
	}
}

func TestShardedData_RegisterAddedDataHandlerReallyAddsAhandler(t *testing.T) {
	t.Parallel()

	f := func(key []byte) {
	}

	sd, _ := shardedData.NewShardedData(defaultTestConfig)

	sd.RegisterHandler(f)

	assert.Equal(t, 1, len(sd.AddedDataHandlers()))
}

func TestShardedData_RegisterAddedDataHandlerNotAddedShouldNotCall(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	chDone := make(chan bool, 0)

	f := func(key []byte) {
		wg.Done()
	}

	go func() {
		wg.Wait()
		chDone <- true
	}()

	sd, _ := shardedData.NewShardedData(defaultTestConfig)

	//first add, no call
	sd.AddData([]byte("aaaa"), "bbbb", 0)
	sd.RegisterHandler(f)
	//second add, should not call as the data was found
	sd.AddData([]byte("aaaa"), "bbbb", 0)

	select {
	case <-chDone:
		assert.Fail(t, "should have not been called")
		return
	case <-time.After(timeoutWaitForWaitGroups):
	}

	assert.Equal(t, 1, len(sd.AddedDataHandlers()))
}

func TestShardedData_SearchFirstDataNotFoundShouldRetNilAndFalse(t *testing.T) {
	t.Parallel()

	sd, _ := shardedData.NewShardedData(defaultTestConfig)

	value, ok := sd.SearchFirstData([]byte("aaaa"))
	assert.Nil(t, value)
	assert.False(t, ok)
}

func TestShardedData_SearchFirstDataFoundShouldRetResults(t *testing.T) {
	t.Parallel()

	sd, _ := shardedData.NewShardedData(defaultTestConfig)

	sd.AddData([]byte("aaa"), "a1", 0)
	sd.AddData([]byte("aaaa"), "a2", 4)
	sd.AddData([]byte("aaaa"), "a3", 5)

	value, ok := sd.SearchFirstData([]byte("aaa"))
	assert.NotNil(t, value)
	assert.True(t, ok)
}
