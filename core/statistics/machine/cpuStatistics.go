package machine

import (
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/cpu"
)

var durationSecond = time.Second

// CpuStatistics can compute the cpu usage percent
type CpuStatistics struct {
	numCpu          int
	cpuPercentUsage uint64
}

// NewCpuStatistics will create a CpuStatistics object
func NewCpuStatistics() *CpuStatistics {
	numCpu, _ := cpu.Counts(true)

	return &CpuStatistics{
		cpuPercentUsage: 0,
		numCpu:          numCpu,
	}
}

// ComputeStatistics computes the current cpu usage. It should be called on a go routine as it is a blocking
// call for a bounded time (1 second)
func (cs *CpuStatistics) ComputeStatistics() {
	currentProcess, err := GetCurrentProcess()
	if err != nil {
		return
	}

	percent, err := currentProcess.Percent(time.Second)
	if err != nil {
		return
	}

	cpuUsagePercent := percent / float64(cs.numCpu)

	atomic.StoreUint64(&cs.cpuPercentUsage, uint64(cpuUsagePercent))
}

func (cs *CpuStatistics) setZeroStatsAndWait() {
	atomic.StoreUint64(&cs.cpuPercentUsage, 0)
	time.Sleep(durationSecond)
}

// CpuPercentUsage will return the cpu percent usage. Concurrent safe.
func (cs *CpuStatistics) CpuPercentUsage() uint64 {
	return atomic.LoadUint64(&cs.cpuPercentUsage)
}
