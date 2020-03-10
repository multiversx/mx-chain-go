package mock

// NodesCoordinatorCacheStub -
type NodesCoordinatorCacheStub struct {
	PutCalled func(key []byte, value interface{}) (evicted bool)
	GetCalled func(key []byte) (value interface{}, ok bool)
}

// Put -
func (rm *NodesCoordinatorCacheStub) Put(key []byte, value interface{}) (evicted bool) {
	if rm.PutCalled != nil {
		return rm.PutCalled(key, value)
	}
	return false
}

// Get -
func (rm *NodesCoordinatorCacheStub) Get(key []byte) (value interface{}, ok bool) {
	if rm.GetCalled != nil {
		return rm.GetCalled(key)
	}
	return nil, false
}
