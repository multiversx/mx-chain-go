package leavesRetriever_test

import (
	"context"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	trieTest "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/trie/leavesRetriever"
	"github.com/stretchr/testify/assert"
)

func TestNewLeavesRetriever(t *testing.T) {
	t.Parallel()

	t.Run("nil db", func(t *testing.T) {
		t.Parallel()

		lr, err := leavesRetriever.NewLeavesRetriever(nil, &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{}, 100)
		assert.Nil(t, lr)
		assert.Equal(t, leavesRetriever.ErrNilDB, err)
	})
	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		lr, err := leavesRetriever.NewLeavesRetriever(testscommon.NewMemDbMock(), nil, &hashingMocks.HasherMock{}, 100)
		assert.Nil(t, lr)
		assert.Equal(t, leavesRetriever.ErrNilMarshaller, err)
	})
	t.Run("nil hasher", func(t *testing.T) {
		t.Parallel()

		lr, err := leavesRetriever.NewLeavesRetriever(testscommon.NewMemDbMock(), &marshallerMock.MarshalizerMock{}, nil, 100)
		assert.Nil(t, lr)
		assert.Equal(t, leavesRetriever.ErrNilHasher, err)
	})
	t.Run("new leaves retriever", func(t *testing.T) {
		t.Parallel()

		var lr common.TrieLeavesRetriever
		assert.True(t, check.IfNil(lr))

		lr, err := leavesRetriever.NewLeavesRetriever(testscommon.NewMemDbMock(), &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{}, 100)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(lr))
	})
}

