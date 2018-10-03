package leveldb_test

import (
	"ElrondNetwork/elrond-go-sandbox/storage"
	"ElrondNetwork/elrond-go-sandbox/storage/leveldb"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestLevelDB(t *testing.T) {
	dir, err := ioutil.TempDir("", "leveldb_temp")
	lvdb, err := leveldb.NewDB(dir)

	if err != nil {
		t.Fatalf("Failed creating leveldb database file")
	}

	Suite(t, lvdb)
}

func Suite(t *testing.T, p storage.Persister) {
	TestingInitNoError(t, p)
	TestingPutNoError(t, p)
	TestingGetPresent(t, p)
	TestingGetNotPresent(t, p)
	TestingHasPresent(t, p)
	TestingHasNotPresent(t, p)
	TestingDeletePresent(t, p)
	TestingDeleteNotPresent(t, p)
	TestingClose(t, p)
	TestingDestroy(t, p)
}

func TestingInitNoError(t *testing.T, p storage.Persister) {
	err := p.Init()
	if err != nil {
		t.Fatalf("error initializing db")
	}
}

func TestingPutNoError(t *testing.T, p storage.Persister) {
	key, val := []byte("key"), []byte("value")

	err := p.Put(key, val)

	if err != nil {
		t.Errorf("error saving in db")
	}
}

func TestingGetPresent(t *testing.T, p storage.Persister) {
	key, val := []byte("key1"), []byte("value1")

	err := p.Put(key, val)

	if err != nil {
		t.Errorf("error saving in db")
	}

	v, err := p.Get(key)

	if err != nil {
		t.Errorf(err.Error())
	}

	if !reflect.DeepEqual(v, val) {
		t.Errorf("read:%s but expected: %s", v, val)
	}
}

func TestingGetNotPresent(t *testing.T, p storage.Persister) {
	key := []byte("key2")

	v, err := p.Get(key)

	if err == nil {
		t.Errorf("error expected but got nil, value %s", v)
	}
}

func TestingHasPresent(t *testing.T, p storage.Persister) {
	key, val := []byte("key3"), []byte("value3")

	err := p.Put(key, val)

	if err != nil {
		t.Errorf("error saving in db")
	}

	has, err := p.Has(key)

	if err != nil {
		t.Errorf(err.Error())
	}

	if !has {
		t.Errorf("value expected but not found")
	}
}

func TestingHasNotPresent(t *testing.T, p storage.Persister) {
	key := []byte("key4")

	has, err := p.Has(key)

	if err != nil {
		t.Errorf("no error expected but got %s", err)
	}

	if has {
		t.Errorf("value not expected but found")
	}
}

func TestingDeletePresent(t *testing.T, p storage.Persister) {
	key, val := []byte("key5"), []byte("value5")

	err := p.Put(key, val)

	if err != nil {
		t.Errorf("error saving in db")
	}

	err = p.Delete(key)

	if err != nil {
		t.Errorf("no error expected but got %s", err)
	}

	has, err := p.Has(key)

	if has {
		t.Errorf("element not expected as already deleted")
	}
}

func TestingDeleteNotPresent(t *testing.T, p storage.Persister) {
	key := []byte("key6")

	err := p.Delete(key)

	if err != nil {
		t.Errorf("no error expected but got %s", err)
	}
}

func TestingClose(t *testing.T, p storage.Persister) {
	err := p.Close()

	if err != nil {
		t.Errorf("no error expected but got %s", err)
	}
}

func TestingDestroy(t *testing.T, p storage.Persister) {
	err := p.Destroy()
	if err != nil {
		t.Errorf("no error expected but got %s", err)
	}
}
