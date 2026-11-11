package dfsTrieIterator

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"math"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/state/parsers"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	trieTest "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/storageManager"
	"github.com/multiversx/mx-chain-go/trie"
	"github.com/stretchr/testify/assert"
)

var maxSize = uint64(math.MaxUint64)

// getLeaves collects the leaves of the iteration as hex encoded key-value pairs, counting and charging every new key
func getLeaves(it *dfsIterator, numLeaves int, maxSize uint64, ctx context.Context) (map[string]string, error) {
	leaves := make(map[string]string)
	err := it.IterateLeaves(numLeaves, maxSize, parsers.NewMainTrieLeafParser(), ctx, func(leaf core.KeyValueHolder) (bool, uint64, error) {
		hexKey := hex.EncodeToString(leaf.Key())
		hexData := hex.EncodeToString(leaf.Value())
		numLeavesBefore := len(leaves)
		leaves[hexKey] = hexData
		return len(leaves) > numLeavesBefore, uint64(len(hexKey) + len(hexData)), nil
	})
	if err != nil {
		return nil, err
	}

	return leaves, nil
}

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
	t.Run("path symbol that is not a nibble, terminator accepted as the last symbol only", func(t *testing.T) {
		t.Parallel()

		for _, path := range [][]byte{{1, 17}, {16, 1}, {1, 16, 2}} {
			state := append(bytes.Repeat([]byte{0}, 32), path...)
			iterator, err := NewIterator([][]byte{state}, testscommon.NewMemDbMock(), &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{})
			assert.Nil(t, iterator, "%v", path)
			assert.Equal(t, trie.ErrInvalidIteratorState, err, "%v", path)
		}
		state := append(bytes.Repeat([]byte{0}, 32), 1, 16)
		iterator, err := NewIterator([][]byte{state}, testscommon.NewMemDbMock(), &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{})
		assert.Nil(t, err)
		assert.NotNil(t, iterator)
	})
	t.Run("path longer than the maximum is invalid", func(t *testing.T) {
		t.Parallel()

		nodeHash := bytes.Repeat([]byte{1}, 32)
		tooLong := append(append([]byte{}, nodeHash...), make([]byte, maxPathLength+1)...)
		iterator, err := NewIterator([][]byte{nodeHash, tooLong}, testscommon.NewMemDbMock(), &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{})
		assert.Nil(t, iterator)
		assert.Equal(t, trie.ErrInvalidIteratorState, err)

		longest := append(append([]byte{}, nodeHash...), make([]byte, maxPathLength)...)
		iterator, err = NewIterator([][]byte{nodeHash, longest}, testscommon.NewMemDbMock(), &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{})
		assert.Nil(t, err)
		assert.Equal(t, [][]byte{nodeHash, longest}, iterator.GetIteratorState())
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

		assert.Equal(t, 2, len(iterator.nextNodes))
	})
}

func TestDfsIterator_IterateLeaves_Collecting(t *testing.T) {
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

		trieData, err := getLeaves(iterator, numLeaves, maxSize, ctx)
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

		trieData, err := getLeaves(iterator, numLeaves, maxSize, context.Background())
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

		trieData, err := getLeaves(iterator, 17, maxSize, context.Background())
		assert.Nil(t, err)
		assert.Equal(t, expectedNumRetrievedLeaves, len(trieData))
	})
	t.Run("num leaves 0 iterates until maxSize reached", func(t *testing.T) {
		t.Parallel()

		tr := trieTest.GetNewTrie()
		numLeaves := 25
		trieTest.AddDataToTrie(tr, numLeaves)
		rootHash, _ := tr.RootHash()

		_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
		iterator, _ := NewIterator([][]byte{rootHash}, tr.GetStorageManager(), marshaller, hasher)

		trieData, err := getLeaves(iterator, 0, 200, context.Background())
		assert.Nil(t, err)
		assert.Equal(t, 8, len(trieData))
		assert.Equal(t, 8, len(iterator.nextNodes))
	})
	t.Run("max size reached returns retrieved leaves and saves iterator context", func(t *testing.T) {
		t.Parallel()

		tr := trieTest.GetNewTrie()
		numLeaves := 25
		trieTest.AddDataToTrie(tr, numLeaves)
		rootHash, _ := tr.RootHash()

		_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
		iterator, _ := NewIterator([][]byte{rootHash}, tr.GetStorageManager(), marshaller, hasher)

		iteratorMaxSize := uint64(200)
		trieData, err := getLeaves(iterator, numLeaves, iteratorMaxSize, context.Background())
		assert.Nil(t, err)
		assert.Equal(t, 8, len(trieData))
		assert.Equal(t, 8, len(iterator.nextNodes))
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
			trieData, err := getLeaves(iterator, 5, maxSize, context.Background())
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
		expectedNumRetrievedLeaves := 0
		trieTest.AddDataToTrie(tr, numLeaves)
		rootHash, _ := tr.RootHash()

		_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
		iterator, _ := NewIterator([][]byte{rootHash}, tr.GetStorageManager(), marshaller, hasher)

		trieData, err := getLeaves(iterator, numLeaves, maxSize, nil)
		assert.Nil(t, err)
		assert.Equal(t, expectedNumRetrievedLeaves, len(trieData))
	})
}

