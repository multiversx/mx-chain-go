package machine

import (
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/mem"
)

type memStatistics struct {
	memPercentUsage uint64
	totalMemory     uint64
}

func (ms *memStatistics) getMemStatistics() {
	vms, err := mem.VirtualMemory()
	if err != nil {
		ms.setZeroStatsAndWait()
		return
	}

	atomic.StoreUint64(&ms.totalMemory, vms.Total)
	atomic.StoreUint64(&ms.memPercentUsage, uint64(vms.UsedPercent))
	time.Sleep(durationSecond)
}

func (ms *memStatistics) setZeroStatsAndWait() {
	atomic.StoreUint64(&ms.memPercentUsage, 0)
	atomic.StoreUint64(&ms.totalMemory, 0)
	time.Sleep(durationSecond)
}

// MemPercentUsage will return the memory percent usage. Concurrent safe.
func (ms *memStatistics) MemPercentUsage() uint64 {
	return atomic.LoadUint64(&ms.memPercentUsage)
}

// TotalMemory will return the total memory available in bytes. Concurrent safe.
func (ms *memStatistics) TotalMemory() uint64 {
	return atomic.LoadUint64(&ms.totalMemory)
}
