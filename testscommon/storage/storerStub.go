package storage

import "github.com/ElrondNetwork/elrond-go/common"

// StorerStub -
type StorerStub struct {
	PutCalled                    func(key, data []byte, priority common.StorageAccessType) error
	PutInEpochCalled             func(key, data []byte, epoch uint32, priority common.StorageAccessType) error
	GetCalled                    func(key []byte, priority common.StorageAccessType) ([]byte, error)
	HasCalled                    func(key []byte, priority common.StorageAccessType) error
	SearchFirstCalled            func(key []byte, priority common.StorageAccessType) ([]byte, error)
	RemoveFromCurrentEpochCalled func(key []byte, priority common.StorageAccessType) error
	RemoveCalled                 func(key []byte, priority common.StorageAccessType) error
	ClearCacheCalled             func()
	DestroyUnitCalled            func() error
	GetFromEpochCalled           func(key []byte, epoch uint32, priority common.StorageAccessType) ([]byte, error)
	GetBulkFromEpochCalled       func(keys [][]byte, epoch uint32, priority common.StorageAccessType) (map[string][]byte, error)
	GetOldestEpochCalled         func() (uint32, error)
	RangeKeysCalled              func(handler func(key []byte, val []byte) bool)
	CloseCalled                  func() error
}

// Put -
func (ss *StorerStub) Put(key, data []byte, priority common.StorageAccessType) error {
	if ss.PutCalled != nil {
		return ss.PutCalled(key, data, priority)
	}
	return nil
}

// PutInEpoch -
func (ss *StorerStub) PutInEpoch(key, data []byte, epoch uint32, priority common.StorageAccessType) error {
	if ss.PutInEpochCalled != nil {
		return ss.PutInEpochCalled(key, data, epoch, priority)
	}
	return nil
}

// Get -
func (ss *StorerStub) Get(key []byte, priority common.StorageAccessType) ([]byte, error) {
	if ss.GetCalled != nil {
		return ss.GetCalled(key, priority)
	}
	return nil, nil
}

// Has -
func (ss *StorerStub) Has(key []byte, priority common.StorageAccessType) error {
	if ss.HasCalled != nil {
		return ss.HasCalled(key, priority)
	}
	return nil
}

// SearchFirst -
func (ss *StorerStub) SearchFirst(key []byte, priority common.StorageAccessType) ([]byte, error) {
	if ss.SearchFirstCalled != nil {
		return ss.SearchFirstCalled(key, priority)
	}
	return nil, nil
}

// RemoveFromCurrentEpoch -
func (ss *StorerStub) RemoveFromCurrentEpoch(key []byte, priority common.StorageAccessType) error {
	if ss.RemoveFromCurrentEpochCalled != nil {
		return ss.RemoveFromCurrentEpochCalled(key, priority)
	}
	return nil
}

// Remove -
func (ss *StorerStub) Remove(key []byte, priority common.StorageAccessType) error {
	if ss.RemoveCalled != nil {
		return ss.RemoveCalled(key, priority)
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

// GetFromEpoch -
func (ss *StorerStub) GetFromEpoch(key []byte, epoch uint32, priority common.StorageAccessType) ([]byte, error) {
	if ss.GetFromEpochCalled != nil {
		return ss.GetFromEpochCalled(key, epoch, priority)
	}
	return nil, nil
}

// GetBulkFromEpoch -
func (ss *StorerStub) GetBulkFromEpoch(keys [][]byte, epoch uint32, priority common.StorageAccessType) (map[string][]byte, error) {
	if ss.GetBulkFromEpochCalled != nil {
		return ss.GetBulkFromEpochCalled(keys, epoch, priority)
	}
	return nil, nil
}

// GetOldestEpoch -
func (ss *StorerStub) GetOldestEpoch() (uint32, error) {
	if ss.GetOldestEpochCalled != nil {
		return ss.GetOldestEpochCalled()
	}
	return 0, nil
}

// RangeKeys -
func (ss *StorerStub) RangeKeys(handler func(key []byte, val []byte) bool) {
	if ss.RangeKeysCalled != nil {
		ss.RangeKeysCalled(handler)
	}
}

// Close -
func (ss *StorerStub) Close() error {
	if ss.CloseCalled != nil {
		return ss.CloseCalled()
	}
	return nil
}

// IsInterfaceNil -
func (ss *StorerStub) IsInterfaceNil() bool {
	return ss == nil
}
