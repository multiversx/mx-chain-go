package shardedData

import (
	"bytes"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/stretchr/testify/assert"
)

var timeoutWaitForWaitGroups = time.Second * 2

var defaultTestConfig = storageunit.CacheConfig{
	Capacity:    75000,
	SizeInBytes: 104857600,
	Shards:      1,
}

func TestNewShardedData_BadConfigShouldErr(t *testing.T) {
	t.Parallel()

	cacheConfigBad := storageunit.CacheConfig{
		Capacity: 0,
	}

	sd, err := NewShardedData("", cacheConfigBad)
	assert.NotNil(t, err)
	assert.Nil(t, sd)
}

func TestNewShardedData_GoodConfigShouldWork(t *testing.T) {
	t.Parallel()

	sd, err := NewShardedData("", defaultTestConfig)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(sd))
}

func TestShardedData_AddData(t *testing.T) {
	t.Parallel()

	sd, _ := NewShardedData("", defaultTestConfig)

	keyTx1 := []byte("hash_tx1")
	shardID1 := "1"

	keyTx2 := []byte("hash_tx2")
	shardID2 := "2"

	sd.AddData(keyTx1, &transaction.Transaction{Nonce: 1}, 0, shardID1)

	shardStore := sd.ShardDataStore("1")
	has := shardStore.Has(keyTx1)
	assert.True(t, has, "Key was not added to minipool")
	assert.True(t, shardStore.Len() == 1,
		"Transaction pool length is not 1 after one element was added")

	sd.AddData(keyTx2, &transaction.Transaction{Nonce: 2}, 0, shardID2)

	assert.False(t, shardStore.Has(keyTx2))
	assert.True(t, sd.ShardDataStore(shardID2).Has(keyTx2))
}

func TestShardedData_StorageEvictsData(t *testing.T) {
	t.Parallel()

	sd, _ := NewShardedData("", defaultTestConfig)

	for i := 1; i < int(defaultTestConfig.Capacity+100); i++ {
		key := []byte(strconv.Itoa(i))
		sd.AddData(key, &transaction.Transaction{Nonce: uint64(i)}, 0, "1")
	}

	assert.Less(t, sd.ShardDataStore("1").Len(), int(defaultTestConfig.Capacity),
		"Transaction pool entries exceeds the maximum configured number")
}

func TestShardedData_NoDuplicates(t *testing.T) {
	t.Parallel()

	sd, _ := NewShardedData("", defaultTestConfig)

	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 0, "1")
	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 0, "1")
	assert.Equal(t, 1, sd.ShardDataStore("1").Len(),
		"Transaction pool should not contain duplicates")
}

func TestShardedData_AddDataInParallel(t *testing.T) {
	t.Parallel()

	sd, _ := NewShardedData("", defaultTestConfig)

	wg := sync.WaitGroup{}

	vals := int(defaultTestConfig.Capacity)
	wg.Add(vals)

	for i := 0; i < vals; i++ {
		key := []byte(strconv.Itoa(i))
		go func(i int, wg *sync.WaitGroup) {
			sd.AddData(key, &transaction.Transaction{Nonce: uint64(i)}, 0, "1")
			wg.Done()
		}(i, &wg)
	}

	wg.Wait()

	//checking
	for i := 0; i < vals; i++ {
		key := []byte(strconv.Itoa(i))
		assert.True(t, sd.shardStore("1").cache.Has(key), fmt.Sprintf("for val %d", i))
	}
}

