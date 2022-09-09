package enablers

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFlagsHolder_NilFlagShouldPanic(t *testing.T) {
	t.Parallel()

	fh := newEpochFlagsHolder()
	require.NotNil(t, fh)

	fh.scDeployFlag = nil
	require.Panicsf(t, func() { fh.IsSCDeployFlagEnabled() }, "")
}

func TestFlagsHolder_ResetPenalizedTooMuchGasFlag(t *testing.T) {
	t.Parallel()

	fh := newEpochFlagsHolder()
	require.NotNil(t, fh)

	fh.penalizedTooMuchGasFlag.SetValue(true)
	require.True(t, fh.IsPenalizedTooMuchGasFlagEnabled())
	fh.ResetPenalizedTooMuchGasFlag()
	require.False(t, fh.IsPenalizedTooMuchGasFlagEnabled())
}
