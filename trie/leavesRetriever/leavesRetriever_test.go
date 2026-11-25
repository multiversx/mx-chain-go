package leavesRetriever_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	trieTest "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/multiversx/mx-chain-go/trie/leavesRetriever"
	"github.com/stretchr/testify/assert"
)

func TestNewLeavesRetriever(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		lr, err := leavesRetriever.NewLeavesRetriever(nil, &hashingMocks.HasherMock{}, 100)
		assert.Nil(t, lr)
		assert.Equal(t, leavesRetriever.ErrNilMarshaller, err)
	})
	t.Run("nil hasher", func(t *testing.T) {
		t.Parallel()

		lr, err := leavesRetriever.NewLeavesRetriever(&marshallerMock.MarshalizerMock{}, nil, 100)
		assert.Nil(t, lr)
		assert.Equal(t, leavesRetriever.ErrNilHasher, err)
	})
	t.Run("new leaves retriever", func(t *testing.T) {
		t.Parallel()

		var lr common.TrieLeavesRetriever
		assert.True(t, check.IfNil(lr))

		lr, err := leavesRetriever.NewLeavesRetriever(&marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{}, 100)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(lr))
	})
}

func newBufferingStorage(storage common.StorageManager) (common.StorageManager, *int, *int) {
	numEntered := 0
	numExited := 0
	stub := &storageManager.StorageManagerStub{
		GetCalled: func(key []byte) ([]byte, error) {
			if numEntered != numExited+1 {
				return nil, fmt.Errorf("storage read outside pruning buffering mode: entered %d, exited %d", numEntered, numExited)
			}

			return storage.Get(key)
		},
		EnterPruningBufferingModeCalled: func() {
			numEntered++
		},
		ExitPruningBufferingModeCalled: func() {
			numExited++
		},
	}

	return stub, &numEntered, &numExited
}

func TestLeavesRetriever_IterateLeaves(t *testing.T) {
	t.Parallel()

	tr := trieTest.GetNewTrie()
	trieTest.AddDataToTrie(tr, 25)
	rootHash, _ := tr.RootHash()
	leafParser := &trieMock.TrieLeafParserStub{
		ParseLeafCalled: func(key []byte, val []byte, version core.TrieNodeVersion) (core.KeyValueHolder, error) {
			return keyValStorage.NewKeyValStorage(key, val), nil
		},
	}
	lr, _ := leavesRetriever.NewLeavesRetriever(&marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{}, 100000)

	t.Run("nil leaves parser", func(t *testing.T) {
		t.Parallel()

		newState, err := lr.IterateLeaves(10, 0, [][]byte{rootHash}, tr.GetStorageManager(), nil, context.Background(), nil)
		assert.NotNil(t, err)
		assert.Nil(t, newState)
	})
	t.Run("invalid iterator state", func(t *testing.T) {
		t.Parallel()

		newState, err := lr.IterateLeaves(10, 0, [][]byte{[]byte("short")}, tr.GetStorageManager(), leafParser, context.Background(), nil)
		assert.NotNil(t, err)
		assert.Nil(t, newState)
	})
	t.Run("iterator state node not in storage", func(t *testing.T) {
		t.Parallel()

		missingNode := bytes.Repeat([]byte{1}, len(rootHash))
		newState, err := lr.IterateLeaves(10, 0, [][]byte{missingNode}, tr.GetStorageManager(), leafParser, context.Background(), nil)
		assert.True(t, errors.Is(err, trie.ErrIteratorNodeUnavailable))
		assert.Nil(t, newState)
	})
	t.Run("parser error returns no iterator state", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		failingParser := &trieMock.TrieLeafParserStub{
			ParseLeafCalled: func(key []byte, val []byte, version core.TrieNodeVersion) (core.KeyValueHolder, error) {
				return nil, expectedErr
			},
		}
		newState, err := lr.IterateLeaves(10, 0, [][]byte{rootHash}, tr.GetStorageManager(), failingParser, context.Background(), nil)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, newState)
	})
	t.Run("iterates the trie from the given storage in multiple calls", func(t *testing.T) {
		t.Parallel()

		leaves := make(map[string]string)
		visit := func(leaf core.KeyValueHolder) (bool, uint64, error) {
			leaves[string(leaf.Key())] = string(leaf.Value())
			return true, 0, nil
		}

		newState, err := lr.IterateLeaves(10, 0, [][]byte{rootHash}, tr.GetStorageManager(), leafParser, context.Background(), visit)
		assert.Nil(t, err)
		assert.Equal(t, 10, len(leaves))
		assert.Equal(t, 8, len(newState))

		newState, err = lr.IterateLeaves(0, 0, newState, tr.GetStorageManager(), leafParser, context.Background(), visit)
		assert.Nil(t, err)
		assert.Equal(t, 25, len(leaves))
		assert.Nil(t, newState)
	})
	t.Run("buffers pruning of the given storage while iterating", func(t *testing.T) {
		t.Parallel()

		storage, numEntered, numExited := newBufferingStorage(tr.GetStorageManager())
		visit := func(leaf core.KeyValueHolder) (bool, uint64, error) {
			assert.Equal(t, 1, *numEntered)
			assert.Equal(t, 0, *numExited)
			return true, 0, nil
		}
		_, err := lr.IterateLeaves(10, 0, [][]byte{rootHash}, storage, leafParser, context.Background(), visit)
		assert.Nil(t, err)
		assert.Equal(t, 1, *numEntered)
		assert.Equal(t, 1, *numExited)

		failingVisit := func(leaf core.KeyValueHolder) (bool, uint64, error) {
			return false, 0, errors.New("expected error")
		}
		_, err = lr.IterateLeaves(10, 0, [][]byte{rootHash}, storage, leafParser, context.Background(), failingVisit)
		assert.NotNil(t, err)
		assert.Equal(t, 2, *numEntered)
		assert.Equal(t, 2, *numExited)

		_, err = lr.IterateLeaves(10, 0, nil, storage, leafParser, context.Background(), visit)
		assert.NotNil(t, err)
		assert.Equal(t, 2, *numEntered)
		assert.Equal(t, 2, *numExited)
	})
}
