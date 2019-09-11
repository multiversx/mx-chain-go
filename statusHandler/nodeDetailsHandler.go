package statusHandler

import (
	"sync"
)

// NodeDetailsHandler will handle displaying at /node/details all metrics already collected for other status handlers
type NodeDetailsHandler struct {
	nodeMetrics *sync.Map
}

// NewNodeDetailsHandler will return an instance of the struct
func NewNodeDetailsHandler() *NodeDetailsHandler {
	return &NodeDetailsHandler{
		nodeMetrics: &sync.Map{},
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (nd *NodeDetailsHandler) IsInterfaceNil() bool {
	if nd == nil {
		return true
	}
	return false
}

// Increment method increment a metric
func (nd *NodeDetailsHandler) Increment(key string) {
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
func (nd *NodeDetailsHandler) Decrement(key string) {
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
func (nd *NodeDetailsHandler) SetInt64Value(key string, value int64) {
	nd.nodeMetrics.Store(key, value)
}

// SetUInt64Value method - sets an uint64 value for a key
func (nd *NodeDetailsHandler) SetUInt64Value(key string, value uint64) {
	nd.nodeMetrics.Store(key, value)
}

// SetStringValue method - sets a string value for a key
func (nd *NodeDetailsHandler) SetStringValue(key string, value string) {
	nd.nodeMetrics.Store(key, value)
}

// Close method - won't do anything
func (nd *NodeDetailsHandler) Close() {
}

// DetailsMap will return all metrics in a map
func (nd *NodeDetailsHandler) DetailsMap() (map[string]interface{}, error) {
	nodeDetailsMap := make(map[string]interface{})
	nd.nodeMetrics.Range(func(key, value interface{}) bool {
		nodeDetailsMap[key.(string)] = value
		return true
	})

	return nodeDetailsMap, nil
}
