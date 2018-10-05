package memorydb_test

import (
	"testing"

	"ElrondNetwork/elrond-go-sandbox/storage/memorydb"

	"github.com/stretchr/testify/assert"
)

func TestInitNoError(t *testing.T) {
	mdb, err := memorydb.New()

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	err = mdb.Init()
	assert.Nil(t, err, "error initializing db")
}

func TestPutNoError(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	mdb, err := memorydb.New()

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	err = mdb.Put(key, val)

	assert.Nil(t, err, "error saving in db")
}

func TestGetPresent(t *testing.T) {
	key, val := []byte("key1"), []byte("value1")
	mdb, err := memorydb.New()

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	err = mdb.Put(key, val)

	assert.Nil(t, err, "error saving in db")

	v, err := mdb.Get(key)

	assert.Nil(t, err, "error not expected but got %s", err)
	assert.Equal(t, val, v, "expected %s but got %s", val, v)
}

func TestGetNotPresent(t *testing.T) {
	key := []byte("key2")
	mdb, err := memorydb.New()

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	v, err := mdb.Get(key)

	assert.NotNil(t, err, "error expected but got nil, value %s", v)
}

func TestHasPresent(t *testing.T) {
	key, val := []byte("key3"), []byte("value3")
	mdb, err := memorydb.New()

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	err = mdb.Put(key, val)

	assert.Nil(t, err, "error saving in db")

	has, err := mdb.Has(key)

	assert.Nil(t, err, "error not expected but got %s", err)
	assert.True(t, has, "value expected but not found")
}

func TestHasNotPresent(t *testing.T) {
	key := []byte("key4")
	mdb, err := memorydb.New()

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	has, err := mdb.Has(key)

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.False(t, has, "value not expected but found")
}

func TestDeletePresent(t *testing.T) {
	key, val := []byte("key5"), []byte("value5")
	mdb, err := memorydb.New()

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	err = mdb.Put(key, val)

	assert.Nil(t, err, "error saving in db")

	err = mdb.Remove(key)

	assert.Nil(t, err, "no error expected but got %s", err)

	has, err := mdb.Has(key)

	assert.False(t, has, "element not expected as already deleted")
}

func TestDeleteNotPresent(t *testing.T) {
	key := []byte("key6")
	mdb, err := memorydb.New()

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	err = mdb.Remove(key)

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestClose(t *testing.T) {
	mdb, err := memorydb.New()

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	err = mdb.Close()

	assert.Nil(t, err, "no error expected but got %s", err)
}

func TestDestroy(t *testing.T) {
	mdb, err := memorydb.New()

	assert.Nil(t, err, "failed to create memorydb: %s", err)

	err = mdb.Destroy()

	assert.Nil(t, err, "no error expected but got %s", err)
}
