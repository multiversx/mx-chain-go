package leveldb_test

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/leveldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createLevelDb(t *testing.T, batchDelaySeconds int, maxBatchSize int, maxOpenFiles int) (p *leveldb.DB) {
	dir, _ := ioutil.TempDir("", "leveldb_temp")
	lvdb, err := leveldb.NewDB(dir, batchDelaySeconds, maxBatchSize, maxOpenFiles)

	assert.Nil(t, err, "Failed creating leveldb database file")
	return lvdb
}

func TestDB_InitNoError(t *testing.T) {
	ldb := createLevelDb(t, 10, 1, 10)

	err := ldb.Init()

	assert.Nil(t, err, "error initializing DB")
}

func TestDB_CorruptdeDBShouldRecover(t *testing.T) {
	dir, _ := ioutil.TempDir("", "leveldb_temp")
	defer func() {
		_ = os.RemoveAll(dir)
	}()
	db, err := leveldb.NewDB(dir, 10, 1, 10)
	require.Nil(t, err)

	key := []byte("key")
	val := []byte("val")
	err = db.Put(key, val)
	require.Nil(t, err)
	_ = db.Close()

	err = os.Remove(path.Join(dir, "MANIFEST-000000"))
	require.Nil(t, err)

	dbRecovered, err := leveldb.NewDB(dir, 10, 1, 10)
	if err != nil {
		assert.Fail(t, fmt.Sprintf("should have not errored %s", err.Error()))
		return
	}

	valRecovered, err := dbRecovered.Get(key)
	assert.Nil(t, err)
	_ = dbRecovered.Close()

	assert.Equal(t, val, valRecovered)
}

func TestDB_DoubleOpenShouldError(t *testing.T) {
	dir, _ := ioutil.TempDir("", "leveldb_temp")
	lvdb1, err := leveldb.NewDB(dir, 10, 1, 10)
	require.Nil(t, err)

	defer func() {
		_ = lvdb1.Close()
		_ = os.RemoveAll(dir)
	}()

	_, err = leveldb.NewDB(dir, 10, 1, 10)
	assert.NotNil(t, err)
}

func TestDB_DoubleOpenButClosedInTimeShouldWork(t *testing.T) {
	dir, _ := ioutil.TempDir("", "leveldb_temp")
	lvdb1, err := leveldb.NewDB(dir, 10, 1, 10)
	require.Nil(t, err)

	defer func() {
		_ = lvdb1.Close()
		_ = os.RemoveAll(dir)
	}()

	go func() {
		time.Sleep(time.Second * 3)
		_ = lvdb1.Close()
	}()

	lvdb2, err := leveldb.NewDB(dir, 10, 1, 10)
	assert.Nil(t, err)
	assert.NotNil(t, lvdb2)

	_ = lvdb2.Close()
}

func TestDB_PutNoError(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createLevelDb(t, 10, 1, 10)

	err := ldb.Put(key, val)

	assert.Nil(t, err, "error saving in DB")
}

func TestDB_GetErrorAfterPutBeforeTimeout(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createLevelDb(t, 1, 100, 10)

	err := ldb.Put(key, val)
	assert.Nil(t, err)
	v, err := ldb.Get(key)
	assert.Equal(t, val, v)
	assert.Nil(t, err)
}

func TestDB_GetOKAfterPutWithTimeout(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createLevelDb(t, 1, 100, 10)

	err := ldb.Put(key, val)
	assert.Nil(t, err)
	time.Sleep(time.Second * 3)

	v, err := ldb.Get(key)
	assert.Nil(t, err)
	assert.Equal(t, val, v)
}

func TestDB_GetErrorOnFail(t *testing.T) {
	ldb := createLevelDb(t, 1, 100, 10)
	_ = ldb.Close()

	v, err := ldb.Get([]byte("key"))
	assert.Nil(t, v)
	assert.NotNil(t, err)
}

func TestDB_RemoveBeforeTimeoutOK(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	key, val := []byte("key"), []byte("value")
	ldb := createLevelDb(t, 1, 100, 10)

	err := ldb.Put(key, val)
	assert.Nil(t, err)

	_ = ldb.Remove(key)
	time.Sleep(time.Second * 2)

	v, err := ldb.Get(key)
	assert.Nil(t, v)
	assert.Equal(t, storage.ErrKeyNotFound, err)
}

