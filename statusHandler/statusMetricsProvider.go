package statusHandler

import (
	"sync"
)

// StatusMetricsProvider will handle displaying at /node/details all metrics already collected for other status handlers
type StatusMetricsProvider struct {
	nodeMetrics *sync.Map
}

// NewStatusMetricsProvider will return an instance of the struct
func NewStatusMetricsProvider() *StatusMetricsProvider {
	return &StatusMetricsProvider{
		nodeMetrics: &sync.Map{},
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (nd *StatusMetricsProvider) IsInterfaceNil() bool {
	if nd == nil {
		return true
	}
	return false
}

// Increment method increment a metric
func (nd *StatusMetricsProvider) Increment(key string) {
	keyValueI, ok := nd.nodeMetrics.Load(key)
	if !ok {
		return
	}

	keyValue, ok := keyValueI.(uint64)
	if !ok {
		return
	}

	keyValue++
	nd.nodeMetrics.Store(key, keyValue)
}

// Decrement method - decrement a metric
func (nd *StatusMetricsProvider) Decrement(key string) {
	keyValueI, ok := nd.nodeMetrics.Load(key)
	if !ok {
		return
	}

	keyValue, ok := keyValueI.(uint64)
	if !ok {
		return
	}

	keyValue--
	nd.nodeMetrics.Store(key, keyValue)
}

// SetInt64Value method - sets an int64 value for a key
func (nd *StatusMetricsProvider) SetInt64Value(key string, value int64) {
	nd.nodeMetrics.Store(key, value)
}

// SetUInt64Value method - sets an uint64 value for a key
func (nd *StatusMetricsProvider) SetUInt64Value(key string, value uint64) {
	nd.nodeMetrics.Store(key, value)
}

// SetStringValue method - sets a string value for a key
func (nd *StatusMetricsProvider) SetStringValue(key string, value string) {
	nd.nodeMetrics.Store(key, value)
}

// Close method - won't do anything
func (nd *StatusMetricsProvider) Close() {
}

// StatusMetricsMap will return all metrics in a map
func (nd *StatusMetricsProvider) StatusMetricsMap() (map[string]interface{}, error) {
	statusMetricsMap := make(map[string]interface{})
	nd.nodeMetrics.Range(func(key, value interface{}) bool {
		statusMetricsMap[key.(string)] = value
		return true
	})

	return statusMetricsMap, nil
}
