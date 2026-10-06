package mock

import (
	"errors"

	"github.com/multiversx/mx-chain-go/storage"
)

// PersisterFactoryStub -
type PersisterFactoryStub struct {
	CreateCalled          func(path string) (storage.Persister, error)
	CreateDisabledCalled  func() storage.Persister
	PersisterExistsCalled func(path string) (bool, error)
}

// Create -
func (pfs *PersisterFactoryStub) Create(path string) (storage.Persister, error) {
	if pfs.CreateCalled != nil {
		return pfs.CreateCalled(path)
	}

	return nil, errors.New("not implemented")
}

// CreateDisabled -
func (pfs *PersisterFactoryStub) CreateDisabled() storage.Persister {
	if pfs.CreateDisabledCalled != nil {
		return pfs.CreateDisabledCalled()
	}
	return nil
}

// PersisterExists -
func (pfs *PersisterFactoryStub) PersisterExists(path string) (bool, error) {
	if pfs.PersisterExistsCalled != nil {
		return pfs.PersisterExistsCalled(path)
	}

	return true, nil
}

// IsInterfaceNil -
func (pfs *PersisterFactoryStub) IsInterfaceNil() bool {
	return pfs == nil
}
