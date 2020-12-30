//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. metrics.proto
package metrics

// ListFromMap returns a list of values from the provided map
func ListFromMap(metricsMap map[string]interface{}) *MetricsList {
	r := &MetricsList{
		Metrics: make([]Metric, 0, len(metricsMap)),
	}
	for key, value := range metricsMap {
		m := Metric{
			Key: key,
		}
		switch v := value.(type) {
		case uint64:
			m.Value = &Metric_ValUint64{ValUint64: v}
		case string:
			m.Value = &Metric_ValString{ValString: v}
		default:
			continue
		}
		r.Metrics = append(r.Metrics, m)
	}
	return r
}

// MapFromList returns a map with the values from the list
func MapFromList(l *MetricsList) map[string]interface{} {
	ret := make(map[string]interface{}, len(l.Metrics))
	for _, m := range l.Metrics {
		switch v := m.Value.(type) {
		case *Metric_ValUint64:
			ret[m.Key] = v.ValUint64
		case *Metric_ValString:
			ret[m.Key] = v.ValString
		default:
			continue
		}
	}
	return ret
}
