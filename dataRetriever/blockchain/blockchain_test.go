package blockchain

import (
	"fmt"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestBlockChain_SetCurrentBlockHeader(t *testing.T) {
	t.Parallel()

	bc, _ := NewBlockChain(&mock.AppStatusHandlerStub{})

	hdr := &block.Header{
		Nonce: 4,
	}
	hdrHash := []byte("hash")

	bc.SetCurrentBlockHeaderHash(hdrHash)

	err := bc.SetCurrentBlockHeader(hdr)
	assert.Nil(t, err)

	assert.Equal(t, hdr, bc.GetCurrentBlockHeader())
	assert.False(t, hdr == bc.GetCurrentBlockHeader())

	assert.Equal(t, hdrHash, bc.GetCurrentBlockHeaderHash())
}

func TestBlockChain_Concurrency(t *testing.T) {
	t.Parallel()

	bc, _ := NewBlockChain(&mock.AppStatusHandlerStub{})

	numCalls := 100

	header := &block.HeaderV3{}
	rootHash := []byte("rootHash")

	var wg sync.WaitGroup
	wg.Add(numCalls)

	for i := range numCalls {
		go func(i int) {
			switch i % 4 {
			case 0:
				_ = bc.GetCurrentBlockRootHash()
			case 1:
				_ = bc.SetCurrentBlockHeader(header)
			case 2:
				_ = bc.SetCurrentBlockHeaderAndRootHash(header, rootHash)
			case 3:
				_ = bc.SetGenesisHeader(header)
			default:
				require.Fail(t, "should have not been called")
			}
			wg.Done()
		}(i)
	}

	wg.Wait()
}
