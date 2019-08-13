package machine

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCpuStatisticsUsagePercent(t *testing.T) {
	t.Parallel()

	cs := &cpuStatistics{}

	cs.getCpuStatistics()
	cpuUsagePercentValue := cs.CpuPercentUsage()
	fmt.Printf("CPU usage: %d%%\n", cpuUsagePercentValue)

	assert.True(t, cpuUsagePercentValue >= 0)
	assert.True(t, cpuUsagePercentValue <= 100)
}
