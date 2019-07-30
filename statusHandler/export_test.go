package statusHandler

import (
	"errors"
	"github.com/prometheus/client_golang/prometheus"
)

func GetMetricByKey(key string) (prometheus.Gauge, error) {
	value, ok := prometheusGaugeMetrics.Load(key)
	if ok {
		return value.(prometheus.Gauge), nil
	}
	return nil, errors.New("metric does not exist")
}
