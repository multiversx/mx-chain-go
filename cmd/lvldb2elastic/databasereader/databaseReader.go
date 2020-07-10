package databasereader

import (
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/factory"
)

type databaseReader struct {
	directoryReader storage.DirectoryReaderHandler
	dbFilePath      string
}

// NewDatabaseReader will return a new instance of databaseReader
func NewDatabaseReader(dbFilePath string) (*databaseReader, error) {
	if len(dbFilePath) == 0 {
		return nil, ErrEmptyDbFilePath
	}

	return &databaseReader{
		directoryReader: factory.NewDirectoryReader(),
		dbFilePath:      dbFilePath,
	}, nil
}

// TODO : finish
