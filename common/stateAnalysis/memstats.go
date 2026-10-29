package stateAnalysis

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func processCPUTime() time.Duration {
	var ru syscall.Rusage
	err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru)
	if err != nil {
		return 0
	}
	return time.Duration(ru.Utime.Nano() + ru.Stime.Nano())
}

func residentSetBytes() int64 {
	data, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(data))
	if len(fields) < 2 {
		return 0
	}
	pages, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return 0
	}
	return pages * int64(os.Getpagesize())
}

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
