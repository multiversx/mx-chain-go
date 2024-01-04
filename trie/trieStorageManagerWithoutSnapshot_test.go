package trie_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieStorageManagerWithoutSnapshot(t *testing.T) {
	t.Parallel()

	t.Run("nil trie storage manager should error", func(t *testing.T) {
		t.Parallel()

		ts, err := trie.NewTrieStorageManagerWithoutSnapshot(nil)
		assert.Equal(t, trie.ErrNilTrieStorage, err)
		assert.Nil(t, ts)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tsm, _ := trie.NewTrieStorageManager(trie.GetDefaultTrieStorageManagerParameters())
		ts, err := trie.NewTrieStorageManagerWithoutSnapshot(tsm)
		assert.Nil(t, err)
		assert.NotNil(t, ts)
	})
}

func TestTrieStorageManagerWithoutSnapshot_GetFromCurrentEpoch(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
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

	args := trie.GetDefaultTrieStorageManagerParameters()
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

	args := trie.GetDefaultTrieStorageManagerParameters()
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

	args := trie.GetDefaultTrieStorageManagerParameters()
	tsm, _ := trie.NewTrieStorageManager(args)
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(tsm)

	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	ts.TakeSnapshot("", nil, nil, iteratorChannels, nil, &trieMock.MockStatistics{}, 10)

	select {
	case <-iteratorChannels.LeavesChan:
	default:
		assert.Fail(t, "unclosed channel")
	}
}

func TestTrieStorageManagerWithoutSnapshot_GetLatestStorageEpoch(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
	tsm, _ := trie.NewTrieStorageManager(args)
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(tsm)

	epoch, err := ts.GetLatestStorageEpoch()
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), epoch)
}

func TestTrieStorageManagerWithoutSnapshot_SetEpochForPutOperationDoesNotPanic(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
	tsm, _ := trie.NewTrieStorageManager(args)
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(tsm)

	ts.SetEpochForPutOperation(5)
}

func TestTrieStorageManagerWithoutSnapshot_ShouldTakeSnapshot(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
	tsm, _ := trie.NewTrieStorageManager(args)
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(tsm)

	ok := ts.ShouldTakeSnapshot()
	assert.False(t, ok)
}

func TestTrieStorageManagerWithoutSnapshot_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var ts common.StorageManager
	assert.True(t, check.IfNil(ts))

	args := trie.GetDefaultTrieStorageManagerParameters()
	tsm, _ := trie.NewTrieStorageManager(args)
	ts, _ = trie.NewTrieStorageManagerWithoutSnapshot(tsm)
	assert.False(t, check.IfNil(ts))
}

func TestTrieStorageManagerWithoutSnapshot_IsSnapshotSupportedShouldReturnFalse(t *testing.T) {
	t.Parallel()

	args := trie.GetDefaultTrieStorageManagerParameters()
	tsm, _ := trie.NewTrieStorageManager(args)
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(tsm)

	assert.False(t, ts.IsSnapshotSupported())
}
