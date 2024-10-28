package state

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStateComponents_ParseStateChangesTypesToCollect(t *testing.T) {
	t.Parallel()

	collectRead, collectWrite, err := parseStateChangesTypesToCollect([]string{"read", "write"})
	require.NoError(t, err)
	require.True(t, collectRead)
	require.True(t, collectWrite)

	collectRead, collectWrite, err = parseStateChangesTypesToCollect([]string{"read", "read", "write", "write"})
	require.NoError(t, err)
	require.True(t, collectRead)
	require.True(t, collectWrite)

	collectRead, collectWrite, err = parseStateChangesTypesToCollect([]string{"Read", "read", "Write", "write"})
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
}
