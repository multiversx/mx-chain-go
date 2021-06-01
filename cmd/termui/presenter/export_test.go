package presenter

import "errors"

func (psh *PresenterStatusHandler) GetPresenterMetricByKey(key string) (interface{}, error) {
	psh.mutPresenterMap.RLock()
	defer psh.mutPresenterMap.RUnlock()

	value, ok := psh.presenterMetrics[key]
	if ok {
		return value, nil
	}
	return nil, errors.New("metric does not exist")
}
