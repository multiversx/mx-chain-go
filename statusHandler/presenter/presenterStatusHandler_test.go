package presenter_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/statusHandler/presenter"
	"github.com/stretchr/testify/assert"
)

func TestPresenterStatusHandler_NewPresenterStatusHandler(t *testing.T) {
	t.Parallel()

	presenterStatusHandler := presenter.NewPresenterStatusHandler()

	assert.NotNil(t, presenterStatusHandler)
}

func TestPresenterStatusHandler_TestIncrement(t *testing.T) {
	t.Parallel()

	presenterStatusHandler := presenter.NewPresenterStatusHandler()

	presenterStatusHandler.SetUInt64Value(core.MetricNonce, 0)
	presenterStatusHandler.Increment(core.MetricNonce)
	result := presenterStatusHandler.GetNonce()

	assert.Equal(t, uint64(1), result)
}

func TestPresenterStatusHandler_WrongKeyIncrementShouldDoNothing(t *testing.T) {
	t.Parallel()

	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricNonce, 0)
	presenterStatusHandler.Increment("dummyKey")
	result := presenterStatusHandler.GetNonce()

	assert.Equal(t, uint64(0), result)
}

func TestPresenterStatusHandler_WrongTypeIncrementShouldDoNothing(t *testing.T) {
	t.Parallel()

	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(core.MetricNonce, "0")
	presenterStatusHandler.Increment(core.MetricNonce)
	result := presenterStatusHandler.GetNonce()

	assert.Equal(t, uint64(0), result)
}

func TestPresenterStatusHandler_TestSetInt64(t *testing.T) {
	t.Parallel()

	var intValue = int64(100)
	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.SetInt64Value(core.MetricNonce, intValue)
	valueI, err := presenterStatusHandler.GetPresenterMetricByKey(core.MetricNonce)
	assert.Nil(t, err)

	result := valueI.(int64)
	assert.Equal(t, intValue, result)
}

func TestPresenterStatusHandler_TestSetUInt64(t *testing.T) {
	t.Parallel()

	var intValue = uint64(200)

	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricNonce, intValue)
	result := presenterStatusHandler.GetNonce()

	assert.Equal(t, intValue, result)
}

func TestPresenterStatusHandler_TestSetString(t *testing.T) {
	t.Parallel()

	var stringValue = "KEY"

	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(core.MetricPublicKeyBlockSign, stringValue)
	result := presenterStatusHandler.GetPublicKeyBlockSign()

	assert.Equal(t, stringValue, result)
}

func TestPresenterStatusHandler_Write(t *testing.T) {
	t.Parallel()

	logLine := "Hello"
	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	logLineLen, err := presenterStatusHandler.Write([]byte(logLine))

	assert.Nil(t, err)
	assert.Equal(t, len(logLine), logLineLen)

}

func TestPresenterStatusHandler_GetLogLine(t *testing.T) {
	t.Parallel()

	logLine := "Hello"
	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	logLineLen, err := presenterStatusHandler.Write([]byte(logLine))

	assert.Nil(t, err)
	assert.Equal(t, len(logLine), logLineLen)

	time.Sleep(10 * time.Millisecond)
	logLines := presenterStatusHandler.GetLogLines()

	assert.Equal(t, 1, len(logLines))
	assert.Equal(t, logLine, logLines[0])
}

func TestPresenterStatusHandler_Increment(t *testing.T) {
	t.Parallel()

	countConsensus := uint64(0)
	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(core.MetricCountConsensus, countConsensus)
	presenterStatusHandler.Increment(core.MetricCountConsensus)
	result := presenterStatusHandler.GetCountConsensus()

	assert.Equal(t, countConsensus+1, result)
}
