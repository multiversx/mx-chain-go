package mock

// StorerStub -
type StorerStub struct {
	PutCalled              func(key, data []byte) error
	GetCalled              func(key []byte) ([]byte, error)
	HasCalled              func(key []byte) error
	RemoveCalled           func(key []byte) error
	GetFromEpochCalled     func(key []byte, epoch uint32) ([]byte, error)
	HasInEpochCalled       func(key []byte, epoch uint32) error
	ClearCacheCalled       func()
	DestroyUnitCalled      func() error
	RangeKeysCalled        func(handler func(key []byte, val []byte) bool)
	CloseCalled            func() error
	GetBulkFromEpochCalled func(keys [][]byte, epoch uint32) (map[string][]byte, error)
}

// SearchFirst -
func (ss *StorerStub) SearchFirst(_ []byte) ([]byte, error) {
	return nil, nil
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

// GetFromEpoch -
func (ss *StorerStub) GetFromEpoch(key []byte, epoch uint32) ([]byte, error) {
	if ss.GetFromEpochCalled != nil {
		return ss.GetFromEpochCalled(key, epoch)
	}

	return nil, nil
}

// GetBulkFromEpoch -
func (ss *StorerStub) GetBulkFromEpoch(keys [][]byte, epoch uint32) (map[string][]byte, error) {
	return ss.GetBulkFromEpochCalled(keys, epoch)
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

// RangeKeys -
func (ss *StorerStub) RangeKeys(handler func(key []byte, val []byte) bool) {
	if ss.RangeKeysCalled != nil {
		ss.RangeKeysCalled(handler)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ss *StorerStub) IsInterfaceNil() bool {
	return ss == nil
}
