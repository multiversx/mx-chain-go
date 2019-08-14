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

func TestTermuiStatusHandler_TestIncrement(t *testing.T) {
	t.Parallel()

	termuiStatusHandler := statusHandler.NewTermuiStatusHandler()

	termuiStatusHandler.SetUInt64Value(core.MetricNonce, 0)
	termuiStatusHandler.Increment(core.MetricNonce)
	valueI, err := termuiStatusHandler.GetTermuiMetricByKey(core.MetricNonce)
	assert.Nil(t, err)

	result := valueI.(uint64)
	assert.Equal(t, uint64(1), result)
}

func TestTermuiStatusHandler_TestSetInt64(t *testing.T) {
	t.Parallel()

	var intValue = int64(100)

	termuiStatusHandler := statusHandler.NewTermuiStatusHandler()

	termuiStatusHandler.SetInt64Value(core.MetricNonce, intValue)
	valueI, err := termuiStatusHandler.GetTermuiMetricByKey(core.MetricNonce)
	assert.Nil(t, err)

	result := valueI.(int64)
	assert.Equal(t, intValue, result)

}

func TestTermuiStatusHandler_TestSetUInt64(t *testing.T) {
	t.Parallel()

	var intValue = uint64(200)

	termuiStatusHandler := statusHandler.NewTermuiStatusHandler()

	termuiStatusHandler.SetUInt64Value(core.MetricNonce, intValue)
	valueI, err := termuiStatusHandler.GetTermuiMetricByKey(core.MetricNonce)
	assert.Nil(t, err)

	result := valueI.(uint64)
	assert.Equal(t, intValue, result)

}

func TestTermuiStatusHandler_TestSetString(t *testing.T) {
	t.Parallel()

	var stringValue = "KEY"

	termuiStatusHandler := statusHandler.NewTermuiStatusHandler()

	termuiStatusHandler.SetStringValue(core.MetricPublicKeyBlockSign, stringValue)
	valueI, err := termuiStatusHandler.GetTermuiMetricByKey(core.MetricPublicKeyBlockSign)
	assert.Nil(t, err)

	result := valueI.(string)
	assert.Equal(t, stringValue, result)
}
