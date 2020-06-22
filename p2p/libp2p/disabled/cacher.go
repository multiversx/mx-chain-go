package disabled

// Cacher is a mock implementation of p2p.Cacher interface
type Cacher struct {
}

// HasOrAdd does nothing and returns (false, false)
func (c *Cacher) HasOrAdd(_ []byte, _ interface{}, _ int) (has, added bool) {
	return false, false
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *Cacher) IsInterfaceNil() bool {
	return c == nil
}
