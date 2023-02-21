package factory_test

import (
	"os"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/stretchr/testify/require"
)

func createDefaultDBConfig() config.DBConfig {
	return config.DBConfig{
		Type:                "LvlDBSerial",
		BatchDelaySeconds:   2,
		MaxBatchSize:        100,
		MaxOpenFiles:        10,
		UseTmpAsFilePath:    false,
		ShardIDProviderType: "BinarySplit",
		NumShards:           4,
	}
}

func TestNewPersisterFactory(t *testing.T) {
	t.Parallel()

	pf, err := factory.NewPersisterFactory(createDefaultDBConfig())
	require.NotNil(t, pf)
	require.Nil(t, err)
}

func TestPersisterFactory_Create(t *testing.T) {
	t.Parallel()

	t.Run("invalid file path, should fail", func(t *testing.T) {
		t.Parallel()

		pf, _ := factory.NewPersisterFactory(createDefaultDBConfig())

		p, err := pf.Create("")
		require.Nil(t, p)
		require.Equal(t, storage.ErrInvalidFilePath, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		pf, _ := factory.NewPersisterFactory(createDefaultDBConfig())

		dir := t.TempDir()

		p, err := pf.Create(dir)
		require.NotNil(t, p)
		require.Nil(t, err)
	})
}

func TestPersisterFactory_GetDBConfig(t *testing.T) {
	t.Parallel()

	t.Run("load db config from toml config file", func(t *testing.T) {
		t.Parallel()

		pf, _ := factory.NewPersisterFactory(createDefaultDBConfig())

		dirPath := t.TempDir()
		configPath := factory.GetPersisterConfigFilePath(dirPath)

		expectedDBConfig := config.DBConfig{
			FilePath:          "filepath1",
			Type:              "type1",
			BatchDelaySeconds: 1,
			MaxBatchSize:      2,
			MaxOpenFiles:      3,
			NumShards:         4,
		}

		err := core.SaveTomlFile(expectedDBConfig, configPath)
		require.Nil(t, err)

		conf, err := pf.GetDBConfig(dirPath)
		require.Nil(t, err)
		require.Equal(t, &expectedDBConfig, conf)
	})

	t.Run("not empty dir, load default db config", func(t *testing.T) {
		t.Parallel()

		pf, _ := factory.NewPersisterFactory(createDefaultDBConfig())

		dirPath := t.TempDir()

		f, err := core.CreateFile(core.ArgCreateFileArgument{
			Directory:     dirPath,
			Prefix:        "test",
			FileExtension: "log",
		})
		require.Nil(t, err)

		defer func() {
			_ = f.Close()
		}()

		expectedDBConfig := factory.GetDefaultDBConfig()

		conf, err := pf.GetDBConfig(dirPath)
		require.Nil(t, err)
		require.Equal(t, expectedDBConfig, conf)
	})

	t.Run("empty dir, load db config from main config", func(t *testing.T) {
		t.Parallel()

		expectedDBConfig := createDefaultDBConfig()

		pf, _ := factory.NewPersisterFactory(createDefaultDBConfig())

		dirPath := t.TempDir()

		conf, err := pf.GetDBConfig(dirPath)
		require.Nil(t, err)
		require.Equal(t, &expectedDBConfig, conf)
	})
}

func TestCreatePersisterConfigFile(t *testing.T) {
	t.Parallel()

	t.Run("no path present, in memory db, should not fail", func(t *testing.T) {
		t.Parallel()

		dbConfig := createDefaultDBConfig()

		err := factory.CreatePersisterConfigFile("no/valid/path", &dbConfig)
		require.Nil(t, err)
	})

	t.Run("config file already present, should not fail", func(t *testing.T) {
		t.Parallel()

		dbConfig := createDefaultDBConfig()
		dirPath := t.TempDir()
		configPath := factory.GetPersisterConfigFilePath(dirPath)

		err := core.SaveTomlFile(dbConfig, configPath)
		require.Nil(t, err)

		err = factory.CreatePersisterConfigFile(dirPath, &dbConfig)
		require.Nil(t, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dbConfig := createDefaultDBConfig()
		dirPath := t.TempDir()

		err := factory.CreatePersisterConfigFile(dirPath, &dbConfig)
		require.Nil(t, err)

		configPath := factory.GetPersisterConfigFilePath(dirPath)
		_, err = os.Stat(configPath)
		require.False(t, os.IsNotExist(err))
	})
}
