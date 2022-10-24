package disabled

import "github.com/ElrondNetwork/elrond-go/trie/statistics"

type trieStatisticsCollector struct {
}

// NewTrieStatisticsCollector returns a new instance of the disabled trieStatisticsCollector
func NewTrieStatisticsCollector() *trieStatisticsCollector {
	return &trieStatisticsCollector{}
}

// Add does nothing as it is disabled
func (t trieStatisticsCollector) Add(_ *statistics.TrieStatsDTO) {
}

// UpdateMetricAndPrintStatistics does nothing as it is disabled
func (t trieStatisticsCollector) UpdateMetricAndPrintStatistics(_ string) {
}
