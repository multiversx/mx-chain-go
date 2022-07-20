package trie_test

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go/trie"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrieFactory_CreateWithoutPruning(t *testing.T) {
	t.Parallel()

	tsm, err := trie.CreateTrieStorageManager(getNewTrieStorageManagerArgs(), false, true, true)
	assert.Nil(t, err)
	assert.Equal(t, "*trie.trieStorageManagerWithoutPruning", fmt.Sprintf("%T", tsm))
}

func TestTrieFactory_CreateWithoutSnapshot(t *testing.T) {
	t.Parallel()

	tsm, err := trie.CreateTrieStorageManager(getNewTrieStorageManagerArgs(), true, false, true)
	assert.Nil(t, err)
	assert.Equal(t, "*trie.trieStorageManagerWithoutSnapshot", fmt.Sprintf("%T", tsm))
}

func TestTrieFactory_CreateWithoutCheckpoints(t *testing.T) {
	t.Parallel()

	tsm, err := trie.CreateTrieStorageManager(getNewTrieStorageManagerArgs(), true, true, false)
	assert.Nil(t, err)
	assert.Equal(t, "*trie.trieStorageManagerWithoutCheckpoints", fmt.Sprintf("%T", tsm))
}

func TestTrieFactory_CreateNormal(t *testing.T) {
	t.Parallel()

	tsm, err := trie.CreateTrieStorageManager(getNewTrieStorageManagerArgs(), true, true, true)
	assert.Nil(t, err)
	assert.Equal(t, "*trie.trieStorageManager", fmt.Sprintf("%T", tsm))
}
