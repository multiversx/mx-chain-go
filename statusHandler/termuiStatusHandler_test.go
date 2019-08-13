package statusHandler_test

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/stretchr/testify/assert"
	"testing"
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
	nonceI, err := termuiStatusHandler.GetTermuiMetricByKey(core.MetricNonce)
	nonce := nonceI.(int)

	assert.Equal(t, 0, nonce)
	assert.Nil(t, err)
}

func TestTermuiStatusHandler_TestIncrement(t *testing.T) {
	t.Parallel()

	var metricKey = core.MetricNonce

	termuiStatusHandler := statusHandler.NewTermuiStatusHandler()

	termuiStatusHandler.Increment(metricKey)
	valueI, err := termuiStatusHandler.GetTermuiMetricByKey(metricKey)
	assert.Nil(t, err)

	result := valueI.(int)
	assert.Equal(t, 1, result)
}

func TestTermuiStatusHandler_TestSetInt64(t *testing.T) {
	t.Parallel()

	var metricKey = core.MetricNonce
	var intValue = 100

	termuiStatusHandler := statusHandler.NewTermuiStatusHandler()

	termuiStatusHandler.SetInt64Value(metricKey, int64(intValue))
	valueI, err := termuiStatusHandler.GetTermuiMetricByKey(metricKey)
	assert.Nil(t, err)

	result := valueI.(int)
	assert.Equal(t, intValue, result)

}

func TestTermuiStatusHandler_TestSetUInt64(t *testing.T) {
	t.Parallel()

	var metricKey = core.MetricNonce
	var intValue = 200

	termuiStatusHandler := statusHandler.NewTermuiStatusHandler()

	termuiStatusHandler.SetUInt64Value(metricKey, uint64(intValue))
	valueI, err := termuiStatusHandler.GetTermuiMetricByKey(metricKey)
	assert.Nil(t, err)

	result := valueI.(int)
	assert.Equal(t, intValue, result)

}

func TestTermuiStatusHandler_TestSetString(t *testing.T) {
	t.Parallel()

	var metricKey = core.MetricPublicKey
	var stringValue = "KEY"

	termuiStatusHandler := statusHandler.NewTermuiStatusHandler()

	termuiStatusHandler.SetStringValue(metricKey, stringValue)
	valueI, err := termuiStatusHandler.GetTermuiMetricByKey(metricKey)
	assert.Nil(t, err)

	result := valueI.(string)
	assert.Equal(t, stringValue, result)
}
