package trie

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/trie/evictionWaitingList"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var marshalizer = &mock.MarshalizerMock{}
var hasher = &mock.HasherMock{}

func createMemUnit() storage.Storer {
	capacity := uint32(10)
	shards := uint32(1)
	sizeInBytes := uint64(0)
	cache, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: capacity, Shards: shards, SizeInBytes: sizeInBytes})
	persist, _ := memorydb.NewlruDB(100000)
	unit, _ := storageUnit.NewStorageUnit(cache, persist)

	return unit
}

// CreateTrieStorageManager creates the trie storage manager for the tests
func createTrieStorageManager(store storage.Storer) (data.StorageManager, storage.Storer) {
	ewl, _ := evictionWaitingList.NewEvictionWaitingList(100, memorydb.New(), marshalizer)
	tempDir, _ := ioutil.TempDir("", "trie")
	cfg := config.DBConfig{
		FilePath:          tempDir,
		Type:              string(storageUnit.LvlDBSerial),
		BatchDelaySeconds: 4,
		MaxBatchSize:      10000,
		MaxOpenFiles:      10,
	}
	generalCfg := config.TrieStorageManagerConfig{
		PruningBufferLen:   1000,
		SnapshotsBufferLen: 10,
		MaxSnapshots:       2,
	}
	tsm, _ := NewTrieStorageManager(store, marshalizer, hasher, cfg, ewl, generalCfg)

	return tsm, store
}

func createInMemoryTrie() (data.Trie, storage.Storer) {
	memUnit := createMemUnit()
	tsm, _ := createTrieStorageManager(memUnit)
	tr, _ := NewTrie(tsm, marshalizer, hasher, 6)

	return tr, memUnit
}

func createInMemoryTrieFromDB(db storage.Persister) (data.Trie, storage.Storer) {
	capacity := uint32(10)
	shards := uint32(1)
	sizeInBytes := uint64(0)
	cache, _ := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: capacity, Shards: shards, SizeInBytes: sizeInBytes})
	unit, _ := storageUnit.NewStorageUnit(cache, db)

	tsm, _ := createTrieStorageManager(unit)
	tr, _ := NewTrie(tsm, marshalizer, hasher, 6)

	return tr, unit
}

func addDataToTrie(numKeysValues int, tr data.Trie) {
	for i := 0; i < numKeysValues; i++ {
		keyVal := hasher.Compute(fmt.Sprintf("%d", i))

		_ = tr.Update(keyVal, keyVal)
	}
}

func createRequesterResolver(completeTrie data.Trie, interceptedNodes storage.Cacher, exceptionHashes [][]byte) RequestHandler {
	return &mock.RequestHandlerStub{
		RequestTrieNodesCalled: func(destShardID uint32, hashes [][]byte, topic string) {
			for _, hash := range hashes {
				if hashInList(hash, exceptionHashes) {
					continue
				}

				buff, err := completeTrie.GetSerializedNode(hash)
				if err != nil {
					continue
				}

				var n *InterceptedTrieNode
				n, err = NewInterceptedTrieNode(buff, marshalizer, hasher)
				if err != nil {
					continue
				}

				interceptedNodes.Put(hash, n, 0)
			}
		},
	}
}

func hashInList(hash []byte, list [][]byte) bool {
	for _, h := range list {
		if bytes.Equal(h, hash) {
			return true
		}
	}

	return false
}

func TestNewDoubleListTrieSyncer_InvalidParametersShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.RequestHandler = nil
	d, err := NewDoubleListTrieSyncer(arg)
	assert.True(t, check.IfNil(d))
	assert.Equal(t, ErrNilRequestHandler, err)
}

func TestNewDoubleListTrieSyncer(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	d, err := NewDoubleListTrieSyncer(arg)
	assert.False(t, check.IfNil(d))
	assert.Nil(t, err)
}

func TestDoubleListTrieSyncer_StartSyncingNilRootHashShouldReturnNil(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	d, _ := NewDoubleListTrieSyncer(arg)
	err := d.StartSyncing(nil, context.Background())

	assert.Nil(t, err)
}

