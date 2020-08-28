package leveldb_test

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/leveldb"
	"github.com/stretchr/testify/assert"
)

func createSerialLevelDb(t *testing.T, batchDelaySeconds int, maxBatchSize int, maxOpenFiles int) (p *leveldb.SerialDB) {
	dir, _ := ioutil.TempDir("", "leveldb_temp")
	lvdb, err := leveldb.NewSerialDB(dir, batchDelaySeconds, maxBatchSize, maxOpenFiles)

	assert.Nil(t, err, "Failed creating leveldb database file")
	return lvdb
}

func TestSerialDB_InitNoError(t *testing.T) {
	ldb := createSerialLevelDb(t, 10, 1, 10)

	err := ldb.Init()

	assert.Nil(t, err, "error initializing DB")
}

func TestSerialDB_PutNoError(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createSerialLevelDb(t, 10, 1, 10)

	err := ldb.Put(key, val)

	assert.Nil(t, err, "error saving in DB")
}

func TestSerialDB_GetErrorAfterPutBeforeTimeout(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createSerialLevelDb(t, 1, 100, 10)

	_ = ldb.Put(key, val)
	v, err := ldb.Get(key)

	assert.Equal(t, val, v)
	assert.Nil(t, err)
}

func TestSerialDB_GetErrorOnFail(t *testing.T) {
	ldb := createSerialLevelDb(t, 10, 1, 10)
	_ = ldb.Destroy()

	v, err := ldb.Get([]byte("key"))
	assert.Nil(t, v)
	assert.NotNil(t, err)
}

func TestSerialDB_CallsNotBlockingAfterCloseOrDestroy(t *testing.T) {
	ldb := createSerialLevelDb(t, 10, 1, 10)
	_ = ldb.Destroy()

	_, err := ldb.Get([]byte("key"))
	assert.Equal(t, storage.ErrSerialDBIsClosed, err)

	err = ldb.Has([]byte("key"))
	assert.Equal(t, storage.ErrSerialDBIsClosed, err)

	err = ldb.Remove([]byte("key"))
	assert.Equal(t, storage.ErrSerialDBIsClosed, err)

	err = ldb.Put([]byte("key"), []byte("val"))
	assert.Equal(t, storage.ErrSerialDBIsClosed, err)
}

func TestSerialDB_GetOKAfterPutWithTimeout(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createSerialLevelDb(t, 1, 100, 10)

	_ = ldb.Put(key, val)
	time.Sleep(time.Second * 3)
	v, err := ldb.Get(key)

	assert.Nil(t, err)
	assert.Equal(t, val, v)
}

func TestSerialDB_RemoveBeforeTimeoutOK(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createSerialLevelDb(t, 1, 100, 10)

	_ = ldb.Put(key, val)
	_ = ldb.Remove(key)
	time.Sleep(time.Second * 2)
	v, err := ldb.Get(key)

	assert.Nil(t, v)
	assert.Equal(t, storage.ErrKeyNotFound, err)
}

func TestSerialDB_RemoveAfterTimeoutOK(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createSerialLevelDb(t, 1, 100, 10)

	_ = ldb.Put(key, val)
	time.Sleep(time.Second * 2)
	_ = ldb.Remove(key)
	v, err := ldb.Get(key)

	assert.Nil(t, v)
	assert.Equal(t, storage.ErrKeyNotFound, err)
}

func TestSerialDB_GetPresent(t *testing.T) {
	key, val := []byte("key1"), []byte("value1")
	ldb := createSerialLevelDb(t, 10, 1, 10)

	_ = ldb.Put(key, val)
	v, err := ldb.Get(key)

	assert.Nil(t, err, "error not expected, but got %s", err)
	assert.Equalf(t, v, val, "read:%s but expected: %s", v, val)
}

func TestSerialDB_GetNotPresent(t *testing.T) {
	key := []byte("key2")
	ldb := createSerialLevelDb(t, 10, 1, 10)

	v, err := ldb.Get(key)

	assert.NotNil(t, err, "error expected but got nil, value %s", v)
}

func TestSerialDB_HasPresent(t *testing.T) {
	key, val := []byte("key3"), []byte("value3")
	ldb := createSerialLevelDb(t, 10, 1, 10)

	_ = ldb.Put(key, val)
	err := ldb.Has(key)

	assert.Nil(t, err)
}

func TestSerialDB_HasNotPresent(t *testing.T) {
	key := []byte("key4")
	ldb := createSerialLevelDb(t, 10, 1, 10)

	err := ldb.Has(key)

	assert.NotNil(t, err)
	assert.Equal(t, err, storage.ErrKeyNotFound)
}

func TestSerialDB_RemovePresent(t *testing.T) {
	key, val := []byte("key5"), []byte("value5")
	ldb := createSerialLevelDb(t, 10, 1, 10)

	_ = ldb.Put(key, val)
	_ = ldb.Remove(key)
	err := ldb.Has(key)

	assert.NotNil(t, err)
	assert.Equal(t, err, storage.ErrKeyNotFound)
}

func TestSerialDB_RemoveNotPresent(t *testing.T) {
	key := []byte("key6")
	ldb := createSerialLevelDb(t, 10, 1, 10)

	err := ldb.Remove(key)

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestSerialDB_Close(t *testing.T) {
	ldb := createSerialLevelDb(t, 10, 1, 10)

	err := ldb.Close()

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestSerialDB_CloseTwice(t *testing.T) {
	ldb := createSerialLevelDb(t, 10, 1, 10)

	_ = ldb.Close()
	err := ldb.Close()

	assert.Nil(t, err)
}

func TestSerialDB_Destroy(t *testing.T) {
	ldb := createSerialLevelDb(t, 10, 1, 10)

	err := ldb.Destroy()

	assert.Nil(t, err, "no error expected but got %s", err)
}
