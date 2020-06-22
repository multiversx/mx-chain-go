package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getArgs() TrieFactoryArgs {
	return TrieFactoryArgs{
		Marshalizer: &mock.MarshalizerMock{},
		Hasher:      &mock.HasherMock{},
		PathManager: &mock.PathManagerStub{},
	}
}

func createTrieStorageCfg() config.StorageConfig {
	return config.StorageConfig{
		Cache: config.CacheConfig{Type: "LRU", Capacity: 1000},
		DB:    config.DBConfig{Type: string(storageUnit.MemoryDB)},
		Bloom: config.BloomFilterConfig{},
	}
}

func TestNewTrieFactory_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Marshalizer = nil
	tf, err := NewTrieFactory(args)

	assert.Nil(t, tf)
	assert.Equal(t, trie.ErrNilMarshalizer, err)
}

func TestNewTrieFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Hasher = nil
	tf, err := NewTrieFactory(args)

	assert.Nil(t, tf)
	assert.Equal(t, trie.ErrNilHasher, err)
}

func TestNewTrieFactory_NilPathManagerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.PathManager = nil
	tf, err := NewTrieFactory(args)

	assert.Nil(t, tf)
	assert.Equal(t, trie.ErrNilPathManager, err)
}

func TestNewTrieFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgs()

	tf, err := NewTrieFactory(args)
	require.Nil(t, err)
	require.False(t, check.IfNil(tf))
}

func TestTrieFactory_CreateNotSupportedCacheType(t *testing.T) {
	t.Parallel()

	args := getArgs()
	tf, _ := NewTrieFactory(args)
	trieStorageCfg := config.StorageConfig{}

	maxTrieLevelInMemory := uint(5)
	_, tr, err := tf.Create(trieStorageCfg, "0", false, maxTrieLevelInMemory)
	require.Nil(t, tr)
	require.Equal(t, storage.ErrNotSupportedCacheType, err)
}

func TestTrieFactory_CreateWithoutPrunningWork(t *testing.T) {
	t.Parallel()

	args := getArgs()
	tf, _ := NewTrieFactory(args)
	trieStorageCfg := createTrieStorageCfg()

	maxTrieLevelInMemory := uint(5)
	_, tr, err := tf.Create(trieStorageCfg, "0", false, maxTrieLevelInMemory)
	require.NotNil(t, tr)
	require.Nil(t, err)
}

func TestTrieFactory_CreateWithPrunningWrongDbType(t *testing.T) {
	t.Parallel()

	args := getArgs()
	tf, _ := NewTrieFactory(args)
	trieStorageCfg := createTrieStorageCfg()

	maxTrieLevelInMemory := uint(5)
	_, tr, err := tf.Create(trieStorageCfg, "0", true, maxTrieLevelInMemory)
	require.Nil(t, tr)
	require.Equal(t, storage.ErrNotSupportedDBType, err)
}

func TestTrieFactory_CreateInvalidCacheSize(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.EvictionWaitingListCfg = config.EvictionWaitingListConfig{
		DB: config.DBConfig{Type: string(storageUnit.MemoryDB)},
	}
	tf, _ := NewTrieFactory(args)
	trieStorageCfg := createTrieStorageCfg()

	maxTrieLevelInMemory := uint(5)
	_, tr, err := tf.Create(trieStorageCfg, "0", true, maxTrieLevelInMemory)
	require.Nil(t, tr)
	require.Equal(t, data.ErrInvalidCacheSize, err)
}

func TestTrieFactory_CreateWithPRunningShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.EvictionWaitingListCfg = config.EvictionWaitingListConfig{
		DB:   config.DBConfig{Type: string(storageUnit.MemoryDB)},
		Size: 100,
	}
	tf, _ := NewTrieFactory(args)
	trieStorageCfg := createTrieStorageCfg()

	maxTrieLevelInMemory := uint(5)
	_, tr, err := tf.Create(trieStorageCfg, "0", true, maxTrieLevelInMemory)
	require.NotNil(t, tr)
	require.Nil(t, err)
}
