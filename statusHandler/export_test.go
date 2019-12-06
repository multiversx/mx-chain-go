package statusHandler

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
)

func (psh *PrometheusStatusHandler) GetPrometheusMetricByKey(key string) (prometheus.Gauge, error) {
	value, ok := psh.prometheusGaugeMetrics.Load(key)
	if ok {
		return value.(prometheus.Gauge), nil
	}
	return nil, errors.New("metric does not exist")
}

// StatusMetricsMap will return all metrics in a map
func (nd *statusMetrics) StatusMetricsMap() map[string]interface{} {
	statusMetricsMap := make(map[string]interface{})
	nd.nodeMetrics.Range(func(key, value interface{}) bool {
		statusMetricsMap[key.(string)] = value
		return true
	})

	return statusMetricsMap
}
