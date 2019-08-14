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

func (tsh *TermuiStatusHandler) GetTermuiMetricByKey(key string) (interface{}, error) {
	value, ok := tsh.termuiConsoleMetrics.Load(key)
	if ok {
		return value, nil
	}
	return nil, errors.New("metric does not exist")
}

func (tsh *TermuiStatusHandler) GetMetricsCount() int {
	count := 0
	tsh.termuiConsoleMetrics.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	return count
}
