package spos

import "sync"

type invalidSignersCache struct {
	invalidSignersMap *sync.Map
}

// NewInvalidSignersCache returns a new instance of invalidSignersCache
func NewInvalidSignersCache() *invalidSignersCache {
	return &invalidSignersCache{
		invalidSignersMap: &sync.Map{},
	}
}

// AddInvalidSigners adds the provided hash into the internal map if it does not exist
func (cache *invalidSignersCache) AddInvalidSigners(hash string) {
	if len(hash) == 0 {
		return
	}

	cache.invalidSignersMap.Store(hash, struct{}{})
}

// HasInvalidSigners check whether the provided hash exists in the internal map or not
func (cache *invalidSignersCache) HasInvalidSigners(hash string) bool {
	_, has := cache.invalidSignersMap.Load(hash)
	return has
}

// Reset clears the internal map
func (cache *invalidSignersCache) Reset() {
	cache.invalidSignersMap = &sync.Map{}
}

// IsInterfaceNil returns true if there is no value under the interface
func (cache *invalidSignersCache) IsInterfaceNil() bool {
	return cache == nil
}
