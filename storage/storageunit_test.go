package storage_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/stretchr/testify/assert"
)

func initStorageUnit(t *testing.T, cSize int) *storage.Unit {
	mdb, err1 := memorydb.New()
	cache, err2 := lrucache.NewCache(cSize)

	assert.Nil(t, err1, "failed creating db: %s", err1)
	assert.Nil(t, err2, "no error expected but got %s", err2)

	sUnit, err := storage.NewStorageUnit(cache, mdb)

	assert.Nil(t, err, "failed to create storage unit")

	return sUnit
}

func TestStorageUnitNilPersister(t *testing.T) {
	cache, err := lrucache.NewCache(10)

	assert.Nil(t, err, "no error expected but got %s", err)

	_, err = storage.NewStorageUnit(cache, nil)

	assert.NotNil(t, err, "expected failure")
}

func TestStorageUnitNilCacher(t *testing.T) {
	mdb, err1 := memorydb.New()

	assert.Nil(t, err1, "failed creating db")

	_, err1 = storage.NewStorageUnit(nil, mdb)

	assert.NotNil(t, err1, "expected failure")
}

func TestPutNotPresent(t *testing.T) {
	key, val := []byte("key0"), []byte("value0")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	has, err := s.Has(key)

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.True(t, has, "expected to find key %s, but not found", key)
}

func TestPutNotPresentCache(t *testing.T) {
	key, val := []byte("key1"), []byte("value1")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	s.ClearCache()

	has, err := s.Has(key)

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.True(t, has, "expected to find key %s, but not found", key)
}

func TestPutPresent(t *testing.T) {
	key, val := []byte("key2"), []byte("value2")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	// put again same value, no error expected
	err = s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestGetNotPresent(t *testing.T) {
	key := []byte("key3")
	s := initStorageUnit(t, 10)
	v, err := s.Get(key)

	assert.NotNil(t, err, "expected to find no value, but found %s", v)
}

func TestGetNotPresentCache(t *testing.T) {
	key, val := []byte("key4"), []byte("value4")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	s.ClearCache()

	v, err := s.Get(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.Equal(t, val, v, "expected %s but got %s", val, v)
}

func TestGetPresent(t *testing.T) {
	key, val := []byte("key5"), []byte("value4")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	v, err := s.Get(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.Equal(t, val, v, "expected %s but got %s", val, v)
}

func TestHasNotPresent(t *testing.T) {
	key := []byte("key6")
	s := initStorageUnit(t, 10)
	has, err := s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.False(t, has, "not expected to find value")
}

func TestHasNotPresentCache(t *testing.T) {
	key, val := []byte("key7"), []byte("value7")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	s.ClearCache()

	has, err := s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.True(t, has, "expected to find key but not found")
}

func TestHasPresent(t *testing.T) {
	key, val := []byte("key8"), []byte("value8")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	has, err := s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.True(t, has, "expected to find key but not found")
}

func TestHasOrAddNotPresent(t *testing.T) {
	key, val := []byte("key9"), []byte("value9")
	s := initStorageUnit(t, 10)
	has, err := s.HasOrAdd(key, val)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.False(t, has, "not expected to find value")

	has, err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.True(t, has, "expected to find key but not found")
}

func TestHasOrAddNotPresentCache(t *testing.T) {
	key, val := []byte("key10"), []byte("value10")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)

	s.ClearCache()

	has, err := s.HasOrAdd(key, val)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.True(t, has, "expected to find value")
}

func TestHasOrAddPresent(t *testing.T) {
	key, val := []byte("key11"), []byte("value11")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)

	has, err := s.HasOrAdd(key, val)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.True(t, has, "expected to find value")
}

func TestDeleteNotPresent(t *testing.T) {
	key := []byte("key12")
	s := initStorageUnit(t, 10)
	err := s.Remove(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
}

func TestDeleteNotPresentCache(t *testing.T) {
	key, val := []byte("key13"), []byte("value13")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)
	assert.Nil(t, err, "Could not put value in storage unit")

	has, err := s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.True(t, has, "expected to find key")

	s.ClearCache()

	err = s.Remove(key)
	assert.Nil(t, err, "expected no error, but got %s", err)

	has, err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.False(t, has, "not expected to find value")
}

func TestDeletePresent(t *testing.T) {
	key, val := []byte("key14"), []byte("value14")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)
	assert.Nil(t, err, "Could not put value in storage unit")

	has, err := s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.True(t, has, "expected to find key")

	err = s.Remove(key)

	assert.Nil(t, err, "expected no error, but got %s", err)

	has, err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.False(t, has, "not expected to find value")
}

func TestClearCacheNotAffectPersist(t *testing.T) {
	key, val := []byte("key15"), []byte("value15")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)
	assert.Nil(t, err, "Could not put value in storage unit")
	s.ClearCache()

	has, err := s.Has(key)

	assert.Nil(t, err, "no error expected, but got %s", err)
	assert.True(t, has, "expected to find key")
}

func TestDestroyUnitNoError(t *testing.T) {
	s := initStorageUnit(t, 10)
	err := s.DestroyUnit()
	assert.Nil(t, err, "no error expected, but got %s", err)
}

func TestCreateCacheFromConfWrongType(t *testing.T) {

	cacher, err := storage.NewCache("NotLRU", 100)

	assert.NotNil(t, err, "error expected")
	assert.Nil(t, cacher, "cacher expected to be nil, but got %s", cacher)
}

func TestCreateCacheFromConfOK(t *testing.T) {

	cacher, err := storage.NewCache(storage.LRUCache, 10)

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, cacher, "valid cacher expected but got nil")
}

func TestCreateDBFromConfWrongType(t *testing.T) {
	persister, err := storage.NewDB("NotLvlDB", "test")

	assert.NotNil(t, err, "error expected")
	assert.Nil(t, persister, "persister expected to be nil, but got %s", persister)
}

func TestCreateDBFromConfWrongFileName(t *testing.T) {
	persister, err := storage.NewDB(storage.LvlDB, "")
	assert.NotNil(t, err, "error expected")
	assert.Nil(t, persister, "persister expected to be nil, but got %s", persister)
}

func TestCreateDBFromConfOk(t *testing.T) {
	persister, err := storage.NewDB(storage.LvlDB, "tmp")
	assert.Nil(t, err, "no error expected")
	assert.NotNil(t, persister, "valid persister expected but got nil")

	err = persister.Destroy()
	assert.Nil(t, err, "no error expected destroying the persister")
}

func TestNewStorageUnitFromConfWrongCacheConfig(t *testing.T) {

	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: "NotLRU",
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     storage.LvlDB,
	})

	assert.NotNil(t, err, "error expected")
	assert.Nil(t, storer, "storer expected to be nil but got %s", storer)
}

func TestNewStorageUnitFromConfWrongDBConfig(t *testing.T) {
	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: storage.LRUCache,
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     "NotLvlDB",
	})

	assert.NotNil(t, err, "error expected")
	assert.Nil(t, storer, "storer expected to be nil but got %s", storer)
}

func TestNewStorageUnitFromConfOk(t *testing.T) {
	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: storage.LRUCache,
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     storage.LvlDB,
	})

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, storer, "valid storer expected but got nil")
	err = storer.DestroyUnit()
	assert.Nil(t, err, "no error expected destroying the persister")
}