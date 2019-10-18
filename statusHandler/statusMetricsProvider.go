package statusHandler

import (
	"sync"
)

// statusMetrics will handle displaying at /node/details all metrics already collected for other status handlers
type statusMetrics struct {
	nodeMetrics *sync.Map
}

// NewStatusMetrics will return an instance of the struct
func NewStatusMetrics() *statusMetrics {
	return &statusMetrics{
		nodeMetrics: &sync.Map{},
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (nd *statusMetrics) IsInterfaceNil() bool {
	if nd == nil {
		return true
	}
	return false
}

// Increment method increment a metric
func (nd *statusMetrics) Increment(key string) {
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

// AddUint64 method increase a metric with a specific value
func (nd *statusMetrics) AddUint64(key string, val uint64) {
	keyValueI, ok := nd.nodeMetrics.Load(key)
	if !ok {
		return
	}

	keyValue, ok := keyValueI.(uint64)
	if !ok {
		return
	}

	keyValue += val
	nd.nodeMetrics.Store(key, keyValue)
}

// Decrement method - decrement a metric
func (nd *statusMetrics) Decrement(key string) {
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
func (nd *statusMetrics) SetInt64Value(key string, value int64) {
	nd.nodeMetrics.Store(key, value)
}

// SetUInt64Value method - sets an uint64 value for a key
func (nd *statusMetrics) SetUInt64Value(key string, value uint64) {
	nd.nodeMetrics.Store(key, value)
}

// SetStringValue method - sets a string value for a key
func (nd *statusMetrics) SetStringValue(key string, value string) {
	nd.nodeMetrics.Store(key, value)
}

// Close method - won't do anything
func (nd *statusMetrics) Close() {
}

// StatusMetricsMap will return all metrics in a map
func (nd *statusMetrics) StatusMetricsMap() (map[string]interface{}, error) {
	statusMetricsMap := make(map[string]interface{})
	nd.nodeMetrics.Range(func(key, value interface{}) bool {
		statusMetricsMap[key.(string)] = value
		return true
	})

	return statusMetricsMap, nil
}