func TestDB_RemoveAfterTimeoutOK(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createLevelDb(t, 1, 100, 10)

	err := ldb.Put(key, val)
	assert.Nil(t, err)
	time.Sleep(time.Second * 2)

	_ = ldb.Remove(key)

	v, err := ldb.Get(key)
	assert.Nil(t, v)
	assert.Equal(t, storage.ErrKeyNotFound, err)
}

func TestDB_GetPresent(t *testing.T) {
	key, val := []byte("key1"), []byte("value1")
	ldb := createLevelDb(t, 10, 1, 10)

	err := ldb.Put(key, val)

	assert.Nil(t, err, "error saving in DB")

	v, err := ldb.Get(key)

	assert.Nil(t, err, "error not expected, but got %s", err)
	assert.Equalf(t, v, val, "read:%s but expected: %s", v, val)
}

func TestDB_GetNotPresent(t *testing.T) {
	key := []byte("key2")
	ldb := createLevelDb(t, 10, 1, 10)

	v, err := ldb.Get(key)

	assert.NotNil(t, err, "error expected but got nil, value %s", v)
}

func TestDB_HasPresent(t *testing.T) {
	key, val := []byte("key3"), []byte("value3")
	ldb := createLevelDb(t, 10, 1, 10)

	err := ldb.Put(key, val)

	assert.Nil(t, err, "error saving in DB")

	err = ldb.Has(key)

	assert.Nil(t, err)
}

func TestDB_HasNotPresent(t *testing.T) {
	key := []byte("key4")
	ldb := createLevelDb(t, 10, 1, 10)

	err := ldb.Has(key)

	assert.NotNil(t, err)
	assert.Equal(t, err, storage.ErrKeyNotFound)
}

func TestDB_RemovePresent(t *testing.T) {
	key, val := []byte("key5"), []byte("value5")
	ldb := createLevelDb(t, 10, 1, 10)

	err := ldb.Put(key, val)

	assert.Nil(t, err, "error saving in DB")

	err = ldb.Remove(key)

	assert.Nil(t, err, "no error expected but got %s", err)

	err = ldb.Has(key)

	assert.NotNil(t, err)
	assert.Equal(t, err, storage.ErrKeyNotFound)
}

func TestDB_RemoveNotPresent(t *testing.T) {
	key := []byte("key6")
	ldb := createLevelDb(t, 10, 1, 10)

	err := ldb.Remove(key)

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestDB_Close(t *testing.T) {
	ldb := createLevelDb(t, 10, 1, 10)

	err := ldb.Close()

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestDB_Destroy(t *testing.T) {
	ldb := createLevelDb(t, 10, 1, 10)

	err := ldb.Destroy()

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestDB_RangeKeys(t *testing.T) {
	ldb := createLevelDb(t, 1, 1, 10)
	defer func() {
		_ = ldb.Close()
	}()

	keysVals := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
		"key4": []byte("value4"),
		"key5": []byte("value5"),
		"key6": []byte("value6"),
		"key7": []byte("value7"),
	}

	for key, val := range keysVals {
		_ = ldb.Put([]byte(key), val)
	}

	time.Sleep(time.Second * 2)

	recovered := make(map[string][]byte)

	handler := func(key []byte, val []byte) bool {
		recovered[string(key)] = val
		return true
	}

	ldb.RangeKeys(handler)

	assert.Equal(t, keysVals, recovered)
}

func TestDB_PutGetLargeValue(t *testing.T) {
	t.Parallel()

	buffLargeValue := make([]byte, 32*1000000) //equivalent to ~1000000 hashes
	key := []byte("key")
	_, _ = rand.Read(buffLargeValue)

	ldb := createLevelDb(t, 1, 1, 10)
	defer func() {
		_ = ldb.Close()
	}()

	err := ldb.Put(key, buffLargeValue)
	assert.Nil(t, err)

	time.Sleep(time.Second * 2)

	recovered, err := ldb.Get(key)
	assert.Nil(t, err)

	assert.Equal(t, buffLargeValue, recovered)
}
