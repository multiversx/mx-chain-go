package mock

// StorerStub -
type StorerStub struct {
	PutCalled              func(key, data []byte) error
	GetCalled              func(key []byte) ([]byte, error)
	GetFromEpochCalled     func(key []byte, epoch uint32) ([]byte, error)
	GetBulkFromEpochCalled func(keys [][]byte, epoch uint32) (map[string][]byte, error)
	HasCalled              func(key []byte) error
	HasInEpochCalled       func(key []byte, epoch uint32) error
	SearchFirstCalled      func(key []byte) ([]byte, error)
	RemoveCalled           func(key []byte) error
	ClearCacheCalled       func()
	DestroyUnitCalled      func() error
	RangeKeysCalled        func(handler func(key []byte, val []byte) bool)
	PutInEpochCalled       func(key, data []byte, epoch uint32) error
	GetOldestEpochCalled   func() (uint32, error)
	CloseCalled            func() error
}

// PutInEpoch -
func (ss *StorerStub) PutInEpoch(key, data []byte, epoch uint32) error {
	if ss.PutInEpochCalled != nil {
		return ss.PutInEpochCalled(key, data, epoch)
	}

	return nil
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
	if ss.GetBulkFromEpochCalled != nil {
		return ss.GetBulkFromEpochCalled(keys, epoch)
	}

	return nil, nil
}

// SearchFirst -
func (ss *StorerStub) SearchFirst(key []byte) ([]byte, error) {
	if ss.SearchFirstCalled != nil {
		return ss.SearchFirstCalled(key)
	}

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
	if ss.PutCalled != nil {
		return ss.PutCalled(key, data)
	}

	return nil
}

// Get -
func (ss *StorerStub) Get(key []byte) ([]byte, error) {
	if ss.GetCalled != nil {
		return ss.GetCalled(key)
	}

	return nil, nil
}

// Has -
func (ss *StorerStub) Has(key []byte) error {
	if ss.HasCalled != nil {
		return ss.HasCalled(key)
	}

	return nil
}

// Remove -
func (ss *StorerStub) Remove(key []byte) error {
	if ss.RemoveCalled != nil {
		return ss.RemoveCalled(key)
	}

	return nil
}

// ClearCache -
func (ss *StorerStub) ClearCache() {
	if ss.ClearCacheCalled != nil {
		ss.ClearCacheCalled()
	}
}

// DestroyUnit -
func (ss *StorerStub) DestroyUnit() error {
	if ss.DestroyUnitCalled != nil {
		return ss.DestroyUnitCalled()
	}

	return nil
}

// RangeKeys -
func (ss *StorerStub) RangeKeys(handler func(key []byte, val []byte) bool) {
	if ss.RangeKeysCalled != nil {
		ss.RangeKeysCalled(handler)
	}
}

// GetOldestEpoch -
func (ss *StorerStub) GetOldestEpoch() (uint32, error) {
	if ss.GetOldestEpochCalled != nil {
		return ss.GetOldestEpochCalled()
	}

	return 0, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ss *StorerStub) IsInterfaceNil() bool {
	return ss == nil
}
