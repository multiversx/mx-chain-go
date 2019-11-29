package sync

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

func updateMetricsFromStorage(
	store dataRetriever.StorageService,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
	marshalizer marshal.Marshalizer,
	statusHandler core.AppStatusHandler,
	nonce uint64,
) (numTxs uint64, numHdrs uint64) {
	uint64Metrics, stringMetrics := loadMetricsFromDb(store, uint64ByteSliceConverter, marshalizer, nonce)

	saveUint64Metrics(statusHandler, uint64Metrics)
	saveStringMetrics(statusHandler, stringMetrics)

	return getTotalTxsAndHdrs(uint64Metrics)
}

func getTotalTxsAndHdrs(metrics map[string]uint64) (uint64, uint64) {
	numTxs, ok := metrics[core.MetricNumProcessedTxs]
	if !ok {
		numTxs = 0
	}

	numHdrs, ok := metrics[core.MetricNumShardHeadersProcessed]
	if !ok {
		numHdrs = 0
	}

	return numTxs, numHdrs
}

// LoadMetricsFromDb will load from storage metrics
func loadMetricsFromDb(store dataRetriever.StorageService, uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter, marshalizer marshal.Marshalizer, nonce uint64,
) (map[string]uint64, map[string]string) {
	nonceBytes := uint64ByteSliceConverter.ToByteSlice(nonce)
	storer := store.GetStorer(dataRetriever.StatusMetricsUnit)
	statusMetricsDbBytes, err := storer.Get(nonceBytes)
	if err != nil {
		log.Info("cannot load persistent metrics from storage", err)
		return nil, nil
	}

	metricsMap := make(map[string]interface{})
	err = marshalizer.Unmarshal(&metricsMap, statusMetricsDbBytes)
	if err != nil {
		log.Info("cannot unmarshal persistent metrics", err)
		return nil, nil
	}

	return prepareMetricMaps(metricsMap)
}

func saveUint64Metrics(statusHandler core.AppStatusHandler, metrics map[string]uint64) {
	if metrics == nil {
		return
	}

	for key, value := range metrics {
		statusHandler.SetUInt64Value(key, value)
	}
}

func saveStringMetrics(statusHandler core.AppStatusHandler, metrics map[string]string) {
	if metrics == nil {
		return
	}

	for key, value := range metrics {
		statusHandler.SetStringValue(key, value)
	}
}

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
