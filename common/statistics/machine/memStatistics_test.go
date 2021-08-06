package machine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStatisticsUsage(t *testing.T) {
	t.Parallel()

	mem := AcquireMemStatistics()

	assert.True(t, mem.PercentUsed <= 100)
	assert.True(t, mem.Total > 0)
	assert.True(t, mem.UsedByGolang > 0)
	assert.True(t, mem.UsedBySystem > 0)
	assert.True(t, mem.HeapInUse > 0)
	assert.True(t, mem.StackInUse > 0)
}
