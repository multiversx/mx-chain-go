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
	dir := t.TempDir()
	db, err := createPersister(dir, "sharded")
	require.Nil(t, err)

	err = db.Put([]byte("aaa"), []byte("aaaval"))
	require.Nil(t, err)

	err = db.Put([]byte("aab"), []byte("aabval"))
	require.Nil(t, err)

	err = db.Put([]byte("aac"), []byte("aacval"))
	require.Nil(t, err)

	err = db.Close()
	require.Nil(t, err)

	db, err = createPersister(dir, "sharded")
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

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, key, persisterPath, "simple")
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, key, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersisterOneKey2mil(b *testing.B) {
	entries, keys := generateKeys(2000000)
	randIndex := randInt(0, len(keys)-1)
	key := []byte(keys[randIndex])

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, key, persisterPath, "simple")
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, key, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersisterOneKey4mil(b *testing.B) {
	entries, keys := generateKeys(4000000)
	randIndex := randInt(0, len(keys)-1)
	key := []byte(keys[randIndex])

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, key, persisterPath, "simple")
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, key, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersisterOneKey8mil(b *testing.B) {
	entries, keys := generateKeys(8000000)
	randIndex := randInt(0, len(keys)-1)
	key := []byte(keys[randIndex])

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, key, persisterPath, "simple")
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersisterOneKey(b, key, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister1milAllKeys(b *testing.B) {
	entries, keys := generateKeys(1000000)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersister(b, keys, persisterPath, "simple")
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersister(b, keys, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister1mil10kKeys(b *testing.B) {
	entries, keys := generateKeys(1000000)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersister(b, keys[0:10000], persisterPath, "simple")
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersister(b, keys[0:10000], shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister1mil100kKeys(b *testing.B) {
	entries, keys := generateKeys(1000000)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersister(b, keys[0:100000], persisterPath, "simple")
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersister(b, keys[0:100000], shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister2milAllKeys(b *testing.B) {
	entries, keys := generateKeys(2000000)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersister(b, keys, persisterPath, "simple")
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersister(b, keys, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister4milAllKeys(b *testing.B) {
	entries, keys := generateKeys(4000000)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersister(b, keys, persisterPath, "simple")
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersister(b, keys, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister8milAllKeys(b *testing.B) {
	entries, keys := generateKeys(8000000)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(t *testing.B) {
		benchmarkPersister(b, keys, persisterPath, "simple")
	})

	b.Run("sharded persister", func(t *testing.B) {
		benchmarkPersister(b, keys, shardedPersisterPath, "sharded")
	})
}

func createPersister(path string, id string) (storage.Persister, error) {
	switch id {
	case "simple":
		return NewLevelDB(path, 2, 1000, 1000)
	case "sharded":
		shardCoordinator, _ := sharding.NewMultiShardCoordinator(4, 0)
		return NewShardedPersister(path, 2, 1000, 1000, shardCoordinator)
	default:
		return NewLevelDB(path, 2, 1000, 10)
	}
}

func createAndPopulatePersister(path string, entries map[string][]byte) storage.Persister {
	db, err := createPersister(path, "simple")
	if err != nil {
		log.Error("createAndPopulatePersister: failed to create persister", "error", err.Error())
		return nil
	}

	for key, val := range entries {
		_ = db.Put([]byte(key), val)
	}

	err = db.Close()
	if err != nil {
		log.Error("createAndPopulatePersister: failed to close persister", "error", err.Error())
		return nil
	}

	return db
}

func createAndPopulateShardedPersister(path string, entries map[string][]byte) storage.Persister {
	db, err := createPersister(path, "sharded")
	if err != nil {
		log.Error("createAndPopulateShardedPersister: failed to create persister", "error", err.Error())
		return nil
	}

	for key, val := range entries {
		_ = db.Put([]byte(key), val)
	}

	err = db.Close()
	if err != nil {
		log.Error("createAndPopulateShardedPersister: failed to close persister", "error", err.Error())
		return nil
	}

	return db
}

func benchmarkPersisterOneKey(b *testing.B, key []byte, path string, id string) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getKey(b, key, path, id)
	}
}

func benchmarkPersister(
	b *testing.B,
	keys []string,
	path string,
	id string,
) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getKeys(b, keys, path, id)
	}
}

func getKeys(
	b *testing.B,
	keys []string,
	path string,
	id string,
) {
	b.StopTimer()
	db, err := createPersister(path, id)
	if err != nil {
		log.Error("failed to create persister", "error", err.Error())
		return
	}

	defer func() {
		err := db.Close()
		if err != nil {
			log.Error("failed to close persister", "error", err.Error())
		}
	}()
	b.StartTimer()

	wg := sync.WaitGroup{}
	wg.Add(len(keys))
	for _, key := range keys {
		go func(key string) {
			defer func() {
				wg.Done()
			}()

			_, err = db.Get([]byte(key))
			require.Nil(b, err)
		}(key)
	}

	wg.Wait()
}

func getKey(
	b *testing.B,
	key []byte,
	path string,
	id string,
) {
	go func(key []byte) {
		b.StopTimer()
		db, err := createPersister(path, id)
		if err != nil {
			log.Error("failed to create persister", "error", err.Error())
			return
		}

		defer func() {
			db.Close()
		}()
		b.StartTimer()

		_, err = db.Get(key)
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
