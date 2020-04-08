package lrucache

func (c *LRUCache) AddedDataHandlers() []func(key []byte, value interface{}) {
	return c.addedDataHandlers
}
