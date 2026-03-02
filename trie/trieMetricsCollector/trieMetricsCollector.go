package trieMetricsCollector

type trieMetricsCollector struct {
	currentDepth    int
	maxDepth        int
	sizeLoadedInMem int
}

// NewTrieMetricsCollector creates a new instance of trieMetricsCollector
func NewTrieMetricsCollector() *trieMetricsCollector {
	return &trieMetricsCollector{
		maxDepth:        0,
		sizeLoadedInMem: 0,
	}
}

// SetDepth sets the maxDepth to the provided value if it is greater than the current maxDepth
func (tmc *trieMetricsCollector) SetDepth(depth uint32) {
	tmc.currentDepth = int(depth)
	if depth <= uint32(tmc.maxDepth) {
		return
	}

	tmc.maxDepth = int(depth)
}

// GetCurrentDepth returns the current depth stored in the collector
func (tmc *trieMetricsCollector) GetCurrentDepth() uint32 {
	return uint32(tmc.currentDepth)
}

// GetMaxDepth returns the collected maxDepth
func (tmc *trieMetricsCollector) GetMaxDepth() uint32 {
	return uint32(tmc.maxDepth)
}

// AddSizeLoadedInMem adds the size of the loaded data in memory to the collector
func (tmc *trieMetricsCollector) AddSizeLoadedInMem(size int) {
	tmc.sizeLoadedInMem += size
}

// GetSizeLoadedInMem returns the total size of data loaded in memory
func (tmc *trieMetricsCollector) GetSizeLoadedInMem() int {
	return tmc.sizeLoadedInMem
}
