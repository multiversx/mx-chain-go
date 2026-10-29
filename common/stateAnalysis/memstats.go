package stateAnalysis

import (
	"runtime"
)

func (c *collector) sampleMemory() {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	c.SetGauge(GMemHeapAlloc, int64(ms.HeapAlloc))
	c.SetGauge(GMemHeapInuse, int64(ms.HeapInuse))
	c.SetGauge(GMemHeapObjects, int64(ms.HeapObjects))
	c.SetGauge(GMemSys, int64(ms.Sys))
	c.SetGauge(GMemNumGC, int64(ms.NumGC))
	c.SetGauge(GMemGCPauseTotal, int64(ms.PauseTotalNs))
	c.SetGauge(GMemRSS, residentSetBytes())
}
