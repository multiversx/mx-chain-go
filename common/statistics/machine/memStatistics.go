package machine

import (
	"fmt"
	"runtime"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/shirou/gopsutil/mem"
)

// MemStatistics holds memory statistics
type MemStatistics struct {
	Total        uint64
	PercentUsed  uint64
	UsedByGolang uint64
	UsedBySystem uint64
	HeapInUse    uint64
	StackInUse   uint64
}

func (stats *MemStatistics) String() string {
	return fmt.Sprintf("total:%s, percent:%d%%, go:%s, sys:%s, heap:%s, stack:%s",
		core.ConvertBytes(stats.Total),
		stats.PercentUsed,
		core.ConvertBytes(stats.UsedByGolang),
		core.ConvertBytes(stats.UsedBySystem),
		core.ConvertBytes(stats.HeapInUse),
		core.ConvertBytes(stats.StackInUse),
	)
}

// AcquireMemStatistics acquires memory statistics
func AcquireMemStatistics() MemStatistics {
	var runtimeMemStats runtime.MemStats
	runtime.ReadMemStats(&runtimeMemStats)

	vms, err := mem.VirtualMemory()
	if err != nil {
		return MemStatistics{}
	}

	currentProcess, err := GetCurrentProcess()
	if err != nil {
		return MemStatistics{}
	}

	processMemoryInfo, err := currentProcess.MemoryInfo()
	if err != nil {
		return MemStatistics{}
	}

	percentUsed, err := currentProcess.MemoryPercent()
	if err != nil {
		return MemStatistics{}
	}

	result := MemStatistics{
		Total:        vms.Total,
		PercentUsed:  uint64(percentUsed),
		UsedByGolang: processMemoryInfo.RSS,
		UsedBySystem: runtimeMemStats.Sys,
		HeapInUse:    runtimeMemStats.HeapInuse,
		StackInUse:   runtimeMemStats.StackInuse,
	}

	log.Trace("AcquireMemStatistics", "stats", result.String())
	return result
}