func TestShardedData_RemoveData(t *testing.T) {
	t.Parallel()

	sd, _ := NewShardedData("", defaultTestConfig)

	sd.RemoveData([]byte{}, "missing_cache_id") // coverage

	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 0, "1")
	assert.Equal(t, 1, sd.ShardDataStore("1").Len(),
		"AddData failed, length should be 1")
	sd.RemoveData([]byte("tx_hash1"), "1")
	assert.Equal(t, 0, sd.ShardDataStore("1").Len(),
		"RemoveData failed, length should be 0")

	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 0, "1")
	sd.AddData([]byte("tx_hash2"), &transaction.Transaction{Nonce: 2}, 0, "2")
	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 0, "2")
	assert.Equal(t, 1, sd.ShardDataStore("1").Len(),
		"AddData failed, length should be 1")
	assert.Equal(t, 2, sd.ShardDataStore("2").Len(),
		"AddData failed, length should be 2")

	sd.RemoveDataFromAllShards([]byte("tx_hash1"))
	assert.Equal(t, 0, sd.ShardDataStore("1").Len(),
		"FindAndRemoveData failed, length should be 0 in shard 1")
	assert.Equal(t, 1, sd.ShardDataStore("2").Len(),
		"FindAndRemoveData failed, length should be 1 in shard 2")
}

func TestShardedData_ClearShardStore(t *testing.T) {
	t.Parallel()

	sd, _ := NewShardedData("", defaultTestConfig)

	sd.ClearShardStore("missing_cache_id") // coverage

	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 0, "1")
	sd.AddData([]byte("tx_hash2"), &transaction.Transaction{Nonce: 2}, 0, "2")
	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 0, "2")

	sd.ClearShardStore("2")
	assert.Equal(t, 0, sd.ShardDataStore("2").Len(),
		"Mini pool for shard 2 should be empty after clear")
	assert.Equal(t, 1, sd.ShardDataStore("1").Len(),
		"Mini pool for shard 1 should still have one element")

	sd.Clear()
	assert.Nil(t, sd.shardStore("1"), "Shard 1 should not be in the store anymore")
	assert.Nil(t, sd.shardStore("2"), "Shard 2 should not be in the store anymore")
}

func TestShardedData_MergeShardStores(t *testing.T) {
	t.Parallel()

	sd, _ := NewShardedData("", defaultTestConfig)

	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 0, "1")
	sd.AddData([]byte("tx_hash2"), &transaction.Transaction{Nonce: 2}, 0, "2")
	sd.AddData([]byte("tx_hash3"), &transaction.Transaction{Nonce: 3}, 0, "2")

	sd.MergeShardStores("1", "2")
	assert.Equal(t, 3, sd.ShardDataStore("2").Len(),
		"Mini pool for shard 1 should have 3 elements")
	assert.Nil(t, sd.ShardDataStore("1"))
}

func TestShardedData_RegisterAddedDataHandlerNilHandlerShouldIgnore(t *testing.T) {
	t.Parallel()

	sd, _ := NewShardedData("", defaultTestConfig)

	sd.RegisterOnAdded(nil)

	assert.Equal(t, 0, len(sd.addedDataHandlers))
}

func TestShardedData_RegisterAddedDataHandlerShouldWork(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	chDone := make(chan bool)

	f := func(key []byte, value interface{}) {
		if !bytes.Equal([]byte("aaaa"), key) {
			return
		}

		wg.Done()
	}

	go func() {
		wg.Wait()
		chDone <- true
	}()

	sd, _ := NewShardedData("", defaultTestConfig)

	sd.RegisterOnAdded(f)
	sd.AddData([]byte("aaaa"), "bbbb", 4, "0")

	select {
	case <-chDone:
	case <-time.After(timeoutWaitForWaitGroups):
		assert.Fail(t, "should have been called")
		return
	}
}

func TestShardedData_RegisterAddedDataHandlerReallyAddsHandler(t *testing.T) {
	t.Parallel()

	f := func(key []byte, value interface{}) {
	}

	sd, _ := NewShardedData("", defaultTestConfig)

	sd.RegisterOnAdded(f)

	assert.Equal(t, 1, len(sd.addedDataHandlers))
}

