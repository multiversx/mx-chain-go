package state

import (
	"errors"
	"strconv"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

// TODO test case:
// 1 - get trie, uncollapse multiple from trie, check total trie size matches maxTriesSize
// 2 - put with eviction
// 3 - eviction but oldest is dirty
// 4 - eviction but oldest is newest

const (
	dthSize = 2 * 1024 * 1024 // 2MB
	oneKB   = 1 * 1024
	oneMB   = oneKB * 1024
)

type testTries struct {
	key  []byte
	trie common.Trie
}

func getTestTries(numTries int) []testTries {
	tries := make([]testTries, 0)
	for i := 0; i < numTries; i++ {
		tr := &trieMock.TrieStub{
			SizeInMemoryCalled: func() int {
				return oneKB
			},
		}
		key := []byte("trie" + strconv.Itoa(i))
		tries = append(tries, testTries{key: key, trie: tr})
	}
	return tries
}

func TestNewDataTriesHolder(t *testing.T) {
	t.Parallel()

	t.Run(" invalid max size", func(t *testing.T) {
		t.Parallel()

		dth, err := NewDataTriesHolder(512 * 1024) // less than 1MB
		assert.True(t, errors.Is(err, ErrInvalidMaxTrieSizeValue))
		assert.True(t, check.IfNil(dth))
	})

	t.Run("should create new instance", func(t *testing.T) {
		t.Parallel()

		dth, err := NewDataTriesHolder(dthSize)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(dth))
		assert.Equal(t, dth.maxTriesSize, uint64(dthSize))
	})
}

