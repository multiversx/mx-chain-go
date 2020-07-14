package dataindexer

import (
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/databasereader"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type DatabaseReaderHandler interface {
	GetDatabaseInfo() ([]*databasereader.DatabaseInfo, error)
	GetHeaders(dbInfo *databasereader.DatabaseInfo) ([]data.HeaderHandler, error)
	LoadPersister(dbInfo *databasereader.DatabaseInfo, unit string) (storage.Persister, error)
	IsInterfaceNil() bool
}
