package factory_test

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-storage-go/common"
	"github.com/multiversx/mx-chain-storage-go/pebbledb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	t.Run("with tmp file path, should work", func(t *testing.T) {
		t.Parallel()

		conf := createDefaultDBConfig()
		conf.UseTmpAsFilePath = true

		pf, _ := factory.NewPersisterFactory(conf)

		dir := t.TempDir()

		p, err := pf.Create(dir)
		require.NotNil(t, p)
		require.Nil(t, err)

		// config.toml will be created in tmp path, but cannot be easily checked since
		// the file path is not created deterministically

		// should not find in the dir created initially.
		_, err = os.Stat(dir + "/config.toml")
		require.Error(t, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		pf, _ := factory.NewPersisterFactory(createDefaultDBConfig())

		dir := t.TempDir()

		p, err := pf.Create(dir)
		require.NotNil(t, p)
		require.Nil(t, err)

		// check config.toml file exists
		_, err = os.Stat(dir + "/config.toml")
		require.Nil(t, err)
	})
}

func TestPersisterFactory_CreateWithRetries(t *testing.T) {
	t.Parallel()

	t.Run("wrong config should error", func(t *testing.T) {
		t.Parallel()

		path := path.Join(t.TempDir(), "TEST")
		dbConfig := createDefaultDBConfig()
		dbConfig.Type = "invalid type"

		persisterFactory, err := factory.NewPersisterFactory(dbConfig)
		assert.Nil(t, err)

		db, err := persisterFactory.CreateWithRetries(path)
		assert.True(t, check.IfNil(db))
		assert.Equal(t, common.ErrNotSupportedDBType, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		path := path.Join(t.TempDir(), "TEST")
		dbConfig := createDefaultDBConfig()
		dbConfig.FilePath = path

		persisterFactory, err := factory.NewPersisterFactory(dbConfig)
		assert.Nil(t, err)

		db, err := persisterFactory.CreateWithRetries(path)
		assert.False(t, check.IfNil(db))
		assert.Nil(t, err)
		_ = db.Close()
	})
}

func TestPersisterFactory_Create_ConfigSaveToFilePath(t *testing.T) {
	t.Parallel()

	t.Run("should write toml config file for leveldb", func(t *testing.T) {
		t.Parallel()

		dbConfig := createDefaultBasePersisterConfig()
		dbConfig.Type = string(storageunit.LvlDB)
		pf, _ := factory.NewPersisterFactory(dbConfig)

		dir := t.TempDir()
		path := dir + "storer/"

		p, err := pf.Create(path)
		require.NotNil(t, p)
		require.Nil(t, err)

		configPath := factory.GetPersisterConfigFilePath(path)
		_, err = os.Stat(configPath)
		require.False(t, os.IsNotExist(err))
	})

	t.Run("should write toml config file for serial leveldb", func(t *testing.T) {
		t.Parallel()

		dbConfig := createDefaultBasePersisterConfig()
		dbConfig.Type = string(storageunit.LvlDBSerial)
		pf, _ := factory.NewPersisterFactory(dbConfig)

		dir := t.TempDir()
		path := dir + "storer/"

		p, err := pf.Create(path)
		require.NotNil(t, p)
		require.Nil(t, err)

		configPath := factory.GetPersisterConfigFilePath(path)
		_, err = os.Stat(configPath)
		require.False(t, os.IsNotExist(err))
	})

	t.Run("should not write toml config file for memory db", func(t *testing.T) {
		t.Parallel()

		dbConfig := createDefaultBasePersisterConfig()
		dbConfig.Type = string(storageunit.MemoryDB)
		pf, _ := factory.NewPersisterFactory(dbConfig)

		dir := t.TempDir()
		path := dir + "storer/"

		p, err := pf.Create(path)
		require.NotNil(t, p)
		require.Nil(t, err)

		configPath := factory.GetPersisterConfigFilePath(path)
		_, err = os.Stat(configPath)
		require.True(t, os.IsNotExist(err))
	})

	t.Run("should not create path dir for memory db", func(t *testing.T) {
		t.Parallel()

		dbConfig := createDefaultBasePersisterConfig()
		dbConfig.Type = string(storageunit.MemoryDB)
		pf, _ := factory.NewPersisterFactory(dbConfig)

		dir := t.TempDir()
		path := dir + "storer/"

		p, err := pf.Create(path)
		require.NotNil(t, p)
		require.Nil(t, err)

		_, err = os.Stat(path)
		require.True(t, os.IsNotExist(err))
	})
}

func TestPersisterFactory_Create_ConfigBeforeEngine(t *testing.T) {
	t.Parallel()

	t.Run("legacy goleveldb dir without config file gets the config written", func(t *testing.T) {
		t.Parallel()

		pf, _ := factory.NewPersisterFactory(createDefaultBasePersisterConfig())
		dir := t.TempDir()

		p, err := pf.Create(dir)
		require.Nil(t, err)
		require.Nil(t, p.Close())

		configPath := factory.GetPersisterConfigFilePath(dir)
		require.Nil(t, os.Remove(configPath))

		p, err = pf.Create(dir)
		require.Nil(t, err)
		require.Nil(t, p.Close())

		_, err = os.Stat(configPath)
		require.Nil(t, err)
	})

	t.Run("pebble dir without config file is reopened with pebble even if the main config says leveldb", func(t *testing.T) {
		t.Parallel()

		key, val := []byte("key"), []byte("value")
		dir := t.TempDir()

		pebbleConfig := createDefaultBasePersisterConfig()
		pebbleConfig.Type = string(storageunit.PebbleDB)
		pf, _ := factory.NewPersisterFactory(pebbleConfig)
		p, err := pf.Create(dir)
		require.Nil(t, err)
		require.Nil(t, p.Put(key, val))
		require.Nil(t, p.Close())
		require.Nil(t, os.Remove(factory.GetPersisterConfigFilePath(dir)))

		pf, _ = factory.NewPersisterFactory(createDefaultBasePersisterConfig())
		p, err = pf.Create(dir)
		require.Nil(t, err)
		require.Equal(t, "*pebbledb.DB", fmt.Sprintf("%T", p))
		res, err := p.Get(key)
		require.Nil(t, err)
		require.Equal(t, val, res)
		require.Nil(t, p.Close())

		conf, err := factory.NewDBConfigHandler(createDefaultBasePersisterConfig()).GetDBConfig(dir)
		require.Nil(t, err)
		require.Equal(t, string(storageunit.PebbleDB), conf.Type)
	})

	t.Run("unsupported type leaves no config file behind", func(t *testing.T) {
		t.Parallel()

		dbConfig := createDefaultDBConfig()
		dbConfig.Type = "invalid type"
		pf, _ := factory.NewPersisterFactory(dbConfig)
		dir := path.Join(t.TempDir(), "storer")

		p, err := pf.Create(dir)
		require.Nil(t, p)
		require.Equal(t, common.ErrNotSupportedDBType, err)

		_, err = os.Stat(dir)
		require.True(t, os.IsNotExist(err))
	})
}

func TestPersisterFactory_MixedEngines(t *testing.T) {
	t.Parallel()

	t.Run("plain persisters", func(t *testing.T) {
		t.Parallel()

		testMixedEngines(t, createDefaultBasePersisterConfig(), "*leveldb.SerialDB", "*pebbledb.DB")
	})
	t.Run("sharded persisters", func(t *testing.T) {
		t.Parallel()

		testMixedEngines(t, createDefaultDBConfig(), "*sharded.shardedPersister", "*sharded.shardedPersister")
	})
}

// the operator switches the unit to pebble: existing directories keep their engine, new ones get pebble
func testMixedEngines(t *testing.T, leveldbConfig config.DBConfig, oldEpochType string, newEpochType string) {
	key, val := []byte("key"), []byte("value")
	oldEpochDir := path.Join(t.TempDir(), "Epoch_1")
	newEpochDir := path.Join(t.TempDir(), "Epoch_2")

	leveldbFactory, _ := factory.NewPersisterFactory(leveldbConfig)
	p, err := leveldbFactory.Create(oldEpochDir)
	require.Nil(t, err)
	require.Nil(t, p.Put(key, val))
	require.Nil(t, p.Close())

	pebbleConfig := leveldbConfig
	pebbleConfig.Type = string(storageunit.PebbleDB)
	pebbleConfig.PebbleProfile = pebbledb.HeavyWriteProfile
	resources, err := pebbledb.NewSharedResources(8<<20, 100)
	require.Nil(t, err)
	pebbleFactory, _ := factory.NewPersisterFactoryWithResources(pebbleConfig, resources)

	oldEpoch, err := pebbleFactory.Create(oldEpochDir)
	require.Nil(t, err)
	require.Equal(t, oldEpochType, fmt.Sprintf("%T", oldEpoch))
	res, err := oldEpoch.Get(key)
	require.Nil(t, err)
	require.Equal(t, val, res)

	newEpoch, err := pebbleFactory.Create(newEpochDir)
	require.Nil(t, err)
	require.Equal(t, newEpochType, fmt.Sprintf("%T", newEpoch))
	require.Nil(t, newEpoch.Put(key, val))
	res, err = newEpoch.Get(key)
	require.Nil(t, err)
	require.Equal(t, val, res)

	require.Nil(t, oldEpoch.Close())
	require.Nil(t, newEpoch.Close())
	resources.Close()

	for dir, expectedType := range map[string]string{oldEpochDir: leveldbConfig.Type, newEpochDir: string(storageunit.PebbleDB)} {
		conf, errGet := factory.NewDBConfigHandler(pebbleConfig).GetDBConfig(dir)
		require.Nil(t, errGet)
		require.Equal(t, expectedType, conf.Type, dir)
	}
}

func TestPersisterFactory_CreateDisabled(t *testing.T) {
	t.Parallel()

	factoryInstance, err := factory.NewPersisterFactory(createDefaultDBConfig())
	require.Nil(t, err)

	persisterInstance := factoryInstance.CreateDisabled()
	assert.NotNil(t, persisterInstance)
	assert.Equal(t, "*disabled.errorDisabledPersister", fmt.Sprintf("%T", persisterInstance))
}

func TestPersisterFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	pf, _ := factory.NewPersisterFactory(createDefaultDBConfig())
	require.False(t, pf.IsInterfaceNil())
}

func TestGetTmpFilePath(t *testing.T) {
	t.Parallel()

	pathSeparator := "/"

	tmpDir := os.TempDir()
	tmpBasePath := path.Join(tmpDir, pathSeparator)

	tmpPath, err := factory.GetTmpFilePath("aaaa/bbbb/cccc")
	require.Nil(t, err)
	require.True(t, strings.Contains(tmpPath, path.Join(tmpBasePath, "cccc")))

	tmpPath, _ = factory.GetTmpFilePath("aaaa")
	require.True(t, strings.Contains(tmpPath, path.Join(tmpBasePath, "aaaa")))

	tmpPath, _ = factory.GetTmpFilePath("")
	require.True(t, strings.Contains(tmpPath, path.Join(tmpBasePath, "")))

	tmpPath, _ = factory.GetTmpFilePath("/")
	require.True(t, strings.Contains(tmpPath, path.Join(tmpBasePath, "")))
}
