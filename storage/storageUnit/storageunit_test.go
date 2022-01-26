package storageUnit_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

func logError(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}

func initStorageUnit(tb testing.TB, cSize int) *storageUnit.Unit {
	mdb := memorydb.New()
	cache, err2 := lrucache.NewCache(cSize)
	assert.Nil(tb, err2, "no error expected but got %s", err2)

	sUnit, err := storageUnit.NewStorageUnit(cache, mdb)
	assert.Nil(tb, err, "failed to create storage unit")

	return sUnit
}

func TestStorageUnitNilPersister(t *testing.T) {
	cache, err1 := lrucache.NewCache(10)

	assert.Nil(t, err1, "no error expected but got %s", err1)

	_, err := storageUnit.NewStorageUnit(cache, nil)

	assert.NotNil(t, err, "expected failure")
}

func TestStorageUnitNilCacher(t *testing.T) {
	mdb := memorydb.New()

	_, err1 := storageUnit.NewStorageUnit(nil, mdb)
	assert.NotNil(t, err1, "expected failure")
}

func TestStorageUnit(t *testing.T) {
	cache, err1 := lrucache.NewCache(10)
	mdb := memorydb.New()

	assert.Nil(t, err1, "no error expected but got %s", err1)

	_, err := storageUnit.NewStorageUnit(cache, mdb)
	assert.Nil(t, err, "did not expect failure")
}

func TestPutNotPresent(t *testing.T) {
	key, val := []byte("key0"), []byte("value0")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	err = s.Has(key)

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestPutNotPresentCache(t *testing.T) {
	key, val := []byte("key1"), []byte("value1")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	s.ClearCache()

	err = s.Has(key)

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestPutPresentShouldOverwriteValue(t *testing.T) {
	key, val := []byte("key2"), []byte("value2")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	newVal := []byte("value5")
	err = s.Put(key, newVal)
	assert.Nil(t, err, "no error expected but got %s", err)

	returnedVal, err := s.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, newVal, returnedVal)
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
	err := s.Has(key)

	assert.NotNil(t, err)
	assert.Equal(t, err, storage.ErrKeyNotFound)
}

func TestHasNotPresentCache(t *testing.T) {
	key, val := []byte("key7"), []byte("value7")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	s.ClearCache()

	err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
}

func TestHasPresent(t *testing.T) {
	key, val := []byte("key8"), []byte("value8")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
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

	err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)

	s.ClearCache()

	err = s.Remove(key)
	assert.Nil(t, err, "expected no error, but got %s", err)

	err = s.Has(key)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "key not found")
}

func TestDeletePresent(t *testing.T) {
	key, val := []byte("key14"), []byte("value14")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)
	assert.Nil(t, err, "Could not put value in storage unit")

	err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)

	err = s.Remove(key)

	assert.Nil(t, err, "expected no error, but got %s", err)

	err = s.Has(key)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "key not found")
}

func TestClearCacheNotAffectPersist(t *testing.T) {
	key, val := []byte("key15"), []byte("value15")
	s := initStorageUnit(t, 10)
	err := s.Put(key, val)
	assert.Nil(t, err, "Could not put value in storage unit")
	s.ClearCache()

	err = s.Has(key)

	assert.Nil(t, err, "no error expected, but got %s", err)
}

func TestDestroyUnitNoError(t *testing.T) {
	s := initStorageUnit(t, 10)
	err := s.DestroyUnit()
	assert.Nil(t, err, "no error expected, but got %s", err)
}

func TestCreateCacheFromConfWrongType(t *testing.T) {

	cacher, err := storageUnit.NewCache(storageUnit.CacheConfig{Type: "NotLRU", Capacity: 100, Shards: 1, SizeInBytes: 0})

	assert.NotNil(t, err, "error expected")
	assert.Nil(t, cacher, "cacher expected to be nil, but got %s", cacher)
}

func TestCreateCacheFromConfOK(t *testing.T) {

	cacher, err := storageUnit.NewCache(storageUnit.CacheConfig{Type: storageUnit.LRUCache, Capacity: 10, Shards: 1, SizeInBytes: 0})

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, cacher, "valid cacher expected but got nil")
}

func TestCreateDBFromConfWrongType(t *testing.T) {
	arg := storageUnit.ArgDB{
		DBType:            "NotLvlDB",
		Path:              "test",
		BatchDelaySeconds: 10,
		MaxBatchSize:      10,
		MaxOpenFiles:      10,
	}
	persister, err := storageUnit.NewDB(arg)

	assert.NotNil(t, err, "error expected")
	assert.Nil(t, persister, "persister expected to be nil, but got %s", persister)
}

func TestCreateDBFromConfWrongFileNameLvlDB(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	arg := storageUnit.ArgDB{
		DBType:            storageUnit.LvlDB,
		Path:              "",
		BatchDelaySeconds: 10,
		MaxBatchSize:      10,
		MaxOpenFiles:      10,
	}
	persister, err := storageUnit.NewDB(arg)
	assert.NotNil(t, err, "error expected")
	assert.Nil(t, persister, "persister expected to be nil, but got %s", persister)
}

