package storageunit

import (
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common/stateAnalysis"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-storage-go/common"
	"github.com/multiversx/mx-chain-storage-go/factory"
	"github.com/multiversx/mx-chain-storage-go/storageCacherAdapter"
	"github.com/multiversx/mx-chain-storage-go/storageUnit"
)

// Unit represents a storer's data bank
// holding the cache and persistence unit
type Unit = storageUnit.Unit

// CacheConfig holds the configurable elements of a cache
type CacheConfig = common.CacheConfig

// DBConfig holds the configurable elements of a database
type DBConfig = common.DBConfig

// NilStorer resembles a disabled implementation of the Storer interface
type NilStorer = storageUnit.NilStorer

// CacheType represents the type of the supported caches
type CacheType = common.CacheType

// DBType represents the type of the supported databases
type DBType = common.DBType

// ShardIDProviderType represents the type of the supported shard id providers
type ShardIDProviderType = common.ShardIDProviderType

// ArgDB is a structure that is used to create a new storage.Persister implementation
type ArgDB = factory.ArgDB

// NewStorageUnit is the constructor for the storage unit, creating a new storage unit
// from the given cacher and persister.
func NewStorageUnit(c storage.Cacher, p storage.Persister) (*Unit, error) {
	return storageUnit.NewStorageUnit(c, p)
}

// NewCache creates a new cache from a cache config
func NewCache(config CacheConfig) (storage.Cacher, error) {
	return factory.NewCache(config)
}

// NewDB creates a new database from database config
func NewDB(args ArgDB) (storage.Persister, error) {
	return factory.NewDB(args)
}

// NewStorageUnitFromConf creates a new storage unit from a storage unit config
func NewStorageUnitFromConf(cacheConf CacheConfig, dbConf DBConfig, persisterFactory storage.PersisterFactoryHandler) (*Unit, error) {
	if dbConf.MaxBatchSize > int(cacheConf.Capacity) {
		return nil, common.ErrCacheSizeIsLowerThanBatchSize
	}

	cache, err := NewCache(cacheConf)
	if err != nil {
		return nil, err
	}

	db, err := persisterFactory.CreateWithRetries(dbConf.FilePath)
	if err != nil {
		return nil, err
	}

	return NewStorageUnit(cache, db)
}

// NewTrieStorageUnitFromConf creates the storage unit used by a trie storer when snapshots (and thus pruning
// storers) are disabled. When the state analysis collector is enabled, the persister is timed and the cache hits,
// persister hits, misses, puts and cache size are reported under the unit derived from the db path
func NewTrieStorageUnitFromConf(cacheConf CacheConfig, dbConf DBConfig, persisterFactory storage.PersisterFactoryHandler) (storage.Storer, error) {
	if dbConf.MaxBatchSize > int(cacheConf.Capacity) {
		return nil, common.ErrCacheSizeIsLowerThanBatchSize
	}

	cache, err := NewCache(cacheConf)
	if err != nil {
		return nil, err
	}

	db, err := persisterFactory.CreateWithRetries(dbConf.FilePath)
	if err != nil {
		return nil, err
	}

	unit, err := NewStorageUnit(cache, stateAnalysis.WrapPersister(db, dbConf.FilePath))
	if err != nil {
		return nil, err
	}

	analysisUnit := stateAnalysis.UnitFromIdentifier(dbConf.FilePath)
	if !stateAnalysis.Enabled() || analysisUnit == stateAnalysis.UnitOther {
		return unit, nil
	}

	stateAnalysis.RegisterCacheProbe(dbConf.FilePath, func() (int, int, int) {
		return cache.Len(), int(cache.SizeInBytesContained()), 1
	})

	return &analysedUnit{Unit: unit, cacher: cache, analysisUnit: analysisUnit}, nil
}

// analysedUnit counts cache hits, persister hits, misses and puts of a static trie storage unit
type analysedUnit struct {
	*Unit
	cacher       storage.Cacher
	analysisUnit int
}

// Get searches the key in the cache and then in the persister, recording where it was found
func (au *analysedUnit) Get(key []byte) ([]byte, error) {
	_, inCache := au.cacher.Peek(key)
	val, err := au.Unit.Get(key)
	switch {
	case inCache:
		stateAnalysis.Inc(stateAnalysis.CStPsCacheHit.Unit(au.analysisUnit))
	case err == nil:
		stateAnalysis.Inc(stateAnalysis.CStPsPersisterHit.PersisterBucket(au.analysisUnit, 0))
	default:
		stateAnalysis.Inc(stateAnalysis.CStPsMiss.Unit(au.analysisUnit))
	}

	return val, err
}

// Put adds data to both cache and persistence medium
func (au *analysedUnit) Put(key, data []byte) error {
	stateAnalysis.Inc(stateAnalysis.CStPsPut.Unit(au.analysisUnit))
	return au.Unit.Put(key, data)
}

// IsInterfaceNil returns true if there is no value under the interface
func (au *analysedUnit) IsInterfaceNil() bool {
	return au == nil
}

// NewNilStorer will return a nil storer
func NewNilStorer() *NilStorer {
	return storageUnit.NewNilStorer()
}

// NewStorageCacherAdapter creates a new storageCacherAdapter
func NewStorageCacherAdapter(
	cacher storage.AdaptedSizedLRUCache,
	db storage.Persister,
	storedDataFactory storage.StoredDataFactory,
	marshaller marshal.Marshalizer,
) (storage.Cacher, error) {
	return storageCacherAdapter.NewStorageCacherAdapter(cacher, db, storedDataFactory, marshaller)
}
