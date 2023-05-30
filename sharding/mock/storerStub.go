package mock

import (
	"github.com/multiversx/mx-chain-core-go/storage"
)

// StorerStub -
type StorerStub struct {
	PutCalled                    func(key, data []byte) error
	PutInEpochCalled             func(key, data []byte, epoch uint32) error
	GetCalled                    func(key []byte) ([]byte, error)
	HasCalled                    func(key []byte) error
	SearchFirstCalled            func(key []byte) ([]byte, error)
	RemoveFromCurrentEpochCalled func(key []byte) error
	RemoveCalled                 func(key []byte) error
	ClearCacheCalled             func()
	DestroyUnitCalled            func() error
	GetFromEpochCalled           func(key []byte, epoch uint32) ([]byte, error)
	GetBulkFromEpochCalled       func(keys [][]byte, epoch uint32) ([]storage.KeyValuePair, error)
	GetOldestEpochCalled         func() (uint32, error)
	RangeKeysCalled              func(handler func(key []byte, val []byte) bool)
	GetIdentifierCalled          func() string
	CloseCalled                  func() error
}

// Put -
func (ss *StorerStub) Put(key, data []byte) error {
	if ss.PutCalled != nil {
		return ss.PutCalled(key, data)
	}
	return nil
}

// PutInEpoch -
func (ss *StorerStub) PutInEpoch(key, data []byte, epoch uint32) error {
	if ss.PutInEpochCalled != nil {
		return ss.PutInEpochCalled(key, data, epoch)
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

// SearchFirst -
func (ss *StorerStub) SearchFirst(key []byte) ([]byte, error) {
	if ss.SearchFirstCalled != nil {
		return ss.SearchFirstCalled(key)
	}
	return nil, nil
}

// RemoveFromCurrentEpoch -
func (ss *StorerStub) RemoveFromCurrentEpoch(key []byte) error {
	if ss.RemoveFromCurrentEpochCalled != nil {
		return ss.RemoveFromCurrentEpochCalled(key)
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

// GetFromEpoch -
func (ss *StorerStub) GetFromEpoch(key []byte, epoch uint32) ([]byte, error) {
	if ss.GetFromEpochCalled != nil {
		return ss.GetFromEpochCalled(key, epoch)
	}
	return nil, nil
}

// GetBulkFromEpoch -
func (ss *StorerStub) GetBulkFromEpoch(keys [][]byte, epoch uint32) ([]storage.KeyValuePair, error) {
	if ss.GetBulkFromEpochCalled != nil {
		return ss.GetBulkFromEpochCalled(keys, epoch)
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

// GetIdentifier -
func (ss *StorerStub) GetIdentifier() string {
	if ss.GetIdentifierCalled != nil {
		return ss.GetIdentifierCalled()
	}
	return ""
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
