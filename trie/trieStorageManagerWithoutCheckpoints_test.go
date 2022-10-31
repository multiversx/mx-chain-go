package trie_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common"
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

	tsm, _ := trie.NewTrieStorageManager(getNewTrieStorageManagerArgs())
	ts, _ := trie.NewTrieStorageManagerWithoutCheckpoints(tsm)

	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: nil,
		ErrChan:    make(chan error, 1),
	}
	ts.SetCheckpoint([]byte("rootHash"), make([]byte, 0), iteratorChannels, nil, &trieMock.MockStatistics{})
	assert.Equal(t, uint32(0), ts.PruningBlockingOperations())

	iteratorChannels = &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    make(chan error, 1),
	}
	ts.SetCheckpoint([]byte("rootHash"), make([]byte, 0), iteratorChannels, nil, &trieMock.MockStatistics{})
	assert.Equal(t, uint32(0), ts.PruningBlockingOperations())

	select {
	case <-iteratorChannels.LeavesChan:
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
