package leveldb_test

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/leveldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createSerialLevelDb(tb testing.TB, batchDelaySeconds int, maxBatchSize int, maxOpenFiles int) (p *leveldb.SerialDB) {
	lvdb, err := leveldb.NewSerialDB(tb.TempDir(), batchDelaySeconds, maxBatchSize, maxOpenFiles)

	assert.Nil(tb, err, "Failed creating leveldb database file")
	return lvdb
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

	_, err := ldb.Get([]byte("key1"))
	assert.Equal(t, storage.ErrDBIsClosed, err)

	err = ldb.Has([]byte("key2"))
	assert.Equal(t, storage.ErrDBIsClosed, err)

	err = ldb.Remove([]byte("key3"))
	assert.Equal(t, storage.ErrDBIsClosed, err)

	err = ldb.Put([]byte("key4"), []byte("val"))
	assert.Equal(t, storage.ErrDBIsClosed, err)

	ldb.RangeKeys(func(key []byte, value []byte) bool {
		require.Fail(t, "should have not called range")
		return false
	})
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

func TestSerialDB_SpecialValueTest(t *testing.T) {
	t.Parallel()

	ldb := createSerialLevelDb(t, 100, 100, 10)
	key := []byte("key")
	removedValue := []byte("removed") // in old implementations we had a check against this value
	randomValue := []byte("random")
	t.Run("operations: put -> get of 'removed' value", func(t *testing.T) {
		err := ldb.Put(key, removedValue)
		require.Nil(t, err)

		recovered, err := ldb.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, removedValue, recovered)
	})
	t.Run("operations: put -> remove -> get of 'removed' value", func(t *testing.T) {
		err := ldb.Put(key, removedValue)
		require.Nil(t, err)

		err = ldb.Remove(key)
		require.Nil(t, err)

		recovered, err := ldb.Get(key)
		assert.Equal(t, storage.ErrKeyNotFound, err)
		assert.Nil(t, recovered)
	})
	t.Run("operations: put -> remove -> put -> get of 'removed' value", func(t *testing.T) {
		err := ldb.Put(key, removedValue)
		require.Nil(t, err)

		err = ldb.Remove(key)
		require.Nil(t, err)

		err = ldb.Put(key, removedValue)
		require.Nil(t, err)

		recovered, err := ldb.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, removedValue, recovered)
	})
	t.Run("operations: put -> remove -> put -> get of random value", func(t *testing.T) {
		err := ldb.Put(key, randomValue)
		require.Nil(t, err)

		err = ldb.Remove(key)
		require.Nil(t, err)

		err = ldb.Put(key, randomValue)
		require.Nil(t, err)

		recovered, err := ldb.Get(key)
		assert.Nil(t, err)
		assert.Equal(t, randomValue, recovered)
	})

	_ = ldb.Close()
}

func BenchmarkSerialDB_SpecialValueTest(b *testing.B) {
	ldb := createSerialLevelDb(b, 10000, 10000000, 10)
	key := []byte("key")
	removedValue := []byte("removed") // in old implementations we had a check against this value

	b.Run("put -> remove -> get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ldb.Put(key, removedValue)
			_ = ldb.Remove(key)
			_, _ = ldb.Get(key)
		}
	})
	b.Run("put -> get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ldb.Put(key, removedValue)
			_, _ = ldb.Get(key)
		}
	})
	b.Run("put -> remove -> get with different keys", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			testKey := append(key, big.NewInt(int64(i)).Bytes()...)

			_ = ldb.Put(testKey, removedValue)
			_ = ldb.Remove(testKey)
			_, _ = ldb.Get(testKey)
		}
	})
	b.Run("put -> get with different keys", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			testKey := append(key, big.NewInt(int64(i)).Bytes()...)

			_ = ldb.Put(testKey, removedValue)
			_, _ = ldb.Get(testKey)
		}
	})
}
