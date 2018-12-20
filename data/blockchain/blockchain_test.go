package blockchain_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/stretchr/testify/assert"
)

type blockChainUnits struct {
	txBadBlockCache storage.Cacher
	headerUnit      *storage.Unit
	peerBlockUnit   *storage.Unit
	stateBlockUnit  *storage.Unit
	txBlockUnit     *storage.Unit
	txUnit          *storage.Unit
}

func blockchainConfig() *blockchain.Config {
	cacher := storage.CacheConfig{Type: storage.LRUCache, Size: 100}
	bloom := storage.BloomConfig{Size: 2048, HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv}}
	persisterTxBlockBodyStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "TxBlockBodyStorage"}
	persisterStateBlockBodyStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "StateBlockBodyStorage"}
	persisterPeerBlockBodyStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "PeerBlockBodyStorage"}
	persisterBlockHeaderStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "BlockHeaderStorage"}
	persisterTxStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "TxStorage"}
	return &blockchain.Config{
		TxBlockBodyStorage:    storage.UnitConfig{CacheConf: cacher, DBConf: persisterTxBlockBodyStorage, BloomConf: bloom},
		StateBlockBodyStorage: storage.UnitConfig{CacheConf: cacher, DBConf: persisterStateBlockBodyStorage, BloomConf: bloom},
		PeerBlockBodyStorage:  storage.UnitConfig{CacheConf: cacher, DBConf: persisterPeerBlockBodyStorage, BloomConf: bloom},
		BlockHeaderStorage:    storage.UnitConfig{CacheConf: cacher, DBConf: persisterBlockHeaderStorage, BloomConf: bloom},
		TxStorage:             storage.UnitConfig{CacheConf: cacher, DBConf: persisterTxStorage, BloomConf: bloom},
		TxPoolStorage:         cacher,
		TxBadBlockBodyCache:   cacher,
	}
}

func createUnitsFromConfig(blConfig *blockchain.Config) *blockChainUnits {
	blUnits := &blockChainUnits{}

	txBadBlockCache, _ := storage.NewCache(blConfig.TxBadBlockBodyCache.Type, blConfig.TxBadBlockBodyCache.Size)

	blUnits.txBadBlockCache = txBadBlockCache

	blUnits.txUnit, _ = storage.NewStorageUnitFromConf(
		blConfig.TxStorage.CacheConf,
		blConfig.TxStorage.DBConf,
		blConfig.TxStorage.BloomConf)

	blUnits.txBlockUnit, _ = storage.NewStorageUnitFromConf(
		blConfig.TxBlockBodyStorage.CacheConf,
		blConfig.TxBlockBodyStorage.DBConf,
		blConfig.TxBlockBodyStorage.BloomConf)

	blUnits.stateBlockUnit, _ = storage.NewStorageUnitFromConf(
		blConfig.StateBlockBodyStorage.CacheConf,
		blConfig.StateBlockBodyStorage.DBConf,
		blConfig.StateBlockBodyStorage.BloomConf)

	blUnits.peerBlockUnit, _ = storage.NewStorageUnitFromConf(
		blConfig.PeerBlockBodyStorage.CacheConf,
		blConfig.PeerBlockBodyStorage.DBConf,
		blConfig.PeerBlockBodyStorage.BloomConf)

	blUnits.headerUnit, _ = storage.NewStorageUnitFromConf(
		blConfig.BlockHeaderStorage.CacheConf,
		blConfig.BlockHeaderStorage.DBConf,
		blConfig.BlockHeaderStorage.BloomConf)

	return blUnits
}

func (blUnits *blockChainUnits) cleanupBlockchainUnits() {
	// cleanup
	if blUnits.headerUnit != nil {
		blUnits.headerUnit.DestroyUnit()
	}
	if blUnits.peerBlockUnit != nil {
		blUnits.peerBlockUnit.DestroyUnit()
	}
	if blUnits.stateBlockUnit != nil {
		blUnits.stateBlockUnit.DestroyUnit()
	}
	if blUnits.txBlockUnit != nil {
		blUnits.txBlockUnit.DestroyUnit()
	}
	if blUnits.txUnit != nil {
		blUnits.txUnit.DestroyUnit()
	}
}

