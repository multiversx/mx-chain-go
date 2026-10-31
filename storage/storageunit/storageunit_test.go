package storageunit_test

import (
	"compress/gzip"
	"encoding/csv"
	"os"
	"path"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-storage-go/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/common/stateAnalysis"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage/factory"
	"github.com/multiversx/mx-chain-go/storage/mock"
	"github.com/multiversx/mx-chain-go/storage/storageunit"
	"github.com/multiversx/mx-chain-go/testscommon/cache"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
)

func TestNewStorageUnit(t *testing.T) {
	t.Parallel()

	cacher := &cache.CacherStub{}
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

func TestNewTrieStorageUnitFromConf_StateAnalysisCounters(t *testing.T) {
	// this test enables the global state analysis collector, so it must not run in parallel with other tests

	dir := t.TempDir()
	require.Nil(t, stateAnalysis.Init(stateAnalysis.Config{
		Enabled:                  true,
		Label:                    "static-unit",
		BlocksPerRow:             1,
		MemorySampleEveryNBlocks: 1,
	}, 0, dir))

	dbConfig := storageunit.DBConfig{
		FilePath:          path.Join(t.TempDir(), "AccountsTrie"),
		Type:              "LvlDBSerial",
		BatchDelaySeconds: 5,
		MaxBatchSize:      1,
		MaxOpenFiles:      10,
	}
	persisterFactory, err := factory.NewPersisterFactory(config.DBConfig{
		Type:              string(dbConfig.Type),
		BatchDelaySeconds: dbConfig.BatchDelaySeconds,
		MaxBatchSize:      dbConfig.MaxBatchSize,
		MaxOpenFiles:      dbConfig.MaxOpenFiles,
	})
	require.Nil(t, err)

	unit, err := storageunit.NewTrieStorageUnitFromConf(storageunit.CacheConfig{Type: "LRU", Capacity: 2}, dbConfig, persisterFactory)
	require.Nil(t, err)
	defer func() { _ = unit.Close() }()

	stateAnalysis.BeginBlock(0, 0, 1, 1, 0, false)

	require.Nil(t, unit.Put([]byte("k1"), []byte("v1")))
	require.Nil(t, unit.Put([]byte("k2"), []byte("v2")))
	require.Nil(t, unit.Put([]byte("k3"), []byte("v3"))) // evicts k1 from the cache

	val, err := unit.Get([]byte("k3")) // cache hit
	require.Nil(t, err)
	require.Equal(t, []byte("v3"), val)
	val, err = unit.Get([]byte("k1")) // persister hit
	require.Nil(t, err)
	require.Equal(t, []byte("v1"), val)
	_, err = unit.Get([]byte("missing")) // miss
	require.NotNil(t, err)

	stateAnalysis.EndBlock()
	stateAnalysis.Close()

	columns, row := readBlocksCsv(t, path.Join(dir, "state-analysis", "shard_0", "blocks.csv.gz"))
	get := func(name string) string {
		idx, found := columns[name]
		require.True(t, found, name)
		return row[idx]
	}
	assert.Equal(t, "3", get("st_ps_put_accounts"))
	assert.Equal(t, "3", get("st_db_put_accounts"))
	assert.Equal(t, "1", get("st_ps_cache_hit_accounts"))
	assert.Equal(t, "1", get("st_ps_persister_hit_accounts_b0"))
	assert.Equal(t, "1", get("st_ps_miss_accounts"))
	assert.Equal(t, "2", get("st_db_get_accounts"))
	assert.Equal(t, "1", get("st_db_get_hit_accounts"))
	assert.Equal(t, "2", get("st_cache_entries_accounts"))
	assert.Equal(t, "1", get("st_active_persisters_accounts"))
	assert.Equal(t, "0", get("st_ps_put_other"))
}

func readBlocksCsv(t *testing.T, filePath string) (map[string]int, []string) {
	f, err := os.Open(filePath)
	require.Nil(t, err)
	defer func() { _ = f.Close() }()
	gz, err := gzip.NewReader(f)
	require.Nil(t, err)
	rows, err := csv.NewReader(gz).ReadAll()
	require.Nil(t, err)
	require.GreaterOrEqual(t, len(rows), 2)

	columns := make(map[string]int)
	for i, name := range rows[0] {
		columns[name] = i
	}
	return columns, rows[len(rows)-1]
}
