package mock

type NodesCoordinatorCacheMock struct {
	PutCalled func(key []byte, value interface{}) (evicted bool)
	GetCalled func(key []byte) (value interface{}, ok bool)
}

func (rm *NodesCoordinatorCacheMock) Put(key []byte, value interface{}) (evicted bool) {
	if rm.PutCalled != nil {
		return rm.PutCalled(key, value)
	}
	return false
}

func (rm *NodesCoordinatorCacheMock) Get(key []byte) (value interface{}, ok bool) {
	if rm.GetCalled != nil {
		return rm.GetCalled(key)
	}
	return nil, false
}
