package maps

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConcurrentMap(t *testing.T) {
	myMap := NewConcurrentMap(4)
	require.Equal(t, uint32(4), myMap.nChunks)
	require.Equal(t, 4, len(myMap.chunks))

	// 1 is minimum number of chunks
	myMap = NewConcurrentMap(0)
	require.Equal(t, uint32(1), myMap.nChunks)
	require.Equal(t, 1, len(myMap.chunks))
}

func TestConcurrentMap_Get(t *testing.T) {
	myMap := NewConcurrentMap(4)
	myMap.Set("a", "foo")
	myMap.Set("b", 42)

	a, ok := myMap.Get("a")
	require.True(t, ok)
	require.Equal(t, "foo", a)

	b, ok := myMap.Get("b")
	require.True(t, ok)
	require.Equal(t, 42, b)
}

func TestConcurrentMap_Count(t *testing.T) {
	myMap := NewConcurrentMap(4)
	myMap.Set("a", "a")
	myMap.Set("b", "b")
	myMap.Set("c", "c")

	require.Equal(t, 3, myMap.Count())
}

func TestConcurrentMap_Keys(t *testing.T) {
	myMap := NewConcurrentMap(4)
	myMap.Set("1", 0)
	myMap.Set("2", 0)
	myMap.Set("3", 0)
	myMap.Set("4", 0)

	require.Equal(t, 4, len(myMap.Keys()))
}

func TestConcurrentMap_Has(t *testing.T) {
	myMap := NewConcurrentMap(4)
	myMap.SetIfAbsent("a", "a")
	myMap.SetIfAbsent("b", "b")

	require.True(t, myMap.Has("a"))
	require.True(t, myMap.Has("b"))
	require.False(t, myMap.Has("c"))
}

func TestConcurrentMap_Remove(t *testing.T) {
	myMap := NewConcurrentMap(4)
	myMap.SetIfAbsent("a", "a")
	myMap.SetIfAbsent("b", "b")

	myMap.Remove("b")

	require.True(t, myMap.Has("a"))
	require.False(t, myMap.Has("b"))
}

func TestConcurrentMap_Clear(t *testing.T) {
	myMap := NewConcurrentMap(4)
	myMap.SetIfAbsent("a", "a")
	myMap.SetIfAbsent("b", "b")

	myMap.Clear()

	require.Equal(t, 0, myMap.Count())
}

func TestConcurrentMap_IterCb(t *testing.T) {
	myMap := NewConcurrentMap(4)

	myMap.Set("a", "a")
	myMap.Set("b", "b")
	myMap.Set("c", "c")

	i := 0
	myMap.IterCb(func(key string, value interface{}) {
		i++
	})

	require.Equal(t, 3, i)
}
