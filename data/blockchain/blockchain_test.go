package blockchain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

func blockchainConfig() *blockchain.Config {
	cacher := storage.CacheConfig{Type: storage.LRUCache, Size: 100}
	persisterBlockStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "BlockStorage"}
	persisterBlockHeaderStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "BlockHeaderStorage"}
	persisterTxStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "TxStorage"}
	return &blockchain.Config{
		BlockStorage:       storage.UnitConfig{CacheConf: cacher, DBConf: persisterBlockStorage},
		BlockHeaderStorage: storage.UnitConfig{CacheConf: cacher, DBConf: persisterBlockHeaderStorage},
		TxStorage:          storage.UnitConfig{CacheConf: cacher, DBConf: persisterTxStorage},
		TxPoolStorage:      cacher,
		BlockCache:         cacher,
	}
}

func failOnPanic(t *testing.T) {
	if r := recover(); r != nil {
		t.Errorf("the code entered panic")
	}
}

func TestNewBlockchainErrOnTxStorageCreationShouldError(t *testing.T) {
	cfg := blockchainConfig()
	// e.g change the config to a not supported cache type
	cfg.TxStorage.CacheConf.Type = "NotLRU"
	_, err := blockchain.NewBlockChain(cfg)
	assert.NotNil(t, err, "NewBlockchain should error on not supported type")
}

func TestNewBlockchainErrOnBlockStorageCreationShouldError(t *testing.T) {
	cfg := blockchainConfig()
	// e.g change the config to a not supported cache type
	cfg.BlockStorage.CacheConf.Type = "NotLRU"
	_, err := blockchain.NewBlockChain(cfg)
	assert.NotNil(t, err, "NewBlockchain should error on not supported cache type for block storage")
}

func TestNewBlockchainErrOnBlockHeaderStorageCreationShouldError(t *testing.T) {
	cfg := blockchainConfig()
	// e.g change the config to a not supported cache type
	cfg.BlockHeaderStorage.CacheConf.Type = "NotLRU"
	_, err := blockchain.NewBlockChain(cfg)
	assert.NotNil(t, err, "NewBlockchain should error on not supported cache type for block header")
}

func TestNewDataErrOnBlockCacheCreationShouldError(t *testing.T) {
	cfg := blockchainConfig()
	// e.g change the config to a not supported cache type
	cfg.BlockCache.Type = "NotLRU"
	_, err := blockchain.NewBlockChain(cfg)
	assert.NotNil(t, err, "NewBlockchain should error on not supported cache type for block cache")
}

func TestNewDataDefaultConfigOK(t *testing.T) {
	defer failOnPanic(t)
	cfg := blockchainConfig()
	b, err := blockchain.NewBlockChain(cfg)
	defer func() {
		err := b.Destroy()
		assert.Nil(t, err, "Unable to destroy blockchain")
	}()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")
}

func TestHasFalseOnWrongUnitType(t *testing.T) {
	cfg := blockchainConfig()
	b, err := blockchain.NewBlockChain(cfg)
	defer func() {
		err := b.Destroy()
		assert.Nil(t, err, "Unable to destroy blockchain")
	}()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	_ = b.Put(blockchain.TransactionUnit, []byte("key1"), []byte("aaa"))
	has, err := b.Has(100, []byte("key1"))

	assert.NotNil(t, err, "expected error but got nil")
	assert.False(t, has, "not expected to find key")
}

func TestHasOk(t *testing.T) {
	cfg := blockchainConfig()
	b, err := blockchain.NewBlockChain(cfg)
	defer func() {
		err := b.Destroy()
		assert.Nil(t, err, "Unable to destroy blockchain")
	}()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	_ = b.Put(blockchain.TransactionUnit, []byte("key1"), []byte("aaa"))
	has, err := b.Has(blockchain.BlockUnit, []byte("key1"))

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.False(t, has, "not expected to find key")

	_ = b.Put(blockchain.BlockUnit, []byte("key1"), []byte("bbb"))
	has, err = b.Has(blockchain.BlockHeaderUnit, []byte("key1"))

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.False(t, has, "not expected to find key")

	_ = b.Put(blockchain.BlockHeaderUnit, []byte("key1"), []byte("ccc"))
	has, err = b.Has(blockchain.TransactionUnit, []byte("key1"))

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.True(t, has, "expected to find key")

	has, err = b.Has(blockchain.BlockUnit, []byte("key1"))

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.True(t, has, "expected to find key")

	has, err = b.Has(blockchain.BlockHeaderUnit, []byte("key1"))

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.True(t, has, "expected to find key")
}

