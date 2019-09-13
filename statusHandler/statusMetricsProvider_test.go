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
}

func TestStatusMetricsProvider_Increment(t *testing.T) {
	t.Parallel()

	ndh := statusHandler.NewStatusMetricsProvider()
	key1 := "test-key1"
	ndh.SetUInt64Value(key1, 0)
	ndh.Increment(key1)

	retMap, err := ndh.StatusMetricsMap()
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), retMap[key1])
}

func TestStatusMetricsProvider_Decrement(t *testing.T) {
	t.Parallel()

	ndh := statusHandler.NewStatusMetricsProvider()
	key := "test-key2"
	ndh.SetUInt64Value(key, 2)
	ndh.Decrement(key)

	retMap, err := ndh.StatusMetricsMap()
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), retMap[key])
}

func TestStatusMetricsProvider_SetInt64Value(t *testing.T) {
	t.Parallel()

	ndh := statusHandler.NewStatusMetricsProvider()
	key := "test-key3"
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
	key := "test-key4"
	value := "value"
	ndh.SetStringValue(key, value)
	ndh.Decrement(key)

	retMap, err := ndh.StatusMetricsMap()
	assert.Nil(t, err)
	assert.Equal(t, value, retMap[key])
}
