package dfsTrieIterator

import (
	"context"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	trieTest "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewIterator(t *testing.T) {
	t.Parallel()

	t.Run("nil db", func(t *testing.T) {
		t.Parallel()

		iterator, err := NewIterator([]byte("hash"), nil, &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{})
		assert.Nil(t, iterator)
		assert.Equal(t, trie.ErrNilDatabase, err)
	})
	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		iterator, err := NewIterator([]byte("hash"), testscommon.NewMemDbMock(), nil, &hashingMocks.HasherMock{})
		assert.Nil(t, iterator)
		assert.Equal(t, trie.ErrNilMarshalizer, err)
	})
	t.Run("nil hasher", func(t *testing.T) {
		t.Parallel()

		iterator, err := NewIterator([]byte("hash"), testscommon.NewMemDbMock(), &marshallerMock.MarshalizerMock{}, nil)
		assert.Nil(t, iterator)
		assert.Equal(t, trie.ErrNilHasher, err)
	})
	t.Run("invalid hash", func(t *testing.T) {
		t.Parallel()

		iterator, err := NewIterator([]byte("invalid hash"), testscommon.NewMemDbMock(), &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{})
		assert.Nil(t, iterator)
		assert.NotNil(t, err)
	})
	t.Run("initialize iterator with a valid hash", func(t *testing.T) {
		t.Parallel()

		tr := trieTest.GetNewTrie()
		_ = tr.Update([]byte("key1"), []byte("value1"))
		_ = tr.Commit()
		rootHash, _ := tr.RootHash()

		_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
		iterator, err := NewIterator(rootHash, tr.GetStorageManager(), marshaller, hasher)
		assert.Nil(t, err)

		assert.Equal(t, rootHash, iterator.rootHash)
		assert.Equal(t, uint64(15), iterator.size)
		assert.Equal(t, 1, len(iterator.nextNodes))
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
		iterator, _ := NewIterator(rootHash, dbWrapper, marshaller, hasher)

		trieData, err := iterator.GetLeaves(numLeaves, ctx)
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
		iterator, _ := NewIterator(rootHash, tr.GetStorageManager(), marshaller, hasher)

		trieData, err := iterator.GetLeaves(numLeaves, context.Background())
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
		iterator, _ := NewIterator(rootHash, tr.GetStorageManager(), marshaller, hasher)

		trieData, err := iterator.GetLeaves(17, context.Background())
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
		iterator, _ := NewIterator(rootHash, tr.GetStorageManager(), marshaller, hasher)

		numRetrievedLeaves := 0
		numIterations := 0
		for numRetrievedLeaves < numLeaves {
			trieData, err := iterator.GetLeaves(5, context.Background())
			assert.Nil(t, err)

			numRetrievedLeaves += len(trieData)
			numIterations++
		}

		assert.Equal(t, numLeaves, numRetrievedLeaves)
		assert.Equal(t, 5, numIterations)
	})
}

func TestDfsIterator_GetIteratorId(t *testing.T) {
	t.Parallel()

	tr := trieTest.GetNewTrie()
	numLeaves := 25
	trieTest.AddDataToTrie(tr, numLeaves)
	rootHash, _ := tr.RootHash()
	_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
	iterator, _ := NewIterator(rootHash, tr.GetStorageManager(), marshaller, hasher)

	numRetrievedLeaves := 0
	for numRetrievedLeaves < numLeaves {
		iteratorId := hasher.Compute(string(append(rootHash, iterator.nextNodes[0].GetData()...)))
		assert.Equal(t, iteratorId, iterator.GetIteratorId())

		trieData, err := iterator.GetLeaves(5, context.Background())
		assert.Nil(t, err)

		numRetrievedLeaves += len(trieData)
	}

	assert.Equal(t, numLeaves, numRetrievedLeaves)
	assert.Nil(t, iterator.GetIteratorId())
}

func TestDfsIterator_Clone(t *testing.T) {
	t.Parallel()

	tr := trieTest.GetNewTrie()
	numLeaves := 25
	trieTest.AddDataToTrie(tr, numLeaves)
	rootHash, _ := tr.RootHash()
	_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
	iterator, _ := NewIterator(rootHash, tr.GetStorageManager(), marshaller, hasher)

	clonedIterator := iterator.Clone()

	nextNodesMemAddr := fmt.Sprintf("%p", iterator.nextNodes)
	clonedNextNodesMemAddr := fmt.Sprintf("%p", clonedIterator.(*dfsIterator).nextNodes)
	assert.NotEqual(t, nextNodesMemAddr, clonedNextNodesMemAddr)
	assert.Equal(t, iterator.rootHash, clonedIterator.(*dfsIterator).rootHash)
	assert.Equal(t, iterator.size, clonedIterator.(*dfsIterator).size)
}

func TestDfsIterator_FinishedIteration(t *testing.T) {
	t.Parallel()

	tr := trieTest.GetNewTrie()
	numLeaves := 25
	trieTest.AddDataToTrie(tr, numLeaves)
	rootHash, _ := tr.RootHash()
	_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
	iterator, _ := NewIterator(rootHash, tr.GetStorageManager(), marshaller, hasher)

	numRetrievedLeaves := 0
	for numRetrievedLeaves < numLeaves {
		assert.False(t, iterator.FinishedIteration())
		trieData, err := iterator.GetLeaves(5, context.Background())
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

	// branch node size = 33
	// root hash size = 32
	// extension nodes size = 34
	// leaf nodes size = 35
	iterator, _ := NewIterator(rootHash, tr.GetStorageManager(), marshaller, hasher)
	assert.Equal(t, uint64(362), iterator.Size()) // 10 branch nodes + 1 root hash

	_, err := iterator.GetLeaves(5, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint64(331), iterator.Size()) // 8 branch nodes + 1 leaf node + 1 root hash

	_, err = iterator.GetLeaves(5, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint64(300), iterator.Size()) // 6 branch nodes + 2 leaf node + 1 root hash

	_, err = iterator.GetLeaves(5, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint64(197), iterator.Size()) // 5 branch nodes  + 1 root hash

	_, err = iterator.GetLeaves(5, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint64(133), iterator.Size()) // 2 branch nodes  + 1 leaf node + 1 root hash

	_, err = iterator.GetLeaves(5, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint64(32), iterator.Size()) // 1 root hash
}
