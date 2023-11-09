package trie_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTrieStorageManagerWithoutCheckpoints(t *testing.T) {
	t.Parallel()

	t.Run("nil storage manager should error", func(t *testing.T) {
		t.Parallel()

		ts, err := trie.NewTrieStorageManagerWithoutCheckpoints(nil)
		require.Equal(t, trie.ErrNilTrieStorage, err)
		require.Nil(t, ts)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tsm, _ := trie.NewTrieStorageManager(trie.GetDefaultTrieStorageManagerParameters())
		ts, err := trie.NewTrieStorageManagerWithoutCheckpoints(tsm)
		assert.Nil(t, err)
		assert.NotNil(t, ts)
	})
}

func TestTrieStorageManagerWithoutCheckpoints_SetCheckpoint(t *testing.T) {
	t.Parallel()

	tsm, _ := trie.NewTrieStorageManager(trie.GetDefaultTrieStorageManagerParameters())
	ts, _ := trie.NewTrieStorageManagerWithoutCheckpoints(tsm)

	iteratorChannels := &common.TrieIteratorChannels{
		LeavesChan: nil,
		ErrChan:    errChan.NewErrChanWrapper(),
	}
	ts.SetCheckpoint([]byte("rootHash"), make([]byte, 0), iteratorChannels, nil, &trieMock.MockStatistics{})
	assert.Equal(t, uint32(0), ts.PruningBlockingOperations())

	iteratorChannels = &common.TrieIteratorChannels{
		LeavesChan: make(chan core.KeyValueHolder),
		ErrChan:    errChan.NewErrChanWrapper(),
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

	tsm, _ := trie.NewTrieStorageManager(trie.GetDefaultTrieStorageManagerParameters())
	ts, _ := trie.NewTrieStorageManagerWithoutCheckpoints(tsm)

	assert.False(t, ts.AddDirtyCheckpointHashes([]byte("rootHash"), nil))
}
