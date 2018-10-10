package blockchain_test

import (
	"ElrondNetwork/elrond-go-sandbox/config"
	"ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"testing"

	"github.com/stretchr/testify/assert"
)

func failOnNoPanic(t *testing.T) {
	if r := recover(); r == nil {
		t.Errorf("the code did not panic")
	}
}

func failOnPanic(t *testing.T) {
	if r := recover(); r != nil {
		t.Errorf("the code entered panic")
	}
}

func TestNewDataErrOnTxStorageCreationShouldPanic(t *testing.T) {
	defer failOnNoPanic(t)
	cfg := &config.TestnetBlockchainConfig.TxStorage.CacheConf.Type
	val := config.TestnetBlockchainConfig.TxStorage.CacheConf.Type

	// restore default config
	defer func(cacheType *config.CacheType, val config.CacheType) {
		*cacheType = val
	}(cfg, val)

	// e.g change the config to a not supported cache type
	config.TestnetBlockchainConfig.TxStorage.CacheConf.Type = 100
	blockchain.NewData()
}

func TestNewDataErrOnBlockStorageCreationShouldPanic(t *testing.T) {
	defer failOnNoPanic(t)

	cfg := &config.TestnetBlockchainConfig.BlockStorage.CacheConf.Type
	val := config.TestnetBlockchainConfig.BlockStorage.CacheConf.Type

	// restore default config
	defer func(cacheType *config.CacheType, val config.CacheType) {
		*cacheType = val
	}(cfg, val)

	// e.g change the config to a not supported cache type
	config.TestnetBlockchainConfig.BlockStorage.CacheConf.Type = 100
	blockchain.NewData()
}

func TestNewDataErrOnBlockHeaderStorageCreationShouldPanic(t *testing.T) {
	defer failOnNoPanic(t)

	cfg := &config.TestnetBlockchainConfig.BlockHeaderStorage.CacheConf.Type
	val := config.TestnetBlockchainConfig.BlockHeaderStorage.CacheConf.Type

	// restore default config
	defer func(cacheType *config.CacheType, val config.CacheType) {
		*cacheType = val
	}(cfg, val)

	// e.g change the config to a not supported cache type
	config.TestnetBlockchainConfig.BlockHeaderStorage.CacheConf.Type = 100
	blockchain.NewData()
}

func TestNewDataErrOnBBlockCacheCreationShouldPanic(t *testing.T) {
	defer failOnNoPanic(t)

	cfg := &config.TestnetBlockchainConfig.BBlockCache.Type
	val := config.TestnetBlockchainConfig.BBlockCache.Type

	// restore default config
	defer func(cacheType *config.CacheType, val config.CacheType) {
		*cacheType = val
	}(cfg, val)

	// e.g change the config to a not supported cache type
	config.TestnetBlockchainConfig.BBlockCache.Type = 100
	blockchain.NewData()
}

func TestNewDataDefaultConfigOK(t *testing.T) {
	defer failOnPanic(t)

	b, err := blockchain.NewData()
	defer b.Destroy()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")
}

func TestHasFalseOnWrongUnitType(t *testing.T) {
	b, err := blockchain.NewData()
	defer b.Destroy()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	b.Put(blockchain.TransactionUnit, []byte("key1"), []byte("aaa"))
	has, err := b.Has(100, []byte("key1"))

	assert.NotNil(t, err, "expected error but got nil")
	assert.False(t, has, "not expected to find key")
}

func TestHasOk(t *testing.T) {
	b, err := blockchain.NewData()
	defer b.Destroy()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	b.Put(blockchain.TransactionUnit, []byte("key1"), []byte("aaa"))
	has, err := b.Has(blockchain.BlockUnit, []byte("key1"))

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.False(t, has, "not expected to find key")

	b.Put(blockchain.BlockUnit, []byte("key1"), []byte("bbb"))
	has, err = b.Has(blockchain.BlockHeaderUnit, []byte("key1"))

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.False(t, has, "not expected to find key")

	b.Put(blockchain.BlockHeaderUnit, []byte("key1"), []byte("ccc"))
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
	b, err := blockchain.NewData()
	defer b.Destroy()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	b.Put(blockchain.TransactionUnit, []byte("key1"), []byte("aaa"))
	val, err := b.Get(100, []byte("key1"))

	assert.NotNil(t, err, "expected error but got nil")
	assert.Nil(t, val, "not expected to find key")
}

func TestGetOk(t *testing.T) {
	b, err := blockchain.NewData()
	defer b.Destroy()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	b.Put(blockchain.TransactionUnit, []byte("key1"), []byte("aaa"))
	val, err := b.Get(blockchain.BlockUnit, []byte("key1"))

	assert.NotNil(t, err, "expected error but got nil")
	assert.Nil(t, val, "not expected to find key")

	val, err = b.Get(blockchain.TransactionUnit, []byte("key1"))

	assert.Nil(t, err, "expected error but got nil")
	assert.NotNil(t, val, "expected to find key")
	assert.Equal(t, val, []byte("aaa"))
}

func TestPutErrOnWrongUnitType(t *testing.T) {
	b, err := blockchain.NewData()
	defer b.Destroy()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	err = b.Put(100, []byte("key1"), []byte("aaa"))

	assert.NotNil(t, err, "expected error but got nil")
}

func TestPutOk(t *testing.T) {
	b, err := blockchain.NewData()
	defer b.Destroy()

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
	b, err := blockchain.NewData()
	defer b.Destroy()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	keys := [][]byte{[]byte("key1"), []byte("key2")}

	m, err := b.GetAll(100, keys)

	assert.NotNil(t, err, "expected error but got nil")
	assert.Nil(t, m, "expected nil map but got %s", m)
}

func TestGetAllOk(t *testing.T) {
	b, err := blockchain.NewData()
	defer b.Destroy()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	b.Put(blockchain.TransactionUnit, []byte("key1"), []byte("value1"))
	b.Put(blockchain.TransactionUnit, []byte("key2"), []byte("value2"))

	keys := [][]byte{[]byte("key1"), []byte("key2")}

	m, err := b.GetAll(blockchain.TransactionUnit, keys)

	assert.Nil(t, err, "no expected error but got %s", err)
	assert.NotNil(t, m, "expected valid map but got nil")
	assert.Equal(t, len(m), 2)
}
