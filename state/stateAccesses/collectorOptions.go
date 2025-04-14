package stateAccesses

// CollectorOption specifies the possible options for the collector
type CollectorOption func(*collector)

// WithCollectRead will enable collecting read action types
func WithCollectRead() func(c *collector) {
	return func(c *collector) {
		c.collectRead = true
	}
}

// WithCollectWrite will enable collecting write action types
func WithCollectWrite() func(c *collector) {
	return func(c *collector) {
		c.collectWrite = true
	}
}
