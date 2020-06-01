package lrucache

func (c *lruCache) AddedDataHandlers() []func(key []byte, value interface{}) {
	return c.addedDataHandlers
}
