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
	shardCoordinator, err := sharding.NewMultiShardCoordinator(2, 0)
	require.Nil(t, err)

	dir := t.TempDir()
	db, err := NewShardedPersister(dir, 2, 1000, 10, shardCoordinator)
	require.Nil(t, err)
	require.NotNil(t, db)

	err = db.Put([]byte("aaa"), []byte("aaaval"))
	require.Nil(t, err)

	err = db.Put([]byte("aab"), []byte("aabval"))
	require.Nil(t, err)

	err = db.Put([]byte("aac"), []byte("aacval"))
	require.Nil(t, err)

	val, err := db.Get([]byte("aaa"))
	require.Nil(t, err)
	require.NotNil(t, val)

	val, err = db.Get([]byte("aab"))
	require.Nil(t, err)
	require.NotNil(t, val)

	val, err = db.Get([]byte("aac"))
	require.Nil(t, err)
	require.NotNil(t, val)
}

func TestNewPersister(t *testing.T) {
	dir := t.TempDir()
	db, err := NewLevelDB(dir, 2, 1000, 10)
	require.Nil(t, err)
	require.NotNil(t, db)
}

func BenchmarkPersisterOneKey1mil(b *testing.B) {
	entries, keys := generateKeys(1000000)
	randIndex := randInt(0, len(keys)-1)
	key := []byte(keys[randIndex])

	b.Run("persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, entries, key, createPersister(b))
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, entries, key, createShardedPersister(b))
	})
}

func BenchmarkPersisterOneKey2mil(b *testing.B) {
	entries, keys := generateKeys(2000000)
	randIndex := randInt(0, len(keys)-1)
	key := []byte(keys[randIndex])

	b.Run("persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, entries, key, createPersister(b))
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, entries, key, createShardedPersister(b))
	})
}

func BenchmarkPersisterOneKey4mil(b *testing.B) {
	entries, keys := generateKeys(4000000)
	randIndex := randInt(0, len(keys)-1)
	key := []byte(keys[randIndex])

	b.Run("persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, entries, key, createPersister(b))
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, entries, key, createShardedPersister(b))
	})
}

func BenchmarkPersisterOneKey8mil(b *testing.B) {
	entries, keys := generateKeys(8000000)
	randIndex := randInt(0, len(keys)-1)
	key := []byte(keys[randIndex])

	b.Run("persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, entries, key, createPersister(b))
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, entries, key, createShardedPersister(b))
	})
}

func BenchmarkPersister1milAllKeys(b *testing.B) {
	entries, keys := generateKeys(1000000)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersister(b, entries, createPersister(b), keys)
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersister(b, entries, createShardedPersister(b), keys)
	})
}

func BenchmarkPersister4milAllKeys(b *testing.B) {
	entries, keys := generateKeys(4000000)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersister(b, entries, createPersister(b), keys)
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersister(b, entries, createShardedPersister(b), keys)
	})
}

func BenchmarkPersister8milAllKeys(b *testing.B) {
	entries, keys := generateKeys(4000000)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersister(b, entries, createPersister(b), keys)
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersister(b, entries, createShardedPersister(b), keys)
	})
}

func BenchmarkPersister1mil10kKeys(b *testing.B) {
	entries, keys := generateKeys(1000000)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersister(b, entries, createPersister(b), keys[0:10000])
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersister(b, entries, createShardedPersister(b), keys[0:10000])
	})
}

func BenchmarkPersister1mil100kKeys(b *testing.B) {
	entries, keys := generateKeys(1000000)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersister(b, entries, createPersister(b), keys[0:100000])
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersister(b, entries, createShardedPersister(b), keys[0:100000])
	})
}

func BenchmarkPersister10mil(b *testing.B) {
	entries, keys := generateKeys(10000000)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersister(b, entries, createPersister(b), keys)
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersister(b, entries, createShardedPersister(b), keys)
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
		getKey(b, db, key)
	}
}

func benchmarkPersister(
	b *testing.B,
	entries map[string][]byte,
	db storage.Persister,
	keys []string,
) {
	for key, val := range entries {
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
	keys []string,
) {
	wg := sync.WaitGroup{}
	wg.Add(len(keys))

	for _, key := range keys {
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
) {
	go func(key []byte) {
		_, err := db.Get(key)
		require.Nil(b, err)
	}(key)
}

func generateKeys(numKeys int) (map[string][]byte, []string) {
	entries := make(map[string][]byte)

	keys := make([]string, 0, len(entries))

	for i := 0; i < numKeys; i++ {
		key := generateRandomByteArray(32)
		valueSize := randInt(20, 200)
		value := generateRandomByteArray(valueSize)

		entries[string(key)] = value
		keys = append(keys, string(key))
	}

	return entries, keys
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
