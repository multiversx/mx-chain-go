package maps

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/stretchr/testify/require"
)

type dummyItem struct {
	score      atomic.Uint32
	key        string
	chunk      *MapChunk
	chunkMutex sync.RWMutex
	mutex      sync.Mutex
}

func newDummyItem(key string) *dummyItem {
	return &dummyItem{
		key: key,
	}
}

func newScoredDummyItem(key string, score uint32) *dummyItem {
	item := &dummyItem{
		key: key,
	}
	item.score.Set(score)
	return item
}

func (item *dummyItem) GetKey() string {
	return item.key
}

func (item *dummyItem) GetScoreChunk() *MapChunk {
	item.chunkMutex.RLock()
	defer item.chunkMutex.RUnlock()

	return item.chunk
}

func (item *dummyItem) SetScoreChunk(chunk *MapChunk) {
	item.chunkMutex.Lock()
	defer item.chunkMutex.Unlock()

	item.chunk = chunk
}

func (item *dummyItem) simulateMutationThatChangesScore(myMap *BucketSortedMap) {
	item.mutex.Lock()
	myMap.NotifyScoreChange(item, item.score.Get())
	item.mutex.Unlock()
}

func simulateMutationThatChangesScore(myMap *BucketSortedMap, key string) {
	item, ok := myMap.Get(key)
	if !ok {
		return
	}

	itemAsDummy := item.(*dummyItem)
	itemAsDummy.simulateMutationThatChangesScore(myMap)
}

func TestNewBucketSortedMap(t *testing.T) {
	myMap := NewBucketSortedMap(4, 100)
	require.Equal(t, uint32(4), myMap.nChunks)
	require.Equal(t, 4, len(myMap.chunks))
	require.Equal(t, uint32(100), myMap.nScoreChunks)
	require.Equal(t, 100, len(myMap.scoreChunks))

	// 1 is minimum number of chunks
	myMap = NewBucketSortedMap(0, 0)
	require.Equal(t, uint32(1), myMap.nChunks)
	require.Equal(t, uint32(1), myMap.nScoreChunks)
}

func TestBucketSortedMap_Count(t *testing.T) {
	myMap := NewBucketSortedMap(4, 100)
	myMap.Set(newScoredDummyItem("a", 0))
	myMap.Set(newScoredDummyItem("b", 1))
	myMap.Set(newScoredDummyItem("c", 2))
	myMap.Set(newScoredDummyItem("d", 3))

	simulateMutationThatChangesScore(myMap, "a")
	simulateMutationThatChangesScore(myMap, "b")
	simulateMutationThatChangesScore(myMap, "c")
	simulateMutationThatChangesScore(myMap, "d")

	require.Equal(t, uint32(4), myMap.Count())
	require.Equal(t, uint32(4), myMap.CountSorted())

	counts := myMap.ChunksCounts()
	require.Equal(t, uint32(1), counts[0])
	require.Equal(t, uint32(1), counts[1])
	require.Equal(t, uint32(1), counts[2])
	require.Equal(t, uint32(1), counts[3])

	counts = myMap.ScoreChunksCounts()
	require.Equal(t, uint32(1), counts[0])
	require.Equal(t, uint32(1), counts[1])
	require.Equal(t, uint32(1), counts[2])
	require.Equal(t, uint32(1), counts[3])
}

func TestBucketSortedMap_Keys(t *testing.T) {
	myMap := NewBucketSortedMap(4, 100)
	myMap.Set(newDummyItem("a"))
	myMap.Set(newDummyItem("b"))
	myMap.Set(newDummyItem("c"))

	simulateMutationThatChangesScore(myMap, "a")
	simulateMutationThatChangesScore(myMap, "b")
	simulateMutationThatChangesScore(myMap, "c")

	require.Equal(t, 3, len(myMap.Keys()))
	require.Equal(t, 3, len(myMap.KeysSorted()))
}

func TestBucketSortedMap_KeysSorted(t *testing.T) {
	myMap := NewBucketSortedMap(1, 4)

	myMap.Set(newScoredDummyItem("d", 3))
	myMap.Set(newScoredDummyItem("a", 0))
	myMap.Set(newScoredDummyItem("c", 2))
	myMap.Set(newScoredDummyItem("b", 1))
	myMap.Set(newScoredDummyItem("f", 5))
	myMap.Set(newScoredDummyItem("e", 4))

	simulateMutationThatChangesScore(myMap, "d")
	simulateMutationThatChangesScore(myMap, "e")
	simulateMutationThatChangesScore(myMap, "f")
	simulateMutationThatChangesScore(myMap, "a")
	simulateMutationThatChangesScore(myMap, "b")
	simulateMutationThatChangesScore(myMap, "c")

	keys := myMap.KeysSorted()
	require.Equal(t, "a", keys[0])
	require.Equal(t, "b", keys[1])
	require.Equal(t, "c", keys[2])

	counts := myMap.ScoreChunksCounts()
	require.Equal(t, uint32(1), counts[0])
	require.Equal(t, uint32(1), counts[1])
	require.Equal(t, uint32(1), counts[2])
	require.Equal(t, uint32(3), counts[3])
}

