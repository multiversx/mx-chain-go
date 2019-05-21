package badgerdb_test

import (
	"io/ioutil"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/storage/badgerdb"
	"github.com/stretchr/testify/assert"
)

func createBadgerDb(t *testing.T) (p *badgerdb.DB) {
	dir, err := ioutil.TempDir("", "leveldb_temp")
	bDB, err := badgerdb.NewDB(dir)

	assert.Nil(t, err, "Failed creating leveldb database file")
	return bDB
}

func TestInitNoError(t *testing.T) {
	ldb := createBadgerDb(t)

	err := ldb.Init()

	assert.Nil(t, err, "error initializing db")
}

func TestPutNoError(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createBadgerDb(t)

	err := ldb.Put(key, val)

	assert.Nil(t, err, "error saving in db")
}

func TestGetPresent(t *testing.T) {
	key, val := []byte("key1"), []byte("value1")
	ldb := createBadgerDb(t)

	err := ldb.Put(key, val)

	assert.Nil(t, err, "error saving in db")

	v, err := ldb.Get(key)

	assert.Nil(t, err, "error not expected, but got %s", err)
	assert.Equalf(t, v, val, "read:%s but expected: %s", v, val)
}

func TestGetNotPresent(t *testing.T) {
	key := []byte("key2")
	ldb := createBadgerDb(t)

	v, err := ldb.Get(key)

	assert.NotNil(t, err, "error expected but got nil, value %s", v)
}

func TestHasPresent(t *testing.T) {
	key, val := []byte("key3"), []byte("value3")
	ldb := createBadgerDb(t)

	err := ldb.Put(key, val)

	assert.Nil(t, err, "error saving in db")

	has, err := ldb.Has(key)

	assert.Nil(t, err, "error not expected but got %s", err)
	assert.True(t, has, "value expected but not found")
}

func TestHasNotPresent(t *testing.T) {
	key := []byte("key4")
	ldb := createBadgerDb(t)

	has, err := ldb.Has(key)

	assert.Contains(t, err.Error(), "Key not found")
	assert.False(t, has, "value not expected but found")
}

func TestRemovePresent(t *testing.T) {
	key, val := []byte("key5"), []byte("value5")
	ldb := createBadgerDb(t)

	err := ldb.Put(key, val)

	assert.Nil(t, err, "error saving in db")

	err = ldb.Remove(key)

	assert.Nil(t, err, "no error expected but got %s", err)

	has, err := ldb.Has(key)

	assert.False(t, has, "element not expected as already deleted")
}

func TestRemoveNotPresent(t *testing.T) {
	key := []byte("key6")
	ldb := createBadgerDb(t)

	err := ldb.Remove(key)

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestClose(t *testing.T) {
	ldb := createBadgerDb(t)

	err := ldb.Close()

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestDestroy(t *testing.T) {
	ldb := createBadgerDb(t)

	err := ldb.Destroy()

	assert.Nil(t, err, "no error expected but got %s", err)
}
