package dataprocessor

import (
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/databasereader"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// DatabaseReaderHandler defines the actions that a database reader has to do
type DatabaseReaderHandler interface {
	GetDatabaseInfo() ([]*databasereader.DatabaseInfo, error)
	GetHeaders(dbInfo *databasereader.DatabaseInfo) ([]data.HeaderHandler, error)
	LoadPersister(dbInfo *databasereader.DatabaseInfo, unit string) (storage.Persister, error)
	IsInterfaceNil() bool
}

// NodesCoordinator defines the actions that a nodes' coordinator has to do
type NodesCoordinator interface {
	sharding.NodesCoordinator
	EpochStartPrepare(metaHdr data.HeaderHandler, body data.BodyHandler)
}