func TestDoubleListTrieSyncer_StartSyncingEmptyRootHashShouldReturnNil(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	d, _ := NewDoubleListTrieSyncer(arg)
	err := d.StartSyncing(EmptyTrieHash, context.Background())

	assert.Nil(t, err)
}

func TestDoubleListTrieSyncer_StartSyncingNilContextShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	d, _ := NewDoubleListTrieSyncer(arg)
	err := d.StartSyncing(bytes.Repeat([]byte{1}, len(EmptyTrieHash)), nil)

	assert.Equal(t, ErrNilContext, err)
}

func TestDoubleListTrieSyncer_StartSyncingCanTimeout(t *testing.T) {
	numKeysValues := 10
	trSource, _ := createInMemoryTrie()
	addDataToTrie(numKeysValues, trSource)
	_ = trSource.Commit()
	roothash, _ := trSource.RootHash()
	log.Info("source trie", "root hash", roothash)

	arg := createMockArgument()

	d, _ := NewDoubleListTrieSyncer(arg)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()

	err := d.StartSyncing(roothash, ctx)
	require.Equal(t, ErrContextClosing, err)
}

func TestDoubleListTrieSyncer_StartSyncingNewTrieShouldWork(t *testing.T) {
	numKeysValues := 100
	trSource, _ := createInMemoryTrie()
	addDataToTrie(numKeysValues, trSource)
	_ = trSource.Commit()
	roothash, _ := trSource.RootHash()
	log.Info("source trie", "root hash", roothash)

	arg := createMockArgument()
	arg.RequestHandler = createRequesterResolver(trSource, arg.InterceptedNodes, nil)

	d, _ := NewDoubleListTrieSyncer(arg)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()

	err := d.StartSyncing(roothash, ctx)
	require.Nil(t, err)

	trie, _ := createInMemoryTrieFromDB(arg.DB.(*mock.MemDbMock))
	trie, _ = trie.Recreate(roothash)
	require.False(t, check.IfNil(trie))

	var val []byte
	for i := 0; i < numKeysValues; i++ {
		keyVal := hasher.Compute(fmt.Sprintf("%d", i))
		val, err = trie.Get(keyVal)
		require.Nil(t, err)
		require.Equal(t, keyVal, val)
	}
}

func TestDoubleListTrieSyncer_StartSyncingPartiallyFilledTrieShouldWork(t *testing.T) {
	numKeysValues := 100
	trSource, memUnitSource := createInMemoryTrie()
	addDataToTrie(numKeysValues, trSource)
	_ = trSource.Commit()
	roothash, _ := trSource.RootHash()
	log.Info("source trie", "root hash", roothash)

	arg := createMockArgument()

	exceptionHashes := make([][]byte, 0)
	//copy half of the nodes from source to destination, add them also to exception list and than try to sync the trie
	numKeysCopied := 0
	memUnitSource.RangeKeys(func(key []byte, val []byte) bool {
		if numKeysCopied >= numKeysValues/2 {
			return false
		}
		_ = arg.DB.Put(key, val)
		exceptionHashes = append(exceptionHashes, key)
		numKeysCopied++
		return true
	})

	log.Info("exception list has", "num elements", len(exceptionHashes))

	arg.RequestHandler = createRequesterResolver(trSource, arg.InterceptedNodes, exceptionHashes)

	d, _ := NewDoubleListTrieSyncer(arg)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()

	err := d.StartSyncing(roothash, ctx)
	require.Nil(t, err)

	trie, _ := createInMemoryTrieFromDB(arg.DB.(*mock.MemDbMock))
	trie, _ = trie.Recreate(roothash)
	require.False(t, check.IfNil(trie))

	var val []byte
	for i := 0; i < numKeysValues; i++ {
		keyVal := hasher.Compute(fmt.Sprintf("%d", i))
		val, err = trie.Get(keyVal)
		require.Nil(t, err)
		require.Equal(t, keyVal, val)
	}
}
