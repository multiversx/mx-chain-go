package machine

import (
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/cpu"
)

var durationSecond = time.Second

type cpuStatistics struct {
	cpuPercentUsage uint64
}

func (cs *cpuStatistics) getCpuStatistics() {
	cpuUsagePercent, err := cpu.Percent(durationSecond, false)
	if err != nil {
		cs.setZeroStatsAndWait()
		return
	}
	if len(cpuUsagePercent) == 0 {
		cs.setZeroStatsAndWait()
		return
	}

	atomic.StoreUint64(&cs.cpuPercentUsage, uint64(cpuUsagePercent[0]))
	time.Sleep(durationSecond)
}

func (cs *cpuStatistics) setZeroStatsAndWait() {
	atomic.StoreUint64(&cs.cpuPercentUsage, 0)
	time.Sleep(durationSecond)
}

// CpuPercentUsage will return the cpu percent usage. Concurrent safe.
func (cs *cpuStatistics) CpuPercentUsage() uint64 {
	return atomic.LoadUint64(&cs.cpuPercentUsage)
}
