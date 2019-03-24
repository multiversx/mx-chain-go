package blockchain_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage/memorydb"
	"github.com/stretchr/testify/assert"
)

type blockChainUnits struct {
	txBadBlockCache storage.Cacher
	headerUnit      storage.Storer
	peerBlockUnit   storage.Storer
	stateBlockUnit  storage.Storer
	txBlockUnit     storage.Storer
	txUnit          storage.Storer
	metaChainUnits  storage.Storer
}

func createCacher() storage.Cacher {
	cacher, _ := storage.NewCache(storage.LRUCache, 100)
	return cacher
}

func createMemDb() storage.Persister {
	db, _ := memorydb.New()
	return db
}

func createUnits() *blockChainUnits {
	blUnits := &blockChainUnits{}

	blUnits.txUnit, _ = storage.NewStorageUnit(createCacher(), createMemDb())
	blUnits.txBlockUnit, _ = storage.NewStorageUnit(createCacher(), createMemDb())
	blUnits.peerBlockUnit, _ = storage.NewStorageUnit(createCacher(), createMemDb())
	blUnits.headerUnit, _ = storage.NewStorageUnit(createCacher(), createMemDb())
	blUnits.metaChainUnits, _ = storage.NewStorageUnit(createCacher(), createMemDb())
	blUnits.txBadBlockCache = createCacher()

	return blUnits
}

func failOnPanic(t *testing.T) {
	if r := recover(); r != nil {
		t.Errorf("the code entered panic")
	}
}

func TestNewBlockChain_NilBadBlockCacheShouldError(t *testing.T) {
	t.Parallel()

	blockChainUnits := createUnits()

	_, err := blockchain.NewBlockChain(
		nil,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit,
		blockChainUnits.metaChainUnits,
	)

	assert.Equal(t, err, blockchain.ErrBadBlocksCacheNil)
}

func TestNewBlockChain_NilTxUnitShouldError(t *testing.T) {
	t.Parallel()

	blockChainUnits := createUnits()

	_, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		nil,
		blockChainUnits.txBlockUnit,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit,
		blockChainUnits.metaChainUnits,
	)

	assert.Equal(t, err, blockchain.ErrTxUnitNil)
}

func TestNewBlockChain_NilTxBlockUnitShouldError(t *testing.T) {
	t.Parallel()

	blockChainUnits := createUnits()

	_, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		nil,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit,
		blockChainUnits.metaChainUnits,
	)

	assert.Equal(t, err, blockchain.ErrMiniBlockUnitNil)
}

func TestNewBlockChain_NilPeerBlockUnitShouldError(t *testing.T) {
	t.Parallel()

	blockChainUnits := createUnits()

	_, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		nil,
		blockChainUnits.headerUnit,
		blockChainUnits.metaChainUnits,
	)

	assert.Equal(t, err, blockchain.ErrPeerBlockUnitNil)
}

func TestNewBlockChain_NilHeaderUnitShouldError(t *testing.T) {
	t.Parallel()

	blockChainUnits := createUnits()

	_, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		blockChainUnits.peerBlockUnit,
		nil,
		blockChainUnits.metaChainUnits,
	)

	assert.Equal(t, err, blockchain.ErrHeaderUnitNil)
}

func TestNewBlockChain_NilMetachainHeaderUnitShouldError(t *testing.T) {
	t.Parallel()

	blockChainUnits := createUnits()

	_, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit,
		nil,
	)

	assert.Equal(t, err, blockchain.ErrMetachainHeaderUnitNil)
}

func TestNewBlockChain_ConfigOK(t *testing.T) {
	t.Parallel()

	blockChainUnits := createUnits()

	b, err := blockchain.NewBlockChain(
		blockChainUnits.txBadBlockCache,
		blockChainUnits.txUnit,
		blockChainUnits.txBlockUnit,
		blockChainUnits.peerBlockUnit,
		blockChainUnits.headerUnit,
		blockChainUnits.metaChainUnits,
	)

	assert.Nil(t, err)
	assert.NotNil(t, b)
}

func TestBlockChain_IsBadBlock(t *testing.T) {
	t.Parallel()

	badBlocksStub := &mock.CacherStub{}
	txUnit := &mock.StorerStub{}
	txBlockUnit := &mock.StorerStub{}
	peerBlockUnit := &mock.StorerStub{}
	headerUnit := &mock.StorerStub{}
	metachainHeaderUnit := &mock.StorerStub{}

	hasReturns := true
	badBlocksStub.HasCalled = func(key []byte) bool {
		return hasReturns
	}

	b, _ := blockchain.NewBlockChain(
		badBlocksStub,
		txUnit,
		txBlockUnit,
		peerBlockUnit,
		headerUnit,
		metachainHeaderUnit,
	)

	isBadBlock := b.IsBadBlock([]byte("test"))
	assert.True(t, isBadBlock)
}

func TestBlockChain_PutBadBlock(t *testing.T) {
	t.Parallel()

	badBlocksStub := &mock.CacherStub{}
	txUnit := &mock.StorerStub{}
	txBlockUnit := &mock.StorerStub{}
	peerBlockUnit := &mock.StorerStub{}
	headerUnit := &mock.StorerStub{}
	metachainHeaderUnit := &mock.StorerStub{}

	putCalled := false
	badBlocksStub.PutCalled = func(key []byte, value interface{}) bool {
		putCalled = true
		return true
	}

	b, _ := blockchain.NewBlockChain(
		badBlocksStub,
		txUnit,
		txBlockUnit,
		peerBlockUnit,
		headerUnit,
		metachainHeaderUnit,
	)

	b.PutBadBlock([]byte("test"))
	assert.True(t, putCalled)
}
