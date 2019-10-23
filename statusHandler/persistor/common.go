package persistor

import "github.com/ElrondNetwork/elrond-go/core"

func prepareMetricMaps(metricsMap map[string]interface{}) (map[string]uint64, map[string]string) {
	uint64Map := make(map[string]uint64)
	stringMap := make(map[string]string)

	uint64Map[core.MetricCountConsensus] = getUint64(metricsMap[core.MetricCountConsensus])
	uint64Map[core.MetricCountConsensusAcceptedBlocks] = getUint64(metricsMap[core.MetricCountConsensusAcceptedBlocks])
	uint64Map[core.MetricCountAcceptedBlocks] = getUint64(metricsMap[core.MetricCountAcceptedBlocks])
	uint64Map[core.MetricCountLeader] = getUint64(metricsMap[core.MetricCountLeader])
	uint64Map[core.MetricNumProcessedTxs] = getUint64(metricsMap[core.MetricNumProcessedTxs])
	uint64Map[core.MetricNumShardHeadersProcessed] = getUint64(metricsMap[core.MetricNumShardHeadersProcessed])

	return uint64Map, stringMap
}

func getUint64(data interface{}) uint64 {
	value, ok := data.(float64)
	if !ok {
		return 0
	}

	return uint64(value)
}
