package triesHolder

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

const (
	dthSize = 2 * 1024 * 1024 // 2MB
	oneKB   = 1 * 1024
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
		assert.NotNil(t, dth.evictedBuffer)
		assert.NotNil(t, dth.dirtyTries)
		assert.NotNil(t, dth.touchedTries)
		assert.Equal(t, uint64(0), dth.cacher.SizeInBytesContained())
		assert.Equal(t, 0, dth.cacher.Len())
	})
}

func TestDataTriesHolder_Put(t *testing.T) {
	t.Parallel()

	t.Run("put in empty tries holder", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		entry := getTestTries(1)[0]

		dth.Put(entry.key, entry.trie)

		assert.Equal(t, 1, dth.cacher.Len())
		assert.Equal(t, 1, len(dth.dirtyTries))
		retrievedEntry, ok := dth.cacher.Get(string(entry.key))
		assert.True(t, ok)
		tr, ok := retrievedEntry.(common.Trie)
		assert.Equal(t, tr, entry.trie)
		assert.Equal(t, uint64(oneKB), dth.cacher.SizeInBytesContained())
		assert.Equal(t, 1, len(dth.dirtyTries))
		assert.Equal(t, 1, len(dth.touchedTries))
		assert.Equal(t, 0, len(dth.evictedBuffer))
	})
	t.Run("put in populated tries holder", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		numEntries := 5
		entries := getTestTries(numEntries)
		for i := 0; i < numEntries; i++ {
			dth.Put(entries[i].key, entries[i].trie)
		}

		assert.Equal(t, numEntries, dth.cacher.Len())
		assert.Equal(t, numEntries, len(dth.dirtyTries))
		assert.Equal(t, numEntries, len(dth.touchedTries))
		assert.Equal(t, 0, len(dth.evictedBuffer))
		assert.Equal(t, uint64(numEntries*oneKB), dth.cacher.SizeInBytesContained())
		cacherKeys := dth.cacher.Keys()
		assert.Equal(t, numEntries, len(cacherKeys))
		for i := 0; i < numEntries; i++ {
			retrievedEntry, ok := dth.cacher.Get(cacherKeys[i])
			assert.True(t, ok)
			tr, ok := retrievedEntry.(common.Trie)
			assert.Equal(t, tr, entries[i].trie)
			assert.Equal(t, cacherKeys[i], string(entries[i].key))
		}
	})
	t.Run("put oldest used trie moves to newest used", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		numEntries := 5
		entries := getTestTries(numEntries)
		for i := 0; i < numEntries; i++ {
			dth.Put(entries[i].key, entries[i].trie)
		}

		dth.Put(entries[0].key, entries[0].trie)

		assert.Equal(t, numEntries, dth.cacher.Len())
		assert.Equal(t, numEntries, len(dth.dirtyTries))
		assert.Equal(t, numEntries, len(dth.touchedTries))
		assert.Equal(t, 0, len(dth.evictedBuffer))
		assert.Equal(t, uint64(numEntries*oneKB), dth.cacher.SizeInBytesContained())
		keys := dth.cacher.Keys()
		assert.Equal(t, string(entries[0].key), keys[numEntries-1])
	})
	t.Run("put existing trie moves to newest used", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		numEntries := 5
		entries := getTestTries(numEntries)
		for i := 0; i < numEntries; i++ {
			dth.Put(entries[i].key, entries[i].trie)
		}

		triePos := 2
		dth.Put(entries[triePos].key, entries[triePos].trie)

		assert.Equal(t, numEntries, dth.cacher.Len())
		assert.Equal(t, numEntries, len(dth.dirtyTries))
		assert.Equal(t, numEntries, len(dth.touchedTries))
		assert.Equal(t, 0, len(dth.evictedBuffer))
		assert.Equal(t, uint64(numEntries*oneKB), dth.cacher.SizeInBytesContained())
		keys := dth.cacher.Keys()
		assert.Equal(t, string(entries[triePos].key), keys[numEntries-1])
	})
	t.Run("put with eviction - evicted dirty tries should be in eviction buffer", func(t *testing.T) {
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

		assert.Equal(t, 1, dth.cacher.Len())
		assert.Equal(t, numEntries+1, len(dth.dirtyTries))
		assert.Equal(t, numEntries+1, len(dth.touchedTries))
		assert.Equal(t, numEntries, len(dth.evictedBuffer))
		assert.Equal(t, uint64(dthSize), dth.cacher.SizeInBytesContained())
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

		assert.Equal(t, numEntries-numEvictedTries+1, dth.cacher.Len())
		assert.Equal(t, 1, len(dth.dirtyTries))
		assert.Equal(t, numEntries+1, len(dth.touchedTries))
		assert.Equal(t, 0, len(dth.evictedBuffer))
		assert.Equal(t, uint64(dthSize), dth.cacher.SizeInBytesContained())
	})
}

