package trie_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieStorageManagerWithoutCheckpointsOkVals(t *testing.T) {
	t.Parallel()

	tsm, _ := trie.NewTrieStorageManager(getNewTrieStorageManagerArgs())
	ts, err := trie.NewTrieStorageManagerWithoutCheckpoints(tsm)
	assert.Nil(t, err)
	assert.NotNil(t, ts)
}

func TestTrieStorageManagerWithoutCheckpoints_SetCheckpoint(t *testing.T) {
	t.Parallel()

	errChan := make(chan error, 1)
	tsm, _ := trie.NewTrieStorageManager(getNewTrieStorageManagerArgs())
	ts, _ := trie.NewTrieStorageManagerWithoutCheckpoints(tsm)

	ts.SetCheckpoint([]byte("rootHash"), make([]byte, 0), nil, errChan, &trieMock.MockStatistics{})
	assert.Equal(t, uint32(0), ts.PruningBlockingOperations())

	chLeaves := make(chan core.KeyValueHolder)
	ts.SetCheckpoint([]byte("rootHash"), make([]byte, 0), chLeaves, errChan, &trieMock.MockStatistics{})
	assert.Equal(t, uint32(0), ts.PruningBlockingOperations())

	select {
	case <-chLeaves:
	default:
		assert.Fail(t, "unclosed channel")
	}
}

func TestTrieStorageManagerWithoutCheckpoints_AddDirtyCheckpointHashes(t *testing.T) {
	t.Parallel()

	tsm, _ := trie.NewTrieStorageManager(getNewTrieStorageManagerArgs())
	ts, _ := trie.NewTrieStorageManagerWithoutCheckpoints(tsm)

	assert.False(t, ts.AddDirtyCheckpointHashes([]byte("rootHash"), nil))
}