func TestBucketSortedMap_ItemMovesOnNotifyScoreChange(t *testing.T) {
	myMap := NewBucketSortedMap(4, 100)

	a := newScoredDummyItem("a", 1)
	b := newScoredDummyItem("b", 42)
	myMap.Set(a)
	myMap.Set(b)

	simulateMutationThatChangesScore(myMap, "a")
	simulateMutationThatChangesScore(myMap, "b")

	require.Equal(t, myMap.scoreChunks[1], a.GetScoreChunk())
	require.Equal(t, myMap.scoreChunks[42], b.GetScoreChunk())

	a.score.Set(2)
	b.score.Set(43)
	simulateMutationThatChangesScore(myMap, "a")
	simulateMutationThatChangesScore(myMap, "b")

	require.Equal(t, myMap.scoreChunks[2], a.GetScoreChunk())
	require.Equal(t, myMap.scoreChunks[43], b.GetScoreChunk())
}

func TestBucketSortedMap_Has(t *testing.T) {
	myMap := NewBucketSortedMap(4, 100)
	myMap.Set(newDummyItem("a"))
	myMap.Set(newDummyItem("b"))

	require.True(t, myMap.Has("a"))
	require.True(t, myMap.Has("b"))
	require.False(t, myMap.Has("c"))
}

func TestBucketSortedMap_Remove(t *testing.T) {
	myMap := NewBucketSortedMap(4, 100)
	myMap.Set(newDummyItem("a"))
	myMap.Set(newDummyItem("b"))

	_, ok := myMap.Remove("b")
	require.True(t, ok)
	_, ok = myMap.Remove("x")
	require.False(t, ok)

	require.True(t, myMap.Has("a"))
	require.False(t, myMap.Has("b"))
}

func TestBucketSortedMap_Clear(t *testing.T) {
	myMap := NewBucketSortedMap(4, 100)
	myMap.Set(newDummyItem("a"))
	myMap.Set(newDummyItem("b"))

	myMap.Clear()

	require.Equal(t, uint32(0), myMap.Count())
	require.Equal(t, uint32(0), myMap.CountSorted())
}

func TestBucketSortedMap_IterCb(t *testing.T) {
	myMap := NewBucketSortedMap(4, 100)

	myMap.Set(newScoredDummyItem("a", 15))
	myMap.Set(newScoredDummyItem("b", 101))
	myMap.Set(newScoredDummyItem("c", 3))
	simulateMutationThatChangesScore(myMap, "a")
	simulateMutationThatChangesScore(myMap, "b")
	simulateMutationThatChangesScore(myMap, "c")

	sorted := []string{"c", "a", "b"}

	i := 0
	myMap.IterCbSortedAscending(func(key string, value BucketSortedMapItem) {
		require.Equal(t, sorted[i], key)
		i++
	})

	require.Equal(t, 3, i)

	i = len(sorted) - 1
	myMap.IterCbSortedDescending(func(key string, value BucketSortedMapItem) {
		require.Equal(t, sorted[i], key)
		i--
	})

	require.Equal(t, 0, i+1)
}

func TestBucketSortedMap_GetSnapshotAscending(t *testing.T) {
	myMap := NewBucketSortedMap(4, 100)

	snapshot := myMap.GetSnapshotAscending()
	require.Equal(t, []BucketSortedMapItem{}, snapshot)

	a := newScoredDummyItem("a", 15)
	b := newScoredDummyItem("b", 101)
	c := newScoredDummyItem("c", 3)

	myMap.Set(a)
	myMap.Set(b)
	myMap.Set(c)

	simulateMutationThatChangesScore(myMap, "a")
	simulateMutationThatChangesScore(myMap, "b")
	simulateMutationThatChangesScore(myMap, "c")

	snapshot = myMap.GetSnapshotAscending()
	require.Equal(t, []BucketSortedMapItem{c, a, b}, snapshot)
}

