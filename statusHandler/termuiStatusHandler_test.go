package statusHandler_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestTermuiStatusHandler_NewTermuiStatusHandler(t *testing.T) {
	t.Parallel()

	termuiStatusHandler := statusHandler.NewTermuiStatusHandler()

	assert.NotNil(t, termuiStatusHandler)
}

func TestTermuiStatusHandler_TermuiShouldPass(t *testing.T) {
	t.Parallel()

	termuiStatusHandler := statusHandler.NewTermuiStatusHandler()
	termuiConsole := termuiStatusHandler.Termui()

	assert.NotNil(t, termuiConsole)
}

func TestTermuiStatusHandler_TestIfMetricsAreInitialized(t *testing.T) {
	t.Parallel()

	termuiStatusHandler := statusHandler.NewTermuiStatusHandler()

	// check if nonce metric for example was initialized
	_, err := termuiStatusHandler.GetTermuiMetricByKey(core.MetricNonce)

	assert.Nil(t, err)
	assert.Equal(t, 26, termuiStatusHandler.GetMetricsCount())
}

func TestTermuiStatusHandler_TestIncrement(t *testing.T) {
	t.Parallel()

	var metricKey = core.MetricNonce

	termuiStatusHandler := statusHandler.NewTermuiStatusHandler()

	termuiStatusHandler.Increment(metricKey)
	valueI, err := termuiStatusHandler.GetTermuiMetricByKey(metricKey)
	assert.Nil(t, err)

	result := valueI.(uint64)
	assert.Equal(t, uint64(1), result)
}

func TestTermuiStatusHandler_TestSetInt64(t *testing.T) {
	t.Parallel()

	var metricKey = core.MetricNonce
	var intValue = int64(100)

	termuiStatusHandler := statusHandler.NewTermuiStatusHandler()

	termuiStatusHandler.SetInt64Value(metricKey, intValue)
	valueI, err := termuiStatusHandler.GetTermuiMetricByKey(metricKey)
	assert.Nil(t, err)

	result := valueI.(int64)
	assert.Equal(t, intValue, result)

}

func TestTermuiStatusHandler_TestSetUInt64(t *testing.T) {
	t.Parallel()

	var metricKey = core.MetricNonce
	var intValue = uint64(200)

	termuiStatusHandler := statusHandler.NewTermuiStatusHandler()

	termuiStatusHandler.SetUInt64Value(metricKey, intValue)
	valueI, err := termuiStatusHandler.GetTermuiMetricByKey(metricKey)
	assert.Nil(t, err)

	result := valueI.(uint64)
	assert.Equal(t, intValue, result)

}

func TestTermuiStatusHandler_TestSetString(t *testing.T) {
	t.Parallel()

	var metricKey = core.MetricPublicKeyBlockSign
	var stringValue = "KEY"

	termuiStatusHandler := statusHandler.NewTermuiStatusHandler()

	termuiStatusHandler.SetStringValue(metricKey, stringValue)
	valueI, err := termuiStatusHandler.GetTermuiMetricByKey(metricKey)
	assert.Nil(t, err)

	result := valueI.(string)
	assert.Equal(t, stringValue, result)
}
