package persister

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/metrics"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

var log = logger.GetOrCreate("statusHandler/persister")

// PersistentStatusHandler is a status handler that will save metrics in storage
type PersistentStatusHandler struct {
	mutStore                 sync.RWMutex
	store                    storage.Storer
	persistentMetrics        *sync.Map
	marshalizer              marshal.Marshalizer
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
}

// NewPersistentStatusHandler will return an instance of the persistent status handler
func NewPersistentStatusHandler(
	marshalizer marshal.Marshalizer,
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter,
) (*PersistentStatusHandler, error) {
	if check.IfNil(marshalizer) {
		return nil, statusHandler.ErrNilMarshalizer
	}
	if check.IfNil(uint64ByteSliceConverter) {
		return nil, statusHandler.ErrNilUint64Converter
	}

	psh := new(PersistentStatusHandler)
	psh.store = storageUnit.NewNilStorer()
	psh.uint64ByteSliceConverter = uint64ByteSliceConverter
	psh.marshalizer = marshalizer
	psh.persistentMetrics = &sync.Map{}
	psh.initMap()

	return psh, nil
}

func (psh *PersistentStatusHandler) initMap() {
	initUint := uint64(0)
	zeroString := "0"

	psh.persistentMetrics.Store(common.MetricCountConsensus, initUint)
	psh.persistentMetrics.Store(common.MetricCountConsensusAcceptedBlocks, initUint)
	psh.persistentMetrics.Store(common.MetricCountAcceptedBlocks, initUint)
	psh.persistentMetrics.Store(common.MetricCountLeader, initUint)
	psh.persistentMetrics.Store(common.MetricNumProcessedTxs, initUint)
	psh.persistentMetrics.Store(common.MetricNumShardHeadersProcessed, initUint)
	psh.persistentMetrics.Store(common.MetricNonce, initUint)
	psh.persistentMetrics.Store(common.MetricCurrentRound, initUint)
	psh.persistentMetrics.Store(common.MetricNonceAtEpochStart, initUint)
	psh.persistentMetrics.Store(common.MetricRoundAtEpochStart, initUint)
	psh.persistentMetrics.Store(common.MetricTotalSupply, zeroString)
	psh.persistentMetrics.Store(common.MetricTotalFees, zeroString)
	psh.persistentMetrics.Store(common.MetricDevRewardsInEpoch, zeroString)
	psh.persistentMetrics.Store(common.MetricInflation, zeroString)
	psh.persistentMetrics.Store(common.MetricEpochForEconomicsData, initUint)
}

// SetStorage will set storage for persistent status handler
func (psh *PersistentStatusHandler) SetStorage(store storage.Storer) error {
	if check.IfNil(store) {
		return statusHandler.ErrNilStorage
	}

	psh.mutStore.Lock()
	psh.store = store
	psh.mutStore.Unlock()

	return nil
}

func (psh *PersistentStatusHandler) saveMetricsInDb(nonce uint64) {
	metricsMap := make(map[string]interface{})
	psh.persistentMetrics.Range(func(key, value interface{}) bool {
		keyString, ok := key.(string)
		if !ok {
			return false
		}
		metricsMap[keyString] = value
		return true
	})

	statusMetricsBytes, err := psh.marshalizer.Marshal(metrics.ListFromMap(metricsMap))
	if err != nil {
		log.Debug("cannot marshal metrics map",
			"error", err)
		return
	}

	nonceBytes := psh.uint64ByteSliceConverter.ToByteSlice(nonce)

	psh.mutStore.RLock()
	err = psh.store.Put(nonceBytes, statusMetricsBytes)
	psh.mutStore.RUnlock()
	if err != nil {
		log.Debug("cannot save metrics map in storage",
			"error", err)
		return
	}
}

// SetInt64Value method - will update the value for a key
func (psh *PersistentStatusHandler) SetInt64Value(key string, value int64) {
	if _, ok := psh.persistentMetrics.Load(key); !ok {
		return
	}

	psh.persistentMetrics.Store(key, value)
}

// SetUInt64Value method - will update the value for a key
func (psh *PersistentStatusHandler) SetUInt64Value(key string, value uint64) {
	valueFromMapI, ok := psh.persistentMetrics.Load(key)
	if !ok {
		return
	}

	psh.persistentMetrics.Store(key, value)

	//metrics wil be saved in storage every time when a block is committed successfully
	if key != common.MetricNonce {
		return
	}

	valueFromMap := GetUint64(valueFromMapI)
	if value < valueFromMap {
		return
	}

	if value == 0 {
		// do not write in database when the metrics are initialized. as a side effect, metrics for genesis block won't be saved
		return
	}

	psh.saveMetricsInDb(value)
}

// SetStringValue method - will update the value of a key
func (psh *PersistentStatusHandler) SetStringValue(key string, value string) {
	if _, ok := psh.persistentMetrics.Load(key); !ok {
		return
	}

	psh.persistentMetrics.Store(key, value)
}

// Increment - will increment the value of a key
func (psh *PersistentStatusHandler) Increment(key string) {
	psh.AddUint64(key, 1)
}

// AddUint64 - will increase the value of a key with a value
func (psh *PersistentStatusHandler) AddUint64(key string, value uint64) {
	keyValueI, ok := psh.persistentMetrics.Load(key)
	if !ok {
		return
	}

	keyValue, ok := keyValueI.(uint64)
	if !ok {
		return
	}

	keyValue += value
	psh.persistentMetrics.Store(key, keyValue)
}

// Decrement - will decrement the value of a key
func (psh *PersistentStatusHandler) Decrement(key string) {
	keyValueI, ok := psh.persistentMetrics.Load(key)
	if !ok {
		return
	}

	keyValue, ok := keyValueI.(uint64)
	if !ok {
		return
	}
	if keyValue == 0 {
		return
	}

	keyValue--
	psh.persistentMetrics.Store(key, keyValue)
}

// Close method - won't do anything
func (psh *PersistentStatusHandler) Close() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (psh *PersistentStatusHandler) IsInterfaceNil() bool {
	return psh == nil
}
