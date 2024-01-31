package storageunit_test

import (
	"path"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/mock"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/multiversx/mx-chain-storage-go/common"
	"github.com/stretchr/testify/assert"
)

func TestNewStorageUnit(t *testing.T) {
	t.Parallel()

	cacher := &testscommon.CacherStub{}
	persister := &mock.PersisterStub{}

	t.Run("nil cacher should error", func(t *testing.T) {
		t.Parallel()

		unit, err := storageunit.NewStorageUnit(nil, persister)
		assert.Nil(t, unit)
		assert.Equal(t, common.ErrNilCacher, err)
	})
	t.Run("nil persister should error", func(t *testing.T) {
		t.Parallel()

		unit, err := storageunit.NewStorageUnit(cacher, nil)
		assert.Nil(t, unit)
		assert.Equal(t, common.ErrNilPersister, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		unit, err := storageunit.NewStorageUnit(cacher, persister)
		assert.NotNil(t, unit)
		assert.Nil(t, err)
	})
}

func TestNewCache(t *testing.T) {
	t.Parallel()

	t.Run("wrong config should error", func(t *testing.T) {
		t.Parallel()

		cfg := storageunit.CacheConfig{
			Type:     "invalid type",
			Capacity: 100,
		}
		cache, err := storageunit.NewCache(cfg)
		assert.True(t, check.IfNil(cache))
		assert.Equal(t, common.ErrNotSupportedCacheType, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cfg := storageunit.CacheConfig{
			Type:     "LRU",
			Capacity: 100,
		}
		cache, err := storageunit.NewCache(cfg)
		assert.False(t, check.IfNil(cache))
		assert.Nil(t, err)
	})
}

func TestNewStorageUnitFromConf(t *testing.T) {
	t.Parallel()

	path := path.Join(t.TempDir(), "TEST")
	dbConfig := storageunit.DBConfig{
		FilePath:          path,
		Type:              "LvlDBSerial",
		BatchDelaySeconds: 5,
		MaxBatchSize:      10,
		MaxOpenFiles:      10,
	}

	t.Run("invalid config should error", func(t *testing.T) {
		t.Parallel()

		cacheConfig := storageunit.CacheConfig{
			Type:     "invalid type",
			Capacity: 100,
		}

		dbConf := config.DBConfig{
			Type:              dbConfig.FilePath,
			BatchDelaySeconds: dbConfig.BatchDelaySeconds,
			MaxBatchSize:      dbConfig.MaxBatchSize,
			MaxOpenFiles:      dbConfig.MaxOpenFiles,
		}
		persisterFactory, err := factory.NewPersisterFactory(dbConf)
		assert.Nil(t, err)

		unit, err := storageunit.NewStorageUnitFromConf(cacheConfig, dbConfig, persisterFactory)
		assert.Nil(t, unit)
		assert.Equal(t, common.ErrNotSupportedCacheType, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		cacheConfig := storageunit.CacheConfig{
			Type:     "LRU",
			Capacity: 100,
		}

		dbConf := config.DBConfig{
			Type:              string(dbConfig.Type),
			BatchDelaySeconds: dbConfig.BatchDelaySeconds,
			MaxBatchSize:      dbConfig.MaxBatchSize,
			MaxOpenFiles:      dbConfig.MaxOpenFiles,
		}
		persisterFactory, err := factory.NewPersisterFactory(dbConf)
		assert.Nil(t, err)

		unit, err := storageunit.NewStorageUnitFromConf(cacheConfig, dbConfig, persisterFactory)
		assert.NotNil(t, unit)
		assert.Nil(t, err)
		_ = unit.Close()
	})
}

func TestNewNilStorer(t *testing.T) {
	t.Parallel()

	unit := storageunit.NewNilStorer()
	assert.NotNil(t, unit)
}

func TestNewStorageCacherAdapter(t *testing.T) {
	t.Parallel()

	cacher := &mock.AdaptedSizedLruCacheStub{}
	db := &mock.PersisterStub{}
	storedDataFactory := &storage.StoredDataFactoryStub{}
	marshaller := &marshallerMock.MarshalizerStub{}

	t.Run("nil parameter should error", func(t *testing.T) {
		t.Parallel()

		adaptor, err := storageunit.NewStorageCacherAdapter(nil, db, storedDataFactory, marshaller)
		assert.True(t, check.IfNil(adaptor))
		assert.Equal(t, common.ErrNilCacher, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		adaptor, err := storageunit.NewStorageCacherAdapter(cacher, db, storedDataFactory, marshaller)
		assert.False(t, check.IfNil(adaptor))
		assert.Nil(t, err)
	})
}
