package blockchain_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"github.com/stretchr/testify/assert"
)

type metaChainUnits struct {
	badBlocksCache storage.Cacher
	metaBlockUnit  storage.Storer
	shardDataUnit  storage.Storer
	peerDataUnit   storage.Storer
}

func createMetaUnits() *metaChainUnits {
	blUnits := &metaChainUnits{}

	cacher := storage.CacheConfig{Type: storage.LRUCache, Size: 100}
	bloom := storage.BloomConfig{Size: 2048, HashFunc: []storage.HasherType{storage.Keccak, storage.Blake2b, storage.Fnv}}

	persisterMetaBlocksStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "MetaBlocksStorage"}
	persisterPeerDataStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "PeerDataStorage"}
	persisterShardDataStorage := storage.DBConfig{Type: storage.LvlDB, FilePath: "ShardDataStorage"}

	txBadBlockCache, _ := storage.NewCache(cacher.Type, cacher.Size)

	blUnits.badBlocksCache = txBadBlockCache

	blUnits.metaBlockUnit, _ = storage.NewStorageUnitFromConf(
		cacher,
		persisterMetaBlocksStorage,
		bloom)

	blUnits.shardDataUnit, _ = storage.NewStorageUnitFromConf(
		cacher,
		persisterPeerDataStorage,
		bloom)

	blUnits.peerDataUnit, _ = storage.NewStorageUnitFromConf(
		cacher,
		persisterShardDataStorage,
		bloom)

	return blUnits
}

func (blUnits *metaChainUnits) cleanupMetaChainUnits() {
	// cleanup
	if blUnits.metaBlockUnit != nil {
		_ = blUnits.metaBlockUnit.DestroyUnit()
	}
	if blUnits.shardDataUnit != nil {
		_ = blUnits.shardDataUnit.DestroyUnit()
	}
	if blUnits.peerDataUnit != nil {
		_ = blUnits.peerDataUnit.DestroyUnit()
	}
}

func TestNewMetaChainNilFirstBlockShouldError(t *testing.T) {
	blockChainUnits := createMetaUnits()
	genHash := make([]byte, 2)

	_, err := blockchain.NewMetaChain(
		nil,
		genHash,
		blockChainUnits.badBlocksCache,
		blockChainUnits.metaBlockUnit,
		blockChainUnits.shardDataUnit,
		blockChainUnits.peerDataUnit)

	blockChainUnits.cleanupMetaChainUnits()

	assert.Equal(t, err, blockchain.ErrMetaGenesisBlockNil)
}

func TestNewMetaChainNilFirstBlockHashShouldError(t *testing.T) {
	blockChainUnits := createMetaUnits()
	genBlock := &block.MetaBlock{}

	_, err := blockchain.NewMetaChain(
		genBlock,
		nil,
		blockChainUnits.badBlocksCache,
		blockChainUnits.metaBlockUnit,
		blockChainUnits.shardDataUnit,
		blockChainUnits.peerDataUnit)

	blockChainUnits.cleanupMetaChainUnits()

	assert.Equal(t, err, blockchain.ErrMetaGenesisBlockHashNil)
}

func TestNewMetaChainNilBadBlockCacheShouldError(t *testing.T) {
	blockChainUnits := createMetaUnits()
	genBlock := &block.MetaBlock{}
	genHash := make([]byte, 2)

	_, err := blockchain.NewMetaChain(
		genBlock,
		genHash,
		nil,
		blockChainUnits.metaBlockUnit,
		blockChainUnits.shardDataUnit,
		blockChainUnits.peerDataUnit)

	blockChainUnits.cleanupMetaChainUnits()

	assert.Equal(t, err, blockchain.ErrBadBlocksCacheNil)
}

