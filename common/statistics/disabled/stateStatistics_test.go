package disabled

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/require"
)

func TestNewDisabledStateStatistics(t *testing.T) {
	t.Parallel()

	stats := NewStateStatistics()
	require.False(t, check.IfNil(stats))
}

func TestStateStatistics_MethodsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, fmt.Sprintf("should have not panicked %v", r))
		}
	}()

	stats := NewStateStatistics()

	stats.Reset()
	stats.ResetSnapshot()
	stats.ResetAll()

	stats.IncrCache()
	stats.IncrSnapshotCache()
	stats.IncrSnapshotCache()
	stats.IncrPersister(1)
	stats.IncrSnapshotPersister(1)
	stats.IncrTrie()

	require.Equal(t, uint64(0), stats.Cache())
	require.Equal(t, uint64(0), stats.SnapshotCache())
	require.Equal(t, uint64(0), stats.Persister(1))
	require.Equal(t, uint64(0), stats.SnapshotPersister(1))
	require.Equal(t, uint64(0), stats.Trie())
}
