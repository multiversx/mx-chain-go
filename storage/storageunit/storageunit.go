package storageunit

import (
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-storage-go/storageCacherAdapter"
	"github.com/multiversx/mx-chain-storage-go/storageUnit"
)

// Unit represents a storer's data bank
// holding the cache and persistence unit
type Unit = storageUnit.Unit

// CacheConfig holds the configurable elements of a cache
type CacheConfig = storageUnit.CacheConfig

// ArgDB is a structure that is used to create a new storage.Persister implementation
type ArgDB = storageUnit.ArgDB

// DBConfig holds the configurable elements of a database
type DBConfig = storageUnit.DBConfig

// NilStorer resembles a disabled implementation of the Storer interface
type NilStorer = storageUnit.NilStorer

// CacheType represents the type of the supported caches
type CacheType = storageUnit.CacheType

// DBType represents the type of the supported databases
type DBType = storageUnit.DBType

// NewStorageUnit is the constructor for the storage unit, creating a new storage unit
// from the given cacher and persister.
func NewStorageUnit(c storage.Cacher, p storage.Persister) (*Unit, error) {
	return storageUnit.NewStorageUnit(c, p)
}

// NewCache creates a new cache from a cache config
func NewCache(config CacheConfig) (storage.Cacher, error) {
	return storageUnit.NewCache(config)
}

// NewDB creates a new database from database config
func NewDB(argDB ArgDB) (storage.Persister, error) {
	return storageUnit.NewDB(argDB)
}

// NewStorageUnitFromConf creates a new storage unit from a storage unit config
func NewStorageUnitFromConf(cacheConf CacheConfig, dbConf DBConfig) (*Unit, error) {
	return storageUnit.NewStorageUnitFromConf(cacheConf, dbConf)
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
