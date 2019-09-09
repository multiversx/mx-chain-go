package machine

import (
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/mem"
)

// MemStatistics can compute the mem usage percent and other mem statistics
type MemStatistics struct {
	memPercentUsage  uint64
	totalMemory      uint64
	usedMemoryByNode uint64
}

// ComputeStatistics computes the current memory usage.
func (ms *MemStatistics) ComputeStatistics() {
	vms, err := mem.VirtualMemory()
	if err != nil {
		ms.setZeroStatsAndWait()
		return
	}

	currentProcess, err := GetCurrentProcess()
	if err != nil {
		ms.setZeroStatsAndWait()
		return
	}

	processMemoryUsage, err := currentProcess.MemoryInfo()
	if err != nil {
		ms.setZeroStatsAndWait()
		return
	}

	ramUsagePercent, err := currentProcess.MemoryPercent()
	if err != nil {
		ms.setZeroStatsAndWait()
		return
	}

	atomic.StoreUint64(&ms.totalMemory, vms.Total)
	atomic.StoreUint64(&ms.memPercentUsage, uint64(ramUsagePercent))
	atomic.StoreUint64(&ms.usedMemoryByNode, processMemoryUsage.RSS)
	time.Sleep(durationSecond)
}

func (ms *MemStatistics) setZeroStatsAndWait() {
	atomic.StoreUint64(&ms.memPercentUsage, 0)
	atomic.StoreUint64(&ms.totalMemory, 0)
	atomic.StoreUint64(&ms.usedMemoryByNode, 0)
	time.Sleep(durationSecond)
}

// MemPercentUsage will return the memory percent usage. Concurrent safe.
func (ms *MemStatistics) MemPercentUsage() uint64 {
	return atomic.LoadUint64(&ms.memPercentUsage)
}

// TotalMemory will return the total memory available in bytes. Concurrent safe.
func (ms *MemStatistics) TotalMemory() uint64 {
	return atomic.LoadUint64(&ms.totalMemory)
}

// UsedMemory will return the total memory used by the node in bytes. Concurrent safe
func (ms *MemStatistics) UsedMemory() uint64 {
	return atomic.LoadUint64(&ms.usedMemoryByNode)
}
