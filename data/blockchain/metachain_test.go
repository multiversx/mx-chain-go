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

func TestNewMetaChainNilBadBlockCacheShouldError(t *testing.T) {
	blockChainUnits := createMetaUnits()
	_, err := blockchain.NewMetaChain(
		nil,
		blockChainUnits.metaBlockUnit,
		blockChainUnits.shardDataUnit,
		blockChainUnits.peerDataUnit)

	blockChainUnits.cleanupMetaChainUnits()

	assert.Equal(t, err, blockchain.ErrBadBlocksCacheNil)
}

func TestNewMetaChainNilMetaBlockUnitShouldError(t *testing.T) {
	blockChainUnits := createMetaUnits()

	_, err := blockchain.NewMetaChain(
		blockChainUnits.badBlocksCache,
		nil,
		blockChainUnits.shardDataUnit,
		blockChainUnits.peerDataUnit)

	blockChainUnits.cleanupMetaChainUnits()

	assert.Equal(t, err, blockchain.ErrMetaBlockUnitNil)
}

func TestNewMetaChainNilShardDataUnitShouldError(t *testing.T) {
	blockChainUnits := createMetaUnits()
	_, err := blockchain.NewMetaChain(
		blockChainUnits.badBlocksCache,
		blockChainUnits.metaBlockUnit,
		nil,
		blockChainUnits.peerDataUnit)

	blockChainUnits.cleanupMetaChainUnits()

	assert.Equal(t, err, blockchain.ErrShardDataUnitNil)
}

func TestNewMetaChainNilPeerDataUnitShouldError(t *testing.T) {
	blockChainUnits := createMetaUnits()
	_, err := blockchain.NewMetaChain(
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
	b, err := blockchain.NewMetaChain(
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

	hasReturns := true
	badBlocksStub.HasCalled = func(key []byte) bool {
		return hasReturns
	}

	b, _ := blockchain.NewMetaChain(
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

	putCalled := false
	badBlocksStub.PutCalled = func(key []byte, value interface{}) bool {
		putCalled = true
		return true
	}

	b, _ := blockchain.NewMetaChain(
		badBlocksStub,
		metablockUnit,
		sharddataUnit,
		peerdataUnit)

	b.PutBadBlock([]byte("test"))
	assert.True(t, putCalled)
}

func TestMetaChain_GetCurrentBlockBody(t *testing.T) {
	t.Parallel()

	m := blockchain.MetaChain{}

	assert.Equal(t, nil, m.GetCurrentBlockBody())
}

func TestMetaChain_GetCurrentBlockHeader(t *testing.T) {
	t.Parallel()

	bl := &block.MetaBlock{}
	m := blockchain.MetaChain{CurrentBlock: bl}

	assert.Equal(t, bl, m.GetCurrentBlockHeader())
}

func TestMetaChain_GetCurrentBlockHeaderHash(t *testing.T) {
	t.Parallel()

	h := []byte{10, 11, 12, 13}
	m := blockchain.MetaChain{CurrentBlockHash: h}

	assert.Equal(t, h, m.GetCurrentBlockHeaderHash())
}

func TestMetaChain_GetGenesisBlock(t *testing.T) {
	t.Parallel()

	bl := &block.MetaBlock{}
	m := blockchain.MetaChain{GenesisBlock: bl}

	assert.Equal(t, bl, m.GetGenesisBlock())
}

func TestMetaChain_GetGenesisHeaderHash(t *testing.T) {
	t.Parallel()

	h := []byte{10, 11, 12, 13}
	m := blockchain.MetaChain{GenesisBlockHash: h}

	assert.Equal(t, h, m.GetGenesisHeaderHash())
}

func TestMetaChain_GetLocalHeight(t *testing.T) {
	t.Parallel()

	height := int64(3)
	m := blockchain.MetaChain{LocalHeight: height}

	assert.Equal(t, height, m.GetLocalHeight())
}

func TestMetaChain_GetNetworkHeight(t *testing.T) {
	t.Parallel()

	height := int64(4)
	m := blockchain.MetaChain{NetworkHeight: height}

	assert.Equal(t, height, m.GetNetworkHeight())
}

func TestMetaChain_SetCurrentBlockBody(t *testing.T) {
	t.Parallel()

	m := blockchain.MetaChain{}
	m.SetCurrentBlockBody(&block.Body{})

	assert.Equal(t, nil, m.GetCurrentBlockBody())
}

func TestMetaChain_SetCurrentBlockHeader(t *testing.T) {
	t.Parallel()

	bl := &block.MetaBlock{}
	m := blockchain.MetaChain{}
	m.SetCurrentBlockHeader(bl)

	assert.Equal(t, bl, m.CurrentBlock)
}

func TestMetaChain_SetCurrentBlockHeaderHash(t *testing.T) {
	t.Parallel()

	h := []byte{10, 11, 12, 13}
	m := blockchain.MetaChain{}
	m.SetCurrentBlockHeaderHash(h)

	assert.Equal(t, h, m.CurrentBlockHash)
}

func TestMetaChain_SetGenesisBlock(t *testing.T) {
	t.Parallel()

	bl := &block.MetaBlock{}
	m := blockchain.MetaChain{}
	m.SetGenesisBlock(bl)

	assert.Equal(t, bl, m.GenesisBlock)
}

func TestMetaChain_SetGenesisHeaderHash(t *testing.T) {
	t.Parallel()

	h := []byte{10, 11, 12, 13}
	m := blockchain.MetaChain{}
	m.SetGenesisHeaderHash(h)

	assert.Equal(t, h, m.GenesisBlockHash)
}

func TestMetaChain_SetLocalHeight(t *testing.T) {
	t.Parallel()

	height := int64(3)
	m := blockchain.MetaChain{}
	m.SetLocalHeight(height)

	assert.Equal(t, height, m.LocalHeight)
}

func TestMetaChain_SetNetworkHeight(t *testing.T) {
	t.Parallel()

	height := int64(3)
	m := blockchain.MetaChain{}
	m.SetNetworkHeight(height)

	assert.Equal(t, height, m.NetworkHeight)
}