func failOnPanic(t *testing.T) {
	if r := recover(); r != nil {
		t.Errorf("the code entered panic")
	}
}

func logError(err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
	return
}

func TestNewBlockchainNilBadBlockCacheShouldError(t *testing.T) {
	cfg := blockchainConfig()
	blockChainUnits := createUnitsFromConfig(cfg)

	_, err := blockchain.NewBlockChain(
		nil,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		blockChainUnits.stateBlockUnit,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit)

	blockChainUnits.cleanupBlockchainUnits()

	assert.Equal(t, err, blockchain.ErrBadBlocksCacheNil)
}

func TestNewBlockchainNilTxUnitShouldError(t *testing.T) {
	cfg := blockchainConfig()
	blockChainUnits := createUnitsFromConfig(cfg)

	_, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		nil,
		blockChainUnits.txBlockUnit,
		blockChainUnits.stateBlockUnit,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit)

	blockChainUnits.cleanupBlockchainUnits()

	assert.Equal(t, err, blockchain.ErrTxUnitNil)
}

func TestNewBlockchainNilTxBlockUnitShouldError(t *testing.T) {
	cfg := blockchainConfig()

	blockChainUnits := createUnitsFromConfig(cfg)

	_, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		nil,
		blockChainUnits.stateBlockUnit,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit)

	blockChainUnits.cleanupBlockchainUnits()

	assert.Equal(t, err, blockchain.ErrTxBlockUnitNil)
}

func TestNewBlockchainNilStateBlockUnitShouldError(t *testing.T) {
	cfg := blockchainConfig()
	blockChainUnits := createUnitsFromConfig(cfg)

	_, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		nil,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit)

	blockChainUnits.cleanupBlockchainUnits()
	assert.Equal(t, err, blockchain.ErrStateBlockUnitNil)
}

func TestNewBlockchainNilPeerBlockUnitShouldError(t *testing.T) {
	cfg := blockchainConfig()
	blockChainUnits := createUnitsFromConfig(cfg)

	_, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		blockChainUnits.stateBlockUnit,
		nil,
		blockChainUnits.headerUnit)

	blockChainUnits.cleanupBlockchainUnits()
	assert.Equal(t, err, blockchain.ErrPeerBlockUnitNil)
}

func TestNewBlockchainNilHeaderUnitShouldError(t *testing.T) {
	cfg := blockchainConfig()

	blockChainUnits := createUnitsFromConfig(cfg)

	_, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		blockChainUnits.stateBlockUnit,
		blockChainUnits.peerBlockUnit,
		nil)

	blockChainUnits.cleanupBlockchainUnits()

	assert.Equal(t, err, blockchain.ErrHeaderUnitNil)
}

func TestNewBlockchainConfigOK(t *testing.T) {
	defer failOnPanic(t)
	cfg := blockchainConfig()
	blockChainUnits := createUnitsFromConfig(cfg)

	b, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		blockChainUnits.stateBlockUnit,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit)

	defer func() {
		err := b.Destroy()
		assert.Nil(t, err, "Unable to destroy blockchain")
	}()

	assert.Nil(t, err)
	assert.NotNil(t, b)
}

func TestHasFalseOnWrongUnitType(t *testing.T) {
	cfg := blockchainConfig()
	blockChainUnits := createUnitsFromConfig(cfg)

	b, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		blockChainUnits.stateBlockUnit,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit)

	defer func() {
		err := b.Destroy()
		assert.Nil(t, err, "Unable to destroy blockchain")
	}()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	err = b.Put(blockchain.TransactionUnit, []byte("key1"), []byte("aaa"))
	logError(err)
	has, err := b.Has(100, []byte("key1"))

	assert.NotNil(t, err, "expected error but got nil")
	assert.False(t, has, "not expected to find key")
}

