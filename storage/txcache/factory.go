package txcache

// CreateCache creates a transactions cache
func CreateCache(config CacheConfig) txCache {
	cache, err := NewTxCache(config)
	if err != nil {
		log.Error("CreateCache()", "err", err)
		return NewDisabledCache()
	}

	return cache
}
