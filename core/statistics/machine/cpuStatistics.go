package machine

import (
	"sync/atomic"
	"time"
)

var durationSecond = time.Second

// CpuStatistics can compute the cpu usage percent
type CpuStatistics struct {
	cpuPercentUsage uint64
}

// ComputeStatistics computes the current cpu usage. It should be called on a go routine as it is a blocking
// call for a bounded time (1 second)
func (cs *CpuStatistics) ComputeStatistics() {
	cpuUsagePercent := CalculateCpuLoad()

	atomic.StoreUint64(&cs.cpuPercentUsage, uint64(cpuUsagePercent))
	time.Sleep(durationSecond)
}

func (cs *CpuStatistics) setZeroStatsAndWait() {
	atomic.StoreUint64(&cs.cpuPercentUsage, 0)
	time.Sleep(durationSecond)
}

// CpuPercentUsage will return the cpu percent usage. Concurrent safe.
func (cs *CpuStatistics) CpuPercentUsage() uint64 {
	return atomic.LoadUint64(&cs.cpuPercentUsage)
}
