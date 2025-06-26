package trieMetricsCollector

type trieMetricsCollector struct {
	maxDepth        int
	sizeLoadedInMem uint64
}

// NewTrieMetricsCollector creates a new instance of trieMetricsCollector
func NewTrieMetricsCollector() *trieMetricsCollector {
	return &trieMetricsCollector{
		// initial maxDepth is set to -1 because the root is on level 0, and for every node that is traversed IncreaseDepth() is called.
		// When the rootNode is traversed, maxDepth will be increased, so the root will be on level 0
		maxDepth:        -1,
		sizeLoadedInMem: 0,
	}
}

func (tmc *trieMetricsCollector) IncreaseDepth() {
	tmc.maxDepth++
}

func (tmc *trieMetricsCollector) GetMaxDepth() uint32 {
	return uint32(tmc.maxDepth)
}
