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

func (psh *PresenterStatusHandler) GetPresenterMetricByKey(key string) (interface{}, error) {
	value, ok := psh.presenterMetrics.Load(key)
	if ok {
		return value, nil
	}
	return nil, errors.New("metric does not exist")
}
