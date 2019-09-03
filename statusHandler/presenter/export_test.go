package presenter

import "errors"

func (psh *PresenterStatusHandler) GetPresenterMetricByKey(key string) (interface{}, error) {
	value, ok := psh.presenterMetrics.Load(key)
	if ok {
		return value, nil
	}
	return nil, errors.New("metric does not exist")
}
