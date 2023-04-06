package factory

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgsOpenStorageUnits() ArgsNewOpenStorageUnits {
	return ArgsNewOpenStorageUnits{
		BootstrapDataProvider:     &mock.BootStrapDataProviderStub{},
		LatestStorageDataProvider: &mock.LatestStorageDataProviderStub{},
		DefaultEpochString:        "Epoch",
		DefaultShardString:        "Shard",
	}
}

func TestNewStorageUnitOpenHandler(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		suoh, err := NewStorageUnitOpenHandler(createMockArgsOpenStorageUnits())
		assert.NoError(t, err)
		assert.NotNil(t, suoh)
	})
	t.Run("nil BootstrapDataProvider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsOpenStorageUnits()
		args.BootstrapDataProvider = nil
		suoh, err := NewStorageUnitOpenHandler(args)
		assert.Equal(t, storage.ErrNilBootstrapDataProvider, err)
		assert.Nil(t, suoh)
	})
	t.Run("nil LatestStorageDataProvider should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsOpenStorageUnits()
		args.LatestStorageDataProvider = nil
		suoh, err := NewStorageUnitOpenHandler(args)
		assert.Equal(t, storage.ErrNilLatestStorageDataProvider, err)
		assert.Nil(t, suoh)
	})
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
	dirName, err := suoh.getMostUpToDateDirectory(config.DBConfig{}, path, shardIDsStr, nil)
	assert.NoError(t, err)
	assert.Equal(t, shardIDsStr[1], dirName)
}

func TestGetMostRecentBootstrapStorageUnit_GetParentDirAndLastEpochErr(t *testing.T) {
	t.Parallel()

	localErr := errors.New("localErr")
	args := createMockArgsOpenStorageUnits()
	args.LatestStorageDataProvider = &mock.LatestStorageDataProviderStub{
		GetParentDirAndLastEpochCalled: func() (string, uint32, error) {
			return "", 0, localErr
		},
	}
	suoh, _ := NewStorageUnitOpenHandler(args)

	storer, err := suoh.GetMostRecentStorageUnit(config.DBConfig{})
	assert.Nil(t, storer)
	assert.Equal(t, localErr, err)
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

	storer, err := suoh.GetMostRecentStorageUnit(config.DBConfig{})
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

	storer, err := suoh.GetMostRecentStorageUnit(config.DBConfig{})
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

	storer, err := suoh.GetMostRecentStorageUnit(config.DBConfig{})
	assert.Nil(t, storer)
	assert.Equal(t, storage.ErrNotSupportedDBType, err)
}

func TestGetMostRecentBootstrapStorageUnit(t *testing.T) {
	t.Parallel()

	args := createMockArgsOpenStorageUnits()
	generalConfig := config.Config{BootstrapStorage: config.StorageConfig{
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

	storer, err := suoh.GetMostRecentStorageUnit(generalConfig.BootstrapStorage.DB)
	assert.NoError(t, err)
	assert.NotNil(t, storer)
}

func TestStorageUnitOpenHandler_OpenDB(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	args := createMockArgsOpenStorageUnits()
	args.LatestStorageDataProvider = &mock.LatestStorageDataProviderStub{
		GetParentDirectoryCalled: func() string {
			return tempDir
		},
	}
	suoh, _ := NewStorageUnitOpenHandler(args)

	// do not run these in parallel as they are using the same temp dir
	t.Run("create DB fails, should error", func(t *testing.T) {
		dbConfig := config.DBConfig{
			FilePath:          "Test",
			Type:              "invalid DB type",
			BatchDelaySeconds: 5,
			MaxBatchSize:      100,
			MaxOpenFiles:      10,
			UseTmpAsFilePath:  false,
		}

		storerInstance, err := suoh.OpenDB(dbConfig, 0, 0)
		assert.NotNil(t, err)
		expectedErrorString := "not supported db type"
		assert.Equal(t, expectedErrorString, err.Error())
		assert.Nil(t, storerInstance)
	})
	t.Run("should work", func(t *testing.T) {
		dbConfig := config.DBConfig{
			FilePath:          "Test",
			Type:              "LvlDBSerial",
			BatchDelaySeconds: 5,
			MaxBatchSize:      100,
			MaxOpenFiles:      10,
			UseTmpAsFilePath:  false,
		}

		storerInstance, err := suoh.OpenDB(dbConfig, 0, 0)
		assert.Nil(t, err)
		assert.NotNil(t, storerInstance)

		_ = storerInstance.Close()
	})

}

func TestOldDataCleanerProvider_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var osu *openStorageUnits
	require.True(t, osu.IsInterfaceNil())

	osu, _ = NewStorageUnitOpenHandler(createMockArgsOpenStorageUnits())
	require.False(t, osu.IsInterfaceNil())
}