func TestDataTriesHolder_Put(t *testing.T) {
	t.Parallel()

	t.Run("put in empty tries holder", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		entry := getTestTries(1)[0]

		dth.Put(entry.key, entry.trie)

		assert.Equal(t, 1, len(dth.tries))
		assert.Equal(t, 1, len(dth.dirtyTries))
		retrievedEntry, ok := dth.tries[string(entry.key)]
		assert.True(t, ok)
		assert.Equal(t, retrievedEntry.trie, entry.trie)
		assert.Equal(t, retrievedEntry.key, entry.key)
		assert.Nil(t, retrievedEntry.nextEntry)
		assert.Nil(t, retrievedEntry.prevEntry)

		assert.Equal(t, dth.oldestUsed, dth.newestUsed)
		assert.Equal(t, dth.oldestUsed, retrievedEntry)
		assert.Equal(t, uint64(oneKB), dth.totalTriesSize)
	})
	t.Run("put in populated tries holder", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		numEntries := 5
		entries := getTestTries(numEntries)
		for i := 0; i < numEntries; i++ {
			dth.Put(entries[i].key, entries[i].trie)
		}

		assert.Equal(t, numEntries, len(dth.tries))
		assert.Equal(t, numEntries, len(dth.dirtyTries))
		assert.Equal(t, entries[0].key, dth.oldestUsed.key)
		assert.Equal(t, entries[numEntries-1].key, dth.newestUsed.key)
		assert.Equal(t, uint64(numEntries*oneKB), dth.totalTriesSize)
		for i := 0; i < numEntries; i++ {
			retrievedEntry, ok := dth.tries[string(entries[i].key)]
			assert.True(t, ok)
			assert.Equal(t, retrievedEntry.trie, entries[i].trie)
			assert.Equal(t, retrievedEntry.key, entries[i].key)
			if i == 0 {
				assert.Nil(t, retrievedEntry.prevEntry)
			} else {
				assert.Equal(t, retrievedEntry.prevEntry.key, entries[i-1].key)
			}
			if i == numEntries-1 {
				assert.Nil(t, retrievedEntry.nextEntry)
			} else {
				assert.Equal(t, retrievedEntry.nextEntry.key, entries[i+1].key)
			}
		}
	})
	t.Run("put oldest used trie moves to newestUsed", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		numEntries := 5
		entries := getTestTries(numEntries)
		for i := 0; i < numEntries; i++ {
			dth.Put(entries[i].key, entries[i].trie)
		}

		dth.Put(entries[0].key, entries[0].trie)

		assert.Equal(t, numEntries, len(dth.tries))
		assert.Equal(t, 5, len(dth.dirtyTries))
		assert.Equal(t, uint64(numEntries*oneKB), dth.totalTriesSize)

		assert.Equal(t, entries[1].key, dth.oldestUsed.key)
		assert.Nil(t, dth.oldestUsed.prevEntry)
		assert.Equal(t, entries[2].key, dth.oldestUsed.nextEntry.key)
		assert.Equal(t, entries[0].key, dth.newestUsed.key)
		assert.Nil(t, dth.newestUsed.nextEntry)
		assert.Equal(t, entries[numEntries-1].key, dth.newestUsed.prevEntry.key)
	})
	t.Run("put existing trie moves to newestUsed", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		numEntries := 5
		entries := getTestTries(numEntries)
		for i := 0; i < numEntries; i++ {
			dth.Put(entries[i].key, entries[i].trie)
		}

		triePos := 2
		dth.Put(entries[triePos].key, entries[triePos].trie)

		assert.Equal(t, numEntries, len(dth.tries))
		assert.Equal(t, numEntries, len(dth.dirtyTries))
		assert.Equal(t, uint64(numEntries*oneKB), dth.totalTriesSize)
		assert.Equal(t, entries[0].key, dth.oldestUsed.key)
		assert.Equal(t, entries[triePos].key, dth.newestUsed.key)
		assert.Nil(t, dth.newestUsed.nextEntry)
		assert.Equal(t, entries[numEntries-1].key, dth.newestUsed.prevEntry.key)
		e1 := dth.tries[string(entries[triePos-1].key)]
		e2 := dth.tries[string(entries[triePos+1].key)]
		assert.Equal(t, e1.nextEntry.key, e2.key)
		assert.Equal(t, e2.prevEntry.key, e1.key)
	})
	t.Run("put newest used marks dirty", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		numEntries := 5
		entries := getTestTries(numEntries)
		for i := 0; i < numEntries; i++ {
			dth.Put(entries[i].key, entries[i].trie)
		}
		dth.dirtyTries = make(map[string]struct{}) // reset dirty tries

		lastEntryKey := entries[numEntries-1].key
		dth.Put(lastEntryKey, entries[numEntries-1].trie)
		assert.Equal(t, 1, len(dth.dirtyTries))
		_, exists := dth.dirtyTries[string(lastEntryKey)]
		assert.True(t, exists)
		assert.Equal(t, numEntries, len(dth.tries))
		assert.Equal(t, entries[0].key, dth.oldestUsed.key)
		assert.Equal(t, entries[numEntries-1].key, dth.newestUsed.key)
		assert.Equal(t, uint64(numEntries*oneKB), dth.totalTriesSize)
		for i := 0; i < numEntries; i++ {
			retrievedEntry, ok := dth.tries[string(entries[i].key)]
			assert.True(t, ok)
			assert.Equal(t, retrievedEntry.trie, entries[i].trie)
			assert.Equal(t, retrievedEntry.key, entries[i].key)
			if i == 0 {
				assert.Nil(t, retrievedEntry.prevEntry)
			} else {
				assert.Equal(t, retrievedEntry.prevEntry.key, entries[i-1].key)
			}
			if i == numEntries-1 {
				assert.Nil(t, retrievedEntry.nextEntry)
			} else {
				assert.Equal(t, retrievedEntry.nextEntry.key, entries[i+1].key)
			}
		}
	})
	t.Run("put with eviction - dirty tries should not evict", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		numEntries := 5
		entries := getTestTries(numEntries)
		for i := 0; i < numEntries; i++ {
			dth.Put(entries[i].key, entries[i].trie)
		}

		tr := &trieMock.TrieStub{
			SizeInMemoryCalled: func() int {
				return dthSize
			},
		}
		key := []byte("trieEvict")

		dth.Put(key, tr)

		assert.Equal(t, numEntries+1, len(dth.tries))
		assert.Equal(t, entries[0].key, dth.oldestUsed.key)
		assert.Equal(t, key, dth.newestUsed.key)
		assert.Nil(t, dth.newestUsed.nextEntry)
		assert.Equal(t, entries[numEntries-1].key, dth.newestUsed.prevEntry.key)
		assert.Equal(t, uint64(dthSize+5*oneKB), dth.totalTriesSize)
	})
	t.Run("put with eviction - not dirty tries should evict", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		numEntries := 5
		entries := getTestTries(numEntries)
		for i := 0; i < numEntries; i++ {
			dth.Put(entries[i].key, entries[i].trie)
		}
		dth.dirtyTries = make(map[string]struct{}) // reset dirty tries

		sizeToEvictTwoTries := dthSize - (3 * oneKB)
		tr := &trieMock.TrieStub{
			SizeInMemoryCalled: func() int {
				return sizeToEvictTwoTries
			},
		}
		key := []byte("trieEvict")

		dth.Put(key, tr)
		numEvictedTries := 2

		assert.Equal(t, numEntries-numEvictedTries+1, len(dth.tries))
		assert.Equal(t, entries[2].key, dth.oldestUsed.key)
		assert.Nil(t, dth.oldestUsed.prevEntry)
		assert.Equal(t, key, dth.newestUsed.key)
		assert.Nil(t, dth.newestUsed.nextEntry)
		assert.Equal(t, entries[numEntries-1].key, dth.newestUsed.prevEntry.key)
		assert.Equal(t, dth.newestUsed.prevEntry.nextEntry, dth.newestUsed)
		assert.Equal(t, uint64(dthSize), dth.totalTriesSize)
	})
}

