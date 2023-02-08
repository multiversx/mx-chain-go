package database

import (
	"crypto/rand"
	"math/big"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-go/storage"
	"github.com/stretchr/testify/require"
)

const oneMil = 1_000_000
const twoMil = 2_000_000
const fourMil = 4_000_000
const eightMil = 8_000_000
const tenK = 10_000
const hundredK = 100_000

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
	entries, keys := generateKeys(oneMil)
	randIndex := randInt(0, len(keys)-1)
	key := []byte(keys[randIndex])

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersisterOneKey(b, key, persisterPath, "simple")
	})
	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersisterOneKey(b, key, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersisterOneKey2mil(b *testing.B) {
	entries, keys := generateKeys(twoMil)
	randIndex := randInt(0, len(keys)-1)
	key := []byte(keys[randIndex])

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersisterOneKey(b, key, persisterPath, "simple")
	})

	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersisterOneKey(b, key, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersisterOneKey4mil(b *testing.B) {
	entries, keys := generateKeys(fourMil)
	randIndex := randInt(0, len(keys)-1)
	key := []byte(keys[randIndex])

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersisterOneKey(b, key, persisterPath, "simple")
	})

	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersisterOneKey(b, key, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersisterOneKey8mil(b *testing.B) {
	entries, keys := generateKeys(eightMil)
	randIndex := randInt(0, len(keys)-1)
	key := []byte(keys[randIndex])

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersisterOneKey(b, key, persisterPath, "simple")
	})

	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersisterOneKey(b, key, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister1milAllKeys(b *testing.B) {
	entries, keys := generateKeys(oneMil)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersister(b, keys, persisterPath, "simple")
	})

	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersister(b, keys, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister1mil10kKeys(b *testing.B) {
	entries, keys := generateKeys(oneMil)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersister(b, keys[0:tenK], persisterPath, "simple")
	})

	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersister(b, keys[0:tenK], shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister1mil100kKeys(b *testing.B) {
	entries, keys := generateKeys(oneMil)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersister(b, keys[0:hundredK], persisterPath, "simple")
	})

	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersister(b, keys[0:hundredK], shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister2milAllKeys(b *testing.B) {
	entries, keys := generateKeys(twoMil)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersister(b, keys, persisterPath, "simple")
	})

	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersister(b, keys, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister4milAllKeys(b *testing.B) {
	entries, keys := generateKeys(fourMil)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

	b.Run("persister", func(b *testing.B) {
		benchmarkPersister(b, keys, persisterPath, "simple")
	})

	b.Run("sharded persister", func(b *testing.B) {
		benchmarkPersister(b, keys, shardedPersisterPath, "sharded")
	})
}

func BenchmarkPersister8milAllKeys(b *testing.B) {
	entries, keys := generateKeys(eightMil)

	persisterPath := b.TempDir()
	_ = createAndPopulatePersister(persisterPath, entries)
	shardedPersisterPath := b.TempDir()
	_ = createAndPopulateShardedPersister(shardedPersisterPath, entries)

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
		return NewSerialDB(path, 2, oneMil, 10)
	case "sharded":
		shardCoordinator, _ := NewShardIDProvider(4)
		return NewShardedPersister(path, 2, oneMil, 10, shardCoordinator)
	default:
		return NewSerialDB(path, 2, oneMil, 10)
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
	require.Nil(b, err)
	b.StartTimer()

	defer func() {
		b.StopTimer()
		errDefer := db.Close()
		require.Nil(b, errDefer)
		b.StartTimer()
	}()

	maxRoutines := make(chan struct{}, 400)
	wg := sync.WaitGroup{}
	wg.Add(len(keys))

	for _, key := range keys {
		maxRoutines <- struct{}{}
		go func(key string) {
			_, err = db.Get([]byte(key))
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
	b.StopTimer()
	db, err := createPersister(path, id)
	require.Nil(b, err)
	b.StartTimer()

	defer func() {
		b.StopTimer()
		errDefer := db.Close()
		require.Nil(b, errDefer)
		b.StartTimer()
	}()

	_, err = db.Get(key)
	require.Nil(b, err)
}

func generateKeys(numKeys int) (map[string][]byte, []string) {
	entries := make(map[string][]byte, numKeys)
	keys := make([]string, 0, numKeys)

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
