package holders

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func haveTimeTrue() bool {
	return true
}

func TestNewTxSelectionOptions(t *testing.T) {
	t.Parallel()

	options, err := NewTxSelectionOptions(10_000_000_000, 30_000, 10, nil)
	require.Equal(t, errNilHaveTimeForSelectionFunc, err)

	options, err = NewTxSelectionOptions(10_000_000_000, 30_000, 10, haveTimeTrue)
	require.NoError(t, err)

	assert.Equal(t, uint64(10_000_000_000), options.GetGasRequested())
	assert.Equal(t, 30_000, options.GetMaxNumTxs())
	assert.Equal(t, 10, options.GetLoopDurationCheckInterval())
	assert.True(t, options.HaveTimeForSelection())
}
