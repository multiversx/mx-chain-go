package trie

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDepthFirstTrieSyncer_InvalidParametersShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument(time.Minute)
	arg.RequestHandler = nil
	d, err := NewDepthFirstTrieSyncer(arg)
	assert.True(t, check.IfNil(d))
	assert.Equal(t, ErrNilRequestHandler, err)
}

func TestNewDepthFirstTrieSyncer(t *testing.T) {
	t.Parallel()

	arg := createMockArgument(time.Minute)
	d, err := NewDepthFirstTrieSyncer(arg)
	assert.False(t, check.IfNil(d))
	assert.Nil(t, err)
}

func TestDepthFirstTrieSyncer_StartSyncingNilRootHashShouldReturnNil(t *testing.T) {
	t.Parallel()

	arg := createMockArgument(time.Minute)
	d, _ := NewDepthFirstTrieSyncer(arg)
	err := d.StartSyncing(nil, context.Background())

	assert.Nil(t, err)
}

func TestDepthFirstTrieSyncer_StartSyncingEmptyRootHashShouldReturnNil(t *testing.T) {
	t.Parallel()

	arg := createMockArgument(time.Minute)
	d, _ := NewDepthFirstTrieSyncer(arg)
	err := d.StartSyncing(common.EmptyTrieHash, context.Background())

	assert.Nil(t, err)
}

func TestDepthFirstTrieSyncer_StartSyncingNilContextShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument(time.Minute)
	d, _ := NewDepthFirstTrieSyncer(arg)
	err := d.StartSyncing(bytes.Repeat([]byte{1}, len(common.EmptyTrieHash)), nil)

	assert.Equal(t, ErrNilContext, err)
}

func TestDepthFirstTrieSyncer_StartSyncingCanTimeout(t *testing.T) {
	numKeysValues := 10
	trSource, _ := createInMemoryTrie()
	addDataToTrie(numKeysValues, trSource)
	_ = trSource.Commit()
	roothash, _ := trSource.RootHash()
	log.Info("source trie", "root hash", roothash)

	arg := createMockArgument(time.Minute)

	d, _ := NewDepthFirstTrieSyncer(arg)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()

	err := d.StartSyncing(roothash, ctx)
	require.Equal(t, core.ErrContextClosing, err)
}

func TestDepthFirstTrieSyncer_StartSyncingTimeoutNoNodesReceived(t *testing.T) {
	numKeysValues := 10
	trSource, _ := createInMemoryTrie()
	addDataToTrie(numKeysValues, trSource)
	_ = trSource.Commit()
	roothash, _ := trSource.RootHash()
	log.Info("source trie", "root hash", roothash)

	arg := createMockArgument(time.Second)

	d, _ := NewDepthFirstTrieSyncer(arg)

	err := d.StartSyncing(roothash, context.Background())
	require.Equal(t, ErrTrieSyncTimeout, err)
}

func TestDepthFirstTrieSyncer_StartSyncingNewTrieShouldWork(t *testing.T) {
	numKeysValues := 100
	trSource, _ := createInMemoryTrie()
	addDataToTrie(numKeysValues, trSource)
	_ = trSource.Commit()
	roothash, _ := trSource.RootHash()
	log.Info("source trie", "root hash", roothash)

	arg := createMockArgument(time.Minute)
	arg.RequestHandler = createRequesterResolver(trSource, arg.InterceptedNodes, nil)
	arg.LeavesChan = make(chan core.KeyValueHolder, 110)

	d, _ := NewDepthFirstTrieSyncer(arg)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()

	err := d.StartSyncing(roothash, ctx)
	require.Nil(t, err)

	tsm, _ := arg.DB.(*trieStorageManager)
	db, _ := tsm.mainStorer.(storage.Persister)
	trie, _ := createInMemoryTrieFromDB(db)
	trie, _ = trie.Recreate(holders.NewDefaultRootHashesHolder(roothash))
	require.False(t, check.IfNil(trie))

	var val []byte
	for i := 0; i < numKeysValues; i++ {
		keyVal := hasherMock.Compute(fmt.Sprintf("%d", i))
		val, _, err = trie.Get(keyVal)
		require.Nil(t, err)
		require.Equal(t, keyVal, val)
	}

	assert.Equal(t, uint64(numKeysValues), d.NumLeaves())
	assert.True(t, d.NumTrieNodes() > d.NumLeaves())
	assert.True(t, d.NumBytes() > 0)
	assert.True(t, d.Duration() > 0)

	wg := &sync.WaitGroup{}
	wg.Add(numKeysValues)

	numLeavesOnChan := 0
	go func() {
		for range arg.LeavesChan {
			numLeavesOnChan++
			wg.Done()
		}
	}()

	wg.Wait()

	assert.Equal(t, numKeysValues, numLeavesOnChan)

	log.Info("synced trie",
		"num trie nodes", d.NumTrieNodes(),
		"num leaves", d.NumLeaves(),
		"data size", core.ConvertBytes(d.NumBytes()),
		"duration", d.Duration())
}

func TestDepthFirstTrieSyncer_StartSyncingPartiallyFilledTrieShouldWork(t *testing.T) {
	t.Skip("todo: update this test to work with trie sync that only uses the cache and not the DB (get node from cache only)")

	numKeysValues := 100
	trSource, memUnitSource := createInMemoryTrie()
	addDataToTrie(numKeysValues, trSource)
	_ = trSource.Commit()
	roothash, _ := trSource.RootHash()
	log.Info("source trie", "root hash", roothash)

	arg := createMockArgument(time.Minute)

	exceptionHashes := make([][]byte, 0)
	// copy half of the nodes from source to destination, add them also to exception list and then try to sync the trie
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

	d, _ := NewDepthFirstTrieSyncer(arg)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*30)
	defer cancelFunc()

	err := d.StartSyncing(roothash, ctx)
	require.Nil(t, err)

	tsm, _ := arg.DB.(*trieStorageManager)
	db, _ := tsm.mainStorer.(storage.Persister)
	trie, _ := createInMemoryTrieFromDB(db)
	trie, _ = trie.Recreate(holders.NewDefaultRootHashesHolder(roothash))
	require.False(t, check.IfNil(trie))

	var val []byte
	for i := 0; i < numKeysValues; i++ {
		keyVal := hasherMock.Compute(fmt.Sprintf("%d", i))
		val, _, err = trie.Get(keyVal)
		require.Nil(t, err)
		require.Equal(t, keyVal, val)
	}
}
