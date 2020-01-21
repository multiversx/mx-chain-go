package maps

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dummyItem struct {
	key   string
	score uint32
	chunk *MapChunk
}

func newDummyItem(key string) *dummyItem {
	return &dummyItem{
		key:   key,
		score: 0,
	}
}

func newScoredDummyItem(key string, score uint32) *dummyItem {
	return &dummyItem{
		key:   key,
		score: score,
	}
}

func (item *dummyItem) GetKey() string {
	return item.key
}

func (item *dummyItem) ComputeScore() uint32 {
	return item.score
}

func (item *dummyItem) GetScoreChunk() *MapChunk {
	return item.chunk
}

func (item *dummyItem) SetScoreChunk(chunk *MapChunk) {
	item.chunk = chunk
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
	myMap.Set(newDummyItem("a"))
	myMap.Set(newDummyItem("b"))
	myMap.Set(newDummyItem("c"))

	myMap.OnScoreChangeByKey("a")
	myMap.OnScoreChangeByKey("b")
	myMap.OnScoreChangeByKey("c")

	require.Equal(t, uint32(3), myMap.Count())
	require.Equal(t, uint32(3), myMap.CountSorted())
}

func TestBucketSortedMap_Keys(t *testing.T) {
	myMap := NewBucketSortedMap(4, 100)
	myMap.Set(newDummyItem("a"))
	myMap.Set(newDummyItem("b"))
	myMap.Set(newDummyItem("c"))

	myMap.OnScoreChangeByKey("a")
	myMap.OnScoreChangeByKey("b")
	myMap.OnScoreChangeByKey("c")

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

	myMap.OnScoreChangeByKey("d")
	myMap.OnScoreChangeByKey("e")
	myMap.OnScoreChangeByKey("f")
	myMap.OnScoreChangeByKey("a")
	myMap.OnScoreChangeByKey("b")
	myMap.OnScoreChangeByKey("c")

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

func TestBucketSortedMap_AddManyItems(t *testing.T) {
	myMap := NewBucketSortedMap(16, 100)

	var waitGroup sync.WaitGroup

	for i := 0; i < 100; i++ {
		waitGroup.Add(1)

		go func(i int) {
			for j := 0; j < 1000; j++ {
				key := fmt.Sprintf("%d_%d", i, j)
				item := newScoredDummyItem(key, uint32(j%100))
				myMap.Set(item)
				myMap.OnScoreChangeByKey(key)
			}

			waitGroup.Done()
		}(i)
	}

	waitGroup.Wait()

	assert.Equal(t, uint32(100*1000), myMap.CountSorted())

	counts := myMap.ScoreChunksCounts()
	for i := 0; i < 100; i++ {
		require.Equal(t, uint32(1000), counts[i])
	}
}
