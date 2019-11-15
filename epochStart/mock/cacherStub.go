package mock

type CacherStub struct {
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
	MaxSizeCalled         func() int
	RegisterHandlerCalled func(func(key []byte))
}

func (cs *CacherStub) Clear() {
	cs.ClearCalled()
}

func (cs *CacherStub) Put(key []byte, value interface{}) (evicted bool) {
	return cs.PutCalled(key, value)
}

func (cs *CacherStub) Get(key []byte) (value interface{}, ok bool) {
	return cs.GetCalled(key)
}

func (cs *CacherStub) Has(key []byte) bool {
	return cs.HasCalled(key)
}

func (cs *CacherStub) Peek(key []byte) (value interface{}, ok bool) {
	return cs.PeekCalled(key)
}

func (cs *CacherStub) HasOrAdd(key []byte, value interface{}) (ok, evicted bool) {
	return cs.HasOrAddCalled(key, value)
}

func (cs *CacherStub) Remove(key []byte) {
	cs.RemoveCalled(key)
}

func (cs *CacherStub) RemoveOldest() {
	cs.RemoveOldestCalled()
}

func (cs *CacherStub) Keys() [][]byte {
	return cs.KeysCalled()
}

func (cs *CacherStub) Len() int {
	return cs.LenCalled()
}

func (cs *CacherStub) MaxSize() int {
	return cs.MaxSizeCalled()
}

func (cs *CacherStub) RegisterHandler(handler func(key []byte)) {
	cs.RegisterHandlerCalled(handler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (cs *CacherStub) IsInterfaceNil() bool {
	if cs == nil {
		return true
	}
	return false
}
