package mock

// NodesCoordinatorCacheMock -
type NodesCoordinatorCacheMock struct {
	ClearCalled func()
	PutCalled   func(key []byte, value interface{}, sieInBytes int) (evicted bool)
	GetCalled   func(key []byte) (value interface{}, ok bool)
}

// Clear -
func (rm *NodesCoordinatorCacheMock) Clear() {
	if rm.ClearCalled != nil {
		rm.ClearCalled()
	}
}

// Put -
func (rm *NodesCoordinatorCacheMock) Put(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
	if rm.PutCalled != nil {
		return rm.PutCalled(key, value, sizeInBytes)
	}
	return false
}

// Get -
func (rm *NodesCoordinatorCacheMock) Get(key []byte) (value interface{}, ok bool) {
	if rm.GetCalled != nil {
		return rm.GetCalled(key)
	}
	return nil, false
}
