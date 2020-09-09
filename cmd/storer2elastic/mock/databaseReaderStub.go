package mock

import (
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/databasereader"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// DatabaseReaderStub -
type DatabaseReaderStub struct {
	GetDatabaseInfoCalled       func() ([]*databasereader.DatabaseInfo, error)
	GetStaticDatabaseInfoCalled func() ([]*databasereader.DatabaseInfo, error)
	GetHeadersCalled            func(dbInfo *databasereader.DatabaseInfo) ([]data.HeaderHandler, error)
	LoadPersisterCalled         func(dbInfo *databasereader.DatabaseInfo, unit string) (storage.Persister, error)
	LoadStaticPersisterCalled   func(dbInfo *databasereader.DatabaseInfo, unit string) (storage.Persister, error)
}

// GetDatabaseInfo -
func (d *DatabaseReaderStub) GetDatabaseInfo() ([]*databasereader.DatabaseInfo, error) {
	if d.GetDatabaseInfoCalled != nil {
		return d.GetDatabaseInfoCalled()
	}

	return nil, nil
}

// GetStaticDatabaseInfo -
func (d *DatabaseReaderStub) GetStaticDatabaseInfo() ([]*databasereader.DatabaseInfo, error) {
	if d.GetStaticDatabaseInfoCalled != nil {
		return d.GetStaticDatabaseInfoCalled()
	}

	return nil, nil
}

// GetHeaders -
func (d *DatabaseReaderStub) GetHeaders(dbInfo *databasereader.DatabaseInfo) ([]data.HeaderHandler, error) {
	if d.GetHeadersCalled != nil {
		return d.GetHeadersCalled(dbInfo)
	}

	return nil, nil
}

// LoadPersister -
func (d *DatabaseReaderStub) LoadPersister(dbInfo *databasereader.DatabaseInfo, unit string) (storage.Persister, error) {
	if d.LoadPersisterCalled != nil {
		return d.LoadPersisterCalled(dbInfo, unit)
	}

	return nil, nil
}

// LoadStaticPersister -
func (d *DatabaseReaderStub) LoadStaticPersister(dbInfo *databasereader.DatabaseInfo, unit string) (storage.Persister, error) {
	if d.LoadStaticPersisterCalled != nil {
		return d.LoadStaticPersisterCalled(dbInfo, unit)
	}

	return nil, nil
}

// IsInterfaceNil -
func (d *DatabaseReaderStub) IsInterfaceNil() bool {
	return d == nil
}
