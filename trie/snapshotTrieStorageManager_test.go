package trie

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewSnapshotTrieStorageManagerInvalidStorerType(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()

	stsm, err := newSnapshotTrieStorageManager(trieStorage, 0)
	assert.True(t, check.IfNil(stsm))
	assert.True(t, strings.Contains(err.Error(), "invalid storer, type is"))
}

func TestNewSnapshotTrieStorageManager(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{}
	stsm, err := newSnapshotTrieStorageManager(trieStorage, 0)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(stsm))
}

func TestNewSnapshotTrieStorageManager_GetFromOldEpochsWithoutCache(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()
	getFromOldEpochsWithoutCacheCalled := false
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
		GetFromOldEpochsWithoutAddingToCacheCalled: func(_ []byte) ([]byte, core.OptionalUint32, error) {
			getFromOldEpochsWithoutCacheCalled = true
			return nil, core.OptionalUint32{}, nil
		},
	}
	stsm, _ := newSnapshotTrieStorageManager(trieStorage, 0)

	_, _ = stsm.Get([]byte("key"))
	assert.True(t, getFromOldEpochsWithoutCacheCalled)
}

func TestNewSnapshotTrieStorageManager_PutWithoutCache(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()
	putWithoutCacheCalled := false
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
		PutInEpochWithoutCacheCalled: func(_ []byte, _ []byte, _ uint32) error {
			putWithoutCacheCalled = true
			return nil
		},
	}
	stsm, _ := newSnapshotTrieStorageManager(trieStorage, 0)

	_ = stsm.Put([]byte("key"), []byte("data"))
	assert.True(t, putWithoutCacheCalled)
}

func TestNewSnapshotTrieStorageManager_GetFromLastEpoch(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()
	getFromLastEpochCalled := false
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
		GetFromLastEpochCalled: func(_ []byte) ([]byte, error) {
			getFromLastEpochCalled = true
			return nil, nil
		},
	}
	stsm, _ := newSnapshotTrieStorageManager(trieStorage, 0)

	_, _ = stsm.GetFromLastEpoch([]byte("key"))
	assert.True(t, getFromLastEpochCalled)
}

func TestSnapshotTrieStorageManager_AlsoAddInPreviousEpoch(t *testing.T) {
	t.Parallel()

	t.Run("HasValue is false", func(t *testing.T) {
		val := []byte("val")
		_, trieStorage := newEmptyTrie()
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetFromOldEpochsWithoutAddingToCacheCalled: func(_ []byte) ([]byte, core.OptionalUint32, error) {
				return val, core.OptionalUint32{}, nil
			},
			PutInEpochCalled: func(_ []byte, _ []byte, _ uint32) error {
				assert.Fail(t, "this should not have been called")
				return nil
			},
		}
		stsm, _ := newSnapshotTrieStorageManager(trieStorage, 5)

		returnedVal, _ := stsm.Get([]byte("key"))
		assert.Equal(t, val, returnedVal)
	})
	t.Run("epoch is previous epoch", func(t *testing.T) {
		val := []byte("val")
		_, trieStorage := newEmptyTrie()
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetFromOldEpochsWithoutAddingToCacheCalled: func(_ []byte) ([]byte, core.OptionalUint32, error) {
				epoch := core.OptionalUint32{
					Value:    4,
					HasValue: true,
				}
				return []byte("val"), epoch, nil
			},
			PutInEpochCalled: func(_ []byte, _ []byte, _ uint32) error {
				assert.Fail(t, "this should not have been called")
				return nil
			},
		}
		stsm, _ := newSnapshotTrieStorageManager(trieStorage, 5)

		returnedVal, _ := stsm.Get([]byte("key"))
		assert.Equal(t, val, returnedVal)
	})
	t.Run("epoch is 0", func(t *testing.T) {
		val := []byte("val")
		_, trieStorage := newEmptyTrie()
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetFromOldEpochsWithoutAddingToCacheCalled: func(_ []byte) ([]byte, core.OptionalUint32, error) {
				epoch := core.OptionalUint32{
					Value:    4,
					HasValue: true,
				}
				return []byte("val"), epoch, nil
			},
			PutInEpochCalled: func(_ []byte, _ []byte, _ uint32) error {
				assert.Fail(t, "this should not have been called")
				return nil
			},
		}
		stsm, _ := newSnapshotTrieStorageManager(trieStorage, 0)

		returnedVal, _ := stsm.Get([]byte("key"))
		assert.Equal(t, val, returnedVal)
	})
	t.Run("key is ActiveDBKey", func(t *testing.T) {
		val := []byte("val")
		_, trieStorage := newEmptyTrie()
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetFromOldEpochsWithoutAddingToCacheCalled: func(_ []byte) ([]byte, core.OptionalUint32, error) {
				epoch := core.OptionalUint32{
					Value:    3,
					HasValue: true,
				}
				return []byte("val"), epoch, nil
			},
			PutInEpochCalled: func(_ []byte, _ []byte, _ uint32) error {
				assert.Fail(t, "this should not have been called")
				return nil
			},
		}
		stsm, _ := newSnapshotTrieStorageManager(trieStorage, 5)

		returnedVal, _ := stsm.Get([]byte(common.ActiveDBKey))
		assert.Equal(t, val, returnedVal)
	})
	t.Run("key is TrieSyncedKey", func(t *testing.T) {
		val := []byte("val")
		_, trieStorage := newEmptyTrie()
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetFromOldEpochsWithoutAddingToCacheCalled: func(_ []byte) ([]byte, core.OptionalUint32, error) {
				epoch := core.OptionalUint32{
					Value:    3,
					HasValue: true,
				}
				return []byte("val"), epoch, nil
			},
			PutInEpochCalled: func(_ []byte, _ []byte, _ uint32) error {
				assert.Fail(t, "this should not have been called")
				return nil
			},
		}
		stsm, _ := newSnapshotTrieStorageManager(trieStorage, 5)

		returnedVal, _ := stsm.Get([]byte(common.TrieSyncedKey))
		assert.Equal(t, val, returnedVal)
	})
	t.Run("add in previous epoch", func(t *testing.T) {
		val := []byte("val")
		putInEpochCalled := false
		_, trieStorage := newEmptyTrie()
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetFromOldEpochsWithoutAddingToCacheCalled: func(_ []byte) ([]byte, core.OptionalUint32, error) {
				epoch := core.OptionalUint32{
					Value:    3,
					HasValue: true,
				}
				return []byte("val"), epoch, nil
			},
			PutInEpochCalled: func(_ []byte, _ []byte, _ uint32) error {
				putInEpochCalled = true
				return nil
			},
		}
		stsm, _ := newSnapshotTrieStorageManager(trieStorage, 5)

		returnedVal, _ := stsm.Get([]byte("key"))
		assert.Equal(t, val, returnedVal)
		assert.True(t, putInEpochCalled)
	})
}
