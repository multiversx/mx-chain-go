package mock

// CacherStub -
type CacherStub struct {
	ClearCalled             func()
	PutCalled               func(key []byte, value interface{}) (evicted bool)
	GetCalled               func(key []byte) (value interface{}, ok bool)
	HasCalled               func(key []byte) bool
	PeekCalled              func(key []byte) (value interface{}, ok bool)
	HasOrAddCalled          func(key []byte, value interface{}) (ok, evicted bool)
	RemoveCalled            func(key []byte)
	RemoveOldestCalled      func()
	KeysCalled              func() [][]byte
	LenCalled               func() int
	MaxSizeCalled           func() int
	RegisterHandlerCalled   func(func(key []byte, value interface{}))
	UnRegisterHandlerCalled func(func(key []byte, value interface{}))
}

// UnRegisterHandler -
func (cs *CacherStub) UnRegisterHandler(handler func(key []byte, value interface{})) {
	cs.UnRegisterHandlerCalled(handler)
}

// Clear -
func (cs *CacherStub) Clear() {
	cs.ClearCalled()
}

// Put -
func (cs *CacherStub) Put(key []byte, value interface{}) (evicted bool) {
	return cs.PutCalled(key, value)
}

// Get -
func (cs *CacherStub) Get(key []byte) (value interface{}, ok bool) {
	return cs.GetCalled(key)
}

// Has -
func (cs *CacherStub) Has(key []byte) bool {
	return cs.HasCalled(key)
}

// Peek -
func (cs *CacherStub) Peek(key []byte) (value interface{}, ok bool) {
	return cs.PeekCalled(key)
}

// HasOrAdd -
func (cs *CacherStub) HasOrAdd(key []byte, value interface{}) (ok, evicted bool) {
	return cs.HasOrAddCalled(key, value)
}

// Remove -
func (cs *CacherStub) Remove(key []byte) {
	cs.RemoveCalled(key)
}

// RemoveOldest -
func (cs *CacherStub) RemoveOldest() {
	cs.RemoveOldestCalled()
}

// Keys -
func (cs *CacherStub) Keys() [][]byte {
	return cs.KeysCalled()
}

// Len -
func (cs *CacherStub) Len() int {
	return cs.LenCalled()
}

// MaxSize -
func (cs *CacherStub) MaxSize() int {
	return cs.MaxSizeCalled()
}

// RegisterHandler -
func (cs *CacherStub) RegisterHandler(handler func(key []byte, value interface{})) {
	cs.RegisterHandlerCalled(handler)
}

// IsInterfaceNil returns true if there is no value under the interface
func (cs *CacherStub) IsInterfaceNil() bool {
	return cs == nil
}
