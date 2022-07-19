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

	args := getNewTrieStorageManagerArgs()
	ts, err := trie.NewTrieStorageManagerWithoutSnapshot(args)
	assert.Nil(t, err)
	assert.NotNil(t, ts)
}

func TestTrieStorageManagerWithoutSnapshot_GetFromCurrentEpoch(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(args)

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
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(args)

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
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(args)

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
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(args)

	errChan := make(chan error, 1)
	leavesCh := make(chan core.KeyValueHolder)
	ts.TakeSnapshot(nil, nil, leavesCh, errChan, &trieMock.MockStatistics{}, 10)

	select {
	case <-leavesCh:
	default:
		assert.Fail(t, "unclosed channel")
	}
}

func TestTrieStorageManagerWithoutSnapshot_SetCheckpoint(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(args)

	errChan := make(chan error, 1)
	leavesCh := make(chan core.KeyValueHolder)
	ts.SetCheckpoint(nil, nil, leavesCh, errChan, &trieMock.MockStatistics{})

	select {
	case <-leavesCh:
	default:
		assert.Fail(t, "unclosed channel")
	}
}

func TestTrieStorageManagerWithoutSnapshot_GetLatestStorageEpoch(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(args)

	epoch, err := ts.GetLatestStorageEpoch()
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), epoch)
}

func TestTrieStorageManagerWithoutSnapshot_AddDirtyCheckpointHashes(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(args)

	shouldCheckpoint := ts.AddDirtyCheckpointHashes([]byte("rootHash"), make(map[string]struct{}))
	assert.False(t, shouldCheckpoint)
}

func TestTrieStorageManagerWithoutSnapshot_Remove(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(args)

	key := []byte("key")
	val := []byte("value")
	_ = args.MainStorer.Put(key, val)

	err := ts.Remove(key)
	assert.Nil(t, err)

	returnedVal, err := args.MainStorer.Get(key)
	assert.Nil(t, returnedVal)
	assert.NotNil(t, err)
}

func TestTrieStorageManagerWithoutSnapshot_SetEpochForPutOperationDoesNotPanic(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(args)

	ts.SetEpochForPutOperation(5)
}

func TestTrieStorageManagerWithoutSnapshot_ShouldTakeSnapshot(t *testing.T) {
	t.Parallel()

	args := getNewTrieStorageManagerArgs()
	ts, _ := trie.NewTrieStorageManagerWithoutSnapshot(args)

	ok := ts.ShouldTakeSnapshot()
	assert.False(t, ok)
}

func TestTrieStorageManagerWithoutSnapshot_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var ts common.StorageManager
	assert.True(t, check.IfNil(ts))

	args := getNewTrieStorageManagerArgs()
	ts, _ = trie.NewTrieStorageManagerWithoutSnapshot(args)
	assert.False(t, check.IfNil(ts))
}
