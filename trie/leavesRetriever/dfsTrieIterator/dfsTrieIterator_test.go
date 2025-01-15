package dfsTrieIterator

import (
	"bytes"
	"context"
	"encoding/hex"
	"math"
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	trieTest "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/stretchr/testify/assert"
)

var maxSize = uint64(math.MaxUint64)

func TestNewIterator(t *testing.T) {
	t.Parallel()

	t.Run("nil db", func(t *testing.T) {
		t.Parallel()

		iterator, err := NewIterator([][]byte{[]byte("initial"), []byte("state")}, nil, &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{})
		assert.Nil(t, iterator)
		assert.Equal(t, trie.ErrNilDatabase, err)
	})
	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		iterator, err := NewIterator([][]byte{[]byte("initial"), []byte("state")}, testscommon.NewMemDbMock(), nil, &hashingMocks.HasherMock{})
		assert.Nil(t, iterator)
		assert.Equal(t, trie.ErrNilMarshalizer, err)
	})
	t.Run("nil hasher", func(t *testing.T) {
		t.Parallel()

		iterator, err := NewIterator([][]byte{[]byte("initial"), []byte("state")}, testscommon.NewMemDbMock(), &marshallerMock.MarshalizerMock{}, nil)
		assert.Nil(t, iterator)
		assert.Equal(t, trie.ErrNilHasher, err)
	})
	t.Run("empty initial state", func(t *testing.T) {
		t.Parallel()

		iterator, err := NewIterator([][]byte{}, testscommon.NewMemDbMock(), &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{})
		assert.Nil(t, iterator)
		assert.Equal(t, trie.ErrEmptyInitialIteratorState, err)
	})
	t.Run("invalid initial state", func(t *testing.T) {
		t.Parallel()

		iterator, err := NewIterator([][]byte{[]byte("invalid state")}, testscommon.NewMemDbMock(), &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{})
		assert.Nil(t, iterator)
		assert.Equal(t, trie.ErrInvalidIteratorState, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		initialState := [][]byte{
			bytes.Repeat([]byte{0}, 40),
			bytes.Repeat([]byte{1}, 40),
		}

		db, marshaller, hasher := trieTest.GetDefaultTrieParameters()
		iterator, err := NewIterator(initialState, db, marshaller, hasher)
		assert.Nil(t, err)

		assert.Equal(t, uint64(80), iterator.size)
		assert.Equal(t, 2, len(iterator.nextNodes))
	})
}

func TestDfsIterator_GetLeaves(t *testing.T) {
	t.Parallel()

	t.Run("context done returns retrieved leaves and saves iterator state", func(t *testing.T) {
		t.Parallel()

		tr := trieTest.GetNewTrie()
		numLeaves := 25
		expectedNumLeaves := 9
		numGetCalls := 0
		trieTest.AddDataToTrie(tr, numLeaves)
		rootHash, _ := tr.RootHash()

		ctx, cancel := context.WithCancel(context.Background())

		trieStorage := tr.GetStorageManager()
		dbWrapper := &storageManager.StorageManagerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				if numGetCalls == 15 {
					cancel()
				}
				numGetCalls++
				return trieStorage.Get(key)
			},
			PutCalled: func(key, data []byte) error {
				return trieStorage.Put(key, data)
			},
		}
		_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
		iterator, _ := NewIterator([][]byte{rootHash}, dbWrapper, marshaller, hasher)

		trieData, err := iterator.GetLeaves(numLeaves, maxSize, ctx)
		assert.Nil(t, err)
		assert.Equal(t, expectedNumLeaves, len(trieData))
	})
	t.Run("finishes iteration returns retrieved leaves", func(t *testing.T) {
		t.Parallel()

		tr := trieTest.GetNewTrie()
		numLeaves := 25
		trieTest.AddDataToTrie(tr, numLeaves)
		rootHash, _ := tr.RootHash()

		_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
		iterator, _ := NewIterator([][]byte{rootHash}, tr.GetStorageManager(), marshaller, hasher)

		trieData, err := iterator.GetLeaves(numLeaves, maxSize, context.Background())
		assert.Nil(t, err)
		assert.Equal(t, numLeaves, len(trieData))
	})
	t.Run("num leaves reached returns retrieved leaves and saves iterator context", func(t *testing.T) {
		t.Parallel()

		tr := trieTest.GetNewTrie()
		numLeaves := 25
		expectedNumRetrievedLeaves := 17
		trieTest.AddDataToTrie(tr, numLeaves)
		rootHash, _ := tr.RootHash()

		_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
		iterator, _ := NewIterator([][]byte{rootHash}, tr.GetStorageManager(), marshaller, hasher)

		trieData, err := iterator.GetLeaves(17, maxSize, context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedNumRetrievedLeaves, len(trieData))
	})
	t.Run("max size reached returns retrieved leaves and saves iterator context", func(t *testing.T) {
		t.Parallel()

		tr := trieTest.GetTrieWithData()
		expectedNumRetrievedLeaves := 2
		rootHash, _ := tr.RootHash()

		_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
		iterator, _ := NewIterator([][]byte{rootHash}, tr.GetStorageManager(), marshaller, hasher)

		iteratorMaxSize := uint64(100)
		trieData, err := iterator.GetLeaves(5, iteratorMaxSize, context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedNumRetrievedLeaves, len(trieData))
	})
	t.Run("retrieve all leaves in multiple calls", func(t *testing.T) {
		t.Parallel()

		tr := trieTest.GetNewTrie()
		numLeaves := 25
		trieTest.AddDataToTrie(tr, numLeaves)
		rootHash, _ := tr.RootHash()
		_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
		iterator, _ := NewIterator([][]byte{rootHash}, tr.GetStorageManager(), marshaller, hasher)

		numRetrievedLeaves := 0
		numIterations := 0
		for numRetrievedLeaves < numLeaves {
			trieData, err := iterator.GetLeaves(5, maxSize, context.Background())
			assert.Nil(t, err)

			numRetrievedLeaves += len(trieData)
			numIterations++
		}

		assert.Equal(t, numLeaves, numRetrievedLeaves)
		assert.Equal(t, 5, numIterations)
	})
	t.Run("retrieve leaves with nil context does not panic", func(t *testing.T) {
		t.Parallel()

		tr := trieTest.GetNewTrie()
		numLeaves := 25
		expectedNumRetrievedLeaves := 25
		trieTest.AddDataToTrie(tr, numLeaves)
		rootHash, _ := tr.RootHash()

		_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
		iterator, _ := NewIterator([][]byte{rootHash}, tr.GetStorageManager(), marshaller, hasher)

		trieData, err := iterator.GetLeaves(numLeaves, maxSize, nil)
		assert.Nil(t, err)
		assert.Equal(t, expectedNumRetrievedLeaves, len(trieData))
	})
}

