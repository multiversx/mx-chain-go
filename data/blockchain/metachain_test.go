package blockchain_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/blockchain"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewMetaChain_ShouldWork(t *testing.T) {
	t.Parallel()

	mc := blockchain.NewMetaChain()

	assert.False(t, check.IfNil(mc))
}

func TestMetaChain_SetNilAppStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	mc := blockchain.NewMetaChain()

	err := mc.SetAppStatusHandler(nil)

	assert.Equal(t, blockchain.ErrNilAppStatusHandler, err)
}

func TestMetaChain_SetAppStatusHandlerShouldWork(t *testing.T) {
	t.Parallel()

	mc := blockchain.NewMetaChain()

	ash := &mock.AppStatusHandlerStub{}
	err := mc.SetAppStatusHandler(ash)

	assert.Nil(t, err)
}

func TestMetaChain_SettersAndGetters(t *testing.T) {
	t.Parallel()

	mc := blockchain.NewMetaChain()

	hdr := &block.MetaBlock{
		Nonce: 4,
	}
	genesis := &block.MetaBlock{
		Nonce: 0,
	}
	hdrHash := []byte("hash")
	genesisHash := []byte("genesis hash")

	mc.SetCurrentBlockHeaderHash(hdrHash)
	mc.SetGenesisHeaderHash(genesisHash)

	err := mc.SetGenesisHeader(genesis)
	assert.Nil(t, err)

	err = mc.SetCurrentBlockHeader(hdr)
	assert.Nil(t, err)

	assert.Equal(t, hdr, mc.GetCurrentBlockHeader())
	assert.False(t, hdr == mc.GetCurrentBlockHeader())

	assert.Equal(t, genesis, mc.GetGenesisHeader())
	assert.False(t, genesis == mc.GetGenesisHeader())

	assert.Equal(t, hdrHash, mc.GetCurrentBlockHeaderHash())
	assert.Equal(t, genesisHash, mc.GetGenesisHeaderHash())
}

func TestMetaChain_SettersAndGettersNilValues(t *testing.T) {
	t.Parallel()

	mc := blockchain.NewMetaChain()

	err := mc.SetGenesisHeader(nil)
	assert.Nil(t, err)

	err = mc.SetCurrentBlockHeader(nil)
	assert.Nil(t, err)

	assert.Nil(t, mc.GetGenesisHeader())
	assert.Nil(t, mc.GetCurrentBlockHeader())
}

func TestMetaChain_CreateNewHeader(t *testing.T) {
	t.Parallel()

	mc := blockchain.NewMetaChain()

	assert.Equal(t, &block.MetaBlock{}, mc.CreateNewHeader())
}
