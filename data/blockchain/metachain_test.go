package blockchain_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/stretchr/testify/assert"
)

func TestNewMetaChainNilBadBlockCacheShouldError(t *testing.T) {
	_, err := blockchain.NewMetaChain(
		nil)

	assert.Equal(t, err, blockchain.ErrBadBlocksCacheNil)
}

func TestNewMetachainConfigOK(t *testing.T) {
	badBlocksStub := &mock.CacherStub{}
	b, err := blockchain.NewMetaChain(
		badBlocksStub,
	)

	assert.Nil(t, err)
	assert.NotNil(t, b)
}

func TestMetaChain_IsBadBlock(t *testing.T) {
	badBlocksStub := &mock.CacherStub{}
	hasReturns := true
	badBlocksStub.HasCalled = func(key []byte) bool {
		return hasReturns
	}

	b, _ := blockchain.NewMetaChain(
		badBlocksStub)

	hasBadBlock := b.HasBadBlock([]byte("test"))
	assert.True(t, hasBadBlock)
}

func TestMetaChain_PutBadBlock(t *testing.T) {
	badBlocksStub := &mock.CacherStub{}
	putCalled := false
	badBlocksStub.PutCalled = func(key []byte, value interface{}) bool {
		putCalled = true
		return true
	}

	b, _ := blockchain.NewMetaChain(
		badBlocksStub,
	)

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

func TestMetaChain_GetGenesisBlock(t *testing.T) {
	t.Parallel()

	bl := &block.MetaBlock{}
	m := blockchain.MetaChain{GenesisBlock: bl}

	assert.Equal(t, bl, m.GetGenesisHeader())
}

func TestMetaChain_SetCurrentBlockBody(t *testing.T) {
	t.Parallel()

	m := blockchain.MetaChain{}
	err := m.SetCurrentBlockBody(block.Body{})

	assert.Nil(t, err)
	assert.Equal(t, nil, m.GetCurrentBlockBody())
}

func TestMetaChain_SetCurrentBlockHeader(t *testing.T) {
	t.Parallel()

	bl := &block.MetaBlock{}
	m := blockchain.MetaChain{}
	_ = m.SetAppStatusHandler(statusHandler.NewNilStatusHandler())
	err := m.SetCurrentBlockHeader(bl)

	assert.Nil(t, err)
	assert.Equal(t, bl, m.CurrentBlock)
}

func TestMetaChain_SetCurrentBlockHeaderWrongType(t *testing.T) {
	t.Parallel()

	bl := &block.Header{}
	m := blockchain.MetaChain{}
	err := m.SetCurrentBlockHeader(bl)

	assert.NotNil(t, err)
	assert.Equal(t, blockchain.ErrWrongTypeInSet, err)
}

func TestMetaChain_SetCurrentBlockHeaderHash(t *testing.T) {
	t.Parallel()

	h := []byte{10, 11, 12, 13}
	m := blockchain.MetaChain{}
	m.SetCurrentBlockHeaderHash(h)

	assert.Equal(t, h, m.GetCurrentBlockHeaderHash())
}

func TestMetaChain_SetGenesisBlock(t *testing.T) {
	t.Parallel()

	bl := &block.MetaBlock{}
	m := blockchain.MetaChain{}
	err := m.SetGenesisHeader(bl)

	assert.Nil(t, err)
	assert.Equal(t, bl, m.GenesisBlock)
}

func TestMetaChain_SetGenesisBlockWrongType(t *testing.T) {
	t.Parallel()

	bl := &block.Header{}
	m := blockchain.MetaChain{}
	err := m.SetGenesisHeader(bl)

	assert.Equal(t, blockchain.ErrWrongTypeInSet, err)
}

func TestMetaChain_SetGenesisHeaderHash(t *testing.T) {
	t.Parallel()

	h := []byte{10, 11, 12, 13}
	m := blockchain.MetaChain{}
	m.SetGenesisHeaderHash(h)

	assert.Equal(t, h, m.GetGenesisHeaderHash())
}

func TestMetaChain_SetLocalHeight(t *testing.T) {
	t.Parallel()

	height := int64(3)
	m := blockchain.MetaChain{}
	m.SetLocalHeight(height)

	assert.Equal(t, height, m.GetLocalHeight())
}

func TestMetaChain_SetNetworkHeight(t *testing.T) {
	t.Parallel()

	height := int64(3)
	m := blockchain.MetaChain{}
	m.SetNetworkHeight(height)

	assert.Equal(t, height, m.GetNetworkHeight())
}
