package trie

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTrieStorageManagerInEpochNilStorageManager(t *testing.T) {
	t.Parallel()

	tsmie, err := newTrieStorageManagerInEpoch(nil, 0)
	assert.Nil(t, tsmie)
	assert.Equal(t, ErrNilTrieStorage, err)
}

func TestNewTrieStorageManagerInEpochInvalidStorageManagerType(t *testing.T) {
	t.Parallel()

	trieStorage := &storageManager.StorageManagerStub{}

	tsmie, err := newTrieStorageManagerInEpoch(trieStorage, 0)
	assert.Nil(t, tsmie)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "invalid storage manager, type is"))
}

func TestNewTrieStorageManagerInEpochInvalidStorerType(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()
	trieStorage.mainStorer = database.NewMemDB()

	tsmie, err := newTrieStorageManagerInEpoch(trieStorage, 0)
	assert.Nil(t, tsmie)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "invalid storer, type is"))
}

func TestNewTrieStorageManagerInEpoch(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()

	tsmie, err := newTrieStorageManagerInEpoch(trieStorage, 0)
	assert.NotNil(t, tsmie)
	assert.Nil(t, err)
}

func TestTrieStorageManagerInEpoch_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var tsmie *trieStorageManagerInEpoch
	assert.True(t, tsmie.IsInterfaceNil())

	_, trieStorage := newEmptyTrie()
	tsmie, _ = newTrieStorageManagerInEpoch(trieStorage, 0)
	assert.False(t, tsmie.IsInterfaceNil())
}

func TestTrieStorageManagerInEpoch_GetFromEpoch(t *testing.T) {
	t.Parallel()

	t.Run("closed storage manager should error", func(t *testing.T) {
		t.Parallel()

		_, trieStorage := newEmptyTrie()
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetFromEpochCalled: func(_ []byte, _ uint32) ([]byte, error) {
				require.Fail(t, "should have not been called")
				return nil, nil
			},
		}
		tsmie, _ := newTrieStorageManagerInEpoch(trieStorage, 0)
		_ = tsmie.Close()

		_, err := tsmie.Get([]byte("key"))
		require.Equal(t, core.ErrContextClosing, err)
	})

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

	t.Run("closing error should work", func(t *testing.T) {
		t.Parallel()

		_, trieStorage := newEmptyTrie()
		getFromEpochCalled := false
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetFromEpochCalled: func(_ []byte, _ uint32) ([]byte, error) {
				getFromEpochCalled = true
				return nil, storage.ErrDBIsClosed
			},
		}
		tsmie, _ := newTrieStorageManagerInEpoch(trieStorage, 0)

		_, err := tsmie.Get([]byte("key"))
		assert.Equal(t, ErrKeyNotFound, err)
		assert.True(t, getFromEpochCalled)
	})

	t.Run("other error should work", func(t *testing.T) {
		t.Parallel()

		_, trieStorage := newEmptyTrie()
		getFromEpochCalled := false
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetFromEpochCalled: func(_ []byte, _ uint32) ([]byte, error) {
				getFromEpochCalled = true
				return nil, errors.New("not closing error")
			},
		}
		tsmie, _ := newTrieStorageManagerInEpoch(trieStorage, 0)

		_, err := tsmie.Get([]byte("key"))
		assert.Equal(t, ErrKeyNotFound, err)
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
