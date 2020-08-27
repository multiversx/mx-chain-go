package sync

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/metrics"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/statusHandler/persister"
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

	metricsList := &metrics.MetricsList{}
	err = marshalizer.Unmarshal(metricsList, statusMetricsDbBytes)
	if err != nil {
		log.Info("cannot unmarshal persistent metrics", err)
		return nil, nil
	}

	return prepareMetricMaps(metrics.MapFromList(metricsList))
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

	uint64Map[core.MetricCountConsensus] = persister.GetUint64(metricsMap[core.MetricCountConsensus])
	uint64Map[core.MetricCountConsensusAcceptedBlocks] = persister.GetUint64(metricsMap[core.MetricCountConsensusAcceptedBlocks])
	uint64Map[core.MetricCountAcceptedBlocks] = persister.GetUint64(metricsMap[core.MetricCountAcceptedBlocks])
	uint64Map[core.MetricCountLeader] = persister.GetUint64(metricsMap[core.MetricCountLeader])
	uint64Map[core.MetricNumProcessedTxs] = persister.GetUint64(metricsMap[core.MetricNumProcessedTxs])
	uint64Map[core.MetricNumShardHeadersProcessed] = persister.GetUint64(metricsMap[core.MetricNumShardHeadersProcessed])
	uint64Map[core.MetricRoundAtEpochStart] = persister.GetUint64(metricsMap[core.MetricRoundAtEpochStart])
	uint64Map[core.MetricNonceAtEpochStart] = persister.GetUint64(metricsMap[core.MetricNonceAtEpochStart])

	return uint64Map, stringMap
}
