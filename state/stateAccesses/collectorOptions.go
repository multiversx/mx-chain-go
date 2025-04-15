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

// WithAccountChanges will enable collecting account changes.
// This is a struct that marks which fields of the account have changed
func WithAccountChanges() func(c *collector) {
	return func(c *collector) {
		c.withAccountChanges = true
	}
}
