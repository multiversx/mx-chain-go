package persistor

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/statusHandler/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewPersistentStatusHandler_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	persistentHandler, err := NewPersistentStatusHandler(nil)

	assert.Nil(t, persistentHandler)
	assert.Equal(t, process.ErrNilMarshalizer, err)
}

func TestNewPersistentStatusHandler(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, err := NewPersistentStatusHandler(marshalizer)

	assert.NotNil(t, persistentHandler)
	assert.Nil(t, err)
}

func TestPersistentStatusHandler_SetStorageNilStorageShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer)

	err := persistentHandler.SetStorage(nil)
	assert.Equal(t, process.ErrNilStorage, err)
}

func TestPersistentStatusHandler_SetStorage(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer)
	storer := &mock.StorerStub{}
	err := persistentHandler.SetStorage(storer)

	assert.Nil(t, err)
}

func TestPersistentStatusHandler_SetUInt64ValueIncorrectMetricShouldNotSet(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer)

	key := "key"
	value := uint64(100)
	persistentHandler.SetUInt64Value(key, value)

	valueFromMap, ok := persistentHandler.persistentMetrics.Load(key)
	assert.Nil(t, valueFromMap)
	assert.Equal(t, false, ok)
}

func TestPersistentStatusHandler_SetUInt64Value(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer)

	key := core.MetricCountConsensus
	value := uint64(100)
	persistentHandler.SetUInt64Value(key, value)

	valueFromMap, ok := persistentHandler.persistentMetrics.Load(key)
	assert.Equal(t, value, valueFromMap)
	assert.Equal(t, true, ok)
}

func TestPersistentStatusHandler_IncrementNoMetricShouldReturn(t *testing.T) {
	t.Parallel()

	key := "key"
	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer)
	persistentHandler.Increment(key)

	valueFromMap, ok := persistentHandler.persistentMetrics.Load(key)
	assert.Nil(t, valueFromMap)
	assert.Equal(t, false, ok)
}

func TestPersistentStatusHandler_Increment(t *testing.T) {
	t.Parallel()

	key := core.MetricCountAcceptedBlocks
	value := uint64(100)
	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer)
	persistentHandler.AddUint64(key, value)
	persistentHandler.Increment(key)

	valueFromMap, ok := persistentHandler.persistentMetrics.Load(key)
	assert.Equal(t, value+1, valueFromMap)
	assert.Equal(t, true, ok)
}

func TestPersistentStatusHandler_AddUInt64ValueIncorrectMetricShouldNotSet(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer)

	key := "key"
	value := uint64(100)
	persistentHandler.AddUint64(key, value)

	valueFromMap, ok := persistentHandler.persistentMetrics.Load(key)
	assert.Nil(t, valueFromMap)
	assert.Equal(t, false, ok)
}

func TestPersistentStatusHandler_AddSetUInt64Value(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer)

	key := core.MetricCountConsensus
	value := uint64(100)
	persistentHandler.SetUInt64Value(key, value)
	persistentHandler.AddUint64(key, value)

	valueFromMap, ok := persistentHandler.persistentMetrics.Load(key)
	assert.Equal(t, value+value, valueFromMap)
	assert.Equal(t, true, ok)
}

func TestPersistentStatusHandler_LoadMetricsFromDbCannotGetFromStorerShouldNil(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer)
	storer := &mock.StorerStub{}
	storer.GetCalled = func(key []byte) (bytes []byte, err error) {
		return nil, errors.New("error")
	}
	_ = persistentHandler.SetStorage(storer)

	metricsUint64, metricsString := persistentHandler.LoadMetricsFromDb()
	assert.Nil(t, metricsString)
	assert.Nil(t, metricsUint64)
}

