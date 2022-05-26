package enableEpochs

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewFlagsHolder_NilFlagShouldPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		require.NotNil(t, r)
	}()

	fh := newFlagsHolder()
	require.NotNil(t, fh)

	fh.scDeployFlag = nil
	fh.IsSCDeployFlagEnabled()
}

func TestFlagsHolder_ResetPenalizedTooMuchGasFlag(t *testing.T) {
	t.Parallel()

	fh := newFlagsHolder()
	require.NotNil(t, fh)

	fh.penalizedTooMuchGasFlag.SetValue(true)
	require.True(t, fh.IsPenalizedTooMuchGasFlagEnabled())
	fh.ResetPenalizedTooMuchGasFlag()
	require.False(t, fh.IsPenalizedTooMuchGasFlagEnabled())
}
