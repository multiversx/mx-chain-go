package trie_test

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"testing"

	"github.com/multiversx/mx-chain-go/trie"
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

	tsm, _ := trie.NewTrieStorageManager(trie.GetDefaultTrieStorageManagerParameters())
	ts, err := trie.NewTrieStorageManagerWithoutPruning(tsm)
	assert.Nil(t, err)
	assert.NotNil(t, ts)
}

func TestTrieStorageManagerWithoutPruning_IsPruningEnabled(t *testing.T) {
	t.Parallel()

	tsm, _ := trie.NewTrieStorageManager(trie.GetDefaultTrieStorageManagerParameters())
	ts, _ := trie.NewTrieStorageManagerWithoutPruning(tsm)
	assert.False(t, ts.IsPruningEnabled())
}

func TestTrieStorageManagerWithoutPruning_Remove(t *testing.T) {
	t.Parallel()

	tsm := &trie.StorageManagerExtensionStub{
		StorageManagerStub: &storageManager.StorageManagerStub{
			RemoveCalled: func(_ []byte) error {
				assert.Fail(t, "remove should not have been called")
				return nil
			},
		},
	}
	tsm.GetBaseTrieStorageManagerCalled = func() common.StorageManager {
		return tsm
	}

	ts, _ := trie.NewTrieStorageManagerWithoutPruning(tsm)
	assert.Nil(t, ts.Remove([]byte("key")))
}
