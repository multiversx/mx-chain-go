package mock

type StorerStub struct {
	PutCalled         func(key, data []byte) error
	GetCalled         func(key []byte) ([]byte, error)
	HasCalled         func(key []byte) (bool, error)
	HasOrAddCalled    func(key []byte, value []byte) (bool, error)
	RemoveCalled      func(key []byte) error
	ClearCacheCalled  func()
	DestroyUnitCalled func() error
}

func (ss *StorerStub) Put(key, data []byte) error {
	return ss.PutCalled(key, data)
}

func (ss *StorerStub) Get(key []byte) ([]byte, error) {
	return ss.GetCalled(key)
}

func (ss *StorerStub) Has(key []byte) (bool, error) {
	return ss.HasCalled(key)
}

func (ss *StorerStub) HasOrAdd(key []byte, value []byte) (bool, error) {
	return ss.HasOrAddCalled(key, value)
}

func (ss *StorerStub) Remove(key []byte) error {
	return ss.RemoveCalled(key)
}

func (ss *StorerStub) ClearCache() {
	ss.ClearCacheCalled()
}

func (ss *StorerStub) DestroyUnit() error {
	return ss.DestroyUnitCalled()
}
