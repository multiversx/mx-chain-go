package holders

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTxSelectionOptions(t *testing.T) {
	t.Parallel()

	options := NewTxSelectionOptions(10_000_000_000, 30_000, 250, 10)

	assert.Equal(t, uint64(10_000_000_000), options.GetGasRequested())
	assert.Equal(t, 30_000, options.GetMaxNumTxs())
	assert.Equal(t, 250, options.GetLoopMaximumDuration())
	assert.Equal(t, 10, options.GetLoopDurationCheckInterval())
}
