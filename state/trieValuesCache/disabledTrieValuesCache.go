package trieValuesCache

import "github.com/multiversx/mx-chain-core-go/core"

type disabledTrieValuesCache struct {
}

// NewDisabledTrieValuesCache creates a new instance of disabledTrieValuesCache
func NewDisabledTrieValuesCache() *disabledTrieValuesCache {
	return &disabledTrieValuesCache{}
}

// Get returns an empty struct and false
func (d *disabledTrieValuesCache) Get(_ []byte) (core.TrieData, bool) {
	return core.TrieData{}, false
}

// Put does nothing for this implementation
func (d *disabledTrieValuesCache) Put(_ []byte, _ core.TrieData) {
}

// Clean does nothing for this implementation
func (d *disabledTrieValuesCache) Clean() {
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledTrieValuesCache) IsInterfaceNil() bool {
	return d == nil
}
