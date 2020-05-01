package trie

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/mock"
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
	ts.TakeSnapshot([]byte{})
}

func TestTrieStorageManagerWithoutPruning_SetCheckpointShouldNotPanic(t *testing.T) {
	t.Parallel()

	ts, _ := NewTrieStorageManagerWithoutPruning(mock.NewMemDbMock())
	ts.SetCheckpoint([]byte{})
}

func TestTrieStorageManagerWithoutPruning_PruneShouldNotPanic(t *testing.T) {
	t.Parallel()

	ts, _ := NewTrieStorageManagerWithoutPruning(mock.NewMemDbMock())
	ts.Prune([]byte{}, 0)
}

func TestTrieStorageManagerWithoutPruning_CancelPruneShouldNotPanic(t *testing.T) {
	t.Parallel()

	ts, _ := NewTrieStorageManagerWithoutPruning(mock.NewMemDbMock())
	ts.CancelPrune([]byte{}, 0)
}

func TestTrieStorageManagerWithoutPruning_MarkForEvictionShouldNotPanic(t *testing.T) {
	t.Parallel()

	ts, _ := NewTrieStorageManagerWithoutPruning(mock.NewMemDbMock())
	err := ts.MarkForEviction([]byte{}, map[string]struct{}{})
	assert.Nil(t, err)
}

func TestTrieStorageManagerWithoutPruning_IsPruningEnabled(t *testing.T) {
	t.Parallel()

	ts, _ := NewTrieStorageManagerWithoutPruning(mock.NewMemDbMock())
	assert.False(t, ts.IsPruningEnabled())
}
