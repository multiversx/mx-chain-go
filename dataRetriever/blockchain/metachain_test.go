package blockchain

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewMetaChain_ShouldWork(t *testing.T) {
	t.Parallel()

	mc, err := NewMetaChain(&mock.AppStatusHandlerStub{})

	assert.Nil(t, err)
	assert.False(t, check.IfNil(mc))
}

func TestMetaChain_NilAppStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	mc, err := NewMetaChain(nil)

	assert.Equal(t, ErrNilAppStatusHandler, err)
	assert.Nil(t, mc)
}

func TestMetaChain_SettersAndGetters(t *testing.T) {
	t.Parallel()

	mc, _ := NewMetaChain(&mock.AppStatusHandlerStub{})

	hdr := &block.MetaBlock{
		Nonce: 4,
	}
	genesis := &block.MetaBlock{
		Nonce: 0,
	}
	hdrHash := []byte("hash")
	genesisHash := []byte("genesis hash")
	rootHash := []byte("root hash")

	mc.SetCurrentBlockHeaderHash(hdrHash)
	mc.SetGenesisHeaderHash(genesisHash)

	err := mc.SetGenesisHeader(genesis)
	assert.Nil(t, err)

	err = mc.SetCurrentBlockHeaderAndRootHash(hdr, rootHash)
	assert.Nil(t, err)

	assert.Equal(t, hdr, mc.GetCurrentBlockHeader())
	assert.False(t, hdr == mc.GetCurrentBlockHeader())

	assert.Equal(t, genesis, mc.GetGenesisHeader())
	assert.False(t, genesis == mc.GetGenesisHeader())

	assert.Equal(t, hdrHash, mc.GetCurrentBlockHeaderHash())
	assert.Equal(t, genesisHash, mc.GetGenesisHeaderHash())

	assert.Equal(t, rootHash, mc.GetCurrentBlockRootHash())
	assert.NotEqual(t, fmt.Sprintf("%p", rootHash), fmt.Sprintf("%p", mc.GetCurrentBlockRootHash()))
}

func TestMetaChain_SettersAndGettersNilValues(t *testing.T) {
	t.Parallel()

	mc, _ := NewMetaChain(&mock.AppStatusHandlerStub{})
	_ = mc.SetGenesisHeader(&block.MetaBlock{})
	_ = mc.SetCurrentBlockHeaderAndRootHash(&block.MetaBlock{}, []byte("root hash"))

	err := mc.SetGenesisHeader(nil)
	assert.Nil(t, err)

	err = mc.SetCurrentBlockHeaderAndRootHash(nil, nil)
	assert.Nil(t, err)

	assert.Nil(t, mc.GetGenesisHeader())
	assert.Nil(t, mc.GetCurrentBlockHeader())
	assert.Empty(t, mc.GetCurrentBlockRootHash())
}

func TestMetaChain_SettersInvalidValues(t *testing.T) {
	t.Parallel()

	bc, _ := NewMetaChain(&mock.AppStatusHandlerStub{})
	err := bc.SetGenesisHeader(&block.Header{})
	assert.Equal(t, err, ErrWrongTypeInSet)

	err = bc.SetCurrentBlockHeaderAndRootHash(&block.Header{}, []byte("root hash"))
	assert.Equal(t, err, ErrWrongTypeInSet)
}
