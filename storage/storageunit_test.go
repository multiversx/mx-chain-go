package storage_test

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/fnv"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing/keccak"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/bloom"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/leveldb"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"

	"github.com/stretchr/testify/assert"
)

func logError(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}

func initStorageUnitWithBloomFilter(t *testing.T, cSize int) *storage.Unit {
	mdb, err1 := memorydb.New()
	cache, err2 := lrucache.NewCache(cSize)
	bf := bloom.NewDefaultFilter()

	assert.Nil(t, err1, "failed creating db: %s", err1)
	assert.Nil(t, err2, "no error expected but got %s", err2)

	sUnit, err := storage.NewStorageUnit(cache, mdb, bf)

	assert.Nil(t, err, "failed to create storage unit")

	return sUnit
}

func initStorageUnitWithNilBloomFilter(t *testing.T, cSize int) *storage.Unit {
	mdb, err1 := memorydb.New()
	cache, err2 := lrucache.NewCache(cSize)

	assert.Nil(t, err1, "failed creating db: %s", err1)
	assert.Nil(t, err2, "no error expected but got %s", err2)

	sUnit, err := storage.NewStorageUnit(cache, mdb, nil)

	assert.Nil(t, err, "failed to create storage unit")

	return sUnit
}

func TestStorageUnitNilPersister(t *testing.T) {
	cache, err1 := lrucache.NewCache(10)
	bf := bloom.NewDefaultFilter()

	assert.Nil(t, err1, "no error expected but got %s", err1)

	_, err := storage.NewStorageUnit(cache, nil, bf)

	assert.NotNil(t, err, "expected failure")
}

func TestStorageUnitNilCacher(t *testing.T) {
	mdb, err1 := memorydb.New()
	bf := bloom.NewDefaultFilter()

	assert.Nil(t, err1, "failed creating db")

	_, err1 = storage.NewStorageUnit(nil, mdb, bf)

	assert.NotNil(t, err1, "expected failure")
}

func TestStorageUnitNilBloomFilter(t *testing.T) {
	cache, err1 := lrucache.NewCache(10)
	mdb, err2 := memorydb.New()

	assert.Nil(t, err1, "no error expected but got %s", err1)
	assert.Nil(t, err2, "failed creating db")

	_, err := storage.NewStorageUnit(cache, mdb, nil)

	assert.Nil(t, err, "did not expect failure")
}

func TestPutNotPresent(t *testing.T) {
	key, val := []byte("key0"), []byte("value0")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	has, err := s.Has(key)

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.True(t, has, "expected to find key %s, but not found", key)
}

func TestPutNotPresentWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key0"), []byte("value0")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	has, err := s.Has(key)

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.True(t, has, "expected to find key %s, but not found", key)
}

func TestPutNotPresentCache(t *testing.T) {
	key, val := []byte("key1"), []byte("value1")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	s.ClearCache()

	has, err := s.Has(key)

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.True(t, has, "expected to find key %s, but not found", key)
}

func TestPutNotPresentCacheWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key1"), []byte("value1")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	s.ClearCache()

	has, err := s.Has(key)

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.True(t, has, "expected to find key %s, but not found", key)
}