func TestHasOk(t *testing.T) {
	cfg := blockchainConfig()
	blockChainUnits := createUnitsFromConfig(cfg)

	b, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		blockChainUnits.stateBlockUnit,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit)

	defer func() {
		err := b.Destroy()
		assert.Nil(t, err, "Unable to destroy blockchain")
	}()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	err = b.Put(blockchain.TransactionUnit, []byte("key1"), []byte("aaa"))
	logError(err)
	has, err := b.Has(blockchain.TxBlockBodyUnit, []byte("key1"))

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.False(t, has, "not expected to find key")

	err = b.Put(blockchain.TxBlockBodyUnit, []byte("key1"), []byte("bbb"))
	logError(err)
	has, err = b.Has(blockchain.BlockHeaderUnit, []byte("key1"))

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.False(t, has, "not expected to find key")

	err = b.Put(blockchain.BlockHeaderUnit, []byte("key1"), []byte("ccc"))
	logError(err)
	has, err = b.Has(blockchain.TransactionUnit, []byte("key1"))

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.True(t, has, "expected to find key")

	has, err = b.Has(blockchain.TxBlockBodyUnit, []byte("key1"))

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.True(t, has, "expected to find key")

	has, err = b.Has(blockchain.BlockHeaderUnit, []byte("key1"))

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.True(t, has, "expected to find key")
}

func TestGetErrOnWrongUnitType(t *testing.T) {
	cfg := blockchainConfig()
	blockChainUnits := createUnitsFromConfig(cfg)

	b, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		blockChainUnits.stateBlockUnit,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit)

	defer func() {
		err := b.Destroy()
		assert.Nil(t, err, "Unable to destroy blockchain")
	}()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	err = b.Put(blockchain.TransactionUnit, []byte("key1"), []byte("aaa"))
	logError(err)
	val, err := b.Get(100, []byte("key1"))

	assert.Equal(t, err, blockchain.ErrNoSuchStorageUnit)
	assert.Nil(t, val, "not expected to find key")
}

func TestGetOk(t *testing.T) {
	cfg := blockchainConfig()
	blockChainUnits := createUnitsFromConfig(cfg)

	b, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		blockChainUnits.stateBlockUnit,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit)

	defer func() {
		err := b.Destroy()
		assert.Nil(t, err, "Unable to destroy blockchain")
	}()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	err = b.Put(blockchain.TransactionUnit, []byte("key1"), []byte("aaa"))
	logError(err)
	val, err := b.Get(blockchain.TxBlockBodyUnit, []byte("key1"))

	assert.NotNil(t, err, "expected error but got nil")
	assert.Nil(t, val, "not expected to find key")

	val, err = b.Get(blockchain.TransactionUnit, []byte("key1"))

	assert.Nil(t, err, "expected error but got nil")
	assert.NotNil(t, val, "expected to find key")
	assert.Equal(t, val, []byte("aaa"))
}

func TestPutErrOnWrongUnitType(t *testing.T) {
	cfg := blockchainConfig()
	blockChainUnits := createUnitsFromConfig(cfg)

	b, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		blockChainUnits.stateBlockUnit,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit)

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
	blockChainUnits := createUnitsFromConfig(cfg)

	b, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		blockChainUnits.stateBlockUnit,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit)

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
	blockChainUnits := createUnitsFromConfig(cfg)

	b, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		blockChainUnits.stateBlockUnit,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit)

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
	blockChainUnits := createUnitsFromConfig(cfg)

	b, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		blockChainUnits.stateBlockUnit,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit)

	defer func() {
		err := b.Destroy()
		assert.Nil(t, err, "Unable to destroy blockchain")
	}()

	assert.Nil(t, err, "no error expected but got %s", err)
	assert.NotNil(t, b, "expected valid blockchain but got nil")

	err = b.Put(blockchain.TransactionUnit, []byte("key1"), []byte("value1"))
	logError(err)
	err = b.Put(blockchain.TransactionUnit, []byte("key2"), []byte("value2"))
	logError(err)

	keys := [][]byte{[]byte("key1"), []byte("key2")}

	m, err := b.GetAll(blockchain.TransactionUnit, keys)

	assert.Nil(t, err, "no expected error but got %s", err)
	assert.NotNil(t, m, "expected valid map but got nil")
	assert.Equal(t, len(m), 2)
}