func TestPersistentStatusHandler_LoadMetricsFromDbCannotUnmarshalShouldNil(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerStub{}
	marshalizer.UnmarshalCalled = func(obj interface{}, buff []byte) error {
		return errors.New("error")
	}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer)
	storer := &mock.StorerStub{}
	storer.GetCalled = func(key []byte) (bytes []byte, err error) {
		return nil, nil
	}
	_ = persistentHandler.SetStorage(storer)

	metricsUint64, metricsString := persistentHandler.LoadMetricsFromDb()
	assert.Nil(t, metricsString)
	assert.Nil(t, metricsUint64)
}

func TestPersistentStatusHandler_LoadMetricsFromDb(t *testing.T) {
	t.Parallel()

	metrics := make(map[string]interface{})
	metrics[core.MetricCountAcceptedBlocks] = uint64(100)
	metricsBytes, _ := json.Marshal(&metrics)

	marshalizer := &mock.MarshalizerStub{}
	marshalizer.UnmarshalCalled = func(obj interface{}, buff []byte) error {
		_ = json.Unmarshal(buff, &obj)
		return nil
	}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer)
	storer := &mock.StorerStub{}
	storer.GetCalled = func(key []byte) (bytes []byte, err error) {

		return metricsBytes, nil
	}
	_ = persistentHandler.SetStorage(storer)

	metricsUint64, metricsString := persistentHandler.LoadMetricsFromDb()
	assert.NotNil(t, metricsString)
	assert.Equal(t, metrics[core.MetricCountAcceptedBlocks], metricsUint64[core.MetricCountAcceptedBlocks])
}

func TestPersistentStatusHandler_saveMetricsInDbMarshalError(t *testing.T) {
	t.Parallel()

	flag := 0
	marshalizer := &mock.MarshalizerStub{}
	marshalizer.MarshalCalled = func(obj interface{}) (bytes []byte, err error) {
		flag++
		return nil, errors.New("error")
	}
	storer := &mock.StorerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer)
	_ = persistentHandler.SetStorage(storer)

	persistentHandler.saveMetricsInDb()
	assert.Equal(t, 1, flag)
}

func TestPersistentStatusHandler_saveMetricsInDbPutError(t *testing.T) {
	t.Parallel()

	flag := 0
	marshalizer := &mock.MarshalizerStub{}
	marshalizer.MarshalCalled = func(obj interface{}) (bytes []byte, err error) {
		flag++
		return nil, nil
	}
	storer := &mock.StorerStub{}
	storer.PutCalled = func(key, data []byte) error {
		flag++
		return errors.New("error")
	}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer)
	_ = persistentHandler.SetStorage(storer)

	persistentHandler.saveMetricsInDb()
	assert.Equal(t, 2, flag)
}

func TestPersistentStatusHandler_DecrementNoMetricShouldReturn(t *testing.T) {
	t.Parallel()

	key := "key"
	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer)
	persistentHandler.Decrement(key)

	valueFromMap, ok := persistentHandler.persistentMetrics.Load(key)
	assert.Nil(t, valueFromMap)
	assert.Equal(t, false, ok)
}

func TestPersistentStatusHandler_Decrement(t *testing.T) {
	t.Parallel()

	key := core.MetricCountAcceptedBlocks
	value := uint64(100)
	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer)
	persistentHandler.SetUInt64Value(key, value)
	persistentHandler.Decrement(key)

	valueFromMap, ok := persistentHandler.persistentMetrics.Load(key)
	assert.Equal(t, value-1, valueFromMap)
	assert.Equal(t, true, ok)
}

func TestPersistentStatusHandler_DecrementKeyValueZeroShouldReturn(t *testing.T) {
	t.Parallel()

	key := core.MetricCountAcceptedBlocks
	value := uint64(0)
	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer)
	persistentHandler.SetUInt64Value(key, value)
	persistentHandler.Decrement(key)

	valueFromMap, ok := persistentHandler.persistentMetrics.Load(key)
	assert.Equal(t, value, valueFromMap)
	assert.Equal(t, true, ok)
}
