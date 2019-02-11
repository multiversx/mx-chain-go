package blockchain_test

import (
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

func TestBlockChain_IsBadBlock(t *testing.T) {
	badBlocksStub := &mock.CacherStub{}
	txUnit := &mock.StorerStub{}
	txBlockUnit := &mock.StorerStub{}
	stateBlockUnit := &mock.StorerStub{}
	peerBlockUnit := &mock.StorerStub{}
	headerUnit := &mock.StorerStub{}

	hasReturns := true
	badBlocksStub.HasCalled = func(key []byte) bool {
		return hasReturns
	}

	b, _ := blockchain.NewBlockChain(
		badBlocksStub,
		txUnit,
		txBlockUnit,
		stateBlockUnit,
		peerBlockUnit,
		headerUnit)

	isBadBlock := b.IsBadBlock([]byte("test"))
	assert.True(t, isBadBlock)
}

func TestBlockChain_PutBadBlock(t *testing.T) {
	badBlocksStub := &mock.CacherStub{}
	txUnit := &mock.StorerStub{}
	txBlockUnit := &mock.StorerStub{}
	stateBlockUnit := &mock.StorerStub{}
	peerBlockUnit := &mock.StorerStub{}
	headerUnit := &mock.StorerStub{}

	putCalled := false
	badBlocksStub.PutCalled = func(key []byte, value interface{}) bool {
		putCalled = true
		return true
	}

	b, _ := blockchain.NewBlockChain(
		badBlocksStub,
		txUnit,
		txBlockUnit,
		stateBlockUnit,
		peerBlockUnit,
		headerUnit)

	b.PutBadBlock([]byte("test"))
	assert.True(t, putCalled)
}