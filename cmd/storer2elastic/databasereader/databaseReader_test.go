package databasereader_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/databasereader"
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/mock"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/stretchr/testify/require"
)

func TestNewDatabaseReader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		argsFunc func() databasereader.Args
		exError  error
	}{
		{
			name: "NilDirectoryReader",
			argsFunc: func() databasereader.Args {
				args := getDatabaseReaderArgs()
				args.DirectoryReader = nil
				return args
			},
			exError: databasereader.ErrNilDirectoryReader,
		},
		{
			name: "NilMarshalizer",
			argsFunc: func() databasereader.Args {
				args := getDatabaseReaderArgs()
				args.Marshalizer = nil
				return args
			},
			exError: databasereader.ErrNilMarshalizer,
		},
		{
			name: "NilPersisterFactory",
			argsFunc: func() databasereader.Args {
				args := getDatabaseReaderArgs()
				args.PersisterFactory = nil
				return args
			},
			exError: databasereader.ErrNilPersisterFactory,
		},
		{
			name: "EmptyDbPath",
			argsFunc: func() databasereader.Args {
				args := getDatabaseReaderArgs()
				args.DbPathWithChainID = ""
				return args
			},
			exError: databasereader.ErrEmptyDbFilePath,
		},
		{
			name: "All arguments ok",
			argsFunc: func() databasereader.Args {
				return getDatabaseReaderArgs()
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := databasereader.New(tt.argsFunc())
			require.Equal(t, err, tt.exError)
		})
	}
}

func TestDatabaseReader_GetDatabaseInfo_NoDatabaseFoundShouldErr(t *testing.T) {
	t.Parallel()

	args := getDatabaseReaderArgs()
	args.DirectoryReader = &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(directoryPath string) ([]string, error) {
			return []string{}, nil
		},
	}

	dbr, err := databasereader.New(args)
	require.NoError(t, err)

	dbsInfo, err := dbr.GetDatabaseInfo()
	require.Equal(t, databasereader.ErrNoDatabaseFound, err)
	require.Nil(t, dbsInfo)

}

func TestDatabaseReader_GetDatabaseInfo_ShouldWork(t *testing.T) {
	t.Parallel()

	epoch, shard := uint32(1), uint32(2)
	args := getDatabaseReaderArgs()
	args.DirectoryReader = &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(directoryPath string) ([]string, error) {
			return []string{fmt.Sprintf("Epoch_%d", epoch), fmt.Sprintf("Shard_%d", shard)}, nil
		},
	}

	dbr, err := databasereader.New(args)
	require.NoError(t, err)

	dbsInfo, err := dbr.GetDatabaseInfo()
	require.Nil(t, err)
	dbInfo := dbsInfo[0]
	require.Equal(t, epoch, dbInfo.Epoch)
	require.Equal(t, shard, dbInfo.Shard)
}

func TestDatabaseReader_GetStaticDatabaseInfo_NoDatabaseFoundShouldErr(t *testing.T) {
	t.Parallel()

	args := getDatabaseReaderArgs()
	args.DirectoryReader = &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(directoryPath string) ([]string, error) {
			return []string{}, nil
		},
	}

	dbr, err := databasereader.New(args)
	require.NoError(t, err)

	dbsInfo, err := dbr.GetStaticDatabaseInfo()
	require.Equal(t, databasereader.ErrNoDatabaseFound, err)
	require.Nil(t, dbsInfo)

}

func TestDatabaseReader_GetStaticDatabaseInfo_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getDatabaseReaderArgs()
	shard := uint32(5)
	args.DirectoryReader = &mock.DirectoryReaderStub{
		ListDirectoriesAsStringCalled: func(directoryPath string) ([]string, error) {
			return []string{"Static", fmt.Sprintf("Shard_%d", shard)}, nil
		},
	}

	dbr, err := databasereader.New(args)
	require.NoError(t, err)

	dbsInfo, err := dbr.GetStaticDatabaseInfo()
	require.Nil(t, err)

	dbInfo := dbsInfo[0]
	require.Equal(t, 1, len(dbsInfo))
	require.Equal(t, shard, dbInfo.Shard)
}

func getDatabaseReaderArgs() databasereader.Args {
	return databasereader.Args{
		DirectoryReader:   &mock.DirectoryReaderStub{},
		GeneralConfig:     config.Config{},
		Marshalizer:       &mock.MarshalizerMock{},
		PersisterFactory:  &mock.PersisterFactoryStub{},
		DbPathWithChainID: "path",
	}
}
