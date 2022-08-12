package trie_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieStorageManagerWithoutPruningWithNilStorage(t *testing.T) {
	t.Parallel()

	ts, err := trie.NewTrieStorageManagerWithoutPruning(nil)
	assert.Nil(t, ts)
	assert.Equal(t, trie.ErrNilTrieStorage, err)
}

func TestNewTrieStorageManagerWithoutPruning(t *testing.T) {
	t.Parallel()

	tsm, _ := trie.NewTrieStorageManager(getNewTrieStorageManagerArgs())
	ts, err := trie.NewTrieStorageManagerWithoutPruning(tsm)
	assert.Nil(t, err)
	assert.NotNil(t, ts)
}

func TestTrieStorageManagerWithoutPruning_IsPruningEnabled(t *testing.T) {
	t.Parallel()

	tsm, _ := trie.NewTrieStorageManager(getNewTrieStorageManagerArgs())
	ts, _ := trie.NewTrieStorageManagerWithoutPruning(tsm)
	assert.False(t, ts.IsPruningEnabled())
}

func TestTrieStorageManagerWithoutPruning_Remove(t *testing.T) {
	t.Parallel()

	tsm, _ := trie.NewTrieStorageManager(getNewTrieStorageManagerArgs())
	ts, _ := trie.NewTrieStorageManagerWithoutPruning(tsm)
	assert.Nil(t, ts.Remove([]byte("key")))
}
