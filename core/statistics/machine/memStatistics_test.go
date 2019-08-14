package machine

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
)

func TestMemStatisticsUsagePercent(t *testing.T) {
	t.Parallel()

	mem := &MemStatistics{}

	mem.ComputeStatistics()
	memUsagePercentValue := mem.MemPercentUsage()
	fmt.Printf("Memory usage: %d%%\n", memUsagePercentValue)
	totalMem := mem.TotalMemory()
	fmt.Printf("Total memory: %s\n", core.ConvertBytes(totalMem))

	assert.True(t, memUsagePercentValue >= 0)
	assert.True(t, memUsagePercentValue <= 100)
	assert.True(t, totalMem > 0)
}
