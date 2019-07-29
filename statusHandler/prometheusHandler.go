package statusHandler

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/prometheus/client_golang/prometheus"
	prometheusUtils "github.com/prometheus/client_golang/prometheus/testutil"
)

// PrometheusStatusHandler will define the handler which will update prometheus metrics
type PrometheusStatusHandler struct {
}

var prometheusGaugeMetrics map[string]prometheus.Gauge

// InitMetrics will declare and init all the metrics which should be used for Prometheus
func (psh *PrometheusStatusHandler) InitMetrics() {
	prometheusGaugeMetrics = make(map[string]prometheus.Gauge)

	erdSynchronizedRound := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: core.MetricSynchronizedRound,
		Help: "todo",
	})

	erdNonce := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: core.MetricNonce,
		Help: "todo",
	})
	prometheusGaugeMetrics[core.MetricNonce] = erdNonce

	prometheusGaugeMetrics[core.MetricSynchronizedRound] = erdSynchronizedRound

	erdCurrentRound := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: core.MetricCurrentRound,
		Help: "todo",
	})
	prometheusGaugeMetrics[core.MetricCurrentRound] = erdCurrentRound

	erdNumConnectedPeers := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: core.MetricNumConnectedPeers,
		Help: "todo",
	})
	prometheusGaugeMetrics[core.MetricNumConnectedPeers] = erdNumConnectedPeers

	erdIsSyncing := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: core.MetricIsSyncing,
		Help: "todo",
	})
	prometheusGaugeMetrics[core.MetricIsSyncing] = erdIsSyncing

	for _, collector := range prometheusGaugeMetrics {
		_ = prometheus.Register(collector)
	}

}

// NewPrometheusStatusHandler will return an instance of a PrometheusStatusHandler
func NewPrometheusStatusHandler() *PrometheusStatusHandler {
	psh := new(PrometheusStatusHandler)
	psh.InitMetrics()
	return psh
}

// Increment will be used for incrementing the value for a key
func (psh *PrometheusStatusHandler) Increment(key string) {
	if metric, ok := prometheusGaugeMetrics[key]; ok {
		metric.Inc()
	}
}

// Decrement will be used for decrementing the value for a key
func (psh *PrometheusStatusHandler) Decrement(key string) {
	if metric, ok := prometheusGaugeMetrics[key]; ok {
		metric.Dec()
	}
}

// SetInt64Value method - will update the value for a key
func (psh *PrometheusStatusHandler) SetInt64Value(key string, value int64) {
	if metric, ok := prometheusGaugeMetrics[key]; ok {
		metric.Set(float64(value))
	}
}

// SetUInt64Value method - will update the value for a key
func (psh *PrometheusStatusHandler) SetUInt64Value(key string, value uint64) {
	if metric, ok := prometheusGaugeMetrics[key]; ok {
		metric.Set(float64(value))
	}
}

// GetValue method - will fetch the value for a key - TESTING ONLY
func (psh *PrometheusStatusHandler) GetValue(key string) float64 {
	if metric, ok := prometheusGaugeMetrics[key]; ok {
		return prometheusUtils.ToFloat64(metric)
	}
	return float64(0)
}

// Close will unregister Prometheus metrics
func (psh *PrometheusStatusHandler) Close() {
	for _, collector := range prometheusGaugeMetrics {
		prometheus.Unregister(collector)
	}
	prometheusGaugeMetrics = make(map[string]prometheus.Gauge, 0)
}
