package persister

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/statusHandler"
	"github.com/multiversx/mx-chain-go/statusHandler/mock"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPersistentStatusHandler_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	uit64Converter := &mock.Uint64ByteSliceConverterMock{}
	persistentHandler, err := NewPersistentStatusHandler(nil, uit64Converter)

	assert.Nil(t, persistentHandler)
	assert.Equal(t, statusHandler.ErrNilMarshalizer, err)
}

func TestNewPersistentStatusHandler_NilConverter(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, err := NewPersistentStatusHandler(marshalizer, nil)

	assert.Nil(t, persistentHandler)
	assert.Equal(t, statusHandler.ErrNilUint64Converter, err)
}

func TestNewPersistentStatusHandler(t *testing.T) {
	t.Parallel()

	uit64Converter := &mock.Uint64ByteSliceConverterMock{}
	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, err := NewPersistentStatusHandler(marshalizer, uit64Converter)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(persistentHandler))
}

func TestPersistentStatusHandler_SetStorageNilStorageShouldErr(t *testing.T) {
	t.Parallel()

	uit64Converter := &mock.Uint64ByteSliceConverterMock{}
	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer, uit64Converter)

	err := persistentHandler.SetStorage(nil)
	assert.Equal(t, statusHandler.ErrNilStorage, err)
}

func TestPersistentStatusHandler_SetStorage(t *testing.T) {
	t.Parallel()

	uit64Converter := &mock.Uint64ByteSliceConverterMock{}
	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer, uit64Converter)
	storer := &storageStubs.StorerStub{}

	err := persistentHandler.SetStorage(storer)
	assert.Nil(t, err)
}

func TestPersistentStatusHandler_SetUInt64ValueIncorrectMetricShouldNotSet(t *testing.T) {
	t.Parallel()

	uit64Converter := &mock.Uint64ByteSliceConverterMock{}
	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer, uit64Converter)

	key := "key"
	value := uint64(100)
	persistentHandler.SetUInt64Value(key, value)

	valueFromMap, ok := persistentHandler.persistentMetrics.Load(key)
	assert.Nil(t, valueFromMap)
	assert.Equal(t, false, ok)
}

func TestPersistentStatusHandler_SetUInt64Value(t *testing.T) {
	t.Parallel()

	uit64Converter := &mock.Uint64ByteSliceConverterMock{}
	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer, uit64Converter)

	key := common.MetricCountConsensus
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
	uit64Converter := &mock.Uint64ByteSliceConverterMock{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer, uit64Converter)
	persistentHandler.Increment(key)

	valueFromMap, ok := persistentHandler.persistentMetrics.Load(key)
	assert.Nil(t, valueFromMap)
	assert.Equal(t, false, ok)
}

func TestPersistentStatusHandler_Increment(t *testing.T) {
	t.Parallel()

	key := common.MetricCountAcceptedBlocks
	value := uint64(100)
	marshalizer := &mock.MarshalizerStub{}
	uit64Converter := &mock.Uint64ByteSliceConverterMock{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer, uit64Converter)
	persistentHandler.AddUint64(key, value)
	persistentHandler.Increment(key)

	valueFromMap, ok := persistentHandler.persistentMetrics.Load(key)
	assert.Equal(t, value+1, valueFromMap)
	assert.Equal(t, true, ok)
}

func TestPersistentStatusHandler_AddUInt64ValueIncorrectMetricShouldNotSet(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerStub{}
	uit64Converter := &mock.Uint64ByteSliceConverterMock{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer, uit64Converter)

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
	uit64Converter := &mock.Uint64ByteSliceConverterMock{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer, uit64Converter)

	key := common.MetricCountConsensus
	value := uint64(100)
	persistentHandler.SetUInt64Value(key, value)
	persistentHandler.AddUint64(key, value)

	valueFromMap, ok := persistentHandler.persistentMetrics.Load(key)
	assert.Equal(t, value+value, valueFromMap)
	assert.Equal(t, true, ok)
}

func TestPersistentStatusHandler_saveMetricsInDbMarshalError(t *testing.T) {
	t.Parallel()

	flag := 0
	marshalizer := &mock.MarshalizerStub{}
	marshalizer.MarshalCalled = func(obj interface{}) (bytes []byte, err error) {
		flag++
		return nil, errors.New("error")
	}
	storer := &storageStubs.StorerStub{}
	uit64Converter := &mock.Uint64ByteSliceConverterMock{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer, uit64Converter)
	_ = persistentHandler.SetStorage(storer)

	persistentHandler.saveMetricsInDb(0)
	assert.Equal(t, 1, flag)
}

func TestPersistentStatusHandler_saveMetricsInDbPutError(t *testing.T) {
	t.Parallel()

	flag := 0
	uit64Converter := &mock.Uint64ByteSliceConverterMock{}
	marshalizer := &mock.MarshalizerStub{}
	marshalizer.MarshalCalled = func(obj interface{}) (bytes []byte, err error) {
		flag++
		return nil, nil
	}
	storer := &storageStubs.StorerStub{}
	storer.PutCalled = func(key, data []byte) error {
		flag++
		return errors.New("error")
	}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer, uit64Converter)
	_ = persistentHandler.SetStorage(storer)

	persistentHandler.saveMetricsInDb(0)
	assert.Equal(t, 2, flag)
}

func TestPersistentStatusHandler_DecrementNoMetricShouldReturn(t *testing.T) {
	t.Parallel()

	key := "key"
	uit64Converter := &mock.Uint64ByteSliceConverterMock{}
	marshalizer := &mock.MarshalizerStub{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer, uit64Converter)
	persistentHandler.Decrement(key)

	valueFromMap, ok := persistentHandler.persistentMetrics.Load(key)
	assert.Nil(t, valueFromMap)
	assert.Equal(t, false, ok)
}

func TestPersistentStatusHandler_Decrement(t *testing.T) {
	t.Parallel()

	key := common.MetricCountAcceptedBlocks
	value := uint64(100)
	marshalizer := &mock.MarshalizerStub{}
	uit64Converter := &mock.Uint64ByteSliceConverterMock{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer, uit64Converter)
	persistentHandler.SetUInt64Value(key, value)
	persistentHandler.Decrement(key)

	valueFromMap, ok := persistentHandler.persistentMetrics.Load(key)
	assert.Equal(t, value-1, valueFromMap)
	assert.Equal(t, true, ok)
}

func TestPersistentStatusHandler_DecrementKeyValueZeroShouldReturn(t *testing.T) {
	t.Parallel()

	key := common.MetricCountAcceptedBlocks
	value := uint64(0)
	marshalizer := &mock.MarshalizerStub{}
	uit64Converter := &mock.Uint64ByteSliceConverterMock{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer, uit64Converter)
	persistentHandler.SetUInt64Value(key, value)
	persistentHandler.Decrement(key)

	valueFromMap, ok := persistentHandler.persistentMetrics.Load(key)
	assert.Equal(t, value, valueFromMap)
	assert.Equal(t, true, ok)
}

func TestPersistentStatusHandler_SetMetricNonce(t *testing.T) {
	t.Parallel()

	called := false
	storer := &storageStubs.StorerStub{}
	marshalizer := &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (bytes []byte, err error) {
			called = true
			return nil, errors.New("err")
		},
	}
	uit64Converter := &mock.Uint64ByteSliceConverterMock{}
	persistentHandler, _ := NewPersistentStatusHandler(marshalizer, uit64Converter)
	_ = persistentHandler.SetStorage(storer)
	time.Sleep(2 * time.Second)

	persistentHandler.SetUInt64Value(common.MetricNonce, 1)
	require.True(t, called)
}
