package statusHandler

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusStatusHandler will define the handler which will update prometheus metrics
type PrometheusStatusHandler struct {
	prometheusGaugeMetrics sync.Map
}

// InitMetricsMap will init the map of prometheus metrics
func (psh *PrometheusStatusHandler) InitMetricsMap() {
	psh.prometheusGaugeMetrics = sync.Map{}
}

// InitMetrics will declare and init all the metrics which should be used for Prometheus
func (psh *PrometheusStatusHandler) InitMetrics() {
	psh.InitMetricsMap()

	erdSynchronizedRound := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: core.MetricSynchronizedRound,
		Help: "The round where the synchronized blockchain is",
	})
	psh.prometheusGaugeMetrics.Store(core.MetricSynchronizedRound, erdSynchronizedRound)

	erdNonce := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: core.MetricNonce,
		Help: "The nonce for the node",
	})
	psh.prometheusGaugeMetrics.Store(core.MetricNonce, erdNonce)

	erdCurrentRound := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: core.MetricCurrentRound,
		Help: "The current round where the node is",
	})
	psh.prometheusGaugeMetrics.Store(core.MetricCurrentRound, erdCurrentRound)

	erdNumConnectedPeers := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: core.MetricNumConnectedPeers,
		Help: "The current number of peers connected",
	})
	psh.prometheusGaugeMetrics.Store(core.MetricNumConnectedPeers, erdNumConnectedPeers)

	erdIsSyncing := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: core.MetricIsSyncing,
		Help: "The synchronization state. If it's in process of syncing will be 1 and if it's synchronized will be 0",
	})
	psh.prometheusGaugeMetrics.Store(core.MetricIsSyncing, erdIsSyncing)

	psh.prometheusGaugeMetrics.Range(func(key, value interface{}) bool {
		gauge := value.(prometheus.Gauge)
		_ = prometheus.Register(gauge)
		return true
	})
}

// NewPrometheusStatusHandler will return an instance of a PrometheusStatusHandler
func NewPrometheusStatusHandler() *PrometheusStatusHandler {
	psh := new(PrometheusStatusHandler)
	psh.InitMetrics()
	return psh
}

// Increment will be used for incrementing the value for a key
func (psh *PrometheusStatusHandler) Increment(key string) {
	if metric, ok := psh.prometheusGaugeMetrics.Load(key); ok {
		metric.(prometheus.Gauge).Inc()
	}
}

// Decrement will be used for decrementing the value for a key
func (psh *PrometheusStatusHandler) Decrement(key string) {
	if metric, ok := psh.prometheusGaugeMetrics.Load(key); ok {
		metric.(prometheus.Gauge).Dec()
	}
}

// SetInt64Value method - will update the value for a key
func (psh *PrometheusStatusHandler) SetInt64Value(key string, value int64) {
	if metric, ok := psh.prometheusGaugeMetrics.Load(key); ok {
		metric.(prometheus.Gauge).Set(float64(value))
	}
}

// SetUInt64Value method - will update the value for a key
func (psh *PrometheusStatusHandler) SetUInt64Value(key string, value uint64) {
	if metric, ok := psh.prometheusGaugeMetrics.Load(key); ok {
		metric.(prometheus.Gauge).Set(float64(value))
	}
}

// Close will unregister Prometheus metrics
func (psh *PrometheusStatusHandler) Close() {
	psh.prometheusGaugeMetrics.Range(func(key, value interface{}) bool {
		gauge := value.(prometheus.Gauge)
		prometheus.Unregister(gauge)
		return true
	})
}
