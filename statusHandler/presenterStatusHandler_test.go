package statusHandler_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestPresenterStatusHandler_NewPresenterStatusHandler(t *testing.T) {
	t.Parallel()

	presenterStatusHandler := statusHandler.NewPresenterStatusHandler()

	assert.NotNil(t, presenterStatusHandler)
}

func TestPresenterStatusHandler_TestIncrement(t *testing.T) {
	t.Parallel()

	presenterStatusHandler := statusHandler.NewPresenterStatusHandler()

	presenterStatusHandler.SetUInt64Value(core.MetricNonce, 0)
	presenterStatusHandler.Increment(core.MetricNonce)
	result := presenterStatusHandler.GetNonce()

	assert.Equal(t, uint64(1), result)
}

func TestPresenterStatusHandler_TestSetInt64(t *testing.T) {
	t.Parallel()

	var intValue = int64(100)

	presenterStatusHandler := statusHandler.NewPresenterStatusHandler()

	presenterStatusHandler.SetInt64Value(core.MetricNonce, intValue)
	valueI, err := presenterStatusHandler.GetPresenterMetricByKey(core.MetricNonce)
	assert.Nil(t, err)

	result := valueI.(int64)

	assert.Equal(t, intValue, result)
}

func TestPresenterStatusHandler_TestSetUInt64(t *testing.T) {
	t.Parallel()

	var intValue = uint64(200)

	presenterStatusHandler := statusHandler.NewPresenterStatusHandler()

	presenterStatusHandler.SetUInt64Value(core.MetricNonce, intValue)
	result := presenterStatusHandler.GetNonce()

	assert.Equal(t, intValue, result)
}

func TestPresenterStatusHandler_TestSetString(t *testing.T) {
	t.Parallel()

	var stringValue = "KEY"

	presenterStatusHandler := statusHandler.NewPresenterStatusHandler()

	presenterStatusHandler.SetStringValue(core.MetricPublicKeyBlockSign, stringValue)
	result := presenterStatusHandler.GetPublicKeyBlockSign()

	assert.Equal(t, stringValue, result)
}
