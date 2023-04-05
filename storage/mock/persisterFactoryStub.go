package mock

import (
	"errors"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
)

// PersisterFactoryStub -
type PersisterFactoryStub struct {
	CreateCalled              func(path string) (storage.Persister, error)
	CreateDisabledCalled      func() storage.Persister
	DBConfigWithoutPathCalled func() config.DBConfig
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

// DBConfigWithoutPath -
func (pfs *PersisterFactoryStub) DBConfigWithoutPath() config.DBConfig {
	if pfs.DBConfigWithoutPathCalled != nil {
		return pfs.DBConfigWithoutPathCalled()
	}

	return config.DBConfig{}
}

// IsInterfaceNil -
func (pfs *PersisterFactoryStub) IsInterfaceNil() bool {
	return pfs == nil
}
