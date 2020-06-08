package disabled

// Cacher is a mock implementation of p2p.Cacher interface
type Cacher struct {
}

// HasOrAdd does nothing and returns true
func (c *Cacher) HasOrAdd(_ []byte, _ interface{}, _ int) (added bool) {
	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *Cacher) IsInterfaceNil() bool {
	return c == nil
}
