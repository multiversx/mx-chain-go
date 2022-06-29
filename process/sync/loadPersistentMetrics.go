package sync

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/metrics"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
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
	numTxs, ok := metrics[common.MetricNumProcessedTxs]
	if !ok {
		numTxs = 0
	}

	numHdrs, ok := metrics[common.MetricNumShardHeadersProcessed]
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
		log.Debug("cannot load persistent metrics from storage", "error", err)
		return nil, nil
	}

	metricsList := &metrics.MetricsList{}
	err = marshalizer.Unmarshal(metricsList, statusMetricsDbBytes)
	if err != nil {
		log.Debug("cannot unmarshal persistent metrics", "error", err)
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

	uint64Map[common.MetricCountConsensus] = persister.GetUint64(metricsMap[common.MetricCountConsensus])
	uint64Map[common.MetricCountConsensusAcceptedBlocks] = persister.GetUint64(metricsMap[common.MetricCountConsensusAcceptedBlocks])
	uint64Map[common.MetricCountAcceptedBlocks] = persister.GetUint64(metricsMap[common.MetricCountAcceptedBlocks])
	uint64Map[common.MetricCountLeader] = persister.GetUint64(metricsMap[common.MetricCountLeader])
	uint64Map[common.MetricNumProcessedTxs] = persister.GetUint64(metricsMap[common.MetricNumProcessedTxs])
	uint64Map[common.MetricNumShardHeadersProcessed] = persister.GetUint64(metricsMap[common.MetricNumShardHeadersProcessed])
	uint64Map[common.MetricEpochForEconomicsData] = persister.GetUint64(metricsMap[common.MetricEpochForEconomicsData])
	uint64Map[common.MetricNonceAtEpochStart] = persister.GetUint64(metricsMap[common.MetricNonceAtEpochStart])
	uint64Map[common.MetricRoundAtEpochStart] = persister.GetUint64(metricsMap[common.MetricRoundAtEpochStart])
	stringMap[common.MetricTotalSupply] = persister.GetString(metricsMap[common.MetricTotalSupply])
	stringMap[common.MetricTotalFees] = persister.GetString(metricsMap[common.MetricTotalFees])
	stringMap[common.MetricDevRewardsInEpoch] = persister.GetString(metricsMap[common.MetricDevRewardsInEpoch])
	stringMap[common.MetricInflation] = persister.GetString(metricsMap[common.MetricInflation])

	return uint64Map, stringMap
}
