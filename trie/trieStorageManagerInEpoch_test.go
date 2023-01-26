package trie

import (
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieStorageManagerInEpochNilStorageManager(t *testing.T) {
	t.Parallel()

	tsmie, err := newTrieStorageManagerInEpoch(nil, 0)
	assert.True(t, check.IfNil(tsmie))
	assert.Equal(t, ErrNilTrieStorage, err)
}

func TestNewTrieStorageManagerInEpochInvalidStorageManagerType(t *testing.T) {
	t.Parallel()

	trieStorage := &testscommon.StorageManagerStub{}

	tsmie, err := newTrieStorageManagerInEpoch(trieStorage, 0)
	assert.True(t, check.IfNil(tsmie))
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "invalid storage manager, type is"))
}

func TestNewTrieStorageManagerInEpochInvalidStorerType(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()
	trieStorage.mainStorer = database.NewMemDB()

	tsmie, err := newTrieStorageManagerInEpoch(trieStorage, 0)
	assert.True(t, check.IfNil(tsmie))
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "invalid storer, type is"))
}

func TestNewTrieStorageManagerInEpoch(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()

	tsmie, err := newTrieStorageManagerInEpoch(trieStorage, 0)
	assert.False(t, check.IfNil(tsmie))
	assert.Nil(t, err)
}

func TestTrieStorageManagerInEpoch_GetFromEpoch(t *testing.T) {
	t.Parallel()

	t.Run("epoch 0 does not panic", func(t *testing.T) {
		t.Parallel()

		_, trieStorage := newEmptyTrie()
		getFromEpochCalled := false
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetFromEpochCalled: func(_ []byte, _ uint32) ([]byte, error) {
				getFromEpochCalled = true
				return nil, nil
			},
		}
		tsmie, _ := newTrieStorageManagerInEpoch(trieStorage, 0)

		_, _ = tsmie.Get([]byte("key"))
		assert.True(t, getFromEpochCalled)
	})

	t.Run("getFromEpoch searches more storers", func(t *testing.T) {
		t.Parallel()

		_, trieStorage := newEmptyTrie()
		getFromCurrentEpochCalled := false
		getFromPreviousEpochCalled := false
		currentEpoch := uint32(5)
		expectedKey := []byte("key")
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
				assert.Equal(t, expectedKey, key)
				if epoch == currentEpoch {
					getFromCurrentEpochCalled = true
				}
				if epoch == currentEpoch-1 {
					getFromPreviousEpochCalled = true
				}
				return nil, nil
			},
		}
		tsmie, _ := newTrieStorageManagerInEpoch(trieStorage, 5)

		_, _ = tsmie.Get(expectedKey)
		assert.True(t, getFromCurrentEpochCalled)
		assert.True(t, getFromPreviousEpochCalled)
	})
}