func TestNewMetaChainNilMetaBlockUnitShouldError(t *testing.T) {
	blockChainUnits := createMetaUnits()
	genBlock := &block.MetaBlock{}
	genHash := make([]byte, 2)

	_, err := blockchain.NewMetaChain(
		genBlock,
		genHash,
		blockChainUnits.badBlocksCache,
		nil,
		blockChainUnits.shardDataUnit,
		blockChainUnits.peerDataUnit)

	blockChainUnits.cleanupMetaChainUnits()

	assert.Equal(t, err, blockchain.ErrMetaBlockUnitNil)
}

func TestNewMetaChainNilShardDataUnitShouldError(t *testing.T) {
	blockChainUnits := createMetaUnits()
	genBlock := &block.MetaBlock{}
	genHash := make([]byte, 2)

	_, err := blockchain.NewMetaChain(
		genBlock,
		genHash,
		blockChainUnits.badBlocksCache,
		blockChainUnits.metaBlockUnit,
		nil,
		blockChainUnits.peerDataUnit)

	blockChainUnits.cleanupMetaChainUnits()

	assert.Equal(t, err, blockchain.ErrShardDataUnitNil)
}

func TestNewMetaChainNilPeerDataUnitShouldError(t *testing.T) {
	blockChainUnits := createMetaUnits()
	genBlock := &block.MetaBlock{}
	genHash := make([]byte, 2)

	_, err := blockchain.NewMetaChain(
		genBlock,
		genHash,
		blockChainUnits.badBlocksCache,
		blockChainUnits.metaBlockUnit,
		blockChainUnits.shardDataUnit,
		nil)

	blockChainUnits.cleanupMetaChainUnits()

	assert.Equal(t, err, blockchain.ErrPeerDataUnitNil)
}

func TestNewMetachainConfigOK(t *testing.T) {
	defer failOnPanic(t)

	blockChainUnits := createMetaUnits()
	genBlock := &block.MetaBlock{}
	genHash := make([]byte, 2)

	b, err := blockchain.NewMetaChain(
		genBlock,
		genHash,
		blockChainUnits.badBlocksCache,
		blockChainUnits.metaBlockUnit,
		blockChainUnits.shardDataUnit,
		blockChainUnits.peerDataUnit)

	blockChainUnits.cleanupMetaChainUnits()

	defer func() {
		err := b.Destroy()
		assert.Nil(t, err, "Unable to destroy blockchain")
	}()

	assert.Nil(t, err)
	assert.NotNil(t, b)
}

func TestMetaChain_IsBadBlock(t *testing.T) {
	badBlocksStub := &mock.CacherStub{}
	metablockUnit := &mock.StorerStub{}
	sharddataUnit := &mock.StorerStub{}
	peerdataUnit := &mock.StorerStub{}
	genBlock := &block.MetaBlock{}
	genHash := make([]byte, 2)

	hasReturns := true
	badBlocksStub.HasCalled = func(key []byte) bool {
		return hasReturns
	}

	b, _ := blockchain.NewMetaChain(
		genBlock,
		genHash,
		badBlocksStub,
		metablockUnit,
		sharddataUnit,
		peerdataUnit)

	isBadBlock := b.IsBadBlock([]byte("test"))
	assert.True(t, isBadBlock)
}

func TestMetaChain_PutBadBlock(t *testing.T) {
	badBlocksStub := &mock.CacherStub{}
	metablockUnit := &mock.StorerStub{}
	sharddataUnit := &mock.StorerStub{}
	peerdataUnit := &mock.StorerStub{}
	genBlock := &block.MetaBlock{}
	genHash := make([]byte, 2)

	putCalled := false
	badBlocksStub.PutCalled = func(key []byte, value interface{}) bool {
		putCalled = true
		return true
	}

	b, _ := blockchain.NewMetaChain(
		genBlock,
		genHash,
		badBlocksStub,
		metablockUnit,
		sharddataUnit,
		peerdataUnit)

	b.PutBadBlock([]byte("test"))
	assert.True(t, putCalled)
}
