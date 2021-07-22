package trie

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/mock"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
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

	ts, err := NewTrieStorageManagerWithoutPruning(mock.NewMemDbMock())
	assert.Nil(t, err)
	assert.NotNil(t, ts)
}

func TestTrieStorageManagerWithoutPruning_TakeSnapshotShouldNotPanic(t *testing.T) {
	t.Parallel()

	ts, _ := NewTrieStorageManagerWithoutPruning(mock.NewMemDbMock())
	ts.TakeSnapshot([]byte{}, true, nil)
}

func TestTrieStorageManagerWithoutPruning_SetCheckpointShouldNotPanic(t *testing.T) {
	t.Parallel()

	ts, _ := NewTrieStorageManagerWithoutPruning(mock.NewMemDbMock())
	ts.SetCheckpoint([]byte{}, nil)
}

func TestTrieStorageManagerWithoutPruning_IsPruningEnabled(t *testing.T) {
	t.Parallel()

	ts, _ := NewTrieStorageManagerWithoutPruning(mock.NewMemDbMock())
	assert.False(t, ts.IsPruningEnabled())
}

func TestTrieStorageManagerWithoutPruning_Close(t *testing.T) {
	t.Parallel()

	closeCalled := false
	ts, _ := NewTrieStorageManagerWithoutPruning(&mock.StorerStub{
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

	ts, _ := NewTrieStorageManagerWithoutPruning(mock.NewMemDbMock())
	assert.False(t, ts.AddDirtyCheckpointHashes([]byte("key"), make(temporary.ModifiedHashes)))
}

func TestTrieStorageManagerWithoutPruning_Remove(t *testing.T) {
	t.Parallel()

	ts, _ := NewTrieStorageManagerWithoutPruning(mock.NewMemDbMock())
	assert.Nil(t, ts.Remove([]byte("key")))
}
