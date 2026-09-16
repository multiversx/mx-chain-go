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

		// identity fields come from the file, tuning from the main config
		expectedDBConfig.BatchDelaySeconds = 2
		expectedDBConfig.MaxBatchSize = 100
		expectedDBConfig.MaxOpenFiles = 10

		conf, err := pf.GetDBConfig(dirPath)
		require.Nil(t, err)
		require.Equal(t, &expectedDBConfig, conf)
	})
	t.Run("load db config from toml config file, main config without tuning keeps the file tuning", func(t *testing.T) {
		t.Parallel()

		pf := factory.NewDBConfigHandler(config.DBConfig{Type: "LvlDBSerial", BloomFilterBitsPerKey: 7})

		dirPath := t.TempDir()
		configPath := factory.GetPersisterConfigFilePath(dirPath)

		expectedDBConfig := config.DBConfig{
			Type:              "type1",
			BatchDelaySeconds: 1,
			MaxBatchSize:      2,
			MaxOpenFiles:      3,
		}

		err := core.SaveTomlFile(expectedDBConfig, configPath)
		require.Nil(t, err)

		expectedDBConfig.BloomFilterBitsPerKey = 7

		conf, err := pf.GetDBConfig(dirPath)
		require.Nil(t, err)
		require.Equal(t, &expectedDBConfig, conf)
	})
	t.Run("legacy dir with goleveldb files, load default provided config", func(t *testing.T) {
		t.Parallel()

		testConfig := createDefaultDBConfig()
		testConfig.BloomFilterBitsPerKey = 10
		pf := factory.NewDBConfigHandler(testConfig)

		dirPath := t.TempDir()
		createEmptyFiles(t, dirPath, "CURRENT", "MANIFEST-000000", "000001.log", "000002.ldb")

		conf, err := pf.GetDBConfig(dirPath)
		require.Nil(t, err)
		require.Equal(t, factory.DefaultType, conf.Type)
		require.Equal(t, 10, conf.BloomFilterBitsPerKey)
	})
	t.Run("legacy dir with files of another engine, should error", func(t *testing.T) {
		t.Parallel()

		pf := factory.NewDBConfigHandler(createDefaultDBConfig())

		for _, foreignFile := range []string{"OPTIONS-000003", "000004.sst", "marker.manifest.000001.MANIFEST-000001", "000005.blob"} {
			dirPath := t.TempDir()
			createEmptyFiles(t, dirPath, "MANIFEST-000001", "000002.log", foreignFile)

			conf, err := pf.GetDBConfig(dirPath)
			require.Nil(t, conf, foreignFile)
			require.ErrorContains(t, err, "unsupported db engine", foreignFile)
		}
	})
	t.Run("legacy sharded dir with files of another engine in a shard, should error", func(t *testing.T) {
		t.Parallel()

		pf := factory.NewDBConfigHandler(createDefaultDBConfig())

		dirPath := t.TempDir()
		shardPath := path.Join(dirPath, "1")
		require.Nil(t, os.Mkdir(shardPath, 0700))
		createEmptyFiles(t, shardPath, "000004.sst")

		conf, err := pf.GetDBConfig(dirPath)
		require.Nil(t, conf)
		require.ErrorContains(t, err, "unsupported db engine")
	})
	t.Run("not empty dir, load default provided config", func(t *testing.T) {
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

		// the file keeps the identity, the new main config provides the tuning
		expectedDBConfig.BatchDelaySeconds = newDBConfig.BatchDelaySeconds
		expectedDBConfig.MaxBatchSize = newDBConfig.MaxBatchSize
		expectedDBConfig.MaxOpenFiles = newDBConfig.MaxOpenFiles

		dbConfigHandler = factory.NewDBConfigHandler(newDBConfig)
		conf, err = dbConfigHandler.GetDBConfig(dirPath)
		require.Nil(t, err)
		require.Equal(t, &expectedDBConfig, conf)
	})
}

func createEmptyFiles(t *testing.T, dirPath string, names ...string) {
	for _, name := range names {
		f, err := os.Create(path.Join(dirPath, name))
		require.Nil(t, err)
		require.Nil(t, f.Close())
	}
}

func TestDBConfigHandler_SaveDBConfigToFilePath(t *testing.T) {
	t.Parallel()

	t.Run("missing dir, should create it and write the config", func(t *testing.T) {
		t.Parallel()

		dbConfig := createDefaultDBConfig()
		dirPath := path.Join(t.TempDir(), "nested", "path")

		pf := factory.NewDBConfigHandler(dbConfig)
		err := pf.SaveDBConfigToFilePath(dirPath, &dbConfig)
		require.Nil(t, err)

		loadedDBConfig := &config.DBConfig{}
		err = core.LoadTomlFile(loadedDBConfig, factory.GetPersisterConfigFilePath(dirPath))
		require.Nil(t, err)
		require.Equal(t, dbConfig, *loadedDBConfig)

		entries, err := os.ReadDir(dirPath)
		require.Nil(t, err)
		require.Len(t, entries, 1, "no temporary file must be left behind")
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