func TestPutPresent(t *testing.T) {
	key, val := []byte("key2"), []byte("value2")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	// put again same value, no error expected
	err = s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestPutPresentWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key2"), []byte("value2")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	// put again same value, no error expected
	err = s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestGetNotPresent(t *testing.T) {
	key := []byte("key3")
	s := initStorageUnitWithBloomFilter(t, 10)
	v, err := s.Get(key)

	assert.NotNil(t, err, "expected to find no value, but found %s", v)
}

func TestGetNotPresentWithNilBloomFilter(t *testing.T) {
	key := []byte("key3")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	v, err := s.Get(key)

	assert.NotNil(t, err, "expected to find no value, but found %s", v)
}

func TestGetNotPresentCache(t *testing.T) {
	key, val := []byte("key4"), []byte("value4")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	s.ClearCache()

	v, err := s.Get(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.Equal(t, val, v, "expected %s but got %s", val, v)
}

func TestGetNotPresentCacheWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key4"), []byte("value4")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	s.ClearCache()

	v, err := s.Get(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.Equal(t, val, v, "expected %s but got %s", val, v)
}

func TestGetPresent(t *testing.T) {
	key, val := []byte("key5"), []byte("value4")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	v, err := s.Get(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.Equal(t, val, v, "expected %s but got %s", val, v)
}

func TestGetPresentWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key5"), []byte("value4")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	v, err := s.Get(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.Equal(t, val, v, "expected %s but got %s", val, v)
}

func TestHasNotPresent(t *testing.T) {
	key := []byte("key6")
	s := initStorageUnitWithBloomFilter(t, 10)
	has, err := s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.False(t, has, "not expected to find value")
}

func TestHasNotPresentWithNilBloomFilter(t *testing.T) {
	key := []byte("key6")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	has, err := s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.False(t, has, "not expected to find value")
}

func TestHasNotPresentCache(t *testing.T) {
	key, val := []byte("key7"), []byte("value7")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	s.ClearCache()

	has, err := s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.True(t, has, "expected to find key but not found")
}

func TestHasNotPresentCacheWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key7"), []byte("value7")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	s.ClearCache()

	has, err := s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.True(t, has, "expected to find key but not found")
}

func TestHasPresent(t *testing.T) {
	key, val := []byte("key8"), []byte("value8")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	has, err := s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.True(t, has, "expected to find key but not found")
}

func TestHasPresentWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key8"), []byte("value8")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)

	assert.Nil(t, err, "no error expected but got %s", err)

	has, err := s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.True(t, has, "expected to find key but not found")
}

func TestHasOrAddNotPresent(t *testing.T) {
	key, val := []byte("key9"), []byte("value9")
	s := initStorageUnitWithBloomFilter(t, 10)
	has, err := s.HasOrAdd(key, val)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.False(t, has, "not expected to find value")

	has, err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.True(t, has, "expected to find key but not found")
}

func TestHasOrAddNotPresentWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key9"), []byte("value9")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	has, err := s.HasOrAdd(key, val)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.False(t, has, "not expected to find value")

	has, err = s.Has(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.True(t, has, "expected to find key but not found")
}

func TestHasOrAddNotPresentCache(t *testing.T) {
	key, val := []byte("key10"), []byte("value10")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)

	s.ClearCache()

	has, err := s.HasOrAdd(key, val)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.True(t, has, "expected to find value")
}

func TestHasOrAddNotPresentCacheWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key10"), []byte("value10")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)

	s.ClearCache()

	has, err := s.HasOrAdd(key, val)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.True(t, has, "expected to find value")
}

func TestHasOrAddPresent(t *testing.T) {
	key, val := []byte("key11"), []byte("value11")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)

	has, err := s.HasOrAdd(key, val)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.True(t, has, "expected to find value")
}

func TestHasOrAddPresentWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key11"), []byte("value11")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Put(key, val)

	has, err := s.HasOrAdd(key, val)

	assert.Nil(t, err, "expected no error, but got %s", err)
	assert.True(t, has, "expected to find value")
}

