package txcache

// CreateCache creates a (fail-safe) transactions cache
func CreateCache(config CacheConfig) txCache {
	cache, err := NewTxCache(config)
	if err != nil {
		log.Error("CreateCache()", "err", err)
		return NewDisabledCache()
	}

	cacheFailsafe := newTxCacheFailsafe(cache)
	return cacheFailsafe
}
