package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO: Add more unit tests after moving this to core package (later in time)
func Test_NewConcurrentMap(t *testing.T) {
	myMap := NewConcurrentMap(4)
	require.Equal(t, uint32(4), myMap.nChunks)
	require.Equal(t, 4, len(myMap.chunks))

	// 1 is minimum number of chunks
	myMap = NewConcurrentMap(0)
	require.Equal(t, uint32(1), myMap.nChunks)
	require.Equal(t, 1, len(myMap.chunks))
}

func Test_ConcurrentMapKeys(t *testing.T) {
	myMap := NewConcurrentMap(4)
	myMap.Set("1", 0)
	myMap.Set("2", 0)
	myMap.Set("3", 0)
	myMap.Set("4", 0)

	require.Equal(t, 4, len(myMap.Keys()))
}
