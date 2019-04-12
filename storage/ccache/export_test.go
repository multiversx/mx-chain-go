package ccache

func (c *CCache) AddedDataHandlers() []func(key []byte) {
	return c.addedDataHandlers
}
