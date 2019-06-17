package leveldb_test

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/leveldb"
	"github.com/stretchr/testify/assert"
)

func createSerialLevelDb(t *testing.T, batchDelaySeconds int, maxBatchSize int) (p *leveldb.SerialDB) {
	dir, err := ioutil.TempDir("", "leveldb_temp")
	lvdb, err := leveldb.NewSerialDB(dir, batchDelaySeconds, maxBatchSize)

	assert.Nil(t, err, "Failed creating leveldb database file")
	return lvdb
}

func TestSerialDB_InitNoError(t *testing.T) {
	ldb := createSerialLevelDb(t, 10, 1)

	err := ldb.Init()

	assert.Nil(t, err, "error initializing db")
}

func TestSerialDB_PutNoError(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createSerialLevelDb(t, 10, 1)

	err := ldb.Put(key, val)

	assert.Nil(t, err, "error saving in db")
}

func TestSerialDB_GetErrorAfterPutBeforeTimeout(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createSerialLevelDb(t, 1, 100)

	err := ldb.Put(key, val)
	assert.Nil(t, err)
	v, err := ldb.Get(key)
	assert.Nil(t, v)
	assert.Equal(t, storage.ErrKeyNotFound, err)
}

func TestSerialDB_GetOKAfterPutWithTimeout(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createSerialLevelDb(t, 1, 100)

	err := ldb.Put(key, val)
	assert.Nil(t, err)
	time.Sleep(time.Second * 3)

	v, err := ldb.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, val, v)
}

func TestSerialDB_RemoveBeforeTimeoutOK(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createSerialLevelDb(t, 1, 100)

	err := ldb.Put(key, val)
	assert.Nil(t, err)

	_ = ldb.Remove(key)
	time.Sleep(time.Second * 2)

	v, err := ldb.Get(key)
	assert.Nil(t, v)
	assert.Equal(t, storage.ErrKeyNotFound, err)
}

func TestSerialDB_RemoveAfterTimeoutOK(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createSerialLevelDb(t, 1, 100)

	err := ldb.Put(key, val)
	assert.Nil(t, err)
	time.Sleep(time.Second * 2)

	_ = ldb.Remove(key)

	v, err := ldb.Get(key)
	assert.Nil(t, v)
	assert.Equal(t, storage.ErrKeyNotFound, err)
}

func TestSerialDB_GetPresent(t *testing.T) {
	key, val := []byte("key1"), []byte("value1")
	ldb := createSerialLevelDb(t, 10, 1)

	err := ldb.Put(key, val)

	assert.Nil(t, err, "error saving in db")

	v, err := ldb.Get(key)

	assert.Nil(t, err, "error not expected, but got %s", err)
	assert.Equalf(t, v, val, "read:%s but expected: %s", v, val)
}

func TestSerialDB_GetNotPresent(t *testing.T) {
	key := []byte("key2")
	ldb := createSerialLevelDb(t, 10, 1)

	v, err := ldb.Get(key)

	assert.NotNil(t, err, "error expected but got nil, value %s", v)
}

func TestSerialDB_HasPresent(t *testing.T) {
	key, val := []byte("key3"), []byte("value3")
	ldb := createSerialLevelDb(t, 10, 1)

	err := ldb.Put(key, val)

	assert.Nil(t, err, "error saving in db")

	err = ldb.Has(key)

	assert.Nil(t, err)
}

func TestSerialDB_HasNotPresent(t *testing.T) {
	key := []byte("key4")
	ldb := createSerialLevelDb(t, 10, 1)

	err := ldb.Has(key)

	assert.NotNil(t, err)
	assert.Equal(t, err, storage.ErrKeyNotFound)
}

func TestSerialDB_RemovePresent(t *testing.T) {
	key, val := []byte("key5"), []byte("value5")
	ldb := createSerialLevelDb(t, 10, 1)

	err := ldb.Put(key, val)

	assert.Nil(t, err, "error saving in db")

	err = ldb.Remove(key)

	assert.Nil(t, err, "no error expected but got %s", err)

	err = ldb.Has(key)

	assert.NotNil(t, err)
	assert.Equal(t, err, storage.ErrKeyNotFound)
}

func TestSerialDB_RemoveNotPresent(t *testing.T) {
	key := []byte("key6")
	ldb := createSerialLevelDb(t, 10, 1)

	err := ldb.Remove(key)

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestSerialDB_Close(t *testing.T) {
	ldb := createSerialLevelDb(t, 10, 1)
	err := ldb.Close()
	<-time.After(10 * time.Millisecond)

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestSerialDB_CloseTwice(t *testing.T) {
	ldb := createSerialLevelDb(t, 10, 1)
	err := ldb.Close()
	<-time.After(time.Millisecond)
	err = ldb.Close()

	assert.Nil(t, err)
}

func TestSerialDB_Destroy(t *testing.T) {
	ldb := createSerialLevelDb(t, 10, 1)
	err := ldb.Destroy()
	<-time.After(time.Millisecond)

	assert.Nil(t, err, "no error expected but got %s", err)
}
