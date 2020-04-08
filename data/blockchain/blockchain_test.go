package blockchain_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewBlockChain_ShouldWork(t *testing.T) {
	t.Parallel()

	bc := blockchain.NewBlockChain()

	assert.False(t, check.IfNil(bc))
}

func TestBlockChain_SetNilAppStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	bc := blockchain.NewBlockChain()

	err := bc.SetAppStatusHandler(nil)

	assert.Equal(t, blockchain.ErrNilAppStatusHandler, err)
}

func TestBlockChain_SetAppStatusHandlerShouldWork(t *testing.T) {
	t.Parallel()

	bc := blockchain.NewBlockChain()

	ash := &mock.AppStatusHandlerStub{}
	err := bc.SetAppStatusHandler(ash)

	assert.Nil(t, err)
}

func TestBlockChain_SettersAndGetters(t *testing.T) {
	t.Parallel()

	bc := blockchain.NewBlockChain()

	hdr := &block.Header{
		Nonce: 4,
	}
	genesis := &block.Header{
		Nonce: 0,
	}
	hdrHash := []byte("hash")
	genesisHash := []byte("genesis hash")

	bc.SetCurrentBlockHeaderHash(hdrHash)
	bc.SetGenesisHeaderHash(genesisHash)

	err := bc.SetGenesisHeader(genesis)
	assert.Nil(t, err)

	err = bc.SetCurrentBlockHeader(hdr)
	assert.Nil(t, err)

	assert.Equal(t, hdr, bc.GetCurrentBlockHeader())
	assert.False(t, hdr == bc.GetCurrentBlockHeader())

	assert.Equal(t, genesis, bc.GetGenesisHeader())
	assert.False(t, genesis == bc.GetGenesisHeader())

	assert.Equal(t, hdrHash, bc.GetCurrentBlockHeaderHash())
	assert.Equal(t, genesisHash, bc.GetGenesisHeaderHash())
}

func TestBlockChain_SettersAndGettersNilValues(t *testing.T) {
	t.Parallel()

	bc := blockchain.NewBlockChain()

	err := bc.SetGenesisHeader(nil)
	assert.Nil(t, err)

	err = bc.SetCurrentBlockHeader(nil)
	assert.Nil(t, err)

	assert.Nil(t, bc.GetGenesisHeader())
	assert.Nil(t, bc.GetCurrentBlockHeader())
}

func TestBlockChain_CreateNewHeader(t *testing.T) {
	t.Parallel()

	bc := blockchain.NewBlockChain()

	assert.Equal(t, &block.Header{}, bc.CreateNewHeader())
}
