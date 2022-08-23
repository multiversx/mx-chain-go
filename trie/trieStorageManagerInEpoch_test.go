package trie

import (
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieStorageManagerInEpochInvalidStorageManagerType(t *testing.T) {
	t.Parallel()

	trieStorage := &testscommon.StorageManagerStub{}

	stsm, err := newTrieStorageManagerInEpoch(trieStorage, 0)
	assert.Nil(t, stsm)
	assert.True(t, strings.Contains(err.Error(), "invalid storage manager, type is"))
}

func TestNewTrieStorageManagerInEpochInvalidStorerType(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()
	trieStorage.mainStorer = memorydb.New()

	stsm, err := newTrieStorageManagerInEpoch(trieStorage, 0)
	assert.Nil(t, stsm)
	assert.True(t, strings.Contains(err.Error(), "invalid storer, type is"))
}

func TestNewTrieStorageManagerInEpoch(t *testing.T) {
	t.Parallel()

	_, trieStorage := newEmptyTrie()

	stsm, err := newTrieStorageManagerInEpoch(trieStorage, 0)
	assert.NotNil(t, stsm)
	assert.Nil(t, err)
}

func TestTrieStorageManagerInEpoch_GetFromEpoch(t *testing.T) {
	_, trieStorage := newEmptyTrie()
	getFromEpochCalled := false
	trieStorage.mainStorer = &trie.SnapshotPruningStorerStub{
		GetFromEpochCalled: func(_ []byte, _ uint32) ([]byte, error) {
			getFromEpochCalled = true
			return nil, nil
		},
	}
	stsm, _ := newTrieStorageManagerInEpoch(trieStorage, 0)

	_, _ = stsm.Get([]byte("key"))
	assert.True(t, getFromEpochCalled)
}
