package holders

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTxSelectionOptionsAPI(t *testing.T) {
	t.Parallel()

	options := NewTxSelectionOptions(10_000_000_000, 30_000, 250, 10)
	optionsAPI := NewTxSelectionOptionsAPI(options, "hash,nonce")

	require.Equal(t, uint64(10_000_000_000), optionsAPI.GetGasRequested())
	require.Equal(t, 30_000, optionsAPI.GetMaxNumTxs())
	require.Equal(t, 250, optionsAPI.GetLoopMaximumDurationMs())
	require.Equal(t, 10, optionsAPI.GetLoopDurationCheckInterval())
	require.Equal(t, "hash,nonce", optionsAPI.GetRequestedFields())
}
