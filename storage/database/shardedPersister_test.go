package database

import (
	"crypto/rand"
	"math/big"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/stretchr/testify/require"
)

func TestNewShardedPersister(t *testing.T) {
	t.Parallel()

	shardCoordinator, err := sharding.NewMultiShardCoordinator(2, 0)
	require.Nil(t, err)

	dir := t.TempDir()
	db, err := NewShardedPersister(dir, 2, 1000, 10, shardCoordinator)
	require.Nil(t, err)
	require.NotNil(t, db)

	err = db.Put([]byte("aaa"), []byte("aaaval"))
	require.Nil(t, err)

	val, err := db.Get([]byte("aaa"))
	require.Nil(t, err)
	require.NotNil(t, val)
}

func TestNewPersister(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	db, err := NewLevelDB(dir, 2, 1000, 10)
	require.Nil(t, err)
	require.NotNil(t, db)
}

func BenchmarkPersisterOneKey1mil(b *testing.B) {
	keys, key := generateKeys(1000000)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, keys, key, createPersister(b))
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, keys, key, createShardedPersister(b))
	})
}

func BenchmarkPersister1mil(b *testing.B) {
	keys, _ := generateKeys(1000000)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersister(b, keys, createPersister(b))
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersister(b, keys, createShardedPersister(b))
	})
}

func BenchmarkPersister10mil(b *testing.B) {
	keys, _ := generateKeys(1000000)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersister(b, keys, createPersister(b))
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersister(b, keys, createShardedPersister(b))
	})
}

func BenchmarkPersister100mil(b *testing.B) {
	keys, _ := generateKeys(1000000)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersister(b, keys, createPersister(b))
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersister(b, keys, createShardedPersister(b))
	})
}

func createPersister(b *testing.B) storage.Persister {
	dir := b.TempDir()
	db, _ := NewLevelDB(dir, 2, 1000, 10)

	return db
}

func createShardedPersister(b *testing.B) storage.Persister {
	shardCoordinator, _ := sharding.NewMultiShardCoordinator(3, 0)

	dir := b.TempDir()
	db, _ := NewShardedPersister(dir, 2, 1000, 10, shardCoordinator)

	return db
}

func benchmarkPersisterOneKey(b *testing.B, keys map[string][]byte, key []byte, db storage.Persister) {
	for key, val := range keys {
		err := db.Put([]byte(key), val)
		require.Nil(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getKey(b, db, key, 100)
	}
}

func benchmarkPersister(b *testing.B, keys map[string][]byte, db storage.Persister) {
	for key, val := range keys {
		err := db.Put([]byte(key), val)
		require.Nil(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getKeys(b, db, keys)
	}
}

func getKeys(
	b *testing.B,
	db storage.Persister,
	keys map[string][]byte,
) {
	wg := sync.WaitGroup{}
	wg.Add(len(keys))

	for key := range keys {
		go func(key string) {
			defer func() {
				wg.Done()
			}()

			_, err := db.Get([]byte(key))
			require.Nil(b, err)
		}(key)
	}

	wg.Wait()
}

func getKey(
	b *testing.B,
	db storage.Persister,
	key []byte,
	numOperations int,
) {
	wg := sync.WaitGroup{}
	wg.Add(numOperations)

	for i := 0; i < numOperations; i++ {
		go func(key []byte) {
			defer func() {
				wg.Done()
			}()

			_, err := db.Get(key)
			require.Nil(b, err)
		}(key)
	}

	wg.Wait()
}

func generateKeys(numKeys int) (map[string][]byte, []byte) {
	keys := make(map[string][]byte)
	var lastKey []byte

	for i := 0; i < numKeys; i++ {
		key := generateRandomByteArray(32)
		valueSize := randInt(20, 200)
		value := generateRandomByteArray(valueSize)

		keys[string(key)] = value
		lastKey = key
	}

	return keys, lastKey
}

func randInt(min int, max int) int {
	dd := int64(max - min)
	vv, _ := rand.Int(rand.Reader, big.NewInt(dd))
	return min + int(vv.Int64())
}

func generateRandomByteArray(size int) []byte {
	r := make([]byte, size)
	_, _ = rand.Read(r)
	return r
}
