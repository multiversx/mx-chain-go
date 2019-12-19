package txcache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO: Add more unit tests after moving this to core package (later in time)
func Test_NewConcurrentMap(t *testing.T) {
	m := NewConcurrentMap(4)
	require.Equal(t, uint32(4), m.nChunks)
	require.Equal(t, 4, len(m.chunks))

	// 1 is minimum number of chunks
	m = NewConcurrentMap(0)
	require.Equal(t, uint32(1), m.nChunks)
	require.Equal(t, 1, len(m.chunks))
}
