package machine

import (
	"os"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/process"
)

var durationSecond = time.Second

// CpuStatistics can compute the cpu usage percent
type CpuStatistics struct {
	numCpu          int
	cpuPercentUsage uint64
	currentProcess  *process.Process
}

// NewCpuStatistics will create a CpuStatistics object
func NewCpuStatistics() *CpuStatistics {
	checkPid := os.Getpid()
	currentProcess, _ := process.NewProcess(int32(checkPid))

	numCpu, _ := cpu.Counts(true)

	return &CpuStatistics{
		cpuPercentUsage: 0,
		currentProcess:  currentProcess,
		numCpu:          numCpu,
	}
}

// ComputeStatistics computes the current cpu usage. It should be called on a go routine as it is a blocking
// call for a bounded time (1 second)
func (cs *CpuStatistics) ComputeStatistics() {
	cpuUsagePercent := cs.calculateCpuLoad()

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

func (cs *CpuStatistics) calculateCpuLoad() float64 {
	percent, err := cs.currentProcess.Percent(0)
	if err != nil {
		return 0
	}

	return percent / float64(cs.numCpu)
}
