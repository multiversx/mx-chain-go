package leveldb_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/leveldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createSerialLevelDb(t *testing.T, batchDelaySeconds int, maxBatchSize int, maxOpenFiles int) (p *leveldb.SerialDB) {
	lvdb, err := leveldb.NewSerialDB(t.TempDir(), batchDelaySeconds, maxBatchSize, maxOpenFiles)

	assert.Nil(t, err, "Failed creating leveldb database file")
	assert.False(t, check.IfNil(lvdb))
	return lvdb
}

func TestSerialDB_Put(t *testing.T) {
	t.Run("invalid priority", func(t *testing.T) {
		key, val := []byte("key"), []byte("value")
		ldb := createSerialLevelDb(t, 10, 1, 10)

		err := ldb.Put(key, val, "invalid")

		assert.True(t, errors.Is(err, storage.ErrInvalidPriorityType))
	})
	t.Run("should work", func(t *testing.T) {
		key, val := []byte("key"), []byte("value")
		ldb := createSerialLevelDb(t, 10, 1, 10)

		err := ldb.Put(key, val, common.TestPriority)
		assert.Nil(t, err, "error saving in DB")

		recovered, err := ldb.Get(key, common.TestPriority)
		assert.Nil(t, err)
		assert.Equal(t, val, recovered)
	})
}

func TestSerialDB_GetErrorAfterPutBeforeTimeout(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createSerialLevelDb(t, 1, 100, 10)

	_ = ldb.Put(key, val, common.TestPriority)
	v, err := ldb.Get(key, common.TestPriority)

	assert.Equal(t, val, v)
	assert.Nil(t, err)
}

func TestSerialDB_GetErrorOnFail(t *testing.T) {
	ldb := createSerialLevelDb(t, 10, 1, 10)
	_ = ldb.Destroy()

	v, err := ldb.Get([]byte("key"), common.TestPriority)
	assert.Nil(t, v)
	assert.NotNil(t, err)
}

func TestSerialDB_MethodCallsAfterCloseOrDestroy(t *testing.T) {
	t.Parallel()

	t.Run("when closing", func(t *testing.T) {
		t.Parallel()

		testSerialDbAllMethodsShouldNotPanic(t, func(db *leveldb.SerialDB) {
			_ = db.Close()
		})
	})
	t.Run("when destroying", func(t *testing.T) {
		t.Parallel()

		testSerialDbAllMethodsShouldNotPanic(t, func(db *leveldb.SerialDB) {
			_ = db.Destroy()
		})
	})
}

func testSerialDbAllMethodsShouldNotPanic(t *testing.T, closeHandler func(db *leveldb.SerialDB)) {
	ldb := createSerialLevelDb(t, 10, 1, 10)

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panic %v", r))
		}
	}()

	closeHandler(ldb)

	_, err := ldb.Get([]byte("key1"), common.TestPriority)
	assert.Equal(t, storage.ErrDBIsClosed, err)

	err = ldb.Has([]byte("key2"), common.TestPriority)
	assert.Equal(t, storage.ErrDBIsClosed, err)

	err = ldb.Remove([]byte("key3"), common.TestPriority)
	assert.Equal(t, storage.ErrDBIsClosed, err)

	err = ldb.Put([]byte("key4"), []byte("val"), common.TestPriority)
	assert.Equal(t, storage.ErrDBIsClosed, err)

	ldb.RangeKeys(func(key []byte, value []byte) bool {
		require.Fail(t, "should have not called range")
		return false
	})
}

func TestSerialDB_GetOKAfterPutWithTimeout(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createSerialLevelDb(t, 1, 100, 10)

	_ = ldb.Put(key, val, common.TestPriority)
	time.Sleep(time.Second * 3)
	v, err := ldb.Get(key, common.TestPriority)

	assert.Nil(t, err)
	assert.Equal(t, val, v)
}

