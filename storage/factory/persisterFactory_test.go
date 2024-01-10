package factory_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createPersisterFactory(config config.DBConfig) (storage.PersisterCreator, error) {
	pfh := factory.NewPersisterFactoryHandler(2, 1)
	return pfh.CreatePersisterHandler(config)
}

func TestNewPersisterFactory(t *testing.T) {
	t.Parallel()

	pf, err := createPersisterFactory(createDefaultDBConfig())
	require.NotNil(t, pf)
	require.Nil(t, err)
}

func TestPersisterFactory_Create(t *testing.T) {
	t.Parallel()

	t.Run("invalid file path, should fail", func(t *testing.T) {
		t.Parallel()

		pf, _ := createPersisterFactory(createDefaultDBConfig())

		p, err := pf.Create("")
		require.Nil(t, p)
		require.Equal(t, storage.ErrInvalidFilePath, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		pf, _ := createPersisterFactory(createDefaultDBConfig())

		dir := t.TempDir()

		p, err := pf.Create(dir)
		require.NotNil(t, p)
		require.Nil(t, err)
	})
}

func TestPersisterFactory_CreateWithRetries(t *testing.T) {
	t.Parallel()

	t.Run("invalid file path, should fail", func(t *testing.T) {
		t.Parallel()

		pf, _ := createPersisterFactory(createDefaultDBConfig())

		p, err := pf.CreateWithRetries("")
		require.Nil(t, p)
		require.Equal(t, storage.ErrInvalidFilePath, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		pf, _ := createPersisterFactory(createDefaultDBConfig())

		dir := t.TempDir()

		p, err := pf.CreateWithRetries(dir)
		require.NotNil(t, p)
		require.Nil(t, err)
	})
}

func TestPersisterFactory_Create_ConfigSaveToFilePath(t *testing.T) {
	t.Parallel()

	t.Run("should write toml config file for leveldb", func(t *testing.T) {
		t.Parallel()

		dbConfig := createDefaultBasePersisterConfig()
		dbConfig.Type = string(storageunit.LvlDB)
		pf, _ := createPersisterFactory(dbConfig)

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
		pf, _ := createPersisterFactory(dbConfig)

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
		pf, _ := createPersisterFactory(dbConfig)

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
		pf, _ := createPersisterFactory(dbConfig)

		dir := t.TempDir()
		path := dir + "storer/"

		p, err := pf.Create(path)
		require.NotNil(t, p)
		require.Nil(t, err)

		_, err = os.Stat(path)
		require.True(t, os.IsNotExist(err))
	})
}

func TestPersisterFactory_CreateDisabled(t *testing.T) {
	t.Parallel()

	factoryInstance, err := createPersisterFactory(createDefaultDBConfig())
	require.Nil(t, err)

	persisterInstance := factoryInstance.CreateDisabled()
	assert.NotNil(t, persisterInstance)
	assert.Equal(t, "*disabled.errorDisabledPersister", fmt.Sprintf("%T", persisterInstance))
}

func TestPersisterFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	pf, _ := createPersisterFactory(createDefaultDBConfig())
	require.False(t, pf.IsInterfaceNil())
}
