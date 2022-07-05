package trie

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewSyncTrieStorageManagerInvalidStorerType(t *testing.T) {
	_, trieStorage := newEmptyTrie()

	stsm, err := NewSyncTrieStorageManager(trieStorage)
	assert.Nil(t, stsm)
	assert.True(t, strings.Contains(err.Error(), "invalid storer type"))
}

func TestNewSyncTrieStorageManager(t *testing.T) {
	_, trieStorage := newEmptyTrie()
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{}
	stsm, err := NewSyncTrieStorageManager(trieStorage)
	assert.Nil(t, err)
	assert.NotNil(t, stsm)
}

func TestNewSyncTrieStorageManager_PutInFirstEpoch(t *testing.T) {
	_, trieStorage := newEmptyTrie()
	putInEpochCalled := false
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
		PutInEpochCalled: func(_ []byte, _ []byte, _ uint32) error {
			putInEpochCalled = true
			return nil
		},
		GetLatestStorageEpochCalled: func() (uint32, error) {
			return 0, nil
		},
	}
	stsm, _ := NewSyncTrieStorageManager(trieStorage)

	err := stsm.Put([]byte("key"), []byte("val"))
	assert.Nil(t, err)
	assert.True(t, putInEpochCalled)
}

func TestNewSyncTrieStorageManager_PutInEpoch(t *testing.T) {
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
