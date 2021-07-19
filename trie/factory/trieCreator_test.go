package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/mock"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getArgs() factory.TrieFactoryArgs {
	return factory.TrieFactoryArgs{
		Marshalizer: &mock.MarshalizerMock{},
		Hasher:      &mock.HasherMock{},
		PathManager: &testscommon.PathManagerStub{},
	}
}

func getCreateArgs() data.TrieCreateArgs {
	return data.TrieCreateArgs{
		TrieStorageConfig:  createTrieStorageCfg(),
		ShardID:            "0",
		PruningEnabled:     false,
		CheckpointsEnabled: false,
		MaxTrieLevelInMem:  5,
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
	tf, err := factory.NewTrieFactory(args)

	assert.Nil(t, tf)
	assert.Equal(t, trie.ErrNilMarshalizer, err)
}

func TestNewTrieFactory_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.Hasher = nil
	tf, err := factory.NewTrieFactory(args)

	assert.Nil(t, tf)
	assert.Equal(t, trie.ErrNilHasher, err)
}

func TestNewTrieFactory_NilPathManagerShouldErr(t *testing.T) {
	t.Parallel()

	args := getArgs()
	args.PathManager = nil
	tf, err := factory.NewTrieFactory(args)

	assert.Nil(t, tf)
	assert.Equal(t, trie.ErrNilPathManager, err)
}

func TestNewTrieFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgs()

	tf, err := factory.NewTrieFactory(args)
	require.Nil(t, err)
	require.False(t, check.IfNil(tf))
}

func TestTrieFactory_CreateNotSupportedCacheType(t *testing.T) {
	t.Parallel()

	args := getArgs()
	tf, _ := factory.NewTrieFactory(args)

	createArgs := getCreateArgs()
	createArgs.TrieStorageConfig = config.StorageConfig{}
	_, tr, err := tf.Create(createArgs)
	require.Nil(t, tr)
	require.Equal(t, storage.ErrNotSupportedCacheType, err)
}

func TestTrieFactory_CreateWithoutPruningShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgs()
	tf, _ := factory.NewTrieFactory(args)

	_, tr, err := tf.Create(getCreateArgs())
	require.NotNil(t, tr)
	require.Nil(t, err)
}

func TestTrieCreator_CreateWithPruningShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgs()
	tf, _ := factory.NewTrieFactory(args)

	createArgs := getCreateArgs()
	createArgs.PruningEnabled = true
	_, tr, err := tf.Create(createArgs)
	require.NotNil(t, tr)
	require.Nil(t, err)
}

func TestTrieCreator_CreateWithoutCheckpointShouldWork(t *testing.T) {
	t.Parallel()

	args := getArgs()
	tf, _ := factory.NewTrieFactory(args)

	createArgs := getCreateArgs()
	createArgs.PruningEnabled = true
	createArgs.CheckpointsEnabled = true
	_, tr, err := tf.Create(createArgs)
	require.NotNil(t, tr)
	require.Nil(t, err)
}
