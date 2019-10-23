package persistor

import (
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
