package trieMetricsCollector

type disabledTrieMetricsCollector struct{}

// NewDisabledTrieMetricsCollector returns a new instance of disabledTrieMetricsCollector
func NewDisabledTrieMetricsCollector() *disabledTrieMetricsCollector {
	return &disabledTrieMetricsCollector{}
}

// SetMaxDepth is a no-op for the disabled metrics collector
func (d *disabledTrieMetricsCollector) SetMaxDepth(_ uint32) {
}

// GetMaxDepth returns 0 for the disabled metrics collector
func (d *disabledTrieMetricsCollector) GetMaxDepth() uint32 {
	return 0
}

// AddSizeLoadedInMem is a no-op for the disabled metrics collector
func (d *disabledTrieMetricsCollector) AddSizeLoadedInMem(_ int) {
}

// GetSizeLoadedInMem returns 0 for the disabled metrics collector
func (d *disabledTrieMetricsCollector) GetSizeLoadedInMem() int {
	return 0
}
