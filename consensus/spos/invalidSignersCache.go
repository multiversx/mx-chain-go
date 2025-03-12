package spos

import "sync"

type invalidSignersCache struct {
	sync.RWMutex
	invalidSignersMap map[string]struct{}
}

// NewInvalidSignersCache returns a new instance of invalidSignersCache
func NewInvalidSignersCache() *invalidSignersCache {
	return &invalidSignersCache{
		invalidSignersMap: make(map[string]struct{}),
	}
}

// AddInvalidSigners adds the provided hash into the internal map if it does not exist
func (cache *invalidSignersCache) AddInvalidSigners(hash string) {
	if len(hash) == 0 {
		return
	}

	cache.Lock()
	defer cache.Unlock()

	cache.invalidSignersMap[hash] = struct{}{}
}

// HasInvalidSigners check whether the provided hash exists in int internal map or not
func (cache *invalidSignersCache) HasInvalidSigners(hash string) bool {
	cache.RLock()
	defer cache.RUnlock()

	_, has := cache.invalidSignersMap[hash]
	return has
}

// Reset clears the internal map
func (cache *invalidSignersCache) Reset() {
	cache.Lock()
	defer cache.Unlock()

	cache.invalidSignersMap = make(map[string]struct{})
}

// IsInterfaceNil returns true if there is no value under the interface
func (cache *invalidSignersCache) IsInterfaceNil() bool {
	return cache == nil
}