func TestBucketSortedMap_GetSnapshotDescending(t *testing.T) {
	myMap := NewBucketSortedMap(4, 100)

	snapshot := myMap.GetSnapshotDescending()
	require.Equal(t, []BucketSortedMapItem{}, snapshot)

	a := newScoredDummyItem("a", 15)
	b := newScoredDummyItem("b", 101)
	c := newScoredDummyItem("c", 3)

	myMap.Set(a)
	myMap.Set(b)
	myMap.Set(c)

	simulateMutationThatChangesScore(myMap, "a")
	simulateMutationThatChangesScore(myMap, "b")
	simulateMutationThatChangesScore(myMap, "c")

	snapshot = myMap.GetSnapshotDescending()
	require.Equal(t, []BucketSortedMapItem{b, a, c}, snapshot)
}

func TestBucketSortedMap_AddManyItems(t *testing.T) {
	numGoroutines := 42
	numItemsPerGoroutine := 1000
	numScoreChunks := 100
	numItemsInScoreChunkPerGoroutine := numItemsPerGoroutine / numScoreChunks
	numItemsInScoreChunk := numItemsInScoreChunkPerGoroutine * numGoroutines

	myMap := NewBucketSortedMap(16, uint32(numScoreChunks))

	var waitGroup sync.WaitGroup
	waitGroup.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			for j := 0; j < numItemsPerGoroutine; j++ {
				key := fmt.Sprintf("%d_%d", i, j)
				item := newScoredDummyItem(key, uint32(j%numScoreChunks))
				myMap.Set(item)
				simulateMutationThatChangesScore(myMap, key)
			}

			waitGroup.Done()
		}(i)
	}

	waitGroup.Wait()

	require.Equal(t, uint32(numGoroutines*numItemsPerGoroutine), myMap.CountSorted())

	counts := myMap.ScoreChunksCounts()
	for i := 0; i < numScoreChunks; i++ {
		require.Equal(t, uint32(numItemsInScoreChunk), counts[i])
	}
}

func TestBucketSortedMap_ClearConcurrentWithRead(t *testing.T) {
	numChunks := uint32(4)
	numScoreChunks := uint32(4)
	myMap := NewBucketSortedMap(numChunks, numScoreChunks)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		for j := 0; j < 1000; j++ {
			myMap.Clear()
		}
	}()

	go func() {
		defer wg.Done()

		for j := 0; j < 1000; j++ {
			require.Equal(t, uint32(0), myMap.Count())
			require.Equal(t, uint32(0), myMap.CountSorted())
			require.Len(t, myMap.ChunksCounts(), int(numChunks))
			require.Len(t, myMap.ScoreChunksCounts(), int(numScoreChunks))
			require.Len(t, myMap.Keys(), 0)
			require.Len(t, myMap.KeysSorted(), 0)
			require.Equal(t, false, myMap.Has("foobar"))
			item, ok := myMap.Get("foobar")
			require.Nil(t, item)
			require.False(t, ok)
			require.Len(t, myMap.GetSnapshotAscending(), 0)
			myMap.IterCbSortedAscending(func(key string, item BucketSortedMapItem) {
			})
			myMap.IterCbSortedDescending(func(key string, item BucketSortedMapItem) {
			})
		}
	}()

	wg.Wait()
}

func TestBucketSortedMap_ClearConcurrentWithWrite(t *testing.T) {
	myMap := NewBucketSortedMap(4, 4)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for j := 0; j < 10000; j++ {
			myMap.Clear()
		}

		wg.Done()
	}()

	go func() {
		for j := 0; j < 10000; j++ {
			myMap.Set(newDummyItem("foobar"))
			_, _ = myMap.Remove("foobar")
			myMap.NotifyScoreChange(newDummyItem("foobar"), 42)
			simulateMutationThatChangesScore(myMap, "foobar")
		}

		wg.Done()
	}()

	wg.Wait()
}

func TestBucketSortedMap_NoForgottenItemsOnConcurrentScoreChanges(t *testing.T) {
	// This test helped us to find a memory leak occuring on concurrent score changes (concurrent movements across buckets)

	for i := 0; i < 1000; i++ {
		myMap := NewBucketSortedMap(16, 16)
		a := newScoredDummyItem("a", 0)
		myMap.Set(a)
		simulateMutationThatChangesScore(myMap, "a")

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			a.score.Set(1)
			simulateMutationThatChangesScore(myMap, "a")
			wg.Done()
		}()

		go func() {
			a.score.Set(2)
			simulateMutationThatChangesScore(myMap, "a")
			wg.Done()
		}()

		wg.Wait()

		require.Equal(t, uint32(1), myMap.CountSorted())
		require.Equal(t, uint32(1), myMap.Count())

		_, _ = myMap.Remove("a")

		require.Equal(t, uint32(0), myMap.CountSorted())
		require.Equal(t, uint32(0), myMap.Count())
	}
}
