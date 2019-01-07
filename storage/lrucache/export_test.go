package lrucache

func (c *LRUCache) AddedDataHandlers() []func(key []byte) {
	return c.addedDataHandlers
}
