package fifocache

func (c *FIFOShardedCache) AddedDataHandlers() []func(key []byte) {
	return c.addedDataHandlers
}
