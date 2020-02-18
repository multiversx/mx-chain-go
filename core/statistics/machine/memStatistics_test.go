package machine

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
)

func TestMemStatisticsUsage(t *testing.T) {
	t.Parallel()

	mem := &MemStatistics{}

	mem.ComputeStatistics()
	memUsagePercentValue := mem.MemPercentUsage()
	fmt.Printf("Memory usage: %d%%\n", memUsagePercentValue)
	totalMem := mem.TotalMemory()
	golangMem := mem.MemoryUsedByGolang()
	sysMem := mem.MemoryUsedBySystem()
	fmt.Printf("Total memory: %s\n", core.ConvertBytes(totalMem))
	fmt.Printf("Golang memory usage: %s\n", core.ConvertBytes(golangMem))
	fmt.Printf("System memory usage: %s\n", core.ConvertBytes(sysMem))

	assert.True(t, memUsagePercentValue <= 100)
	assert.True(t, totalMem > 0)
	assert.True(t, golangMem > 0)
	assert.True(t, sysMem > 0)
}

func TestMemStatisticsUsage_ResetShouldWork(t *testing.T) {
	t.Parallel()

	mem := &MemStatistics{}

	mem.ComputeStatistics()

	mem.setZeroStatsAndWait()

	memUsagePercentValue := mem.MemPercentUsage()
	totalMem := mem.TotalMemory()
	golangMem := mem.MemoryUsedByGolang()

	assert.Equal(t, uint64(0), memUsagePercentValue)
	assert.Equal(t, uint64(0), totalMem)
	assert.Equal(t, uint64(0), golangMem)
}
