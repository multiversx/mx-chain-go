package databasereader_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/databasereader"
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/mock"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/require"
)

func TestDatabaseReader_GetHeaders(t *testing.T) {
	t.Parallel()

	args := getDatabaseReaderArgs()

	hdr := &block.Header{Nonce: 5}
	hdrBytes, _ := args.Marshalizer.Marshal(hdr)

	hdrsPersister := mock.NewPersisterMock()
	_ = hdrsPersister.Put([]byte("hdr"), hdrBytes)

	epoch, shard := uint32(111), uint32(222)

	args.PersisterFactory = &mock.PersisterFactoryStub{
		CreateCalled: func(path string) (storage.Persister, error) {
			require.True(t, strings.Contains(path, fmt.Sprintf("%d", epoch)))
			require.True(t, strings.Contains(path, fmt.Sprintf("%d", shard)))

			return hdrsPersister, nil
		},
		CreateDisabledCalled: nil,
	}
	dr, _ := databasereader.New(args)

	dbInfo := databasereader.DatabaseInfo{
		Epoch: epoch,
		Shard: shard,
	}
	hdrs, err := dr.GetHeaders(&dbInfo)
	require.NoError(t, err)
	require.Equal(t, hdr, hdrs[0])
}

func TestDatabaseReader_LoadPersister(t *testing.T) {
	t.Parallel()

	persister := mock.NewPersisterMock()
	args := getDatabaseReaderArgs()
	args.PersisterFactory = &mock.PersisterFactoryStub{
		CreateCalled: func(path string) (storage.Persister, error) {
			return persister, nil
		},
	}

	dr, _ := databasereader.New(args)

	dbInfo := databasereader.DatabaseInfo{
		Epoch: 5,
		Shard: 5,
	}
	pers, err := dr.LoadPersister(&dbInfo, "unit")
	require.NoError(t, err)
	require.Equal(t, persister, pers)
}

func TestDatabaseReader_LoadStaticPersister(t *testing.T) {
	t.Parallel()

	persister := mock.NewPersisterMock()
	args := getDatabaseReaderArgs()
	args.PersisterFactory = &mock.PersisterFactoryStub{
		CreateCalled: func(path string) (storage.Persister, error) {
			return persister, nil
		},
	}

	dr, _ := databasereader.New(args)

	dbInfo := databasereader.DatabaseInfo{
		Epoch: 5,
		Shard: 5,
	}
	pers, err := dr.LoadStaticPersister(&dbInfo, "unit")
	require.NoError(t, err)
	require.Equal(t, persister, pers)
}
