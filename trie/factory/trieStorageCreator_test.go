package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/trie"
	"github.com/stretchr/testify/assert"
)

func createTrieStorageCfg() config.StorageConfig {
	return config.StorageConfig{
		Cache: config.CacheConfig{Type: "LRU", Capacity: 1000},
		DB:    config.DBConfig{Type: string(storageUnit.MemoryDB)},
		Bloom: config.BloomFilterConfig{},
	}
}

func TestNewOldTrieStorageCreator_NilPathManager(t *testing.T) {
	t.Parallel()

	otsc, err := NewOldTrieStorageCreator(nil, config.Config{})
	assert.True(t, otsc.IsInterfaceNil())
	assert.Equal(t, trie.ErrNilPathManager, err)
}

func TestNewOldTrieStorageCreator(t *testing.T) {
	t.Parallel()

	otsc, err := NewOldTrieStorageCreator(&testscommon.PathManagerStub{}, config.Config{
		AccountsTrieStorageOld:     createTrieStorageCfg(),
		PeerAccountsTrieStorageOld: createTrieStorageCfg(),
	})
	assert.False(t, otsc.IsInterfaceNil())
	assert.Nil(t, err)
}

func TestOldTrieStorageCreator_GetStorageForShard(t *testing.T) {
	t.Parallel()

	userAccCfg := createTrieStorageCfg()
	userAccCfg.DB.FilePath = "user/db"
	peerAccCfg := createTrieStorageCfg()
	peerAccCfg.DB.FilePath = "peer/db"

	otsc, _ := NewOldTrieStorageCreator(&testscommon.PathManagerStub{}, config.Config{
		AccountsTrieStorageOld:     userAccCfg,
		PeerAccountsTrieStorageOld: peerAccCfg,
	})

	db, err := otsc.GetStorageForShard("0", UserAccountTrie)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	db, err = otsc.GetStorageForShard("0", PeerAccountTrie)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	db, err = otsc.GetStorageForShard("1", PeerAccountTrie)
	assert.Nil(t, err)
	assert.NotNil(t, db)

	assert.Equal(t, 3, len(otsc.createdStorages))

	db, err = otsc.GetStorageForShard("0", UserAccountTrie)
	assert.Nil(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, 3, len(otsc.createdStorages))

	db, err = otsc.GetStorageForShard("0", "trie")
	assert.Nil(t, db)
	assert.Equal(t, trie.ErrInvalidTrieType, err)
}

func TestOldTrieStorageCreator_GetSnapshotsConfig(t *testing.T) {
	t.Parallel()

	userAccCfg := createTrieStorageCfg()
	userAccCfg.DB.FilePath = "user/db"
	peerAccCfg := createTrieStorageCfg()
	peerAccCfg.DB.FilePath = "peer/db"

	otsc, _ := NewOldTrieStorageCreator(&testscommon.PathManagerStub{}, config.Config{
		AccountsTrieStorageOld:     userAccCfg,
		PeerAccountsTrieStorageOld: peerAccCfg,
		TrieSnapshotDB:             config.DBConfig{FilePath: "snapshots"},
	})

	cfg := otsc.GetSnapshotsConfig("0", UserAccountTrie)
	assert.Equal(t, "Static/Shard_0/user/snapshots", cfg.FilePath)

	cfg = otsc.GetSnapshotsConfig("1", PeerAccountTrie)
	assert.Equal(t, "Static/Shard_1/peer/snapshots", cfg.FilePath)

	cfg = otsc.GetSnapshotsConfig("1", "trie")
	assert.Equal(t, "", cfg.FilePath)
}
