package database

import (
	"crypto/rand"
	"math/big"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-go/storage"
	"github.com/stretchr/testify/require"
)

const (
	_1Mil = 1_000_000
	_2Mil = 2_000_000
	_4Mil = 4_000_000
	_8Mil = 8_000_000
	_10k  = 10_000
	_100k = 100_000

	numShards = 4
)

func TestNewShardedPersister(t *testing.T) {
	dir := t.TempDir()
	db, err := createPersister(dir, "sharded")
	require.Nil(t, err)

	_ = db.Put([]byte("aaa"), []byte("aaaval"))
	_ = db.Put([]byte("aab"), []byte("aabval"))
	_ = db.Put([]byte("aac"), []byte("aacval"))

	err = db.Close()
	require.Nil(t, err)

	db2, err := createPersister(dir, "sharded")
	require.Nil(t, err)

	_, err = db2.Get([]byte("aaa"))
	require.Nil(t, err)

	_, err = db2.Get([]byte("aab"))
	require.Nil(t, err)

	_, err = db2.Get([]byte("aac"))
	require.Nil(t, err)
}

func TestNewPersister(t *testing.T) {
	dir := t.TempDir()
	db, err := NewLevelDB(dir, 2, 1000, 10)
	require.Nil(t, err)
	require.NotNil(t, db)
}

func BenchmarkPersisterOneKey1mil(b *testing.B) {
	entries, keys := generateKeys(_1Mil)
	randIndex := randInt(0, len(keys)-1)
	key := []byte(keys[randIndex])

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, "simple", entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulatePersister(shardedPersisterPath, "sharded", entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersisterOneKey(b, key, persisterPath, "simple")
	})
	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersisterOneKey(b, key, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersisterOneKey2mil(b *testing.B) {
	entries, keys := generateKeys(_2Mil)
	randIndex := randInt(0, len(keys)-1)
	key := []byte(keys[randIndex])

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, "simple", entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulatePersister(shardedPersisterPath, "sharded", entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersisterOneKey(b, key, persisterPath, "simple")
	})

	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersisterOneKey(b, key, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersisterOneKey4mil(b *testing.B) {
	entries, keys := generateKeys(_4Mil)
	randIndex := randInt(0, len(keys)-1)
	key := []byte(keys[randIndex])

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, "simple", entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulatePersister(shardedPersisterPath, "sharded", entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersisterOneKey(b, key, persisterPath, "simple")
	})

	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersisterOneKey(b, key, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersisterOneKey8mil(b *testing.B) {
	entries, keys := generateKeys(_8Mil)
	randIndex := randInt(0, len(keys)-1)
	key := []byte(keys[randIndex])

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, "simple", entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulatePersister(shardedPersisterPath, "sharded", entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersisterOneKey(b, key, persisterPath, "simple")
	})

	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersisterOneKey(b, key, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister1milAllKeys(b *testing.B) {
	entries, keys := generateKeys(_1Mil)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, "simple", entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulatePersister(shardedPersisterPath, "sharded", entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersister(b, keys, persisterPath, "simple")
	})

	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersister(b, keys, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister1mil10kKeys(b *testing.B) {
	entries, keys := generateKeys(_1Mil)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, "simple", entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulatePersister(shardedPersisterPath, "sharded", entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersister(b, keys[0:_10k], persisterPath, "simple")
	})

	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersister(b, keys[0:_10k], shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister1mil100kKeys(b *testing.B) {
	entries, keys := generateKeys(_1Mil)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, "simple", entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulatePersister(shardedPersisterPath, "sharded", entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersister(b, keys[0:_100k], persisterPath, "simple")
	})

	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersister(b, keys[0:_100k], shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister2milAllKeys(b *testing.B) {
	entries, keys := generateKeys(_2Mil)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, "simple", entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulatePersister(shardedPersisterPath, "sharded", entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersister(b, keys, persisterPath, "simple")
	})

	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersister(b, keys, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister4milAllKeys(b *testing.B) {
	entries, keys := generateKeys(_4Mil)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, "simple", entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulatePersister(shardedPersisterPath, "sharded", entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersister(b, keys, persisterPath, "simple")
	})

	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersister(b, keys, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister8milAllKeys(b *testing.B) {
	entries, keys := generateKeys(_8Mil)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, "simple", entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulatePersister(shardedPersisterPath, "sharded", entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersister(b, keys, persisterPath, "simple")
	})

	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersister(b, keys, shardedPersisterPath, "sharded")
	})
}

func createPersister(path string, id string) (storage.Persister, error) {
	switch id {
	case "simple":
		return NewSerialDB(path, 2, _1Mil, 10)
	case "sharded":
		shardCoordinator, _ := NewShardIDProvider(numShards)
		return NewShardedPersister(path, 2, _1Mil, 10, shardCoordinator)
	default:
		return NewSerialDB(path, 2, _1Mil, 10)
	}
}

func createAndPopulatePersister(path string, id string, entries map[string][]byte) storage.Persister {
	db, err := createPersister(path, id)
	if err != nil {
		log.Error("createAndPopulatePersister: failed to create persister", "type", id, "error", err.Error())
		return nil
	}

	for key, val := range entries {
		err = db.Put([]byte(key), val)
		if err != nil {
			log.Error("createAndPopulatePersister: failed to put key", "type", id, "key", key)
		}
	}

	err = db.Close()
	if err != nil {
		log.Error("createAndPopulatePersister: failed to close persister", "type", id, "error", err.Error())
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
	db := createPersisterWithTimerControl(b, path, id)

	defer func() {
		closePersisterWithTimerControl(b, db)
	}()

	maxRoutines := make(chan struct{}, 400)
	wg := sync.WaitGroup{}
	wg.Add(len(keys))

	for _, key := range keys {
		maxRoutines <- struct{}{}
		go func(key string) {
			_, err := db.Get([]byte(key))
			require.Nil(b, err)

			<-maxRoutines
			wg.Done()
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
	db := createPersisterWithTimerControl(b, path, id)

	defer func() {
		closePersisterWithTimerControl(b, db)
	}()

	_, err := db.Get(key)
	require.Nil(b, err)
}

func createPersisterWithTimerControl(b *testing.B, path, id string) storage.Persister {
	b.StopTimer()
	db, err := createPersister(path, id)
	require.Nil(b, err)
	b.StartTimer()

	return db
}

func closePersisterWithTimerControl(b *testing.B, db storage.Persister) {
	b.StopTimer()
	err := db.Close()
	require.Nil(b, err)
	b.StartTimer()
}

func generateKeys(numKeys int) (map[string][]byte, []string) {
	entries := make(map[string][]byte)

	keys := make([]string, 0)

	for i := 0; i < numKeys; i++ {
		key := generateRandomByteArray(32)
		value := generateRandomByteArray(32)

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
