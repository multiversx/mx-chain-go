package trie

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewSnapshotTrieStorageManagerInvalidStorerType(t *testing.T) {
	t.Parallel()

	args := GetDefaultTrieStorageManagerParameters()
	args.MainStorer = testscommon.CreateMemUnit()
	trieStorage, _ := NewTrieStorageManager(args)

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

func TestSnapshotTrieStorageManager_GetWithoutAddingToCache(t *testing.T) {
	t.Parallel()

	t.Run("closed storage manager should error", func(t *testing.T) {
		t.Parallel()

		_, trieStorage := newEmptyTrie()
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetWithoutAddingToCacheCalled: func(key []byte, maxEpochToSearchFrom uint32) ([]byte, core.OptionalUint32, error) {
				return nil, core.OptionalUint32{}, core.ErrContextClosing
			},
		}
		stsm, _ := newSnapshotTrieStorageManager(trieStorage, 0)
		_ = trieStorage.Close()

		val, _, err := stsm.GetWithoutAddingToCache([]byte("key"), 0)
		assert.Equal(t, core.ErrContextClosing, err)
		assert.Nil(t, val)
	})
	t.Run("GetWithoutAddingToCache returns db closed should error", func(t *testing.T) {
		t.Parallel()

		_, trieStorage := newEmptyTrie()
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetWithoutAddingToCacheCalled: func(_ []byte, _ uint32) ([]byte, core.OptionalUint32, error) {
				return nil, core.OptionalUint32{}, storage.ErrDBIsClosed
			},
		}
		stsm, _ := newSnapshotTrieStorageManager(trieStorage, 0)

		val, _, err := stsm.GetWithoutAddingToCache([]byte("key"), 0)
		assert.Equal(t, storage.ErrDBIsClosed, err)
		assert.Nil(t, val)
	})
	t.Run("should work from old epochs without cache", func(t *testing.T) {
		t.Parallel()

		_, trieStorage := newEmptyTrie()
		getFromOldEpochsWithoutCacheCalled := false
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetWithoutAddingToCacheCalled: func(_ []byte, _ uint32) ([]byte, core.OptionalUint32, error) {
				getFromOldEpochsWithoutCacheCalled = true
				return nil, core.OptionalUint32{}, nil
			},
		}
		stsm, _ := newSnapshotTrieStorageManager(trieStorage, 0)

		_, _, _ = stsm.GetWithoutAddingToCache([]byte("key"), 0)
		assert.True(t, getFromOldEpochsWithoutCacheCalled)
	})
}

func TestSnapshotTrieStorageManager_PutInEpochWithoutCache(t *testing.T) {
	t.Parallel()

	t.Run("closed storage manager should error", func(t *testing.T) {
		t.Parallel()

		_, trieStorage := newEmptyTrie()
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			PutInEpochWithoutCacheCalled: func(_ []byte, _ []byte, _ uint32) error {
				return core.ErrContextClosing
			},
		}
		stsm, _ := newSnapshotTrieStorageManager(trieStorage, 0)
		_ = trieStorage.Close()

		err := stsm.PutInEpochWithoutCache([]byte("key"), []byte("data"))
		assert.Equal(t, core.ErrContextClosing, err)
	})
	t.Run("should work", func(t *testing.T) {
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

		_ = stsm.PutInEpochWithoutCache([]byte("key"), []byte("data"))
		assert.True(t, putWithoutCacheCalled)
	})
}

func TestSnapshotTrieStorageManager_GetFromLastEpoch(t *testing.T) {
	t.Parallel()

	t.Run("closed storage manager should error", func(t *testing.T) {
		t.Parallel()

		_, trieStorage := newEmptyTrie()
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetFromLastEpochCalled: func(_ []byte) ([]byte, error) {
				return nil, core.ErrContextClosing
			},
		}
		stsm, _ := newSnapshotTrieStorageManager(trieStorage, 0)
		_ = trieStorage.Close()

		val, err := stsm.GetFromLastEpoch([]byte("key"))
		assert.Equal(t, core.ErrContextClosing, err)
		assert.Nil(t, val)
	})
	t.Run("should work", func(t *testing.T) {
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
	})
}

func TestSnapshotTrieStorageManager_AlsoAddInPreviousEpoch(t *testing.T) {
	t.Parallel()

	t.Run("HasValue is false", func(t *testing.T) {
		val := []byte("val")
		_, trieStorage := newEmptyTrie()
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetWithoutAddingToCacheCalled: func(_ []byte, _ uint32) ([]byte, core.OptionalUint32, error) {
				return val, core.OptionalUint32{}, nil
			},
			PutInEpochCalled: func(_ []byte, _ []byte, _ uint32) error {
				assert.Fail(t, "this should not have been called")
				return nil
			},
		}
		stsm, _ := newSnapshotTrieStorageManager(trieStorage, 5)

		returnedVal, _, _ := stsm.GetWithoutAddingToCache([]byte("key"), 5)
		assert.Equal(t, val, returnedVal)
	})
	t.Run("epoch is previous epoch", func(t *testing.T) {
		val := []byte("val")
		_, trieStorage := newEmptyTrie()
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetWithoutAddingToCacheCalled: func(_ []byte, _ uint32) ([]byte, core.OptionalUint32, error) {
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

		returnedVal, _, _ := stsm.GetWithoutAddingToCache([]byte("key"), 5)
		assert.Equal(t, val, returnedVal)
	})
	t.Run("epoch is 0", func(t *testing.T) {
		val := []byte("val")
		_, trieStorage := newEmptyTrie()
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetWithoutAddingToCacheCalled: func(_ []byte, _ uint32) ([]byte, core.OptionalUint32, error) {
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

		returnedVal, _, _ := stsm.GetWithoutAddingToCache([]byte("key"), 0)
		assert.Equal(t, val, returnedVal)
	})
	t.Run("key is ActiveDBKey", func(t *testing.T) {
		val := []byte("val")
		_, trieStorage := newEmptyTrie()
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetWithoutAddingToCacheCalled: func(_ []byte, _ uint32) ([]byte, core.OptionalUint32, error) {
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

		returnedVal, _, _ := stsm.GetWithoutAddingToCache([]byte(common.ActiveDBKey), 5)
		assert.Equal(t, val, returnedVal)
	})
	t.Run("key is TrieSyncedKey", func(t *testing.T) {
		val := []byte("val")
		_, trieStorage := newEmptyTrie()
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetWithoutAddingToCacheCalled: func(_ []byte, _ uint32) ([]byte, core.OptionalUint32, error) {
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

		returnedVal, _, _ := stsm.GetWithoutAddingToCache([]byte(common.TrieSyncedKey), 5)
		assert.Equal(t, val, returnedVal)
	})
	t.Run("add in previous epoch", func(t *testing.T) {
		val := []byte("val")
		putInEpochCalled := false
		_, trieStorage := newEmptyTrie()
		trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
			GetWithoutAddingToCacheCalled: func(_ []byte, _ uint32) ([]byte, core.OptionalUint32, error) {
				epoch := core.OptionalUint32{
					Value:    3,
					HasValue: true,
				}
				return []byte("val"), epoch, nil
			},
			PutInEpochCalled: func(_ []byte, _ []byte, _ uint32) error {
				putInEpochCalled = true
				return errors.New("error for coverage only")
			},
		}
		stsm, _ := newSnapshotTrieStorageManager(trieStorage, 5)

		returnedVal, _, _ := stsm.GetWithoutAddingToCache([]byte("key"), 5)
		assert.Equal(t, val, returnedVal)
		assert.True(t, putInEpochCalled)
	})
}
