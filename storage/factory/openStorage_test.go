package factory

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/mock"
	"github.com/stretchr/testify/assert"
)

func createMockArgsOpenStorageUnits() ArgsNewOpenStorageUnits {
	return ArgsNewOpenStorageUnits{
		GeneralConfig:             config.Config{},
		Marshalizer:               &mock.MarshalizerMock{},
		BootstrapDataProvider:     &mock.BootStrapDataProviderStub{},
		LatestStorageDataProvider: &mock.LatestStorageDataProviderStub{},
		WorkingDir:                "",
		ChainID:                   "",
		DefaultDBPath:             "",
		DefaultEpochString:        "Epoch",
		DefaultShardString:        "Shard",
	}
}

func TestNewStorageUnitOpenHandler(t *testing.T) {
	t.Parallel()

	suoh, err := NewStorageUnitOpenHandler(createMockArgsOpenStorageUnits())

	assert.NoError(t, err)
	assert.False(t, check.IfNil(suoh))
}

func TestGetMostUpToDateDirectory(t *testing.T) {
	t.Parallel()

	lastRound := int64(100)
	args := createMockArgsOpenStorageUnits()
	args.BootstrapDataProvider = &mock.BootStrapDataProviderStub{
		LoadForPathCalled: func(persisterFactory storage.PersisterFactory, path string) (*bootstrapStorage.BootstrapData, storage.Storer, error) {
			if strings.Contains(path, "Shard_0") {
				return &bootstrapStorage.BootstrapData{}, nil, nil
			} else {
				return &bootstrapStorage.BootstrapData{
					LastRound: lastRound,
				}, nil, nil
			}
		},
	}
	suoh, _ := NewStorageUnitOpenHandler(args)

	shardIDsStr := []string{"0", "1"}
	path := "currPath"
	dirName, err := suoh.getMostUpToDateDirectory(path, shardIDsStr, nil)
	assert.NoError(t, err)
	assert.Equal(t, shardIDsStr[1], dirName)
}

func TestGetMostRecentBootstrapStorageUnit_GetShardsFromDirectoryErr(t *testing.T) {
	t.Parallel()

	localErr := errors.New("localErr")
	args := createMockArgsOpenStorageUnits()
	args.LatestStorageDataProvider = &mock.LatestStorageDataProviderStub{
		GetShardsFromDirectoryCalled: func(path string) ([]string, error) {
			return nil, localErr
		},
	}
	suoh, _ := NewStorageUnitOpenHandler(args)

	storer, err := suoh.GetMostRecentBootstrapStorageUnit()
	assert.Nil(t, storer)
	assert.Equal(t, localErr, err)
}

func TestGetMostRecentBootstrapStorageUnit_CannotGetMostUpToDateDirectory(t *testing.T) {
	t.Parallel()

	args := createMockArgsOpenStorageUnits()
	args.LatestStorageDataProvider = &mock.LatestStorageDataProviderStub{
		GetShardsFromDirectoryCalled: func(path string) ([]string, error) {
			return []string{"0", "1"}, nil
		},
	}
	suoh, _ := NewStorageUnitOpenHandler(args)

	storer, err := suoh.GetMostRecentBootstrapStorageUnit()
	assert.Nil(t, storer)
	assert.Equal(t, storage.ErrBootstrapDataNotFoundInStorage, err)
}

func TestGetMostRecentBootstrapStorageUnit_CannotCreatePersister(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Parallel()

	args := createMockArgsOpenStorageUnits()
	args.LatestStorageDataProvider = &mock.LatestStorageDataProviderStub{
		GetShardsFromDirectoryCalled: func(path string) ([]string, error) {
			return []string{"0", "1"}, nil
		},
	}
	args.BootstrapDataProvider = &mock.BootStrapDataProviderStub{
		LoadForPathCalled: func(persisterFactory storage.PersisterFactory, path string) (*bootstrapStorage.BootstrapData, storage.Storer, error) {
			return &bootstrapStorage.BootstrapData{
				LastRound: 100,
			}, nil, nil
		},
	}
	suoh, _ := NewStorageUnitOpenHandler(args)

	storer, err := suoh.GetMostRecentBootstrapStorageUnit()
	assert.Nil(t, storer)
	assert.Equal(t, storage.ErrNotSupportedDBType, err)
}

func TestGetMostRecentBootstrapStorageUnit(t *testing.T) {
	t.Parallel()

	args := createMockArgsOpenStorageUnits()
	args.GeneralConfig = config.Config{BootstrapStorage: config.StorageConfig{
		DB: config.DBConfig{Type: "MemoryDB"},
	}}
	args.LatestStorageDataProvider = &mock.LatestStorageDataProviderStub{
		GetShardsFromDirectoryCalled: func(path string) ([]string, error) {
			return []string{"0", "1"}, nil
		},
	}
	args.BootstrapDataProvider = &mock.BootStrapDataProviderStub{
		LoadForPathCalled: func(persisterFactory storage.PersisterFactory, path string) (*bootstrapStorage.BootstrapData, storage.Storer, error) {
			return &bootstrapStorage.BootstrapData{
				LastRound: 100,
			}, nil, nil
		},
	}
	suoh, _ := NewStorageUnitOpenHandler(args)

	storer, err := suoh.GetMostRecentBootstrapStorageUnit()
	assert.NoError(t, err)
	assert.NotNil(t, storer)

}
