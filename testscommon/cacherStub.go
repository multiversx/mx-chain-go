package testscommon

// CacherStub -
type CacherStub struct {
	ClearCalled             func()
	PutCalled               func(key []byte, value interface{}, sizeInBytes int) (evicted bool)
	GetCalled               func(key []byte) (value interface{}, ok bool)
	HasCalled               func(key []byte) bool
	PeekCalled              func(key []byte) (value interface{}, ok bool)
	HasOrAddCalled          func(key []byte, value interface{}, sizeInBytes int) (has, added bool)
	RemoveCalled            func(key []byte)
	RemoveOldestCalled      func()
	KeysCalled              func() [][]byte
	LenCalled               func() int
	MaxSizeCalled           func() int
	RegisterHandlerCalled   func(func(key []byte, value interface{}))
	UnRegisterHandlerCalled func(id string)
}

// NewCacherStub -
func NewCacherStub() *CacherStub {
	return &CacherStub{}
}

// Clear -
func (cacher *CacherStub) Clear() {
	if cacher.ClearCalled != nil {
		cacher.ClearCalled()
	}
}

// Put -
func (cacher *CacherStub) Put(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
	if cacher.PutCalled != nil {
		return cacher.PutCalled(key, value, sizeInBytes)
	}

	return false
}

// Get -
func (cacher *CacherStub) Get(key []byte) (value interface{}, ok bool) {
	if cacher.GetCalled != nil {
		return cacher.GetCalled(key)
	}

	return nil, false
}

// Has -
func (cacher *CacherStub) Has(key []byte) bool {
	if cacher.HasCalled != nil {
		return cacher.HasCalled(key)
	}

	return false
}

// Peek -
func (cacher *CacherStub) Peek(key []byte) (value interface{}, ok bool) {
	if cacher.PeekCalled != nil {
		return cacher.PeekCalled(key)
	}

	return nil, false
}

// HasOrAdd -
func (cacher *CacherStub) HasOrAdd(key []byte, value interface{}, sizeInBytes int) (has, added bool) {
	if cacher.HasOrAddCalled != nil {
		return cacher.HasOrAddCalled(key, value, sizeInBytes)
	}

	return false, false
}

// Remove -
func (cacher *CacherStub) Remove(key []byte) {
	if cacher.RemoveCalled != nil {
		cacher.RemoveCalled(key)
	}
}

// Keys -
func (cacher *CacherStub) Keys() [][]byte {
	if cacher.KeysCalled != nil {
		return cacher.KeysCalled()
	}

	return make([][]byte, 0)
}

// Len -
func (cacher *CacherStub) Len() int {
	if cacher.LenCalled != nil {
		return cacher.LenCalled()
	}

	return 0
}

// MaxSize -
func (cacher *CacherStub) MaxSize() int {
	if cacher.MaxSizeCalled != nil {
		return cacher.MaxSizeCalled()
	}

	return 0
}

// RegisterHandler -
func (cacher *CacherStub) RegisterHandler(handler func(key []byte, value interface{}), _ string) {
	if cacher.RegisterHandlerCalled != nil {
		cacher.RegisterHandlerCalled(handler)
	}
}

// UnRegisterHandler -
func (cacher *CacherStub) UnRegisterHandler(id string) {
	if cacher.UnRegisterHandlerCalled != nil {
		cacher.UnRegisterHandlerCalled(id)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (cacher *CacherStub) IsInterfaceNil() bool {
	return cacher == nil
}
