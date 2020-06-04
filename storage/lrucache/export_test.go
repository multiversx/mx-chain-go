package lrucache

func (c *lruCache) AddedDataHandlers() map[string]func(key []byte, value interface{}) {
	return c.mapDataHandlers
}
