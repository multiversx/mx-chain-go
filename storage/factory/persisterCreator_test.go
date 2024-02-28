package factory_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createDefaultBasePersisterConfig() config.DBConfig {
	return config.DBConfig{
		Type:                "LvlDBSerial",
		BatchDelaySeconds:   2,
		MaxBatchSize:        100,
		MaxOpenFiles:        10,
		UseTmpAsFilePath:    false,
		ShardIDProviderType: "BinarySplit",
		NumShards:           1,
	}
}

func TestPersisterCreator_Create(t *testing.T) {
	t.Parallel()

	t.Run("invalid file path, should fail", func(t *testing.T) {
		t.Parallel()

		pc := factory.NewPersisterCreator(createDefaultDBConfig())

		p, err := pc.Create("")
		require.Nil(t, p)
		require.Equal(t, storage.ErrInvalidFilePath, err)
	})

	t.Run("use tmp as file path", func(t *testing.T) {
		t.Parallel()

		conf := createDefaultDBConfig()
		conf.UseTmpAsFilePath = true

		pc := factory.NewPersisterCreator(conf)

		p, err := pc.Create("path1")
		require.Nil(t, err)
		require.NotNil(t, p)
	})

	t.Run("should create non sharded persister", func(t *testing.T) {
		t.Parallel()

		pc := factory.NewPersisterCreator(createDefaultBasePersisterConfig())

		dir := t.TempDir()
		p, err := pc.Create(dir)
		require.NotNil(t, p)
		require.Nil(t, err)

		assert.True(t, strings.Contains(fmt.Sprintf("%T", p), "*leveldb.SerialDB"))
	})

	t.Run("should create sharded persister", func(t *testing.T) {
		t.Parallel()

		pc := factory.NewPersisterCreator(createDefaultDBConfig())

		dir := t.TempDir()
		p, err := pc.Create(dir)
		require.NotNil(t, p)
		require.Nil(t, err)

		assert.True(t, strings.Contains(fmt.Sprintf("%T", p), "*sharded.shardedPersister"))
	})
}

func TestPersisterCreator_CreateBasePersister(t *testing.T) {
	t.Parallel()

	t.Run("not supported type, should fail", func(t *testing.T) {
		t.Parallel()

		dbConfig := createDefaultBasePersisterConfig()
		dbConfig.Type = "not supported type"
		pc := factory.NewPersisterCreator(dbConfig)

		dir := t.TempDir()
		p, err := pc.CreateBasePersister(dir)
		require.Nil(t, p)
		require.Equal(t, storage.ErrNotSupportedDBType, err)
	})

	t.Run("leveldb", func(t *testing.T) {
		t.Parallel()

		dbConfig := createDefaultBasePersisterConfig()
		dbConfig.Type = string(storageunit.LvlDB)
		pc := factory.NewPersisterCreator(dbConfig)

		dir := t.TempDir()
		p, err := pc.CreateBasePersister(dir)
		require.NotNil(t, p)
		require.Nil(t, err)

		assert.True(t, strings.Contains(fmt.Sprintf("%T", p), "*leveldb.DB"))
	})

	t.Run("serial leveldb", func(t *testing.T) {
		t.Parallel()

		pc := factory.NewPersisterCreator(createDefaultBasePersisterConfig())

		dir := t.TempDir()
		p, err := pc.CreateBasePersister(dir)
		require.NotNil(t, p)
		require.Nil(t, err)

		assert.True(t, strings.Contains(fmt.Sprintf("%T", p), "*leveldb.SerialDB"))
	})

	t.Run("memorydb", func(t *testing.T) {
		t.Parallel()

		dbConfig := createDefaultBasePersisterConfig()
		dbConfig.Type = string(storageunit.MemoryDB)
		pc := factory.NewPersisterCreator(dbConfig)

		dir := t.TempDir()
		p, err := pc.CreateBasePersister(dir)
		require.NotNil(t, p)
		require.Nil(t, err)

		assert.True(t, strings.Contains(fmt.Sprintf("%T", p), "*memorydb.DB"))
	})
}

func TestPersisterCreator_CreateShardIDProvider(t *testing.T) {
	t.Parallel()

	t.Run("not supported type, should fail", func(t *testing.T) {
		t.Parallel()

		dbConfig := createDefaultDBConfig()
		dbConfig.ShardIDProviderType = "not supported type"
		pc := factory.NewPersisterCreator(dbConfig)

		p, err := pc.CreateShardIDProvider()
		require.Nil(t, p)
		require.Equal(t, storage.ErrNotSupportedShardIDProviderType, err)
	})

	t.Run("binary split, should work", func(t *testing.T) {
		t.Parallel()

		dbConfig := createDefaultDBConfig()
		pc := factory.NewPersisterCreator(dbConfig)

		p, err := pc.CreateShardIDProvider()
		require.NotNil(t, p)
		require.Nil(t, err)

		assert.True(t, strings.Contains(fmt.Sprintf("%T", p), "*sharded.shardIDProvider"))
	})
}

func TestGetTmpFilePath(t *testing.T) {
	t.Parallel()

	t.Run("invalid path separator, should fail", func(t *testing.T) {
		t.Parallel()

		invalidPathSeparator := ","
		path, err := factory.GetTmpFilePath("aaaa/bbbb/cccc", invalidPathSeparator)
		require.NotNil(t, err)
		require.Equal(t, "", path)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		pathSeparator := "/"

		tmpDir := os.TempDir()
		tmpBasePath := tmpDir + pathSeparator

		path, err := factory.GetTmpFilePath("aaaa/bbbb/cccc", pathSeparator)
		require.Nil(t, err)
		require.True(t, strings.Contains(path, tmpBasePath+"cccc"))

		path, _ = factory.GetTmpFilePath("aaaa", pathSeparator)
		require.True(t, strings.Contains(path, tmpBasePath+"aaaa"))

		path, _ = factory.GetTmpFilePath("", pathSeparator)
		require.True(t, strings.Contains(path, tmpBasePath+""))

		path, _ = factory.GetTmpFilePath("/", pathSeparator)
		require.True(t, strings.Contains(path, tmpBasePath+""))
	})
}