func TestDataTriesHolder_Get(t *testing.T) {
	t.Parallel()

	t.Run("get not existing trie should return nil", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		tr := dth.Get([]byte("notExistingKey"))
		assert.Nil(t, tr)
		assert.Equal(t, 0, dth.cacher.Len())
	})
	t.Run("get existing trie should move to newest used", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		numEntries := 5
		entries := getTestTries(numEntries)
		for i := 0; i < numEntries; i++ {
			dth.Put(entries[i].key, entries[i].trie)
		}

		tr := dth.Get(entries[1].key)
		assert.Equal(t, entries[1].trie, tr)
		assert.Equal(t, numEntries, dth.cacher.Len())
		keys := dth.cacher.Keys()
		assert.Equal(t, string(entries[1].key), keys[numEntries-1])
		assert.Equal(t, uint64(numEntries*oneKB), dth.cacher.SizeInBytesContained())
	})
	t.Run("get from evicted buffer should put back in cache", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		numEntries := 5
		entries := getTestTries(numEntries)
		for i := 0; i < numEntries; i++ {
			dth.Put(entries[i].key, entries[i].trie)
		}

		assert.Equal(t, 0, len(dth.evictedBuffer))
		tr := &trieMock.TrieStub{
			SizeInMemoryCalled: func() int {
				return dthSize
			},
		}
		key := []byte("trieEvict")
		dth.Put(key, tr)
		assert.Equal(t, numEntries, len(dth.evictedBuffer))
		assert.Equal(t, 1, dth.cacher.Len())

		numGetFromTrie := 3
		for i := 0; i < numGetFromTrie; i++ {
			_ = dth.Get(entries[i].key)
		}

		assert.Equal(t, 3, len(dth.evictedBuffer)) // 2 original tries + trieEvict
		assert.Equal(t, numGetFromTrie, dth.cacher.Len())
	})
}

