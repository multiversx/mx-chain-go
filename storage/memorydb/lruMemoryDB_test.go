package memorydb_test

import (
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLruDB_LruDB_InitNoError(t *testing.T) {
	mdb, err := memorydb.NewlruDB(10000)
	assert.Nil(t, err, "failed to create memorydb: %s", err)

	err = mdb.Init()
	assert.Nil(t, err, "error initializing db")
}

func TestLruDB_LruDB_InitBadSize(t *testing.T) {
	mdb, err := memorydb.NewlruDB(0)
	assert.Nil(t, mdb)
	assert.NotNil(t, err)
}

func TestLruDB_PutNoError(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	mdb, err := memorydb.NewlruDB(10000)

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	err = mdb.Put(key, val)

	assert.Nil(t, err, "error saving in db")
}

func TestLruDB_GetPresent(t *testing.T) {
	key, val := []byte("key1"), []byte("value1")
	mdb, err := memorydb.NewlruDB(10000)

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	err = mdb.Put(key, val)

	assert.Nil(t, err, "error saving in db")

	v, err := mdb.Get(key)

	assert.Nil(t, err, "error not expected but got %s", err)
	assert.Equal(t, val, v, "expected %s but got %s", val, v)
}

func TestLruDB_GetNotPresent(t *testing.T) {
	key := []byte("key2")
	mdb, err := memorydb.NewlruDB(10000)

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	v, err := mdb.Get(key)

	assert.NotNil(t, err, "error expected but got nil, value %s", v)
}

func TestLruDB_HasPresent(t *testing.T) {
	key, val := []byte("key3"), []byte("value3")
	mdb, err := memorydb.NewlruDB(10000)

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	err = mdb.Put(key, val)

	assert.Nil(t, err, "error saving in db")

	err = mdb.Has(key)

	assert.Nil(t, err, "error not expected but got %s", err)
}

func TestLruDB_HasNotPresent(t *testing.T) {
	key := []byte("key4")
	mdb, err := memorydb.NewlruDB(10000)

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	err = mdb.Has(key)

	assert.Equal(t, storage.ErrKeyNotFound, err)
}

func TestLruDB_DeletePresent(t *testing.T) {
	key, val := []byte("key5"), []byte("value5")
	mdb, err := memorydb.NewlruDB(10000)

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	err = mdb.Put(key, val)

	assert.Nil(t, err, "error saving in db")

	err = mdb.Remove(key)

	assert.Nil(t, err, "no error expected but got %s", err)

	err = mdb.Has(key)

	assert.Equal(t, storage.ErrKeyNotFound, err)
}

func TestLruDB_DeleteNotPresent(t *testing.T) {
	key := []byte("key6")
	mdb, err := memorydb.NewlruDB(10000)

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	err = mdb.Remove(key)

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestLruDB_Close(t *testing.T) {
	mdb, err := memorydb.NewlruDB(10000)

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	err = mdb.Close()

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestLruDB_Destroy(t *testing.T) {
	mdb, err := memorydb.NewlruDB(10000)

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	err = mdb.Destroy()

	assert.Nil(t, err, "no error expected but got %s", err)
}
