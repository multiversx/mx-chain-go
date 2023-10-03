package disabled

type cache struct {
}

// NewCache returns a new disabled Cacher implementation
func NewCache() *cache {
	return &cache{}
}

// Clear does nothing as it is disabled
func (c *cache) Clear() {
}

// Put returns false as it is disabled
func (c *cache) Put(_ []byte, _ interface{}, _ int) (evicted bool) {
	return false
}

// Get returns nil and false as it is disabled
func (c *cache) Get(_ []byte) (value interface{}, ok bool) {
	return nil, false
}

// Has returns false as it is disabled
func (c *cache) Has(_ []byte) bool {
	return false
}

// Peek returns nil and false as it is disabled
func (c *cache) Peek(_ []byte) (value interface{}, ok bool) {
	return nil, false
}

// HasOrAdd returns false and false as it is disabled
func (c *cache) HasOrAdd(_ []byte, _ interface{}, _ int) (has, added bool) {
	return false, false
}

// Remove does nothing as it is disabled
func (c *cache) Remove(_ []byte) {
}

// Keys returns an empty slice as it is disabled
func (c *cache) Keys() [][]byte {
	return make([][]byte, 0)
}

// Len returns 0 as it is disabled
func (c *cache) Len() int {
	return 0
}

// SizeInBytesContained returns 0 as it is disabled
func (c *cache) SizeInBytesContained() uint64 {
	return 0
}

// MaxSize returns 0 as it is disabled
func (c *cache) MaxSize() int {
	return 0
}

// RegisterHandler does nothing as it is disabled
func (c *cache) RegisterHandler(_ func(key []byte, value interface{}), _ string) {
}

// UnRegisterHandler does nothing as it is disabled
func (c *cache) UnRegisterHandler(_ string) {
}

// Close returns nil as it is disabled
func (c *cache) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *cache) IsInterfaceNil() bool {
	return c == nil
}
