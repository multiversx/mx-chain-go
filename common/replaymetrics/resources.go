package replaymetrics

import (
	"runtime"
	"time"

	"github.com/multiversx/mx-chain-go/common/statistics/machine"
	"github.com/multiversx/mx-chain-go/common/statistics/osLevel"
)

func (r *Recorder) sampleResources() {
	defer close(r.done)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	proc, _ := machine.GetCurrentProcess()
	ending := false
	for {
		p := r.current.Load()
		if p == nil {
			p = r.last.Load()
		}
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		record := map[string]any{
			"type": "resource", "time": time.Now().UTC(), "heap_alloc_bytes": mem.HeapAlloc,
			"heap_inuse_bytes": mem.HeapInuse, "go_allocated_total_bytes": mem.TotalAlloc,
			"gc_count": mem.NumGC, "gc_pause_total_ns": mem.PauseTotalNs,
		}
		var rss uint64
		if memOS, err := osLevel.ReadCurrentMemStats(); err == nil {
			rss = memOS.VmRSS
			record["rss_bytes"] = rss
			record["rss_high_water_bytes"] = memOS.VmHWM
		}
		if proc != nil {
			if cpu, err := proc.Times(); err == nil {
				record["cpu_total_seconds"] = cpu.User + cpu.System
			}
			if io, err := proc.IOCounters(); err == nil {
				record["read_total_bytes"] = io.ReadBytes
				record["write_total_bytes"] = io.WriteBytes
			}
		}
		r.mu.Lock()
		if p != nil {
			record["progress"] = p.progress
			s := r.epoch(p.progress)
			s.peakRSS = max(s.peakRSS, rss)
		}
		r.write(record)
		r.flush()
		r.mu.Unlock()
		if ending {
			return
		}
		select {
		case <-ticker.C:
		case <-r.stop:
			ending = true
		}
	}
}
