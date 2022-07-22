package trie_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/stretchr/testify/assert"
)

func getTrieStorageManagerOptions() trie.StorageManagerOptions {
	return trie.StorageManagerOptions{
		PruningEnabled:     true,
		SnapshotsEnabled:   true,
		CheckpointsEnabled: true,
	}
}

func TestTrieFactory_CreateWithoutPruning(t *testing.T) {
	t.Parallel()

	options := getTrieStorageManagerOptions()
	options.PruningEnabled = false
	tsm, err := trie.CreateTrieStorageManager(getNewTrieStorageManagerArgs(), options)
	assert.Nil(t, err)
	assert.Equal(t, "*trie.trieStorageManagerWithoutPruning", fmt.Sprintf("%T", tsm))
}

func TestTrieFactory_CreateWithoutSnapshot(t *testing.T) {
	t.Parallel()

	options := getTrieStorageManagerOptions()
	options.SnapshotsEnabled = false
	tsm, err := trie.CreateTrieStorageManager(getNewTrieStorageManagerArgs(), options)
	assert.Nil(t, err)
	assert.Equal(t, "*trie.trieStorageManagerWithoutSnapshot", fmt.Sprintf("%T", tsm))
}

func TestTrieFactory_CreateWithoutCheckpoints(t *testing.T) {
	t.Parallel()

	options := getTrieStorageManagerOptions()
	options.CheckpointsEnabled = false
	tsm, err := trie.CreateTrieStorageManager(getNewTrieStorageManagerArgs(), options)
	assert.Nil(t, err)
	assert.Equal(t, "*trie.trieStorageManagerWithoutCheckpoints", fmt.Sprintf("%T", tsm))
}

func TestTrieFactory_CreateNormal(t *testing.T) {
	t.Parallel()

	tsm, err := trie.CreateTrieStorageManager(getNewTrieStorageManagerArgs(), getTrieStorageManagerOptions())
	assert.Nil(t, err)
	assert.Equal(t, "*trie.trieStorageManager", fmt.Sprintf("%T", tsm))
}