func TestShardedData_Keys(t *testing.T) {
	sd, _ := NewShardedData("", defaultTestConfig)

	txsHashes := [][]byte{[]byte("hash-w"), []byte("hash-x"), []byte("hash-y"), []byte("hash-z")}
	sd.AddData(txsHashes[0], nil, 0, "1")
	sd.AddData(txsHashes[1], nil, 0, "1")
	sd.AddData(txsHashes[2], nil, 0, "2")
	sd.AddData(txsHashes[3], nil, 0, "3")

	assert.ElementsMatch(t, txsHashes, sd.Keys())
}

func TestShardedData_RegisterAddedDataHandlerNotAddedShouldNotCall(t *testing.T) {
	t.Parallel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	chDone := make(chan bool)

	f := func(key []byte, value interface{}) {
		wg.Done()
	}

	go func() {
		wg.Wait()
		chDone <- true
	}()

	sd, _ := NewShardedData("", defaultTestConfig)

	//first add, no call
	sd.AddData([]byte("aaaa"), "bbbb", 4, "0")
	sd.RegisterOnAdded(f)
	//second add, should not call as the data was found
	sd.AddData([]byte("aaaa"), "bbbb", 4, "0")

	select {
	case <-chDone:
		assert.Fail(t, "should have not been called")
		return
	case <-time.After(timeoutWaitForWaitGroups):
	}

	assert.Equal(t, 1, len(sd.addedDataHandlers))
}

func TestShardedData_SearchFirstDataNotFoundShouldRetNilAndFalse(t *testing.T) {
	t.Parallel()

	sd, _ := NewShardedData("", defaultTestConfig)

	value, ok := sd.SearchFirstData([]byte("aaaa"))
	assert.Nil(t, value)
	assert.False(t, ok)
}

func TestShardedData_SearchFirstDataFoundShouldRetResults(t *testing.T) {
	t.Parallel()

	sd, _ := NewShardedData("", defaultTestConfig)

	sd.AddData([]byte("aaa"), "a1", 2, "0")
	sd.AddData([]byte("aaaa"), "a2", 2, "4")
	sd.AddData([]byte("aaaa"), "a3", 2, "5")

	value, ok := sd.SearchFirstData([]byte("aaa"))
	assert.NotNil(t, value)
	assert.True(t, ok)
}

func TestShardedData_RemoveSetOfDataFromPool(t *testing.T) {
	t.Parallel()

	sd, _ := NewShardedData("", defaultTestConfig)

	sd.RemoveSetOfDataFromPool([][]byte{}, "missing_cache_id") // coverage

	sd.AddData([]byte("aaa"), "a1", 2, "0")
	_, ok := sd.SearchFirstData([]byte("aaa"))
	assert.True(t, ok)
	sd.RemoveSetOfDataFromPool([][]byte{[]byte("aaa")}, "0")
	_, ok = sd.SearchFirstData([]byte("aaa"))
	assert.False(t, ok)
}

func TestShardedData_ImmunizeSetOfDataAgainstEviction(t *testing.T) {
	t.Parallel()

	sd, _ := NewShardedData("", defaultTestConfig)
	sd.ImmunizeSetOfDataAgainstEviction([][]byte{[]byte("aaa")}, "0")
}

func TestShardedData_GetCounts(t *testing.T) {
	t.Parallel()

	sd, _ := NewShardedData("", defaultTestConfig)

	sd.RemoveSetOfDataFromPool([][]byte{}, "missing_cache_id") // coverage

	sd.AddData([]byte("aaa"), "a1", 2, "0")
	sd.AddData([]byte("bbb"), "b1", 2, "0")
	counts := sd.GetCounts()
	assert.Equal(t, int64(2), counts.GetTotal())
}

func TestShardedData_Diagnose(t *testing.T) {
	t.Parallel()

	sd, _ := NewShardedData("", defaultTestConfig)

	sd.RemoveSetOfDataFromPool([][]byte{}, "missing_cache_id") // coverage

	sd.AddData([]byte("aaa"), "a1", 2, "0")
	sd.AddData([]byte("bbb"), "b1", 2, "0")
	sd.Diagnose(true)
}
