package trie

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewSnapshotTrieStorageManagerInvalidStorerType(t *testing.T) {
	_, trieStorage := newEmptyTrie()

	stsm, err := newSnapshotTrieStorageManager(trieStorage, 0)
	assert.Nil(t, stsm)
	assert.True(t, strings.Contains(err.Error(), "invalid storer type"))
}

func TestNewSnapshotTrieStorageManager(t *testing.T) {
	_, trieStorage := newEmptyTrie()
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{}
	stsm, err := newSnapshotTrieStorageManager(trieStorage, 0)
	assert.Nil(t, err)
	assert.NotNil(t, stsm)
}

func TestNewSnapshotTrieStorageManager_Put(t *testing.T) {
	_, trieStorage := newEmptyTrie()
	putInEpochWithoutCacheCalled := false
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
		PutInEpochWithoutCacheCalled: func(_ []byte, _ []byte, _ uint32) error {
			putInEpochWithoutCacheCalled = true
			return nil
		},
	}
	stsm, _ := newSnapshotTrieStorageManager(trieStorage, 2)

	err := stsm.Put([]byte("key"), []byte("val"))
	assert.Nil(t, err)
	assert.True(t, putInEpochWithoutCacheCalled)

	stsm, _ = newSnapshotTrieStorageManager(trieStorage, 0)
	err = stsm.Put([]byte("key"), []byte("val"))
	assert.NotNil(t, err)
}

func TestNewSnapshotTrieStorageManager_GetFromOldEpochsWithoutCache(t *testing.T) {
	_, trieStorage := newEmptyTrie()
	getFromOldEpochsWithoutCacheCalled := false
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
		GetFromOldEpochsWithoutAddingToCacheCalled: func(_ []byte, _ int) ([]byte, error) {
			getFromOldEpochsWithoutCacheCalled = true
			return nil, nil
		},
	}
	stsm, _ := newSnapshotTrieStorageManager(trieStorage, 0)

	_, _ = stsm.GetFromOldEpochsWithoutAddingToCache([]byte("key"), 1)
	assert.True(t, getFromOldEpochsWithoutCacheCalled)
}

func TestNewSnapshotTrieStorageManager_GetFromEpochWithoutCache(t *testing.T) {
	_, trieStorage := newEmptyTrie()
	getFromEpochWithoutCacheCalled := false
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
		GetFromEpochWithoutCacheCalled: func(_ []byte, _ uint32) ([]byte, error) {
			getFromEpochWithoutCacheCalled = true
			return nil, nil
		},
	}
	stsm, _ := newSnapshotTrieStorageManager(trieStorage, 0)

	_, _ = stsm.GetFromEpochWithoutCache([]byte("key"), 1)
	assert.True(t, getFromEpochWithoutCacheCalled)
}

func TestNewSnapshotTrieStorageManager_RemoveFromCurrentEpoch(t *testing.T) {
	_, trieStorage := newEmptyTrie()
	removeFromCurrentEpochCalled := false
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
		RemoveFromCurrentEpochCalled: func(_ []byte) error {
			removeFromCurrentEpochCalled = true
			return nil
		},
	}
	stsm, _ := newSnapshotTrieStorageManager(trieStorage, 0)

	_ = stsm.RemoveFromCurrentEpoch([]byte("key"))
	assert.True(t, removeFromCurrentEpochCalled)
}

func TestNewSnapshotTrieStorageManager_GetLatestStorageEpoch(t *testing.T) {
	_, trieStorage := newEmptyTrie()
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{}
	stsm, _ := newSnapshotTrieStorageManager(trieStorage, 5)

	epoch, err := stsm.GetLatestStorageEpoch()
	assert.Nil(t, err)
	assert.Equal(t, uint32(5), epoch)
}

func TestNewSnapshotTrieStorageManager_PutWithoutCache(t *testing.T) {
	_, trieStorage := newEmptyTrie()
	putWithoutCacheCalled := false
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
		PutInEpochWithoutCacheCalled: func(_ []byte, _ []byte, _ uint32) error {
			putWithoutCacheCalled = true
			return nil
		},
	}
	stsm, _ := newSnapshotTrieStorageManager(trieStorage, 0)

	_ = stsm.PutInEpochWithoutCache([]byte("key"), []byte("data"), 0)
	assert.True(t, putWithoutCacheCalled)
}
