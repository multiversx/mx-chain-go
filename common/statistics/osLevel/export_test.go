package osLevel

// SumFields -
func (memStats *MemStats) SumFields() uint64 {
	return memStats.VmPeak + memStats.VmSize + memStats.VmLck + memStats.VmPin + memStats.VmHWM +
		memStats.VmRSS + memStats.RssAnon + memStats.RssFile + memStats.RssShmem + memStats.VmData +
		memStats.VmStk + memStats.VmExe + memStats.VmLib + memStats.VmPTE + memStats.VmSwap
}
