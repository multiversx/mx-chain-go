package blockchain

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewBlockChain_ShouldWork(t *testing.T) {
	t.Parallel()

	bc, _ := NewBlockChain(&mock.AppStatusHandlerStub{})

	assert.False(t, check.IfNil(bc))
}

func TestBlockChain_NilAppStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	bc, err := NewBlockChain(nil)

	assert.Equal(t, ErrNilAppStatusHandler, err)
	assert.Nil(t, bc)
}

func TestBlockChain_SettersAndGetters(t *testing.T) {
	t.Parallel()

	bc, _ := NewBlockChain(&mock.AppStatusHandlerStub{})

	hdr := &block.Header{
		Nonce: 4,
	}
	genesis := &block.Header{
		Nonce: 0,
	}
	hdrHash := []byte("hash")
	genesisHash := []byte("genesis hash")
	rootHash := []byte("root hash")

	bc.SetCurrentBlockHeaderHash(hdrHash)
	bc.SetGenesisHeaderHash(genesisHash)

	err := bc.SetGenesisHeader(genesis)
	assert.Nil(t, err)

	err = bc.SetCurrentBlockHeaderAndRootHash(hdr, rootHash)
	assert.Nil(t, err)

	assert.Equal(t, hdr, bc.GetCurrentBlockHeader())
	assert.False(t, hdr == bc.GetCurrentBlockHeader())

	assert.Equal(t, genesis, bc.GetGenesisHeader())
	assert.False(t, genesis == bc.GetGenesisHeader())

	assert.Equal(t, hdrHash, bc.GetCurrentBlockHeaderHash())
	assert.Equal(t, genesisHash, bc.GetGenesisHeaderHash())

	assert.Equal(t, rootHash, bc.GetCurrentBlockRootHash())
	assert.NotEqual(t, fmt.Sprintf("%p", rootHash), fmt.Sprintf("%p", bc.GetCurrentBlockRootHash()))
}

func TestBlockChain_SettersAndGettersNilValues(t *testing.T) {
	t.Parallel()

	bc, _ := NewBlockChain(&mock.AppStatusHandlerStub{})
	_ = bc.SetGenesisHeader(&block.Header{})
	_ = bc.SetCurrentBlockHeaderAndRootHash(&block.Header{}, []byte("root hash"))

	err := bc.SetGenesisHeader(nil)
	assert.Nil(t, err)

	err = bc.SetCurrentBlockHeaderAndRootHash(nil, nil)
	assert.Nil(t, err)

	assert.Nil(t, bc.GetGenesisHeader())
	assert.Nil(t, bc.GetCurrentBlockHeader())
	assert.Empty(t, bc.GetCurrentBlockRootHash())
}

func TestBlockChain_SettersInvalidValues(t *testing.T) {
	t.Parallel()

	bc, _ := NewBlockChain(&mock.AppStatusHandlerStub{})
	err := bc.SetGenesisHeader(&block.MetaBlock{})
	assert.Equal(t, err, data.ErrInvalidHeaderType)

	err = bc.SetCurrentBlockHeaderAndRootHash(&block.MetaBlock{}, []byte("root hash"))
	assert.Equal(t, err, data.ErrInvalidHeaderType)
}