func TestGetErrOnWrongUnitType(t *testing.T) {
	cfg := blockchainConfig()
	b, err := blockchain.NewBlockChain(cfg)
	defer func() {
		err := b.Destroy()
		assert.Nil(t, err, "Unable to destroy blockchain")
	}()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	_ = b.Put(blockchain.TransactionUnit, []byte("key1"), []byte("aaa"))
	val, err := b.Get(100, []byte("key1"))

	assert.NotNil(t, err, "expected error but got nil")
	assert.Nil(t, val, "not expected to find key")
}

func TestGetOk(t *testing.T) {
	cfg := blockchainConfig()
	b, err := blockchain.NewBlockChain(cfg)
	defer func() {
		err := b.Destroy()
		assert.Nil(t, err, "Unable to destroy blockchain")
	}()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	_ = b.Put(blockchain.TransactionUnit, []byte("key1"), []byte("aaa"))
	val, err := b.Get(blockchain.BlockUnit, []byte("key1"))

	assert.NotNil(t, err, "expected error but got nil")
	assert.Nil(t, val, "not expected to find key")

	val, err = b.Get(blockchain.TransactionUnit, []byte("key1"))

	assert.Nil(t, err, "expected error but got nil")
	assert.NotNil(t, val, "expected to find key")
	assert.Equal(t, val, []byte("aaa"))
}

func TestPutErrOnWrongUnitType(t *testing.T) {
	cfg := blockchainConfig()
	b, err := blockchain.NewBlockChain(cfg)
	defer func() {
		err := b.Destroy()
		assert.Nil(t, err, "Unable to destroy blockchain")
	}()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	err = b.Put(100, []byte("key1"), []byte("aaa"))

	assert.NotNil(t, err, "expected error but got nil")
}

func TestPutOk(t *testing.T) {
	cfg := blockchainConfig()
	b, err := blockchain.NewBlockChain(cfg)
	defer func() {
		err := b.Destroy()
		assert.Nil(t, err, "Unable to destroy blockchain")
	}()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	err = b.Put(blockchain.TransactionUnit, []byte("key1"), []byte("aaa"))
	assert.Nil(t, err, "expected error but got nil")

	val, err := b.Get(blockchain.TransactionUnit, []byte("key1"))

	assert.Nil(t, err, "expected error but got nil")
	assert.NotNil(t, val, "expected to find key")
	assert.Equal(t, val, []byte("aaa"))
}

func TestGetAllErrOnWrongUnitType(t *testing.T) {
	cfg := blockchainConfig()
	b, err := blockchain.NewBlockChain(cfg)
	defer func() {
		err := b.Destroy()
		assert.Nil(t, err, "Unable to destroy blockchain")
	}()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	keys := [][]byte{[]byte("key1"), []byte("key2")}

	m, err := b.GetAll(100, keys)

	assert.NotNil(t, err, "expected error but got nil")
	assert.Nil(t, m, "expected nil map but got %s", m)
}

func TestGetAllOk(t *testing.T) {
	cfg := blockchainConfig()
	b, err := blockchain.NewBlockChain(cfg)
	defer func() {
		err := b.Destroy()
		assert.Nil(t, err, "Unable to destroy blockchain")
	}()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	_ = b.Put(blockchain.TransactionUnit, []byte("key1"), []byte("value1"))
	_ = b.Put(blockchain.TransactionUnit, []byte("key2"), []byte("value2"))

	keys := [][]byte{[]byte("key1"), []byte("key2")}

	m, err := b.GetAll(blockchain.TransactionUnit, keys)

	assert.Nil(t, err, "no expected error but got %s", err)
	assert.NotNil(t, m, "expected valid map but got nil")
	assert.Equal(t, len(m), 2)
}
