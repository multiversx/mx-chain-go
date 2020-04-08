package mock

// StorerStub -
type StorerStub struct {
	PutCalled          func(key, data []byte) error
	GetCalled          func(key []byte) ([]byte, error)
	HasCalled          func(key []byte) error
	RemoveCalled       func(key []byte) error
	GetFromEpochCalled func(key []byte, epoch uint32) ([]byte, error)
	HasInEpochCalled   func(key []byte, epoch uint32) error
	ClearCacheCalled   func()
	DestroyUnitCalled  func() error
}

// SearchFirst -
func (ss *StorerStub) SearchFirst(_ []byte) ([]byte, error) {
	return nil, nil
}

// Close -
func (ss *StorerStub) Close() error {
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

// GetFromEpoch -
func (ss *StorerStub) GetFromEpoch(key []byte, epoch uint32) ([]byte, error) {
	if ss.GetFromEpochCalled != nil {
		return ss.GetFromEpochCalled(key, epoch)
	}

	return nil, nil
}

// HasInEpoch -
func (ss *StorerStub) HasInEpoch(key []byte, epoch uint32) error {
	if ss.HasInEpochCalled != nil {
		return ss.HasInEpochCalled(key, epoch)
	}
	return nil
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