func TestDataTriesHolder_Get(t *testing.T) {
	t.Parallel()

	t.Run("get not existing trie should return nil", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		tr := dth.Get([]byte("notExistingKey"))
		assert.Nil(t, tr)
		assert.Equal(t, 0, len(dth.tries))
	})
	t.Run("get existing trie should move to newestUsed", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		numEntries := 5
		entries := getTestTries(numEntries)
		for i := 0; i < numEntries; i++ {
			dth.Put(entries[i].key, entries[i].trie)
		}

		tr := dth.Get(entries[1].key)
		assert.Equal(t, entries[1].trie, tr)
		assert.Equal(t, numEntries, len(dth.tries))
		assert.Equal(t, entries[0].key, dth.oldestUsed.key)
		assert.Equal(t, entries[1].key, dth.newestUsed.key)
		assert.Nil(t, dth.newestUsed.nextEntry)
		assert.Equal(t, entries[numEntries-1].key, dth.newestUsed.prevEntry.key)
		assert.Equal(t, uint64(numEntries*oneKB), dth.totalTriesSize)
		assert.Equal(t, dth.oldestUsed.nextEntry.key, entries[2].key)
		assert.Equal(t, dth.oldestUsed.nextEntry.prevEntry, dth.oldestUsed)
	})
}

func TestDataTriesHolder_GetAllDirtyAndResetFlag(t *testing.T) {
	t.Parallel()

	t.Run("dirty trie not found in tries map does not panic", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		numEntries := 5
		entries := getTestTries(numEntries)
		for i := 0; i < numEntries; i++ {
			dth.Put(entries[i].key, entries[i].trie)
		}
		delete(dth.tries, string(entries[0].key))

		dirtyTries := dth.GetAllDirtyAndResetFlag()
		assert.Equal(t, numEntries-1, len(dirtyTries))
		assert.Equal(t, 0, len(dth.dirtyTries))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		numEntries := 5
		entries := getTestTries(numEntries)
		for i := 0; i < numEntries; i++ {
			dth.Put(entries[i].key, entries[i].trie)
		}

		dirtyTries := dth.GetAllDirtyAndResetFlag()
		assert.Equal(t, numEntries, len(dirtyTries))
		assert.Equal(t, 0, len(dth.dirtyTries))
	})
}

func TestDataTriesHolder_Reset(t *testing.T) {
	t.Parallel()

	dth, _ := NewDataTriesHolder(dthSize)
	numEntries := 5
	entries := getTestTries(numEntries)
	for i := 0; i < numEntries; i++ {
		dth.Put(entries[i].key, entries[i].trie)
	}

	dth.Reset()
	assert.Equal(t, 0, len(dth.tries))
	assert.Equal(t, 0, len(dth.dirtyTries))
	assert.Nil(t, dth.oldestUsed)
	assert.Nil(t, dth.newestUsed)
	assert.Equal(t, uint64(0), dth.totalTriesSize)
}

func TestDataTriesHolder_Concurrency(t *testing.T) {
	t.Parallel()

	dth, _ := NewDataTriesHolder(dthSize)
	numEntries := 5
	entries := getTestTries(numEntries)

	wg := sync.WaitGroup{}
	wg.Add(numEntries)

	for i := 0; i < numEntries; i++ {
		go func(key int) {
			dth.Put(entries[key].key, entries[key].trie)
			wg.Done()
		}(i)
	}

	wg.Wait()

	assert.Equal(t, numEntries, len(dth.tries))
	assert.Equal(t, numEntries, len(dth.dirtyTries))
	assert.Equal(t, uint64(numEntries*oneKB), dth.totalTriesSize)
	assert.NotNil(t, dth.oldestUsed)
	assert.NotNil(t, dth.newestUsed)
}
