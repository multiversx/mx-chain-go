package storage_test

import (
	"ElrondNetwork/elrond-go-sandbox/storage"
	"ElrondNetwork/elrond-go-sandbox/storage/leveldb"
	"ElrondNetwork/elrond-go-sandbox/storage/lrucache"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestStorageUnitNilPersister(t *testing.T) {
	cache, err := lrucache.NewCache(10)

	if err != nil {
		t.Errorf("no error expected but got %s", err)
	}

	_, err = storage.NewStorageUnit(cache, nil)

	if err == nil {
		t.Errorf("expected failure")
	}
}

func TestStorageUnitNilCacher(t *testing.T) {
	dir, err := ioutil.TempDir("", "leveldb_temp")
	lvdb, err1 := leveldb.NewDB(dir)
	if err1 != nil {
		t.Fatalf("failed creating leveldb database file")
	}

	_, err = storage.NewStorageUnit(nil, lvdb)

	if err == nil {
		t.Errorf("expected failure")
	}
}

func TestStorageUnit(t *testing.T) {
	dir, err := ioutil.TempDir("", "leveldb_temp")
	lvdb, err1 := leveldb.NewDB(dir)
	cache, err2 := lrucache.NewCache(10)

	if err1 != nil {
		t.Fatalf("failed creating leveldb database file")
	}

	if err2 != nil {
		t.Errorf("no error expected but got %s", err)
	}

	sUnit, err := storage.NewStorageUnit(cache, lvdb)

	if err != nil {
		t.Errorf("failed to create storage unit")
	}

	Suite(t, sUnit)
}

func Suite(t *testing.T, s *storage.StorageUnit) {
	TestingPutNotPresent(t, s)
	TestingPutNotPresentCache(t, s)
	TestingPutPresent(t, s)
	TestingGetNotPresent(t, s)
	TestingGetNotPresentCache(t, s)
	TestingGetPresent(t, s)
	TestingContainsNotPresent(t, s)
	TestingContainsNotPresentCache(t, s)
	TestingContainsPresent(t, s)
	TestingContainsOrAddNotPresent(t, s)
	TestingContainsOrAddNotPresentCache(t, s)
	TestingContainsOrAddPresent(t, s)
	TestingDeleteNotPresent(t, s)
	TestingDeleteNotPresentCache(t, s)
	TestingDeletePresent(t, s)
	TestingClearCacheNotAffectPersist(t, s)
	TestingDestroyUnitNoError(t, s)
}

func TestingPutNotPresent(t *testing.T, s *storage.StorageUnit) {
	key, val := []byte("key0"), []byte("value0")

	err := s.Put(key, val)

	if err != nil {
		t.Errorf("no error expected but got %s", err)
	}

	has, err := s.Contains(key)

	if err != nil {
		t.Errorf("no error expected but got %s", err)
	}

	if !has {
		t.Errorf("expected to find key %s, but not found", key)
	}
}

func TestingPutNotPresentCache(t *testing.T, s *storage.StorageUnit) {
	key, val := []byte("key1"), []byte("value1")

	err := s.Put(key, val)

	if err != nil {
		t.Errorf("no error expected but got %s", err)
	}

	s.ClearCache()

	has, err := s.Contains(key)

	if err != nil {
		t.Errorf("no error expected but got %s", err)
	}

	if !has {
		t.Errorf("expected to find key %s, but not found", key)
	}
}

func TestingPutPresent(t *testing.T, s *storage.StorageUnit) {
	key, val := []byte("key2"), []byte("value2")

	err := s.Put(key, val)

	if err != nil {
		t.Errorf("no error expected but got %s", err)
	}

	// put again same value, no error expected
	err = s.Put(key, val)

	if err != nil {
		t.Errorf("no error expected but got %s", err)
	}
}

func TestingGetNotPresent(t *testing.T, s *storage.StorageUnit) {
	key := []byte("key3")

	v, err := s.Get(key)

	if err == nil {
		t.Errorf("expected to find no value, but found %s", v)
	}
}

func TestingGetNotPresentCache(t *testing.T, s *storage.StorageUnit) {
	key, val := []byte("key4"), []byte("value4")

	err := s.Put(key, val)

	if err != nil {
		t.Errorf("no error expected but got %s", err)
	}

	s.ClearCache()

	v, err := s.Get(key)

	if err != nil {
		t.Errorf("expected no error, but got %s", err)
	}

	if !reflect.DeepEqual(v, val) {
		t.Errorf("expected %s but got %s", val, v)
	}
}

func TestingGetPresent(t *testing.T, s *storage.StorageUnit) {
	key, val := []byte("key5"), []byte("value4")

	err := s.Put(key, val)

	if err != nil {
		t.Errorf("no error expected but got %s", err)
	}

	v, err := s.Get(key)

	if err != nil {
		t.Errorf("expected no error, but got %s", err)
	}

	if !reflect.DeepEqual(v, val) {
		t.Errorf("expected %s but got %s", val, v)
	}
}

func TestingContainsNotPresent(t *testing.T, s *storage.StorageUnit) {
	key := []byte("key6")

	has, err := s.Contains(key)

	if err != nil {
		t.Errorf("expected no error, but got %s", err)
	}

	if has {
		t.Errorf("not expected to find value")
	}
}

func TestingContainsNotPresentCache(t *testing.T, s *storage.StorageUnit) {
	key, val := []byte("key7"), []byte("value7")

	err := s.Put(key, val)

	if err != nil {
		t.Errorf("no error expected but got %s", err)
	}

	s.ClearCache()

	has, err := s.Contains(key)

	if err != nil {
		t.Errorf("expected no error, but got %s", err)
	}

	if !has {
		t.Errorf("expected to find key but not found")
	}
}

func TestingContainsPresent(t *testing.T, s *storage.StorageUnit) {
	key, val := []byte("key8"), []byte("value8")

	err := s.Put(key, val)

	if err != nil {
		t.Errorf("no error expected but got %s", err)
	}

	has, err := s.Contains(key)

	if err != nil {
		t.Errorf("expected no error, but got %s", err)
	}

	if !has {
		t.Errorf("expected to find key but not found")
	}
}

func TestingContainsOrAddNotPresent(t *testing.T, s *storage.StorageUnit) {
	key, val := []byte("key9"), []byte("value9")

	has, err := s.ContainsOrAdd(key, val)

	if err != nil {
		t.Errorf("expected no error, but got %s", err)
	}

	if has {
		t.Errorf("not expected to find value")
	}

	has, err = s.Contains(key)

	if err != nil {
		t.Errorf("expected no error, but got %s", err)
	}

	if !has {
		t.Errorf("expected to find key but not found")
	}
}

func TestingContainsOrAddNotPresentCache(t *testing.T, s *storage.StorageUnit) {
	key, val := []byte("key10"), []byte("value10")

	err := s.Put(key, val)

	s.ClearCache()

	has, err := s.ContainsOrAdd(key, val)

	if err != nil {
		t.Errorf("expected no error, but got %s", err)
	}

	if !has {
		t.Errorf("expected to find value")
	}
}

func TestingContainsOrAddPresent(t *testing.T, s *storage.StorageUnit) {
	key, val := []byte("key11"), []byte("value11")

	err := s.Put(key, val)

	has, err := s.ContainsOrAdd(key, val)

	if err != nil {
		t.Errorf("expected no error, but got %s", err)
	}

	if !has {
		t.Errorf("expected to find value")
	}
}

func TestingDeleteNotPresent(t *testing.T, s *storage.StorageUnit) {
	key := []byte("key12")

	err := s.Delete(key)

	if err != nil {
		t.Errorf("expected no error, but got %s", err)
	}
}

func TestingDeleteNotPresentCache(t *testing.T, s *storage.StorageUnit) {
	key, val := []byte("key13"), []byte("value13")

	s.Put(key, val)

	has, err := s.Contains(key)

	if !has || err != nil {
		t.Errorf("expected to find key")
	}

	s.ClearCache()

	err = s.Delete(key)

	if err != nil {
		t.Errorf("expected no error, but got %s", err)
	}

	has, err = s.Contains(key)
	if has || err != nil {
		t.Errorf("not expected to find value or get error")
	}
}

func TestingDeletePresent(t *testing.T, s *storage.StorageUnit) {
	key, val := []byte("key14"), []byte("value14")

	s.Put(key, val)

	has, err := s.Contains(key)

	if !has || err != nil {
		t.Errorf("expected to find key")
	}

	err = s.Delete(key)

	if err != nil {
		t.Errorf("expected no error, but got %s", err)
	}

	has, err = s.Contains(key)
	if err != nil {
		t.Errorf("no error expected, but got %s")
	}

	if has {
		t.Errorf("not expected to find value")
	}
}

func TestingClearCacheNotAffectPersist(t *testing.T, s *storage.StorageUnit) {
	key, val := []byte("key15"), []byte("value15")

	s.Put(key, val)
	s.ClearCache()

	has, err := s.Contains(key)

	if err != nil {
		t.Errorf("no error expected, but got %s", err)
	}

	if !has {
		t.Errorf("expected to find key")
	}
}

func TestingDestroyUnitNoError(t *testing.T, s *storage.StorageUnit) {
	err := s.DestroyUnit()
	if err != nil {
		t.Errorf("no error expected, but got %s", err)
	}
}
