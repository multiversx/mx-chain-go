package presenter_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/cmd/termui/presenter"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/stretchr/testify/assert"
)

func TestPresenterStatusHandler_NewPresenterStatusHandler(t *testing.T) {
	t.Parallel()

	presenterStatusHandler := presenter.NewPresenterStatusHandler()

	assert.False(t, check.IfNil(presenterStatusHandler))
}

func TestPresenterStatusHandler_TestIncrement(t *testing.T) {
	t.Parallel()

	presenterStatusHandler := presenter.NewPresenterStatusHandler()

	presenterStatusHandler.SetUInt64Value(common.MetricNonce, 0)
	presenterStatusHandler.Increment(common.MetricNonce)
	result := presenterStatusHandler.GetNonce()

	assert.Equal(t, uint64(1), result)
}

func TestPresenterStatusHandler_WrongKeyIncrementShouldDoNothing(t *testing.T) {
	t.Parallel()

	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(common.MetricNonce, 0)
	presenterStatusHandler.Increment("dummyKey")
	result := presenterStatusHandler.GetNonce()

	assert.Equal(t, uint64(0), result)
}

func TestPresenterStatusHandler_WrongTypeIncrementShouldDoNothing(t *testing.T) {
	t.Parallel()

	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(common.MetricNonce, "0")
	presenterStatusHandler.Increment(common.MetricNonce)
	result := presenterStatusHandler.GetNonce()

	assert.Equal(t, uint64(0), result)
}

func TestPresenterStatusHandler_TestSetInt64(t *testing.T) {
	t.Parallel()

	var intValue = int64(100)
	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.SetInt64Value(common.MetricNonce, intValue)
	valueI, err := presenterStatusHandler.GetPresenterMetricByKey(common.MetricNonce)
	assert.Nil(t, err)

	result := valueI.(int64)
	assert.Equal(t, intValue, result)
}

func TestPresenterStatusHandler_TestSetUInt64(t *testing.T) {
	t.Parallel()

	var intValue = uint64(200)

	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(common.MetricNonce, intValue)
	result := presenterStatusHandler.GetNonce()

	assert.Equal(t, intValue, result)
}

func TestPresenterStatusHandler_TestSetString(t *testing.T) {
	t.Parallel()

	var stringValue = "KEY"

	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(common.MetricPublicKeyBlockSign, stringValue)
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

	waitTimeBeforeRead := 500 * time.Millisecond
	logLine := "Hello"
	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	logLineLen, err := presenterStatusHandler.Write([]byte(logLine))

	assert.Nil(t, err)
	assert.Equal(t, len(logLine), logLineLen)

	time.Sleep(waitTimeBeforeRead)
	logLines := presenterStatusHandler.GetLogLines()

	assert.Equal(t, 1, len(logLines))
	assert.Equal(t, logLine, logLines[0])
}

func TestPresenterStatusHandler_Increment(t *testing.T) {
	t.Parallel()

	countConsensus := uint64(0)
	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(common.MetricCountConsensus, countConsensus)
	presenterStatusHandler.Increment(common.MetricCountConsensus)
	result := presenterStatusHandler.GetCountConsensus()

	assert.Equal(t, countConsensus+1, result)
}

func TestPresenterStatusHandler_WrongTypeDecrement(t *testing.T) {
	t.Parallel()

	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.SetStringValue(common.MetricNonce, "value")
	presenterStatusHandler.Decrement(common.MetricNonce)
	result := presenterStatusHandler.GetNonce()

	assert.Equal(t, uint64(0), result)
}

func TestPresenterStatusHandler_DecrementDoNothing(t *testing.T) {
	t.Parallel()

	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(common.MetricCountConsensus, 0)
	presenterStatusHandler.Decrement(common.MetricCountConsensus)
	result := presenterStatusHandler.GetCountConsensus()

	assert.Equal(t, uint64(0), result)
}

func TestPresenterStatusHandler_WrongKeyDecrement(t *testing.T) {
	t.Parallel()

	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.Decrement("dummy")
	result := presenterStatusHandler.GetCountConsensus()

	assert.Equal(t, uint64(0), result)
}

func TestPresenterStatusHandler_Decrement(t *testing.T) {
	t.Parallel()

	countConsensus := uint64(10)
	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(common.MetricCountConsensus, countConsensus)
	presenterStatusHandler.Decrement(common.MetricCountConsensus)
	result := presenterStatusHandler.GetCountConsensus()

	assert.Equal(t, countConsensus-1, result)
}

func TestPresenterStatusHandler_AddUint64(t *testing.T) {
	t.Parallel()

	countConsensus := uint64(10)
	value := uint64(5)
	presenterStatusHandler := presenter.NewPresenterStatusHandler()
	presenterStatusHandler.SetUInt64Value(common.MetricCountConsensus, countConsensus)
	presenterStatusHandler.AddUint64(common.MetricCountConsensus, value)
	result := presenterStatusHandler.GetCountConsensus()

	assert.Equal(t, countConsensus+value, result)
}

func TestPresenterStatusHandler_ConcurrentOperations(t *testing.T) {
	t.Parallel()

	psh := presenter.NewPresenterStatusHandler()

	for i := 0; i < 1000; i++ {
		go func(i int) {
			modRes := i % 5
			switch modRes {
			case 0:
				psh.SetStringValue("test-key-2", "test-value")
			case 1:
				psh.Increment("test-key")
			case 2:
				psh.Decrement("test-key")
			case 3:
				psh.SetInt64Value("test-key-3", 37)
			case 4:
				psh.InvalidateCache()
			}
		}(i)
	}
}
