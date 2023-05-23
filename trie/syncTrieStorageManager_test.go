package trie

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewSyncTrieStorageManagerNilTsm(t *testing.T) {
	t.Parallel()

	stsm, err := NewSyncTrieStorageManager(nil)
	assert.Nil(t, stsm)
	assert.Equal(t, ErrNilTrieStorage, err)
}

func TestNewSyncTrieStorageManagerInvalidStorerType(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()

	stsm, err := NewSyncTrieStorageManager(trieStorage)
	assert.Nil(t, stsm)
	assert.True(t, strings.Contains(err.Error(), "invalid storer type"))
}

func TestNewSyncTrieStorageManager(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{}
	stsm, err := NewSyncTrieStorageManager(trieStorage)
	assert.Nil(t, err)
	assert.NotNil(t, stsm)
}

func TestNewSyncTrieStorageManager_PutInFirstEpoch(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()
	putInEpochCalled := 0
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
		PutInEpochCalled: func(_ []byte, _ []byte, _ uint32) error {
			putInEpochCalled++
			return nil
		},
		GetLatestStorageEpochCalled: func() (uint32, error) {
			return 0, nil
		},
	}
	stsm, _ := NewSyncTrieStorageManager(trieStorage)

	err := stsm.Put([]byte("key"), []byte("val"))
	assert.Nil(t, err)
	assert.Equal(t, 1, putInEpochCalled)
}

func TestNewSyncTrieStorageManager_PutInEpochError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	_, trieStorage := newEmptyTrie()
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
		PutInEpochCalled: func(_ []byte, _ []byte, _ uint32) error {
			return expectedErr
		},
	}
	stsm, _ := NewSyncTrieStorageManager(trieStorage)

	err := stsm.Put([]byte("key"), []byte("val"))
	assert.Equal(t, expectedErr, err)
}

func TestNewSyncTrieStorageManager_PutInEpoch(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()
	putInEpochCalled := 0
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
		PutInEpochCalled: func(_ []byte, _ []byte, _ uint32) error {
			putInEpochCalled++
			return nil
		},
		GetLatestStorageEpochCalled: func() (uint32, error) {
			return 5, nil
		},
	}
	stsm, _ := NewSyncTrieStorageManager(trieStorage)

	err := stsm.Put([]byte("key"), []byte("val"))
	assert.Nil(t, err)
	assert.Equal(t, 2, putInEpochCalled)
}
