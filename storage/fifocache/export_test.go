package fifocache

func (c *FIFOShardedCache) AddedDataHandlers() []func(key []byte, value interface{}) {
	return c.addedDataHandlers
}
