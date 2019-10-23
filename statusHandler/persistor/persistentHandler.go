package persistor

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
)

const saveInterval = 10 * time.Second
const StatusMetricsDbEntry = "status_metrics_db_entry"

var log = logger.DefaultLogger()

// PersistentStatusHandler is a status handler that will save metrics in storage
type PersistentStatusHandler struct {
	store             storage.Storer
	persistentMetrics *sync.Map
	marshalizer       marshal.Marshalizer
}

// NewPersistentStatusHandler will return an instance of the persistent status handler
func NewPersistentStatusHandler(
	marshalizer marshal.Marshalizer,
) (*PersistentStatusHandler, error) {
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}

	psh := new(PersistentStatusHandler)
	psh.store = storageUnit.NewNilStorer()
	psh.marshalizer = marshalizer
	psh.persistentMetrics = &sync.Map{}
	psh.initMap()
	return psh, nil
}

func (psh *PersistentStatusHandler) initMap() {
	initUint := uint64(0)

	psh.persistentMetrics.Store(core.MetricCountConsensus, initUint)
	psh.persistentMetrics.Store(core.MetricCountConsensusAcceptedBlocks, initUint)
	psh.persistentMetrics.Store(core.MetricCountAcceptedBlocks, initUint)
	psh.persistentMetrics.Store(core.MetricCountLeader, initUint)
	psh.persistentMetrics.Store(core.MetricNumProcessedTxs, initUint)
	psh.persistentMetrics.Store(core.MetricNumShardHeadersProcessed, initUint)
}

// StartStoreMetricsInStorage will start save status metrics in storage at a constant interval
func (psh *PersistentStatusHandler) StartStoreMetricsInStorage() {
	go func() {
		time.Sleep(saveInterval)
		for {
			select {
			case <-time.After(saveInterval):
				psh.saveMetricsInDb()
			}
		}
	}()
}

// LoadMetricsFromDb will load from storage metrics
func (psh *PersistentStatusHandler) LoadMetricsFromDb() (map[string]uint64, map[string]string) {
	statusMetricsDbBytes, err := psh.store.Get([]byte(StatusMetricsDbEntry))
	if err != nil {
		log.Info("cannot load persistent metrics from storage", err)
		return nil, nil
	}

	metricsMap := make(map[string]interface{})
	err = psh.marshalizer.Unmarshal(&metricsMap, statusMetricsDbBytes)
	if err != nil {
		log.Info("cannot unmarshal persistent metrics", err)
		return nil, nil
	}

	return prepareMetricMaps(metricsMap)
}

// SetStorage will set storage for persistent status handler
func (psh *PersistentStatusHandler) SetStorage(store storage.Storer) error {
	if store == nil || store.IsInterfaceNil() {
		return process.ErrNilStorage
	}

	psh.store = store
	return nil
}

func (psh *PersistentStatusHandler) saveMetricsInDb() {
	metricsMap := make(map[string]interface{})
	psh.persistentMetrics.Range(func(key, value interface{}) bool {
		keyString, ok := key.(string)
		if !ok {
			return false
		}
		metricsMap[keyString] = value
		return true
	})

	statusMetricsBytes, err := psh.marshalizer.Marshal(&metricsMap)
	if err != nil {
		log.Info("cannot marshal metrics map", err)
		return
	}

	err = psh.store.Put([]byte(StatusMetricsDbEntry), statusMetricsBytes)
	if err != nil {
		log.Info("cannot save metrics map in storage", err)
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
	if _, ok := psh.persistentMetrics.Load(key); !ok {
		return
	}

	psh.persistentMetrics.Store(key, value)
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
	keyValueI, ok := psh.persistentMetrics.Load(key)
	if !ok {
		return
	}

	keyValue, ok := keyValueI.(uint64)
	if !ok {
		return
	}

	keyValue++
	psh.persistentMetrics.Store(key, keyValue)
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
	if psh == nil {
		return true
	}
	return false
}
