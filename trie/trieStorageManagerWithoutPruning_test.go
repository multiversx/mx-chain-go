package trie

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieStorageManagerWithoutPruningWithNilDb(t *testing.T) {
	t.Parallel()

	ts, err := NewTrieStorageManagerWithoutPruning(nil)
	assert.Nil(t, ts)
	assert.Equal(t, ErrNilDatabase, err)
}

func TestNewTrieStorageManagerWithoutPruning(t *testing.T) {
	t.Parallel()

	ts, err := NewTrieStorageManagerWithoutPruning(testscommon.NewMemDbMock())
	assert.Nil(t, err)
	assert.NotNil(t, ts)
}

func TestTrieStorageManagerWithoutPruning_TakeSnapshotShouldWork(t *testing.T) {
	t.Parallel()

	errChan := make(chan error, 1)
	ts, _ := NewTrieStorageManagerWithoutPruning(testscommon.NewMemDbMock())
	ts.TakeSnapshot([]byte{}, make([]byte, 0), nil, errChan, &trie.MockStatistics{}, 0)

	chLeaves := make(chan core.KeyValueHolder)
	ts.TakeSnapshot([]byte("rootHash"), make([]byte, 0), chLeaves, errChan, &trie.MockStatistics{}, 0)

	select {
	case <-chLeaves:
	default:
		assert.Fail(t, "unclosed channel")
	}
}

func TestTrieStorageManagerWithoutPruning_SetCheckpointShouldWork(t *testing.T) {
	t.Parallel()

	errChan := make(chan error, 1)
	ts, _ := NewTrieStorageManagerWithoutPruning(testscommon.NewMemDbMock())
	ts.SetCheckpoint(make([]byte, 0), make([]byte, 0), nil, errChan, &trie.MockStatistics{})

	chLeaves := make(chan core.KeyValueHolder)
	ts.SetCheckpoint([]byte("rootHash"), make([]byte, 0), chLeaves, errChan, &trie.MockStatistics{})

	select {
	case <-chLeaves:
	default:
		assert.Fail(t, "unclosed channel")
	}
}

func TestTrieStorageManagerWithoutPruning_IsPruningEnabled(t *testing.T) {
	t.Parallel()

	ts, _ := NewTrieStorageManagerWithoutPruning(testscommon.NewMemDbMock())
	assert.False(t, ts.IsPruningEnabled())
}

func TestTrieStorageManagerWithoutPruning_Close(t *testing.T) {
	t.Parallel()

	closeCalled := false
	ts, _ := NewTrieStorageManagerWithoutPruning(&storageStubs.StorerStub{
		CloseCalled: func() error {
			closeCalled = true
			return nil
		},
	})
	err := ts.Close()
	assert.Nil(t, err)
	assert.True(t, closeCalled)
}

func TestTrieStorageManagerWithoutPruning_AddDirtyCheckpointHashes(t *testing.T) {
	t.Parallel()

	ts, _ := NewTrieStorageManagerWithoutPruning(testscommon.NewMemDbMock())
	assert.False(t, ts.AddDirtyCheckpointHashes([]byte("key"), make(common.ModifiedHashes)))
}

func TestTrieStorageManagerWithoutPruning_Remove(t *testing.T) {
	t.Parallel()

	ts, _ := NewTrieStorageManagerWithoutPruning(testscommon.NewMemDbMock())
	assert.Nil(t, ts.Remove([]byte("key")))
}

func TestTrieStorageManagerWithoutPruning_ShouldTakeSnapshot(t *testing.T) {
	t.Parallel()

	ts, _ := NewTrieStorageManagerWithoutPruning(testscommon.NewMemDbMock())
	assert.False(t, ts.ShouldTakeSnapshot())
}
