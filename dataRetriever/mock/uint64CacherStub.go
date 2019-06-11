package mock

type Uint64CacherStub struct {
	ClearCalled           func()
	PutCalled             func(uint64, interface{}) bool
	GetCalled             func(uint64) (interface{}, bool)
	HasCalled             func(uint64) bool
	PeekCalled            func(uint64) (interface{}, bool)
	HasOrAddCalled        func(uint64, interface{}) (bool, bool)
	RemoveCalled          func(uint64)
	RemoveOldestCalled    func()
	KeysCalled            func() []uint64
	LenCalled             func() int
	RegisterHandlerCalled func(handler func(nonce uint64))
}

func (ucs *Uint64CacherStub) Clear() {
	ucs.ClearCalled()
}

func (ucs *Uint64CacherStub) Put(nonce uint64, value interface{}) bool {
	return ucs.PutCalled(nonce, value)
}

func (ucs *Uint64CacherStub) Get(nonce uint64) (interface{}, bool) {
	return ucs.GetCalled(nonce)
}

func (ucs *Uint64CacherStub) Has(nonce uint64) bool {
	return ucs.HasCalled(nonce)
}

func (ucs *Uint64CacherStub) Peek(nonce uint64) (interface{}, bool) {
	return ucs.PeekCalled(nonce)
}

func (ucs *Uint64CacherStub) HasOrAdd(nonce uint64, value interface{}) (bool, bool) {
	return ucs.HasOrAddCalled(nonce, value)
}

func (ucs *Uint64CacherStub) Remove(nonce uint64) {
	ucs.RemoveCalled(nonce)
}

func (ucs *Uint64CacherStub) RemoveOldest() {
	ucs.RemoveOldestCalled()
}

func (ucs *Uint64CacherStub) Keys() []uint64 {
	return ucs.KeysCalled()
}

func (ucs *Uint64CacherStub) Len() int {
	return ucs.LenCalled()
}

func (ucs *Uint64CacherStub) RegisterHandler(handler func(nonce uint64)) {
	ucs.RegisterHandlerCalled(handler)
}
