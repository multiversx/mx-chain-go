package factory_test

import (
	"os"
	"path"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/stretchr/testify/assert"
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

func TestDBConfigHandler_GetDBConfig(t *testing.T) {
	t.Parallel()

	t.Run("load db config from toml config file", func(t *testing.T) {
		t.Parallel()

		pf := factory.NewDBConfigHandler(createDefaultDBConfig())

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

		testConfig := createDefaultDBConfig()
		testConfig.BatchDelaySeconds = 37
		testConfig.MaxBatchSize = 38
		testConfig.MaxOpenFiles = 39
		testConfig.ShardIDProviderType = "BinarySplit"
		testConfig.NumShards = 4
		pf := factory.NewDBConfigHandler(testConfig)

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

		expectedDBConfig := &config.DBConfig{
			FilePath:            "",
			Type:                factory.DefaultType,
			BatchDelaySeconds:   testConfig.BatchDelaySeconds,
			MaxBatchSize:        testConfig.MaxBatchSize,
			MaxOpenFiles:        testConfig.MaxOpenFiles,
			UseTmpAsFilePath:    false,
			ShardIDProviderType: "",
			NumShards:           0,
		}

		conf, err := pf.GetDBConfig(dirPath)
		require.Nil(t, err)
		require.Equal(t, expectedDBConfig, conf)
	})
	t.Run("empty config.toml file, load default db config", func(t *testing.T) {
		t.Parallel()

		testConfig := createDefaultDBConfig()
		testConfig.BatchDelaySeconds = 37
		testConfig.MaxBatchSize = 38
		testConfig.MaxOpenFiles = 39
		testConfig.ShardIDProviderType = "BinarySplit"
		testConfig.NumShards = 4
		pf := factory.NewDBConfigHandler(testConfig)

		dirPath := t.TempDir()

		f, _ := os.Create(path.Join(dirPath, factory.DBConfigFileName))
		_ = f.Close()

		expectedDBConfig := &config.DBConfig{
			FilePath:            "",
			Type:                factory.DefaultType,
			BatchDelaySeconds:   testConfig.BatchDelaySeconds,
			MaxBatchSize:        testConfig.MaxBatchSize,
			MaxOpenFiles:        testConfig.MaxOpenFiles,
			UseTmpAsFilePath:    false,
			ShardIDProviderType: "",
			NumShards:           0,
		}

		conf, err := pf.GetDBConfig(dirPath)
		require.Nil(t, err)
		require.Equal(t, expectedDBConfig, conf)
	})
	t.Run("empty dir, load db config from main config", func(t *testing.T) {
		t.Parallel()

		expectedDBConfig := createDefaultDBConfig()

		pf := factory.NewDBConfigHandler(createDefaultDBConfig())

		dirPath := t.TempDir()

		conf, err := pf.GetDBConfig(dirPath)
		require.Nil(t, err)
		require.Equal(t, &expectedDBConfig, conf)
	})
	t.Run("getDBConfig twice, should load from config file if file available", func(t *testing.T) {
		t.Parallel()

		expectedDBConfig := createDefaultDBConfig()

		dbConfigHandler := factory.NewDBConfigHandler(createDefaultDBConfig())

		dirPath := t.TempDir()

		conf, err := dbConfigHandler.GetDBConfig(dirPath)
		require.Nil(t, err)
		require.Equal(t, &expectedDBConfig, conf)

		newDBConfig := config.DBConfig{
			Type:              "type1",
			BatchDelaySeconds: 1,
			MaxBatchSize:      2,
			MaxOpenFiles:      3,
			NumShards:         4,
		}

		configPath := factory.GetPersisterConfigFilePath(dirPath)

		err = core.SaveTomlFile(expectedDBConfig, configPath)
		require.Nil(t, err)

		dbConfigHandler = factory.NewDBConfigHandler(newDBConfig)
		conf, err = dbConfigHandler.GetDBConfig(dirPath)
		require.Nil(t, err)
		require.Equal(t, &expectedDBConfig, conf)
	})
}

func TestDBConfigHandler_SaveDBConfigToFilePath(t *testing.T) {
	t.Parallel()

	t.Run("no path present, in memory db, should not fail", func(t *testing.T) {
		t.Parallel()

		dbConfig := createDefaultDBConfig()

		pf := factory.NewDBConfigHandler(dbConfig)
		err := pf.SaveDBConfigToFilePath("no/valid/path", &dbConfig)
		require.Nil(t, err)
	})
	t.Run("config file already present, should not fail and should rewrite", func(t *testing.T) {
		t.Parallel()

		dbConfig1 := createDefaultDBConfig()
		dbConfig1.MaxOpenFiles = 37
		dbConfig1.Type = "dbconfig1"
		dirPath := t.TempDir()
		configPath := factory.GetPersisterConfigFilePath(dirPath)

		err := core.SaveTomlFile(dbConfig1, configPath)
		require.Nil(t, err)

		pf := factory.NewDBConfigHandler(dbConfig1)

		dbConfig2 := createDefaultDBConfig()
		dbConfig2.MaxOpenFiles = 38
		dbConfig2.Type = "dbconfig2"

		err = pf.SaveDBConfigToFilePath(dirPath, &dbConfig2)
		require.Nil(t, err)

		loadedDBConfig := &config.DBConfig{}
		err = core.LoadTomlFile(loadedDBConfig, path.Join(dirPath, "config.toml"))
		require.Nil(t, err)

		assert.Equal(t, dbConfig2, *loadedDBConfig)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dbConfig := createDefaultDBConfig()
		dirPath := t.TempDir()

		pf := factory.NewDBConfigHandler(dbConfig)
		err := pf.SaveDBConfigToFilePath(dirPath, &dbConfig)
		require.Nil(t, err)

		configPath := factory.GetPersisterConfigFilePath(dirPath)
		_, err = os.Stat(configPath)
		require.False(t, os.IsNotExist(err))
	})
}
