<<<<<<< HEAD
package factory_test

import (
	"os"
	"testing"

	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/stretchr/testify/require"
)

func TestNewPersisterFactory(t *testing.T) {
	t.Parallel()

	dbConfigHandler := factory.NewDBConfigHandler(createDefaultDBConfig())
	pf, err := factory.NewPersisterFactory(dbConfigHandler)
	require.NotNil(t, pf)
	require.Nil(t, err)
=======
package factory

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDBConfig(dbType string) config.DBConfig {
	return config.DBConfig{
		FilePath:          "TEST",
		Type:              dbType,
		BatchDelaySeconds: 5,
		MaxBatchSize:      100,
		MaxOpenFiles:      10,
		UseTmpAsFilePath:  false,
	}
}

func TestNewPersisterFactory(t *testing.T) {
	t.Parallel()

	factoryInstance := NewPersisterFactory(createDBConfig("LvlDB"))
	assert.NotNil(t, factoryInstance)
>>>>>>> rc/v1.6.0
}

func TestPersisterFactory_Create(t *testing.T) {
	t.Parallel()

<<<<<<< HEAD
	t.Run("invalid file path, should fail", func(t *testing.T) {
		t.Parallel()

		dbConfigHandler := factory.NewDBConfigHandler(createDefaultDBConfig())
		pf, _ := factory.NewPersisterFactory(dbConfigHandler)

		p, err := pf.Create("")
		require.Nil(t, p)
		require.Equal(t, storage.ErrInvalidFilePath, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		dbConfigHandler := factory.NewDBConfigHandler(createDefaultDBConfig())
		pf, _ := factory.NewPersisterFactory(dbConfigHandler)

		dir := t.TempDir()

		p, err := pf.Create(dir)
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
		dbConfigHandler := factory.NewDBConfigHandler(dbConfig)
		pf, _ := factory.NewPersisterFactory(dbConfigHandler)

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
		dbConfigHandler := factory.NewDBConfigHandler(dbConfig)
		pf, _ := factory.NewPersisterFactory(dbConfigHandler)

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
		dbConfigHandler := factory.NewDBConfigHandler(dbConfig)
		pf, _ := factory.NewPersisterFactory(dbConfigHandler)

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
		dbConfigHandler := factory.NewDBConfigHandler(dbConfig)
		pf, _ := factory.NewPersisterFactory(dbConfigHandler)

		dir := t.TempDir()
		path := dir + "storer/"

		p, err := pf.Create(path)
		require.NotNil(t, p)
		require.Nil(t, err)

		_, err = os.Stat(path)
		require.True(t, os.IsNotExist(err))
	})
=======
	t.Run("empty path should error", func(t *testing.T) {
		t.Parallel()

		factoryInstance := NewPersisterFactory(createDBConfig("LvlDB"))
		persisterInstance, err := factoryInstance.Create("")
		assert.True(t, check.IfNil(persisterInstance))
		expectedErrString := "invalid file path"
		assert.Equal(t, expectedErrString, err.Error())
	})
	t.Run("unknown type should error", func(t *testing.T) {
		t.Parallel()

		factoryInstance := NewPersisterFactory(createDBConfig("invalid type"))
		persisterInstance, err := factoryInstance.Create(t.TempDir())
		assert.True(t, check.IfNil(persisterInstance))
		assert.Equal(t, storage.ErrNotSupportedDBType, err)
	})
	t.Run("for LvlDB should work", func(t *testing.T) {
		t.Parallel()

		factoryInstance := NewPersisterFactory(createDBConfig("LvlDB"))
		persisterInstance, err := factoryInstance.Create(t.TempDir())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(persisterInstance))
		assert.Equal(t, "*leveldb.DB", fmt.Sprintf("%T", persisterInstance))
		_ = persisterInstance.Close()
	})
	t.Run("for LvlDBSerial should work", func(t *testing.T) {
		t.Parallel()

		factoryInstance := NewPersisterFactory(createDBConfig("LvlDBSerial"))
		persisterInstance, err := factoryInstance.Create(t.TempDir())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(persisterInstance))
		assert.Equal(t, "*leveldb.SerialDB", fmt.Sprintf("%T", persisterInstance))
		_ = persisterInstance.Close()
	})
	t.Run("for MemoryDB should work", func(t *testing.T) {
		t.Parallel()

		factoryInstance := NewPersisterFactory(createDBConfig("MemoryDB"))
		persisterInstance, err := factoryInstance.Create(t.TempDir())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(persisterInstance))
		assert.Equal(t, "*memorydb.DB", fmt.Sprintf("%T", persisterInstance))
		_ = persisterInstance.Close()
	})
}

func TestPersisterFactory_CreateDisabled(t *testing.T) {
	t.Parallel()

	factoryInstance := NewPersisterFactory(createDBConfig("LvlDB"))
	persisterInstance := factoryInstance.CreateDisabled()
	assert.NotNil(t, persisterInstance)
	assert.Equal(t, "*disabled.errorDisabledPersister", fmt.Sprintf("%T", persisterInstance))
}

func TestPersisterFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var pf *PersisterFactory
	require.True(t, pf.IsInterfaceNil())

	pf = NewPersisterFactory(config.DBConfig{})
	require.False(t, pf.IsInterfaceNil())
>>>>>>> rc/v1.6.0
}
