package enablers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlagsHolder_ResetPenalizedTooMuchGasFlag(t *testing.T) {
	t.Parallel()

	fh := newEpochFlagsHolder()
	require.NotNil(t, fh)

	fh.penalizedTooMuchGasFlag.SetValue(true)
	require.True(t, fh.IsPenalizedTooMuchGasFlagEnabled())
	fh.ResetPenalizedTooMuchGasFlag()
	require.False(t, fh.IsPenalizedTooMuchGasFlagEnabled())
}
