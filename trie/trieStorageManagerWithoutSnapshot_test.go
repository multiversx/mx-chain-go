package trie_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieStorageManagerWithoutSnapshot(t *testing.T) {
	t.Parallel()

	tsm, _ := trie.NewTrieStorageManager(getNewTrieStorageManagerArgs())
	ts, err := trie.NewTrieStorageManagerWithoutSnapshot(tsm)
	assert.Nil(t, err)
	assert.NotNil(t, ts)
}

func TestTrieStorageManagerWithoutSnapshot_GetFromCurrentEpoch(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	tsm, _ := trie.NewTrieStorageManager(args)
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(tsm)

	key := []byte("key")
	val := []byte("value")
	_ = args.MainStorer.Put(key, val)

	returnedVal, err := ts.GetFromCurrentEpoch(key)
	assert.Nil(t, err)
	assert.Equal(t, val, returnedVal)
}

func TestTrieStorageManagerWithoutSnapshot_PutInEpoch(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	tsm, _ := trie.NewTrieStorageManager(args)
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(tsm)

	key := []byte("key")
	val := []byte("value")

	err := ts.PutInEpoch(key, val, 100)
	assert.Nil(t, err)
	returnedVal, err := args.MainStorer.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, val, returnedVal)
}

func TestTrieStorageManagerWithoutSnapshot_PutInEpochWithoutCache(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	tsm, _ := trie.NewTrieStorageManager(args)
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(tsm)

	key := []byte("key")
	val := []byte("value")

	err := ts.PutInEpochWithoutCache(key, val, 100)
	assert.Nil(t, err)
	returnedVal, err := args.MainStorer.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, val, returnedVal)
}

func TestTrieStorageManagerWithoutSnapshot_TakeSnapshot(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	tsm, _ := trie.NewTrieStorageManager(args)
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(tsm)

	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    make(chan error, 1),
	}
	ts.TakeSnapshot(nil, nil, nil, iteratorChannels, nil, &trieMock.MockStatistics{}, 10)

	select {
	case <-iteratorChannels.LeavesChan:
	default:
		assert.Fail(t, "unclosed channel")
	}
}

func TestTrieStorageManagerWithoutSnapshot_GetLatestStorageEpoch(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	tsm, _ := trie.NewTrieStorageManager(args)
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(tsm)

	epoch, err := ts.GetLatestStorageEpoch()
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), epoch)
}

func TestTrieStorageManagerWithoutSnapshot_SetEpochForPutOperationDoesNotPanic(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	tsm, _ := trie.NewTrieStorageManager(args)
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(tsm)

	ts.SetEpochForPutOperation(5)
}

func TestTrieStorageManagerWithoutSnapshot_ShouldTakeSnapshot(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	tsm, _ := trie.NewTrieStorageManager(args)
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(tsm)

	ok := ts.ShouldTakeSnapshot()
	assert.False(t, ok)
}

func TestTrieStorageManagerWithoutSnapshot_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var ts common.StorageManager
	assert.True(t, check.IfNil(ts))

	args := getNewTrieStorageManagerArgs()
	tsm, _ := trie.NewTrieStorageManager(args)
	ts, _ = trie.NewTrieStorageManagerWithoutSnapshot(tsm)
	assert.False(t, check.IfNil(ts))
}
