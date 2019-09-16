package statusHandler_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestNewStatusMetricsProvider(t *testing.T) {
	t.Parallel()

	ndh := statusHandler.NewStatusMetricsProvider()
	assert.NotNil(t, ndh)
	assert.False(t, ndh.IsInterfaceNil())
}

func TestStatusMetricsProvider_IncrementCallNonExistingKey(t *testing.T) {
	t.Parallel()

	ndh := statusHandler.NewStatusMetricsProvider()
	key1 := "test-key1"
	ndh.Increment(key1)

	retMap, err := ndh.StatusMetricsMap()
	assert.Nil(t, err)
	assert.Nil(t, retMap[key1])
}

func TestStatusMetricsProvider_IncrementNonUint64ValueShouldNotWork(t *testing.T) {
	t.Parallel()

	ndh := statusHandler.NewStatusMetricsProvider()
	key1 := "test-key2"
	value1 := "value2"

	// set a key which is initialized with a string and the Increment method won't affect the key
	ndh.SetStringValue(key1, value1)
	ndh.Increment(key1)

	retMap, _ := ndh.StatusMetricsMap()
	assert.Equal(t, value1, retMap[key1])
}

func TestStatusMetricsProvider_IncrementShouldWork(t *testing.T) {
	t.Parallel()

	ndh := statusHandler.NewStatusMetricsProvider()
	key1 := "test-key3"
	ndh.SetUInt64Value(key1, 0)
	ndh.Increment(key1)

	retMap, err := ndh.StatusMetricsMap()
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), retMap[key1])
}

func TestStatusMetricsProvider_Decrement(t *testing.T) {
	t.Parallel()

	ndh := statusHandler.NewStatusMetricsProvider()
	key := "test-key4"
	ndh.SetUInt64Value(key, 2)
	ndh.Decrement(key)

	retMap, err := ndh.StatusMetricsMap()
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), retMap[key])
}

func TestStatusMetricsProvider_SetInt64Value(t *testing.T) {
	t.Parallel()

	ndh := statusHandler.NewStatusMetricsProvider()
	key := "test-key5"
	value := int64(5)
	ndh.SetInt64Value(key, value)
	ndh.Decrement(key)

	retMap, err := ndh.StatusMetricsMap()
	assert.Nil(t, err)
	assert.Equal(t, value, retMap[key])
}

func TestStatusMetricsProvider_SetStringValue(t *testing.T) {
	t.Parallel()

	ndh := statusHandler.NewStatusMetricsProvider()
	key := "test-key6"
	value := "value"
	ndh.SetStringValue(key, value)
	ndh.Decrement(key)

	retMap, err := ndh.StatusMetricsMap()
	assert.Nil(t, err)
	assert.Equal(t, value, retMap[key])
}
