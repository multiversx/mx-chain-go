package blockchain_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/stretchr/testify/assert"
)

type blockChainUnits struct {
	txBadBlockCache storage.Cacher
	headerUnit      storage.Storer
	peerBlockUnit   storage.Storer
	stateBlockUnit  storage.Storer
	txBlockUnit     storage.Storer
	txUnit          storage.Storer
}

func createUnits() *blockChainUnits {
	blUnits := &blockChainUnits{}

	cacher := storage.CacheConfig{Type: storage.LRUCache, Size: 100}
	bloom := storage.BloomConfig{Size: 2048, HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv}}

	persisterTxBlockBodyStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "TxBlockBodyStorage"}
	persisterStateBlockBodyStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "StateBlockBodyStorage"}
	persisterPeerBlockBodyStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "PeerBlockBodyStorage"}
	persisterBlockHeaderStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "BlockHeaderStorage"}
	persisterTxStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "TxStorage"}

	txBadBlockCache, _ := storage.NewCache(cacher.Type, cacher.Size)

	blUnits.txBadBlockCache = txBadBlockCache

	blUnits.txUnit, _ = storage.NewStorageUnitFromConf(
		cacher,
		persisterTxStorage,
		bloom)

	blUnits.txBlockUnit, _ = storage.NewStorageUnitFromConf(
		cacher,
		persisterTxBlockBodyStorage,
		bloom)

	blUnits.stateBlockUnit, _ = storage.NewStorageUnitFromConf(
		cacher,
		persisterStateBlockBodyStorage,
		bloom)

	blUnits.peerBlockUnit, _ = storage.NewStorageUnitFromConf(
		cacher,
		persisterPeerBlockBodyStorage,
		bloom)

	blUnits.headerUnit, _ = storage.NewStorageUnitFromConf(
		cacher,
		persisterBlockHeaderStorage,
		bloom)

	return blUnits
}

func (blUnits *blockChainUnits) cleanupBlockchainUnits() {
	// cleanup
	if blUnits.headerUnit != nil {
		_ = blUnits.headerUnit.DestroyUnit()
	}
	if blUnits.peerBlockUnit != nil {
		_ = blUnits.peerBlockUnit.DestroyUnit()
	}
	if blUnits.stateBlockUnit != nil {
		_ = blUnits.stateBlockUnit.DestroyUnit()
	}
	if blUnits.txBlockUnit != nil {
		_ = blUnits.txBlockUnit.DestroyUnit()
	}
	if blUnits.txUnit != nil {
		_ = blUnits.txUnit.DestroyUnit()
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
	blockChainUnits := createUnits()

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
	blockChainUnits := createUnits()

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
	blockChainUnits := createUnits()

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
	blockChainUnits := createUnits()

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
	blockChainUnits := createUnits()

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
	blockChainUnits := createUnits()

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

	blockChainUnits := createUnits()

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
	blockChainUnits := createUnits()

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
	blockChainUnits := createUnits()

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
	blockChainUnits := createUnits()

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
	blockChainUnits := createUnits()

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
	blockChainUnits := createUnits()

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
	blockChainUnits := createUnits()

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
	blockChainUnits := createUnits()

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
	blockChainUnits := createUnits()

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

func TestBlockChain_GetStorer(t *testing.T) {
	t.Parallel()

	txUnit := &mock.StorerStub{}
	txBlockUnit := &mock.StorerStub{}
	stateBlockUnit := &mock.StorerStub{}
	peerBlockUnit := &mock.StorerStub{}
	headerUnit := &mock.StorerStub{}

	blkc, _ := blockchain.NewBlockChain(
		&mock.CacherStub{},
		txUnit,
		txBlockUnit,
		stateBlockUnit,
		peerBlockUnit,
		headerUnit,
	)

	assert.True(t, txUnit == blkc.GetStorer(blockchain.TransactionUnit))
	assert.True(t, txBlockUnit == blkc.GetStorer(blockchain.TxBlockBodyUnit))
	assert.True(t, stateBlockUnit == blkc.GetStorer(blockchain.StateBlockBodyUnit))
	assert.True(t, peerBlockUnit == blkc.GetStorer(blockchain.PeerBlockBodyUnit))
	assert.True(t, headerUnit == blkc.GetStorer(blockchain.BlockHeaderUnit))

}