func TestDataTriesHolder_GetAll(t *testing.T) {
	t.Parallel()

	t.Run("dirty trie not found in tries map does not panic", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		numEntries := 5
		entries := getTestTries(numEntries)
		for i := 0; i < numEntries; i++ {
			dth.Put(entries[i].key, entries[i].trie)
		}
		dth.cacher.Remove(string(entries[0].key))

		dirtyTries := dth.GetAll()
		assert.Equal(t, numEntries-1, len(dirtyTries))
		assert.Equal(t, 0, len(dth.dirtyTries))
	})
	t.Run("trie size is correctly computed after GetAll and eviction", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		numEntries := 5
		entries := getTestTries(numEntries)
		for i := 0; i < numEntries; i++ {
			dth.Put(entries[i].key, entries[i].trie)
		}

		assert.Equal(t, uint64(numEntries*oneKB), dth.cacher.SizeInBytesContained())

		dirtyTries := dth.GetAll()
		assert.Equal(t, numEntries, len(dirtyTries))
		assert.Equal(t, 0, len(dth.dirtyTries))
		assert.Equal(t, uint64(numEntries*oneKB), dth.cacher.SizeInBytesContained())

		trieSize := dthSize - ((numEntries + 1) * oneKB)
		tr := &trieMock.TrieStub{
			SizeInMemoryCalled: func() int {
				return trieSize
			},
		}
		key := []byte("newTrie")
		dth.Put(key, tr)
		assert.Equal(t, numEntries+1, dth.cacher.Len())
		assert.Equal(t, uint64(numEntries*oneKB+trieSize), dth.cacher.SizeInBytesContained())
		dth.dirtyTries = make(map[string]struct{}) // reset dirty tries

		// get a trie and "resolve" some nodes, thus increasing its size in memory
		_ = dth.Get(key)
		originalSize := trieSize
		trieSize = trieSize + oneKB
		assert.Equal(t, uint64(numEntries*oneKB+originalSize), dth.cacher.SizeInBytesContained()) // size is not updated
		assert.Equal(t, numEntries+1, dth.cacher.Len())                                           // no eviction
		assert.Equal(t, 1, len(dth.touchedTries))

		dirtyTries = dth.GetAll()
		assert.Equal(t, 0, len(dirtyTries))
		assert.Equal(t, 0, len(dth.dirtyTries))
		assert.Equal(t, uint64(dthSize), dth.cacher.SizeInBytesContained()) // size is updated
		assert.Equal(t, numEntries+1, dth.cacher.Len())                     // no eviction
		assert.Equal(t, 0, len(dth.touchedTries))

		// put again the same trie, now with increased size triggering eviction
		_ = dth.Get(key)
		assert.Equal(t, 1, len(dth.touchedTries))
		trieSize = trieSize + oneKB/2
		assert.Equal(t, 0, len(dth.dirtyTries))
		assert.Equal(t, uint64(dthSize), dth.cacher.SizeInBytesContained()) // size is  not updated
		assert.Equal(t, numEntries+1, dth.cacher.Len())                     // no eviction

		dirtyTries = dth.GetAll()
		assert.Equal(t, 0, len(dirtyTries))
		assert.Equal(t, 0, len(dth.dirtyTries))
		assert.Equal(t, uint64(dthSize-oneKB/2), dth.cacher.SizeInBytesContained()) // size is updated
		assert.Equal(t, numEntries, dth.cacher.Len())                               // eviction
		assert.Equal(t, 0, len(dth.touchedTries))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dth, _ := NewDataTriesHolder(dthSize)
		numEntries := 5
		entries := getTestTries(numEntries)
		for i := 0; i < numEntries; i++ {
			dth.Put(entries[i].key, entries[i].trie)
		}

		dirtyTries := dth.GetAll()
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
	assert.Equal(t, 0, dth.cacher.Len())
	assert.Equal(t, uint64(0), dth.cacher.SizeInBytesContained())
	assert.Equal(t, 0, len(dth.dirtyTries))
	assert.Equal(t, 0, len(dth.touchedTries))
	assert.Equal(t, 0, len(dth.evictedBuffer))
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

	assert.Equal(t, numEntries, dth.cacher.Len())
	assert.Equal(t, numEntries, len(dth.dirtyTries))
	assert.Equal(t, uint64(numEntries*oneKB), dth.cacher.SizeInBytesContained())
}

func BenchmarkDataTriesHolder_PutWithEviction(b *testing.B) {
	numEntries := 100000
	dth, _ := NewDataTriesHolder(uint64(numEntries * oneKB / 10)) // set max size to 10% of total size to force evictions
	entries := getTestTries(numEntries)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry := entries[i%numEntries]
		dth.Put(entry.key, entry.trie)
		if i%1000 == 0 {
			dth.dirtyTries = make(map[string]struct{}) // reset dirty tries to allow evictions
		}
	}
}

func BenchmarkDataTriesHolder_PutNoEviction(b *testing.B) {
	numEntries := 100000
	dth, _ := NewDataTriesHolder(uint64(numEntries * oneKB * 2)) // set max size to 200% of total size to avoid evictions
	entries := getTestTries(numEntries)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry := entries[i%numEntries]
		dth.Put(entry.key, entry.trie)
	}
}

func BenchmarkDataTriesHolder_Get(b *testing.B) {
	numEntries := 100000
	dth, _ := NewDataTriesHolder(uint64(numEntries * oneKB * 2))
	entries := getTestTries(numEntries)
	for i := 0; i < numEntries; i++ {
		dth.Put(entries[i].key, entries[i].trie)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry := entries[i%numEntries]
		dth.Get(entry.key)
	}
}

func BenchmarkDataTriesHolder_GetAll(b *testing.B) {
	numEntries := 10000
	dth, _ := NewDataTriesHolder(uint64(numEntries * oneKB * 2))
	entries := getTestTries(numEntries)
	for i := 0; i < numEntries; i++ {
		dth.Put(entries[i].key, entries[i].trie)
	}
	dth.dirtyTries = make(map[string]struct{})
	dth.touchedTries = make(map[string]struct{})
	numDirty := numEntries / 10
	for i := 0; i < numDirty; i++ {
		dth.dirtyTries[string(entries[i].key)] = struct{}{}
		dth.touchedTries[string(entries[i].key)] = struct{}{}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dth.GetAll()
	}
}
