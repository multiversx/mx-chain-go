package statusHandler_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	prometheusUtils "github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestPrometheusStatusHandler_NewPrometheusStatusHandler(t *testing.T) {
	t.Parallel()

	var promStatusHandler core.AppStatusHandler
	promStatusHandler = statusHandler.NewPrometheusStatusHandler()
	assert.NotNil(t, promStatusHandler)
}

func TestPrometheusStatusHandler_TestIfMetricsAreInitialized(t *testing.T) {
	t.Parallel()

	_ = statusHandler.NewPrometheusStatusHandler()

	// check if nonce metric for example was initialized
	_, err := statusHandler.GetMetricByKey(core.MetricNumConnectedPeers)
	assert.Nil(t, err)
}

func TestPrometheusStatusHandler_TestIncrementAndDecrement(t *testing.T) {
	t.Parallel()

	var metricKey = core.MetricNonce

	var promStatusHandler core.AppStatusHandler
	promStatusHandler = statusHandler.NewPrometheusStatusHandler()

	// increment the nonce metric
	promStatusHandler.Increment(metricKey)

	// get the gauge
	gauge, err := statusHandler.GetMetricByKey(metricKey)
	assert.Nil(t, err)

	result := prometheusUtils.ToFloat64(gauge)
	// test if the metric was incremented
	assert.Equal(t, float64(1), result)

	// increment the metric two more times and check if it worked
	promStatusHandler.Increment(metricKey)
	promStatusHandler.Increment(metricKey)

	result = prometheusUtils.ToFloat64(gauge)
	assert.Equal(t, float64(3), result)

	// now decrement the metric
	promStatusHandler.Decrement(metricKey)

	result = prometheusUtils.ToFloat64(gauge)

	assert.Equal(t, float64(2), result)
}

func TestPrometheusStatusHandler_TestSetInt64ValueAndSetUInt64Value(t *testing.T) {
	t.Parallel()

	var metricKey = core.MetricCurrentRound

	var promStatusHandler core.AppStatusHandler
	promStatusHandler = statusHandler.NewPrometheusStatusHandler()

	// set an int64 value
	promStatusHandler.SetInt64Value(metricKey, int64(10))

	gauge, err := statusHandler.GetMetricByKey(metricKey)
	assert.Nil(t, err)

	result := prometheusUtils.ToFloat64(gauge)
	// test if the metric value was updated
	assert.Equal(t, float64(10), result)

	// set an uint64 value
	promStatusHandler.SetUInt64Value(metricKey, uint64(20))

	gauge, err = statusHandler.GetMetricByKey(metricKey)
	assert.Nil(t, err)

	result = prometheusUtils.ToFloat64(gauge)
	// test if the metric value was updated
	assert.Equal(t, float64(20), result)
}