func TestDfsIterator_GetIteratorState(t *testing.T) {
	t.Parallel()

	tr := trieTest.GetNewTrie()
	_ = tr.Update([]byte("doe"), []byte("reindeer"))
	_ = tr.Update([]byte("dog"), []byte("puppy"))
	_ = tr.Update([]byte("ddog"), []byte("cat"))
	_ = tr.Commit()
	rootHash, _ := tr.RootHash()
	_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
	iterator, _ := NewIterator([][]byte{rootHash}, tr.GetStorageManager(), marshaller, hasher)

	leaves, err := iterator.GetLeaves(2, maxSize, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 2, len(leaves))
	val, ok := leaves[hex.EncodeToString([]byte("doe"))]
	assert.True(t, ok)
	assert.Equal(t, hex.EncodeToString([]byte("reindeer")), val)
	val, ok = leaves[hex.EncodeToString([]byte("ddog"))]
	assert.True(t, ok)
	assert.Equal(t, hex.EncodeToString([]byte("cat")), val)

	iteratorState := iterator.GetIteratorState()
	assert.Equal(t, 1, len(iteratorState))
	hash := iteratorState[0][:hasher.Size()]
	key := iteratorState[0][hasher.Size():]
	assert.Equal(t, []byte{0x7, 0x6, 0xf, 0x6, 0x4, 0x6, 0x10}, key)
	leafBytes, err := tr.GetStorageManager().Get(hash)
	assert.Nil(t, err)
	assert.NotNil(t, leafBytes)
}

func TestDfsIterator_FinishedIteration(t *testing.T) {
	t.Parallel()

	tr := trieTest.GetNewTrie()
	numLeaves := 25
	trieTest.AddDataToTrie(tr, numLeaves)
	rootHash, _ := tr.RootHash()
	_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
	iterator, _ := NewIterator([][]byte{rootHash}, tr.GetStorageManager(), marshaller, hasher)

	numRetrievedLeaves := 0
	for numRetrievedLeaves < numLeaves {
		assert.False(t, iterator.FinishedIteration())
		trieData, err := iterator.GetLeaves(5, maxSize, context.Background())
		assert.Nil(t, err)

		numRetrievedLeaves += len(trieData)
	}

	assert.Equal(t, numLeaves, numRetrievedLeaves)
	assert.True(t, iterator.FinishedIteration())
}

func TestDfsIterator_Size(t *testing.T) {
	t.Parallel()

	tr := trieTest.GetNewTrie()
	numLeaves := 25
	trieTest.AddDataToTrie(tr, numLeaves)
	rootHash, _ := tr.RootHash()
	_, marshaller, hasher := trieTest.GetDefaultTrieParameters()

	iterator, _ := NewIterator([][]byte{rootHash}, tr.GetStorageManager(), marshaller, hasher)
	assert.Equal(t, uint64(32), iterator.Size()) // root hash
	assert.False(t, iterator.FinishedIteration())

	_, err := iterator.GetLeaves(5, maxSize, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint64(299), iterator.Size()) // 9 hashes + leaf key(3) + 8 x intermediary nodes key(8 * 1)
	assert.False(t, iterator.FinishedIteration())

	_, err = iterator.GetLeaves(5, maxSize, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint64(268), iterator.Size()) // 8 hashes + 2 x leaf keys(2 * 3) + 6 x intermediary nodes key(6*1)
	assert.False(t, iterator.FinishedIteration())

	_, err = iterator.GetLeaves(5, maxSize, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint64(165), iterator.Size()) // 5 hashes + 5 x intermediary nodes key(5*1)
	assert.False(t, iterator.FinishedIteration())

	_, err = iterator.GetLeaves(5, maxSize, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint64(101), iterator.Size()) // 3 hashes + leaf key(3) + 2 x intermediary nodes key(2*1)
	assert.False(t, iterator.FinishedIteration())

	_, err = iterator.GetLeaves(5, maxSize, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), iterator.Size())
	assert.True(t, iterator.FinishedIteration())
}