func TestCreateDBFromConfLvlDBOk(t *testing.T) {
	dir, _ := ioutil.TempDir("", "leveldb_temp")
	arg := storageUnit.ArgDB{
		DBType:            storageUnit.LvlDB,
		Path:              dir,
		BatchDelaySeconds: 10,
		MaxBatchSize:      10,
		MaxOpenFiles:      10,
	}
	persister, err := storageUnit.NewDB(arg)
	assert.Nil(t, err, "no error expected")
	assert.NotNil(t, persister, "valid persister expected but got nil")

	err = persister.Destroy()
	assert.Nil(t, err, "no error expected destroying the persister")
}

func TestNewStorageUnit_FromConfWrongCacheSizeVsBatchSize(t *testing.T) {

	storer, err := storageUnit.NewStorageUnitFromConf(storageUnit.CacheConfig{
		Capacity: 10,
		Type:     storageUnit.LRUCache,
	}, storageUnit.DBConfig{
		FilePath:          "Blocks",
		Type:              storageUnit.LvlDB,
		MaxBatchSize:      11,
		BatchDelaySeconds: 1,
		MaxOpenFiles:      10,
	})

	assert.NotNil(t, err, "error expected")
	assert.Nil(t, storer, "storer expected to be nil but got %s", storer)
}

func TestNewStorageUnit_FromConfWrongCacheConfig(t *testing.T) {

	storer, err := storageUnit.NewStorageUnitFromConf(storageUnit.CacheConfig{
		Capacity: 10,
		Type:     "NotLRU",
	}, storageUnit.DBConfig{
		FilePath:          "Blocks",
		Type:              storageUnit.LvlDB,
		BatchDelaySeconds: 1,
		MaxBatchSize:      1,
		MaxOpenFiles:      10,
	})

	assert.NotNil(t, err, "error expected")
	assert.Nil(t, storer, "storer expected to be nil but got %s", storer)
}

func TestNewStorageUnit_FromConfWrongDBConfig(t *testing.T) {
	storer, err := storageUnit.NewStorageUnitFromConf(storageUnit.CacheConfig{
		Capacity: 10,
		Type:     storageUnit.LRUCache,
	}, storageUnit.DBConfig{
		FilePath: "Blocks",
		Type:     "NotLvlDB",
	})

	assert.NotNil(t, err, "error expected")
	assert.Nil(t, storer, "storer expected to be nil but got %s", storer)
}

func TestNewStorageUnit_FromConfLvlDBOk(t *testing.T) {
	storer, err := storageUnit.NewStorageUnitFromConf(storageUnit.CacheConfig{
		Capacity: 10,
		Type:     storageUnit.LRUCache,
	}, storageUnit.DBConfig{
		FilePath:          "Blocks",
		Type:              storageUnit.LvlDB,
		MaxBatchSize:      1,
		BatchDelaySeconds: 1,
		MaxOpenFiles:      10,
	})

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, storer, "valid storer expected but got nil")
	err = storer.DestroyUnit()
	assert.Nil(t, err, "no error expected destroying the persister")
}

func TestNewStorageUnit_ShouldWorkLvlDB(t *testing.T) {
	storer, err := storageUnit.NewStorageUnitFromConf(storageUnit.CacheConfig{
		Capacity: 10,
		Type:     storageUnit.LRUCache,
	}, storageUnit.DBConfig{
		FilePath:          "Blocks",
		Type:              storageUnit.LvlDB,
		BatchDelaySeconds: 1,
		MaxBatchSize:      1,
		MaxOpenFiles:      10,
	})

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, storer, "valid storer expected but got nil")
	err = storer.DestroyUnit()
	assert.Nil(t, err, "no error expected destroying the persister")
}

const (
	valuesInDb = 100000
)

func BenchmarkStorageUnit_Put(b *testing.B) {
	b.StopTimer()
	s := initStorageUnit(b, 1)
	defer func() {
		err := s.DestroyUnit()
		logError(err)
	}()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nr := rand.Intn(valuesInDb)
		b.StartTimer()

		err := s.Put([]byte(strconv.Itoa(nr)), []byte(strconv.Itoa(nr)))
		logError(err)
	}
}

func BenchmarkStorageUnit_GetWithDataBeingPresent(b *testing.B) {
	b.StopTimer()
	s := initStorageUnit(b, 1)
	defer func() {
		err := s.DestroyUnit()
		logError(err)
	}()
	for i := 0; i < valuesInDb; i++ {
		err := s.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
		logError(err)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nr := rand.Intn(valuesInDb)
		b.StartTimer()

		_, err := s.Get([]byte(strconv.Itoa(nr)))
		logError(err)
	}
}

func BenchmarkStorageUnit_GetWithDataNotBeingPresent(b *testing.B) {
	b.StopTimer()
	s := initStorageUnit(b, 1)
	defer func() {
		err := s.DestroyUnit()
		logError(err)
	}()
	for i := 0; i < valuesInDb; i++ {
		err := s.Put([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
		logError(err)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		nr := rand.Intn(valuesInDb) + valuesInDb
		b.StartTimer()

		_, err := s.Get([]byte(strconv.Itoa(nr)))
		logError(err)
	}
}
