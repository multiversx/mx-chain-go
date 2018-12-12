package mock

type LRUCacheStub struct {
	ClearCalled           func()
	PutCalled             func(key []byte, value interface{}) (evicted bool)
	GetCalled             func(key []byte) (value interface{}, ok bool)
	HasCalled             func(key []byte) bool
	PeekCalled            func(key []byte) (value interface{}, ok bool)
	HasOrAddCalled        func(key []byte, value interface{}) (ok, evicted bool)
	RemoveCalled          func(key []byte)
	RemoveOldestCalled    func()
	KeysCalled            func() [][]byte
	LenCalled             func() int
	RegisterHandlerCalled func(func(key []byte))
}

func (stub *LRUCacheStub) Clear() {
	stub.ClearCalled()
}

func (stub *LRUCacheStub) Put(key []byte, value interface{}) (evicted bool) {
	return stub.PutCalled(key, value)
}

func (stub *LRUCacheStub) Get(key []byte) (value interface{}, ok bool) {
	return stub.GetCalled(key)
}

func (stub *LRUCacheStub) Has(key []byte) bool {
	return stub.HasCalled(key)
}

func (stub *LRUCacheStub) Peek(key []byte) (value interface{}, ok bool) {
	return stub.PeekCalled(key)
}

func (stub *LRUCacheStub) HasOrAdd(key []byte, value interface{}) (ok, evicted bool) {
	return stub.HasOrAddCalled(key, value)
}

func (stub *LRUCacheStub) Remove(key []byte) {
	stub.RemoveCalled(key)
}

func (stub *LRUCacheStub) RemoveOldest() {
	stub.RemoveOldestCalled()
}

func (stub *LRUCacheStub) Keys() [][]byte {
	return stub.KeysCalled()
}

func (stub *LRUCacheStub) Len() int {
	return stub.LenCalled()
}

func (stub *LRUCacheStub) RegisterHandler(handler func(key []byte)) {
	stub.RegisterHandlerCalled(handler)
}