func TestSerialDB_RemoveBeforeTimeoutOK(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createSerialLevelDb(t, 1, 100, 10)

	_ = ldb.Put(key, val, common.TestPriority)
	_ = ldb.Remove(key, common.TestPriority)
	time.Sleep(time.Second * 2)
	v, err := ldb.Get(key, common.TestPriority)

	assert.Nil(t, v)
	assert.Equal(t, storage.ErrKeyNotFound, err)
}

func TestSerialDB_RemoveAfterTimeoutOK(t *testing.T) {
	key, val := []byte("key"), []byte("value")
	ldb := createSerialLevelDb(t, 1, 100, 10)

	_ = ldb.Put(key, val, common.TestPriority)
	time.Sleep(time.Second * 2)
	_ = ldb.Remove(key, common.TestPriority)
	v, err := ldb.Get(key, common.TestPriority)

	assert.Nil(t, v)
	assert.Equal(t, storage.ErrKeyNotFound, err)
}

func TestSerialDB_Get(t *testing.T) {
	t.Run("data is present in the persister", func(t *testing.T) {
		key, val := []byte("key1"), []byte("value1")
		ldb := createSerialLevelDb(t, 10, 1, 10)

		_ = ldb.Put(key, val, common.TestPriority)
		v, err := ldb.Get(key, common.TestPriority)

		assert.Nil(t, err, "error not expected, but got %s", err)
		assert.Equalf(t, v, val, "read:%s but expected: %s", v, val)
	})
	t.Run("data is missing from the persister", func(t *testing.T) {
		key := []byte("key2")
		ldb := createSerialLevelDb(t, 10, 1, 10)

		v, err := ldb.Get(key, common.TestPriority)

		assert.NotNil(t, err, "error expected but got nil, value %s", v)
		assert.Nil(t, v)
	})
	t.Run("invalid priority", func(t *testing.T) {
		key := []byte("key2")
		ldb := createSerialLevelDb(t, 10, 1, 10)

		v, err := ldb.Get(key, "invalid")

		assert.True(t, errors.Is(err, storage.ErrInvalidPriorityType))
		assert.Nil(t, v)
	})
}

func TestSerialDB_Has(t *testing.T) {
	t.Run("data is present in the persister", func(t *testing.T) {
		key, val := []byte("key3"), []byte("value3")
		ldb := createSerialLevelDb(t, 10, 1, 10)

		_ = ldb.Put(key, val, common.TestPriority)
		err := ldb.Has(key, common.TestPriority)

		assert.Nil(t, err)
	})
	t.Run("data is missing from the persister", func(t *testing.T) {
		key := []byte("key4")
		ldb := createSerialLevelDb(t, 10, 1, 10)

		err := ldb.Has(key, common.TestPriority)

		assert.NotNil(t, err)
		assert.Equal(t, err, storage.ErrKeyNotFound)
	})
	t.Run("invalid priority", func(t *testing.T) {
		key := []byte("key4")
		ldb := createSerialLevelDb(t, 10, 1, 10)

		err := ldb.Has(key, "invalid")

		assert.True(t, errors.Is(err, storage.ErrInvalidPriorityType))
	})
}

func TestSerialDB_Remove(t *testing.T) {
	t.Run("data is present in the persister", func(t *testing.T) {
		key, val := []byte("key5"), []byte("value5")
		ldb := createSerialLevelDb(t, 10, 1, 10)

		_ = ldb.Put(key, val, common.TestPriority)
		_ = ldb.Remove(key, common.TestPriority)
		err := ldb.Has(key, common.TestPriority)

		assert.NotNil(t, err)
		assert.Equal(t, err, storage.ErrKeyNotFound)
	})
	t.Run("data is missing from the persister", func(t *testing.T) {
		key := []byte("key6")
		ldb := createSerialLevelDb(t, 10, 1, 10)

		err := ldb.Remove(key, common.TestPriority)

		assert.Nil(t, err, "no error expected but got %s", err)
	})
	t.Run("invalid priority", func(t *testing.T) {
		key := []byte("key6")
		ldb := createSerialLevelDb(t, 10, 1, 10)

		err := ldb.Remove(key, "invalid")

		assert.True(t, errors.Is(err, storage.ErrInvalidPriorityType))
	})
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
