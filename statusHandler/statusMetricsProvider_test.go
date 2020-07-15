package statusHandler_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestNewStatusMetricsProvider(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	assert.NotNil(t, sm)
	assert.False(t, sm.IsInterfaceNil())
}

func TestStatusMetricsProvider_IncrementCallNonExistingKey(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	key1 := "test-key1"
	sm.Increment(key1)

	retMap := sm.StatusMetricsMap()

	assert.Nil(t, retMap[key1])
}

func TestStatusMetricsProvider_IncrementNonUint64ValueShouldNotWork(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	key1 := "test-key2"
	value1 := "value2"
	// set a key which is initialized with a string and the Increment method won't affect the key
	sm.SetStringValue(key1, value1)
	sm.Increment(key1)

	retMap := sm.StatusMetricsMap()
	assert.Equal(t, value1, retMap[key1])
}

func TestStatusMetricsProvider_IncrementShouldWork(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	key1 := "test-key3"
	sm.SetUInt64Value(key1, 0)
	sm.Increment(key1)

	retMap := sm.StatusMetricsMap()

	assert.Equal(t, uint64(1), retMap[key1])
}

func TestStatusMetricsProvider_Decrement(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	key := "test-key4"
	sm.SetUInt64Value(key, 2)
	sm.Decrement(key)

	retMap := sm.StatusMetricsMap()

	assert.Equal(t, uint64(1), retMap[key])
}

func TestStatusMetricsProvider_SetInt64Value(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	key := "test-key5"
	value := int64(5)
	sm.SetInt64Value(key, value)
	sm.Decrement(key)

	retMap := sm.StatusMetricsMap()

	assert.Equal(t, value, retMap[key])
}

func TestStatusMetricsProvider_SetStringValue(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	key := "test-key6"
	value := "value"
	sm.SetStringValue(key, value)
	sm.Decrement(key)

	retMap := sm.StatusMetricsMap()

	assert.Equal(t, value, retMap[key])
}

func TestStatusMetricsProvider_AddUint64Value(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	key := "test-key6"
	value := uint64(100)
	sm.SetUInt64Value(key, value)
	sm.AddUint64(key, value)

	retMap := sm.StatusMetricsMap()
	assert.Equal(t, value+value, retMap[key])
}

func TestStatusMetrics_StatusMetricsWithoutP2PPrometheusString(t *testing.T) {
	t.Parallel()

	sm := statusHandler.NewStatusMetrics()
	key1, value1 := "test-key7", uint64(100)
	key2, value2 := "test-key8", "value8"
	sm.SetUInt64Value(key1, value1)
	sm.SetStringValue(key2, value2)

	strRes := sm.StatusMetricsWithoutP2PPrometheusString()

	// result should contain only numeric values
	assert.True(t, strings.Contains(strRes, fmt.Sprintf("%s %v", key1, value1)))
}
