package shardedData_test

import (
	"bytes"
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

	_, err := shardedData.NewShardedData(cacheConfigBad)
	assert.NotNil(t, err)
}

func TestShardedData_AddData(t *testing.T) {
	t.Parallel()

	sd, err := shardedData.NewShardedData(defaultTestConfig)
	assert.Nil(t, err)

	sd.AddData([]byte("hash_tx1"), &transaction.Transaction{Nonce: 1}, 1)

	has := sd.ShardDataStore(1).Has([]byte("hash_tx1"))
	assert.True(t, has, "Key was not added to minipool")
	assert.True(t, sd.ShardDataStore(1).Len() == 1,
		"Transaction pool length is not 1 after one element was added")

	sd.AddData([]byte("hash_tx2"), &transaction.Transaction{Nonce: 2}, 2)

	assert.False(t, sd.ShardDataStore(1).Has([]byte("hash_tx2")))
	assert.True(t, sd.ShardDataStore(2).Has([]byte("hash_tx2")))
}

func TestShardedData_StorageEvictsData(t *testing.T) {
	t.Parallel()

	size := 1000

	sd, err := shardedData.NewShardedData(defaultTestConfig)
	assert.Nil(t, err)

	for i := 1; i < size+2; i++ {
		key := []byte(strconv.Itoa(i))
		sd.AddData(key, &transaction.Transaction{Nonce: uint64(i)}, 1)
	}

	assert.Equal(t, size, sd.ShardDataStore(1).Len(),
		"Transaction pool entries excedes the maximum configured number")
}

func TestShardedData_NoDuplicates(t *testing.T) {
	t.Parallel()

	sd, err := shardedData.NewShardedData(defaultTestConfig)
	assert.Nil(t, err)

	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 1)
	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 1)
	assert.Equal(t, 1, sd.ShardDataStore(1).Len(),
		"Transaction pool should not contain duplicates")
}

func TestShardedData_AddDataInParallel(t *testing.T) {
	t.Parallel()

	sd, err := shardedData.NewShardedData(defaultTestConfig)
	assert.Nil(t, err)

	for i := 1; i < 10000+2; i++ {
		key := []byte(strconv.Itoa(i))
		go func(i int) {
			sd.AddData(key, &transaction.Transaction{Nonce: uint64(i)}, 1)
		}(i)
	}
}

func TestShardedData_RemoveData(t *testing.T) {
	t.Parallel()

	sd, err := shardedData.NewShardedData(defaultTestConfig)
	assert.Nil(t, err)

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

	sd, err := shardedData.NewShardedData(defaultTestConfig)
	assert.Nil(t, err)

	sd.Clear()
	sd.ClearMiniPool(1)

	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 1)
	sd.AddData([]byte("tx_hash2"), &transaction.Transaction{Nonce: 2}, 2)
	sd.AddData([]byte("tx_hash1"), &transaction.Transaction{Nonce: 1}, 2)

	sd.ClearMiniPool(2)
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

	sd, err := shardedData.NewShardedData(defaultTestConfig)
	assert.Nil(t, err)

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

	sd, err := shardedData.NewShardedData(defaultTestConfig)
	assert.Nil(t, err)

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

	sd, err := shardedData.NewShardedData(defaultTestConfig)
	assert.Nil(t, err)

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

	sd, err := shardedData.NewShardedData(defaultTestConfig)
	assert.Nil(t, err)

	sd.RegisterHandler(f)
	sd.AddData([]byte("aaaa"), "bbbb", 0)

	select {
	case <-chDone:
	case <-time.After(timeoutWaitForWaitGroups):
		assert.Fail(t, "should have been called")
		return
	}

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

	sd, err := shardedData.NewShardedData(defaultTestConfig)
	assert.Nil(t, err)

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

func TestShardedData_SearchNotFoundShouldRetEmptyMap(t *testing.T) {
	t.Parallel()

	sd, err := shardedData.NewShardedData(defaultTestConfig)
	assert.Nil(t, err)

	resp := sd.SearchData([]byte("aaaa"))
	assert.NotNil(t, resp)
	assert.Equal(t, 0, len(resp))
}

func TestShardedData_SearchFoundShouldRetResults(t *testing.T) {
	t.Parallel()

	sd, err := shardedData.NewShardedData(defaultTestConfig)
	assert.Nil(t, err)

	sd.AddData([]byte("aaa"), struct{}{}, 0)
	sd.AddData([]byte("aaaa"), struct{}{}, 4)
	sd.AddData([]byte("aaaa"), struct{}{}, 5)

	resp := sd.SearchData([]byte("aaaa"))
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp))
	assert.Equal(t, struct{}{}, resp[4])
	assert.Equal(t, struct{}{}, resp[5])
}
