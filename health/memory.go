package health

import (
	"runtime"
)

var _ memory = (*realMemory)(nil)

type realMemory struct {
}

func (m *realMemory) getStats() runtime.MemStats {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats
}
