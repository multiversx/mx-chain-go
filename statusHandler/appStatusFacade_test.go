package statusHandler_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	prometheusUtils "github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestNewAppStatusFacadeWithHandlers_NilHandlersShouldFail(t *testing.T) {
	t.Parallel()

	_, err := statusHandler.NewAppStatusFacadeWithHandlers()
	assert.Equal(t, statusHandler.ErrNilHandlersSlice, err)
}

func TestNewAppStatusFacadeWithHandlers_OkHandlersShouldPass(t *testing.T) {
	t.Parallel()

	_, err := statusHandler.NewAppStatusFacadeWithHandlers(statusHandler.NewNillStatusHandler(),
		statusHandler.NewPrometheusStatusHandler())

	assert.Nil(t, err)
}

func TestAppStatusFacade_IncrementAndDecrementShouldPass(t *testing.T) {
	t.Parallel()

	var metricKey = core.MetricSynchronizedRound

	// we create a new facade which contains a prometheus handler in order to test
	promStatusHandler := statusHandler.NewPrometheusStatusHandler()
	asf, err := statusHandler.NewAppStatusFacadeWithHandlers(statusHandler.NewNillStatusHandler(), promStatusHandler)
	assert.Nil(t, err)

	asf.Increment(metricKey)

	time.Sleep(5 * time.Millisecond)
	gauge, err := promStatusHandler.GetPrometheusMetricByKey(metricKey)
	assert.Nil(t, err)

	result := prometheusUtils.ToFloat64(gauge)
	assert.Equal(t, float64(1), result)

	asf.Decrement(metricKey)

	time.Sleep(5 * time.Millisecond)
	result = prometheusUtils.ToFloat64(gauge)
	assert.Equal(t, float64(0), result)
}

func TestAppStatusFacade_SetInt64ValueAndSetUint64ValueShouldPass(t *testing.T) {
	t.Parallel()

	var metricKey = core.MetricSynchronizedRound

	// we create a new facade which contains a prometheus handler in order to test
	promStatusHandler := statusHandler.NewPrometheusStatusHandler()
	asf, err := statusHandler.NewAppStatusFacadeWithHandlers(statusHandler.NewNillStatusHandler(), promStatusHandler)
	assert.Nil(t, err)

	// set an int64 value
	asf.SetInt64Value(metricKey, int64(10))

	time.Sleep(5 * time.Millisecond)
	gauge, err := promStatusHandler.GetPrometheusMetricByKey(metricKey)
	assert.Nil(t, err)

	result := prometheusUtils.ToFloat64(gauge)
	// test if the metric value was updated
	assert.Equal(t, float64(10), result)

	// set an uint64 value
	asf.SetUInt64Value(metricKey, uint64(20))

	time.Sleep(5 * time.Millisecond)
	gauge, err = promStatusHandler.GetPrometheusMetricByKey(metricKey)
	assert.Nil(t, err)

	result = prometheusUtils.ToFloat64(gauge)
	// test if the metric value was updated
	assert.Equal(t, float64(20), result)
}
