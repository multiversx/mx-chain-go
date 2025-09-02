package trieMetricsCollector

type trieMetricsCollector struct {
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

// SetMaxDepth sets the maxDepth to the provided value if it is greater than the current maxDepth
func (tmc *trieMetricsCollector) SetMaxDepth(depth uint32) {
	if depth <= uint32(tmc.maxDepth) {
		return
	}

	tmc.maxDepth = int(depth)
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
