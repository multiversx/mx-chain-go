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
}

func TestPersisterFactory_Create(t *testing.T) {
	t.Parallel()

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
}