func TestDeleteNotPresent(t *testing.T) {
	key := []byte("key12")
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Remove(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
}

func TestDeleteNotPresentWithNilBloomFilter(t *testing.T) {
	key := []byte("key12")
	s := initStorageUnitWithNilBloomFilter(t, 10)
	err := s.Remove(key)

	assert.Nil(t, err, "expected no error, but got %s", err)
}

func TestDeleteNotPresentCache(t *testing.T) {
	key, val := []byte("key13"), []byte("value13")
	s := initStorageUnitWithBloomFilter(t, 10)
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

func TestDeleteNotPresentCacheWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key13"), []byte("value13")
	s := initStorageUnitWithNilBloomFilter(t, 10)
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
	s := initStorageUnitWithBloomFilter(t, 10)
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

func TestDeletePresentWithNilBloomFilter(t *testing.T) {
	key, val := []byte("key14"), []byte("value14")
	s := initStorageUnitWithNilBloomFilter(t, 10)
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
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.Put(key, val)
	assert.Nil(t, err, "Could not put value in storage unit")
	s.ClearCache()

	has, err := s.Has(key)

	assert.Nil(t, err, "no error expected, but got %s", err)
	assert.True(t, has, "expected to find key")
}

func TestDestroyUnitNoError(t *testing.T) {
	s := initStorageUnitWithBloomFilter(t, 10)
	err := s.DestroyUnit()
	assert.Nil(t, err, "no error expected, but got %s", err)
}

func TestDestroyUnitWithNilBloomFilterNoError(t *testing.T) {
	s := initStorageUnitWithNilBloomFilter(t, 10)
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

func TestCreateBloomFilterFromConfWrongSize(t *testing.T) {
	bfConfig := storage.BloomConfig{
		Size:     2,
		HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv},
	}

	bf, err := storage.NewBloomFilter(bfConfig)

	assert.NotNil(t, err, "error expected")
	assert.Nil(t, bf, "persister expected to be nil, but got %s", bf)
}

func TestCreateBloomFilterFromConfWrongHashFunc(t *testing.T) {
	bfConfig := storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{},
	}

	bf, err := storage.NewBloomFilter(bfConfig)

	assert.NotNil(t, err, "error expected")
	assert.Nil(t, bf, "persister expected to be nil, but got %s", bf)
}

func TestCreateBloomFilterFromConfOk(t *testing.T) {
	bfConfig := storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv},
	}

	bf, err := storage.NewBloomFilter(bfConfig)

	assert.Nil(t, err, "no error expected")
	assert.NotNil(t, bf, "valid persister expected but got nil")
}

func TestNewStorageUnitFromConfWrongCacheConfig(t *testing.T) {

	storer, err := storage.NewStorageUnitFromConf(storage.CacheConfig{
		Size: 10,
		Type: "NotLRU",
	}, storage.DBConfig{
		FilePath: "Blocks",
		Type:     storage.LvlDB,
	}, storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv},
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
	}, storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv},
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
	}, storage.BloomConfig{
		Size:     2048,
		HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv},
	})

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, storer, "valid storer expected but got nil")
	err = storer.DestroyUnit()
	assert.Nil(t, err, "no error expected destroying the persister")
}

const (
	valuesInDb = 100000
	bfSize     = 100000
)

func initSUWithNilBloomFilter(cSize int) *storage.Unit {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
	}

	ldb, err1 := leveldb.NewDB(dir + "/levelDB")
	cache, err2 := lrucache.NewCache(cSize)

	if err1 != nil {
		fmt.Println(err1)
	}
	if err2 != nil {
		fmt.Println(err2)
	}

	sUnit, err := storage.NewStorageUnit(cache, ldb, nil)

	if err != nil {
		fmt.Println(err)
	}

	return sUnit

}

func initSUWithBloomFilter(cSize int, bfSize uint) *storage.Unit {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
	}

	ldb, err1 := leveldb.NewDB(dir + "/levelDB")
	cache, err2 := lrucache.NewCache(cSize)
	bf, err3 := bloom.NewFilter(bfSize, []hashing.Hasher{keccak.Keccak{}, blake2b.Blake2b{}, fnv.Fnv{}})

	if err1 != nil {
		fmt.Println(err1)
	}
	if err2 != nil {
		fmt.Println(err2)
	}
	if err3 != nil {
		fmt.Println(err3)
	}

	sUnit, err := storage.NewStorageUnit(cache, ldb, bf)

	if err != nil {
		fmt.Println(err)
	}

	return sUnit
}

func BenchmarkStorageUnitPutWithNilBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithNilBloomFilter(1)
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

func BenchmarkStorageUnitPutWithBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithBloomFilter(1, bfSize)
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

func BenchmarkStorageUnitGetPresentWithNilBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithNilBloomFilter(1)
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

func BenchmarkStorageUnitGetPresentWithBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithBloomFilter(1, bfSize)
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

func BenchmarkStorageUnitGetNotPresentWithNilBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithNilBloomFilter(1)
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

func BenchmarkStorageUnitGetNotPresentWithBloomFilter(b *testing.B) {
	b.StopTimer()
	s := initSUWithBloomFilter(1, bfSize)
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