func TestDfsIterator_IterateLeaves(t *testing.T) {
	t.Parallel()

	newIterator := func(numLeaves int) *dfsIterator {
		tr := trieTest.GetNewTrie()
		trieTest.AddDataToTrie(tr, numLeaves)
		rootHash, _ := tr.RootHash()
		_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
		iterator, _ := NewIterator([][]byte{rootHash}, tr.GetStorageManager(), marshaller, hasher)
		return iterator
	}

	t.Run("the states of a trie with a key that is the prefix of another are accepted when resubmitted", func(t *testing.T) {
		t.Parallel()

		// "a" is stored in the terminator child of the branch where "ba" goes on, so a state may end in a terminator
		tr := trieTest.GetNewTrie()
		keys := []string{"a", "ba", "ca", "da"}
		for _, key := range keys {
			assert.Nil(t, tr.Update([]byte(key), []byte("value of "+key)))
		}
		assert.Nil(t, tr.Commit())
		rootHash, _ := tr.RootHash()
		_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
		iterator, err := NewIterator([][]byte{rootHash}, tr.GetStorageManager(), marshaller, hasher)
		assert.Nil(t, err)

		seen := make(map[string]struct{})
		terminatorSeen := false
		for page := 0; page < 10; page++ {
			err = iterator.IterateLeaves(1, maxSize, parsers.NewMainTrieLeafParser(), context.Background(), func(leaf core.KeyValueHolder) (bool, uint64, error) {
				seen[string(leaf.Key())] = struct{}{}
				return true, 0, nil
			})
			assert.Nil(t, err)
			state := iterator.GetIteratorState()
			if len(state) == 0 {
				break
			}
			for _, entry := range state {
				terminatorSeen = terminatorSeen || bytes.HasSuffix(entry, []byte{terminator})
			}
			iterator, err = NewIterator(state, tr.GetStorageManager(), marshaller, hasher)
			assert.Nil(t, err, "page %d", page)
		}
		assert.Len(t, seen, len(keys))
		assert.True(t, terminatorSeen)
	})
	t.Run("a tampered path of the state is an invalid state", func(t *testing.T) {
		t.Parallel()

		iterator := newIterator(25)
		err := iterator.IterateLeaves(1, maxSize, parsers.NewMainTrieLeafParser(), context.Background(), func(_ core.KeyValueHolder) (bool, uint64, error) {
			return true, 0, nil
		})
		assert.Nil(t, err)
		state := iterator.GetIteratorState()
		assert.NotEmpty(t, state)
		state[0] = append(state[0], 1) // one more nibble on the path of the first node: its leaf keys do not fit any more

		_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
		tampered, err := NewIterator(state, iterator.db, marshaller, hasher)
		assert.Nil(t, err)
		err = tampered.IterateLeaves(0, maxSize, parsers.NewMainTrieLeafParser(), context.Background(), func(_ core.KeyValueHolder) (bool, uint64, error) {
			return true, 0, nil
		})
		assert.ErrorIs(t, err, trie.ErrInvalidIteratorState)
	})
	t.Run("counts only accepted leaves and saves iterator state", func(t *testing.T) {
		t.Parallel()

		iterator := newIterator(25)
		numVisited := 0
		numAccepted := 0
		err := iterator.IterateLeaves(5, maxSize, parsers.NewMainTrieLeafParser(), context.Background(), func(leaf core.KeyValueHolder) (bool, uint64, error) {
			numVisited++
			if numVisited%2 == 0 {
				return false, 0, nil
			}
			numAccepted++
			return true, 0, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, 5, numAccepted)
		assert.Equal(t, 9, numVisited)
		assert.False(t, iterator.FinishedIteration())
	})
	t.Run("size of accepted leaves stops the iteration", func(t *testing.T) {
		t.Parallel()

		iterator := newIterator(25)
		numVisited := 0
		err := iterator.IterateLeaves(0, 250, parsers.NewMainTrieLeafParser(), context.Background(), func(leaf core.KeyValueHolder) (bool, uint64, error) {
			numVisited++
			return true, 100, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, 3, numVisited)
		assert.False(t, iterator.FinishedIteration())
	})
	t.Run("size of rejected leaves is not charged", func(t *testing.T) {
		t.Parallel()

		iterator := newIterator(25)
		numVisited := 0
		err := iterator.IterateLeaves(0, 250, parsers.NewMainTrieLeafParser(), context.Background(), func(leaf core.KeyValueHolder) (bool, uint64, error) {
			numVisited++
			return false, 100, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, 25, numVisited)
		assert.True(t, iterator.FinishedIteration())
	})
	t.Run("visitor error aborts the iteration", func(t *testing.T) {
		t.Parallel()

		iterator := newIterator(25)
		expectedErr := errors.New("expected error")
		numVisited := 0
		err := iterator.IterateLeaves(0, maxSize, parsers.NewMainTrieLeafParser(), context.Background(), func(leaf core.KeyValueHolder) (bool, uint64, error) {
			numVisited++
			return false, 0, expectedErr
		})
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 1, numVisited)
	})
	t.Run("visits all leaves when no limit is reached", func(t *testing.T) {
		t.Parallel()

		iterator := newIterator(25)
		visited := make(map[string]string)
		err := iterator.IterateLeaves(0, maxSize, parsers.NewMainTrieLeafParser(), context.Background(), func(leaf core.KeyValueHolder) (bool, uint64, error) {
			visited[string(leaf.Key())] = string(leaf.Value())
			return true, 0, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, 25, len(visited))
		assert.Equal(t, "value7", visited["value7"])
		assert.True(t, iterator.FinishedIteration())
		assert.Nil(t, iterator.GetIteratorState())
	})
	t.Run("context done returns without error and the iteration can be resumed", func(t *testing.T) {
		t.Parallel()

		iterator := newIterator(25)
		ctx, cancel := context.WithCancel(context.Background())
		visited := make(map[string]int)
		err := iterator.IterateLeaves(0, maxSize, parsers.NewMainTrieLeafParser(), ctx, func(leaf core.KeyValueHolder) (bool, uint64, error) {
			visited[string(leaf.Key())]++
			if len(visited) == 4 {
				cancel()
			}
			return true, 0, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, 4, len(visited))
		assert.False(t, iterator.FinishedIteration())

		err = iterator.IterateLeaves(0, maxSize, parsers.NewMainTrieLeafParser(), context.Background(), func(leaf core.KeyValueHolder) (bool, uint64, error) {
			visited[string(leaf.Key())]++
			return true, 0, nil
		})
		assert.Nil(t, err)
		assert.Equal(t, 25, len(visited))
		for key, count := range visited {
			assert.Equal(t, 1, count, key)
		}
		assert.True(t, iterator.FinishedIteration())
	})
	t.Run("leaf visited when the visitor fails is visited again on retry", func(t *testing.T) {
		t.Parallel()

		iterator := newIterator(25)
		expectedErr := errors.New("expected error")
		failedKey := ""
		numVisited := 0
		err := iterator.IterateLeaves(0, maxSize, parsers.NewMainTrieLeafParser(), context.Background(), func(leaf core.KeyValueHolder) (bool, uint64, error) {
			numVisited++
			if numVisited == 5 {
				failedKey = string(leaf.Key())
				return false, 0, expectedErr
			}
			return true, 0, nil
		})
		assert.Equal(t, expectedErr, err)

		visited := make(map[string]struct{})
		err = iterator.IterateLeaves(0, maxSize, parsers.NewMainTrieLeafParser(), context.Background(), func(leaf core.KeyValueHolder) (bool, uint64, error) {
			visited[string(leaf.Key())] = struct{}{}
			return true, 0, nil
		})
		assert.Nil(t, err)
		assert.Contains(t, visited, failedKey)
		assert.Equal(t, 21, len(visited))
	})
	t.Run("duplicate keys are not counted twice when the visitor rejects them", func(t *testing.T) {
		t.Parallel()

		tr := trieTest.GetNewTrie()
		trieTest.AddDataToTrie(tr, 1)
		rootHash, _ := tr.RootHash()
		_, marshaller, hasher := trieTest.GetDefaultTrieParameters()
		iterator, _ := NewIterator([][]byte{rootHash, rootHash}, tr.GetStorageManager(), marshaller, hasher)

		trieData, err := getLeaves(iterator, 2, maxSize, context.Background())
		assert.Nil(t, err)
		assert.Equal(t, 1, len(trieData))
		assert.True(t, iterator.FinishedIteration())
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

	leaves, err := getLeaves(iterator, 2, maxSize, context.Background())
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
		trieData, err := getLeaves(iterator, 5, maxSize, context.Background())
		assert.Nil(t, err)

		numRetrievedLeaves += len(trieData)
	}

	assert.Equal(t, numLeaves, numRetrievedLeaves)
	assert.True(t, iterator.FinishedIteration())
}
