package state

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStateComponents_ParseStateChangesTypesToCollect(t *testing.T) {
	t.Parallel()

	t.Run("should parse state changes: 1 read 1 write", func(t *testing.T) {
		t.Parallel()

		collectRead, collectWrite, err := parseStateChangesTypesToCollect([]string{"read", "write"})
		require.NoError(t, err)
		require.True(t, collectRead)
		require.True(t, collectWrite)
	})

	t.Run("should parse state changes: multiple types", func(t *testing.T) {
		t.Parallel()

		collectRead, collectWrite, err := parseStateChangesTypesToCollect([]string{"read", "read", "write", "write"})
		require.NoError(t, err)
		require.True(t, collectRead)
		require.True(t, collectWrite)
	})

	t.Run("should parse state changes: inconsistent strings", func(t *testing.T) {
		collectRead, collectWrite, err := parseStateChangesTypesToCollect([]string{"Read", "read", "Write", "write"})
		require.NoError(t, err)
		require.True(t, collectRead)
		require.True(t, collectWrite)

		collectRead, collectWrite, err = parseStateChangesTypesToCollect([]string{"Read"})
		require.NoError(t, err)
		require.True(t, collectRead)
		require.False(t, collectWrite)

		collectRead, collectWrite, err = parseStateChangesTypesToCollect([]string{"Read", "rEaD"})
		require.NoError(t, err)
		require.True(t, collectRead)
		require.False(t, collectWrite)

		collectRead, collectWrite, err = parseStateChangesTypesToCollect([]string{"Write"})
		require.NoError(t, err)
		require.False(t, collectRead)
		require.True(t, collectWrite)

		collectRead, collectWrite, err = parseStateChangesTypesToCollect([]string{"Write", "write", "wRiTe"})
		require.NoError(t, err)
		require.False(t, collectRead)
		require.True(t, collectWrite)
	})

	t.Run("should parse state changes: no types", func(t *testing.T) {
		t.Parallel()

		collectRead, collectWrite, err := parseStateChangesTypesToCollect([]string{})
		require.NoError(t, err)
		require.False(t, collectRead)
		require.False(t, collectWrite)

		collectRead, collectWrite, err = parseStateChangesTypesToCollect(nil)
		require.NoError(t, err)
		require.False(t, collectRead)
		require.False(t, collectWrite)
	})

	t.Run("should not parse state changes: invalid types", func(t *testing.T) {
		t.Parallel()

		collectRead, collectWrite, err := parseStateChangesTypesToCollect([]string{"r3ad", "writ3"})
		require.ErrorContains(t, err, "unknown action type")
		require.False(t, collectRead)
		require.False(t, collectWrite)
	})
}