func TestLeavesRetriever_GetLeaves(t *testing.T) {
	t.Parallel()

	t.Run("get leaves from new instance", func(t *testing.T) {
		t.Parallel()

		tr := trieTest.GetNewTrie()
		trieTest.AddDataToTrie(tr, 25)
		rootHash, _ := tr.RootHash()

		lr, _ := leavesRetriever.NewLeavesRetriever(tr.GetStorageManager(), &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{}, 100000)
		leaves, iteratorId, err := lr.GetLeaves(10, rootHash, []byte(""), context.Background())
		assert.Nil(t, err)
		assert.Equal(t, 10, len(leaves))
		assert.Equal(t, 32, len(iteratorId))
	})
	t.Run("get leaves from existing instance", func(t *testing.T) {
		t.Parallel()

		tr := trieTest.GetNewTrie()
		trieTest.AddDataToTrie(tr, 25)
		rootHash, _ := tr.RootHash()

		lr, _ := leavesRetriever.NewLeavesRetriever(tr.GetStorageManager(), &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{}, 10000000)
		leaves1, iteratorId1, err := lr.GetLeaves(10, rootHash, []byte(""), context.Background())
		assert.Nil(t, err)
		assert.Equal(t, 10, len(leaves1))
		assert.Equal(t, 32, len(iteratorId1))
		assert.Equal(t, 1, len(lr.GetIterators()))
		assert.Equal(t, 1, len(lr.GetLruIteratorIDs()))

		leaves2, iteratorId2, err := lr.GetLeaves(10, rootHash, iteratorId1, context.Background())
		assert.Nil(t, err)
		assert.Equal(t, 10, len(leaves2))
		assert.Equal(t, 32, len(iteratorId2))
		assert.Equal(t, 2, len(lr.GetIterators()))
		assert.Equal(t, 2, len(lr.GetLruIteratorIDs()))

		assert.NotEqual(t, leaves1, leaves2)
		assert.NotEqual(t, iteratorId1, iteratorId2)
	})
	t.Run("traversing a trie saves all iterator instances", func(t *testing.T) {
		t.Parallel()

		tr := trieTest.GetNewTrie()
		trieTest.AddDataToTrie(tr, 25)
		rootHash, _ := tr.RootHash()

		lr, _ := leavesRetriever.NewLeavesRetriever(tr.GetStorageManager(), &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{}, 10000000)
		leaves1, iteratorId1, err := lr.GetLeaves(10, rootHash, []byte(""), context.Background())
		assert.Nil(t, err)
		assert.Equal(t, 10, len(leaves1))
		assert.Equal(t, 32, len(iteratorId1))
		assert.Equal(t, 1, len(lr.GetIterators()))
		assert.Equal(t, 1, len(lr.GetLruIteratorIDs()))

		leaves2, iteratorId2, err := lr.GetLeaves(10, rootHash, iteratorId1, context.Background())
		assert.Nil(t, err)
		assert.Equal(t, 10, len(leaves2))
		assert.Equal(t, 32, len(iteratorId2))
		assert.Equal(t, 2, len(lr.GetIterators()))
		assert.Equal(t, 2, len(lr.GetLruIteratorIDs()))

		leaves3, iteratorId3, err := lr.GetLeaves(10, rootHash, iteratorId2, context.Background())
		assert.Nil(t, err)
		assert.Equal(t, 5, len(leaves3))
		assert.Equal(t, 0, len(iteratorId3))
		assert.Equal(t, 2, len(lr.GetIterators()))
		assert.Equal(t, 2, len(lr.GetLruIteratorIDs()))
	})
	t.Run("iterator instances are evicted in a lru manner", func(t *testing.T) {
		t.Parallel()

		tr := trieTest.GetNewTrie()
		trieTest.AddDataToTrie(tr, 25)
		rootHash, _ := tr.RootHash()
		maxSize := uint64(1000)

		lr, _ := leavesRetriever.NewLeavesRetriever(tr.GetStorageManager(), &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{}, maxSize)
		iterators := make([][]byte, 0)
		_, id1, _ := lr.GetLeaves(5, rootHash, []byte(""), context.Background())
		iterators = append(iterators, id1)
		_, id2, _ := lr.GetLeaves(5, rootHash, id1, context.Background())
		iterators = append(iterators, id2)
		_, id3, _ := lr.GetLeaves(5, rootHash, id2, context.Background())
		iterators = append(iterators, id3)

		assert.Equal(t, 3, len(lr.GetIterators()))
		for i, id := range lr.GetLruIteratorIDs() {
			assert.Equal(t, iterators[i], id)
		}

		_, id4, _ := lr.GetLeaves(5, rootHash, id3, context.Background())
		iterators = append(iterators, id4)
		assert.Equal(t, 3, len(lr.GetIterators()))
		for i, id := range lr.GetLruIteratorIDs() {
			assert.Equal(t, iterators[i+1], id)
		}
	})
	t.Run("when an iterator instance is used, it is moved in the front of the eviction queue", func(t *testing.T) {
		t.Parallel()

		tr := trieTest.GetNewTrie()
		trieTest.AddDataToTrie(tr, 25)
		rootHash, _ := tr.RootHash()
		maxSize := uint64(100000)

		lr, _ := leavesRetriever.NewLeavesRetriever(tr.GetStorageManager(), &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{}, maxSize)
		iterators := make([][]byte, 0)
		_, id1, _ := lr.GetLeaves(5, rootHash, []byte(""), context.Background())
		iterators = append(iterators, id1)
		leaves1, id2, _ := lr.GetLeaves(5, rootHash, id1, context.Background())
		iterators = append(iterators, id2)
		_, id3, _ := lr.GetLeaves(5, rootHash, id2, context.Background())
		iterators = append(iterators, id3)

		assert.Equal(t, 3, len(lr.GetIterators()))
		for i, id := range lr.GetLruIteratorIDs() {
			assert.Equal(t, iterators[i], id)
		}

		leaves2, id4, _ := lr.GetLeaves(5, rootHash, id1, context.Background())
		assert.Equal(t, leaves1, leaves2)
		assert.Equal(t, id2, id4)

		assert.Equal(t, 3, len(lr.GetIterators()))
		retrievedIterators := lr.GetLruIteratorIDs()
		assert.Equal(t, 3, len(retrievedIterators))
		assert.Equal(t, id2, retrievedIterators[0])
		assert.Equal(t, id3, retrievedIterators[1])
		assert.Equal(t, id1, retrievedIterators[2])
	})
	t.Run("iterator not found", func(t *testing.T) {
		t.Parallel()

		tr := trieTest.GetNewTrie()
		trieTest.AddDataToTrie(tr, 25)
		rootHash, _ := tr.RootHash()
		maxSize := uint64(100000)

		lr, _ := leavesRetriever.NewLeavesRetriever(tr.GetStorageManager(), &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{}, maxSize)
		leaves, id, err := lr.GetLeaves(5, rootHash, []byte("invalid iterator"), context.Background())
		assert.Nil(t, leaves)
		assert.Equal(t, 0, len(id))
		assert.Equal(t, leavesRetriever.ErrIteratorNotFound, err)
	})
}
