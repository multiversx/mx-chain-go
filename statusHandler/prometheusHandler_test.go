package statusHandler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	promStatusHandler := statusHandler.NewPrometheusStatusHandler()

	// check if nonce metric for example was initialized
	_, err := promStatusHandler.GetPrometheusMetricByKey(core.MetricNumConnectedPeers)
	assert.Nil(t, err)
}

func TestPrometheusStatusHandler_TestIncrement(t *testing.T) {
	t.Parallel()

	var metricKey = core.MetricNonce

	promStatusHandler := statusHandler.NewPrometheusStatusHandler()

	// increment the nonce metric
	promStatusHandler.Increment(metricKey)

	// get the gauge
	gauge, err := promStatusHandler.GetPrometheusMetricByKey(metricKey)
	assert.Nil(t, err)

	result := prometheusUtils.ToFloat64(gauge)
	// test if the metric was incremented
	assert.Equal(t, float64(1), result)
}

func TestPrometheusStatusHandler_TestDecrement(t *testing.T) {
	t.Parallel()

	var metricKey = core.MetricNonce

	promStatusHandler := statusHandler.NewPrometheusStatusHandler()

	// get the gauge
	gauge, err := promStatusHandler.GetPrometheusMetricByKey(metricKey)
	assert.Nil(t, err)

	// now decrement the metric
	promStatusHandler.Decrement(metricKey)

	result := prometheusUtils.ToFloat64(gauge)

	assert.Equal(t, float64(-1), result)
}

func TestPrometheusStatusHandler_TestSetInt64Value(t *testing.T) {
	t.Parallel()

	var metricKey = core.MetricCurrentRound

	promStatusHandler := statusHandler.NewPrometheusStatusHandler()

	// set an int64 value
	promStatusHandler.SetInt64Value(metricKey, int64(10))

	gauge, err := promStatusHandler.GetPrometheusMetricByKey(metricKey)
	assert.Nil(t, err)

	result := prometheusUtils.ToFloat64(gauge)
	// test if the metric value was updated
	assert.Equal(t, float64(10), result)
}

func TestPrometheusStatusHandler_TestSetUInt64Value(t *testing.T) {
	t.Parallel()

	var metricKey = core.MetricCurrentRound

	promStatusHandler := statusHandler.NewPrometheusStatusHandler()

	// set an uint64 value
	promStatusHandler.SetUInt64Value(metricKey, uint64(20))

	gauge, err := promStatusHandler.GetPrometheusMetricByKey(metricKey)
	assert.Nil(t, err)

	result := prometheusUtils.ToFloat64(gauge)
	// test if the metric value was updated
	assert.Equal(t, float64(20), result)
}

func BenchmarkPrometheusStatusHandler_Increment(b *testing.B) {
	var promStatusHandler core.AppStatusHandler
	promStatusHandler = statusHandler.NewPrometheusStatusHandler()

	testServer := httptest.NewServer(promhttp.Handler())
	defer testServer.Close()
	b.ResetTimer()

	_, err := http.Get(testServer.URL)
	assert.Nil(b, err)

	for n := 0; n < b.N; n++ {
		promStatusHandler.Increment(core.MetricIsSyncing)
	}
	promStatusHandler.Close()
}

func BenchmarkPrometheusStatusHandler_Decrement(b *testing.B) {
	var promStatusHandler core.AppStatusHandler
	promStatusHandler = statusHandler.NewPrometheusStatusHandler()

	testServer := httptest.NewServer(promhttp.Handler())
	defer testServer.Close()
	b.ResetTimer()

	_, err := http.Get(testServer.URL)
	assert.Nil(b, err)

	for n := 0; n < b.N; n++ {
		promStatusHandler.Decrement(core.MetricIsSyncing)
	}
	promStatusHandler.Close()
}

func BenchmarkPrometheusStatusHandler_SetInt64Value(b *testing.B) {
	var promStatusHandler core.AppStatusHandler

	promStatusHandler = statusHandler.NewPrometheusStatusHandler()

	testServer := httptest.NewServer(promhttp.Handler())
	defer testServer.Close()
	b.ResetTimer()

	_, err := http.Get(testServer.URL)
	assert.Nil(b, err)

	for n := 0; n < b.N; n++ {
		promStatusHandler.SetInt64Value(core.MetricIsSyncing, int64(10))
	}
	promStatusHandler.Close()
}

func BenchmarkPrometheusStatusHandler_SetUInt64Value(b *testing.B) {
	var promStatusHandler core.AppStatusHandler
	promStatusHandler = statusHandler.NewPrometheusStatusHandler()

	testServer := httptest.NewServer(promhttp.Handler())
	defer testServer.Close()
	b.ResetTimer()

	_, err := http.Get(testServer.URL)
	assert.Nil(b, err)

	for n := 0; n < b.N; n++ {
		promStatusHandler.SetUInt64Value(core.MetricIsSyncing, uint64(10))
	}
	promStatusHandler.Close()
}
