package fifocache

func (c *FIFOShardedCache) AddedDataHandlers() map[string]func(key []byte, value interface{}) {
	return c.mapDataHandlers
}
