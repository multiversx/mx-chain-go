package mock

// StorerStub -
type StorerStub struct {
	PutCalled          func(key, data []byte) error
	GetCalled          func(key []byte) ([]byte, error)
	GetFromEpochCalled func(key []byte, epoch uint32) ([]byte, error)
	HasCalled          func(key []byte) error
	HasInEpochCalled   func(key []byte, epoch uint32) error
	SearchFirstCalled  func(key []byte) ([]byte, error)
	RemoveCalled       func(key []byte) error
	ClearCacheCalled   func()
	DestroyUnitCalled  func() error
	CloseCalled        func() error
}

// GetFromEpoch -
func (ss *StorerStub) GetFromEpoch(key []byte, epoch uint32) ([]byte, error) {
	return ss.GetFromEpochCalled(key, epoch)
}

// HasInEpoch -
func (ss *StorerStub) HasInEpoch(key []byte, epoch uint32) error {
	return ss.HasInEpochCalled(key, epoch)
}

// SearchFirst -
func (ss *StorerStub) SearchFirst(key []byte) ([]byte, error) {
	return ss.SearchFirstCalled(key)
}

// Close -
func (ss *StorerStub) Close() error {
	if ss.CloseCalled != nil {
		return ss.CloseCalled()
	}
	return nil
}

// Put -
func (ss *StorerStub) Put(key, data []byte) error {
	return ss.PutCalled(key, data)
}

// Get -
func (ss *StorerStub) Get(key []byte) ([]byte, error) {
	return ss.GetCalled(key)
}

// Has -
func (ss *StorerStub) Has(key []byte) error {
	return ss.HasCalled(key)
}

// Remove -
func (ss *StorerStub) Remove(key []byte) error {
	return ss.RemoveCalled(key)
}

// ClearCache -
func (ss *StorerStub) ClearCache() {
	ss.ClearCacheCalled()
}

// DestroyUnit -
func (ss *StorerStub) DestroyUnit() error {
	return ss.DestroyUnitCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ss *StorerStub) IsInterfaceNil() bool {
	return ss == nil
}
