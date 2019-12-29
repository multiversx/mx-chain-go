package mock

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go/storage"
)

type PersisterFactoryStub struct {
	CreateCalled func(path string) (storage.Persister, error)
}

func (pfs *PersisterFactoryStub) Create(path string) (storage.Persister, error) {
	if pfs.CreateCalled != nil {
		return pfs.CreateCalled(path)
	}

	return nil, errors.New("not implemented")
}

func (pfs *PersisterFactoryStub) IsInterfaceNil() bool {
	return pfs == nil
}
