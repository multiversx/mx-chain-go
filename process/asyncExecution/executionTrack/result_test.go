package executionTrack

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrintCleanStatus(t *testing.T) {
	t.Parallel()

	require.Equal(t, "ok", CleanResultOK.String())
	require.Equal(t, "mismatch", CleanResultMismatch.String())
	require.Equal(t, "not found", CleanResultNotFound.String())
	require.Equal(t, "unknown", CleanResult(10).String())
}
