package synchBlock_test

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/synchBlock"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// WaitTime defines the time in milliseconds until node waits the requested info from the network
const WaitTime = time.Duration(100 * time.Millisecond)

func TestNewBootstrap(t *testing.T) {
	bs := synchBlock.NewBootstrap(nil, nil, nil, nil, WaitTime)
	assert.NotNil(t, bs)
}

func TestBootstrap_SynchBlock(t *testing.T) {

	ebm := mock.ExecBlockMock{}
	ebm.ProcessBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, body *block.Block) error {
		return nil
	}

	hdr := block.Header{Nonce: 1}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	bp := blockPool.NewBlockPool(nil)

	bs := synchBlock.NewBootstrap(bp, &blkc, nil, &ebm, WaitTime)
	bs.OnRequestHeader = func(nounce uint64) {}
	bs.OnRequestBody = func(nounce uint64) {}

	r := bs.SynchBlock()

	assert.Equal(t, execution.ErrMissingHeader, r)

	bs.RequestHeader(2)
	go bs.ReceivedHeader(2)

	r = bs.SynchBlock()
	assert.Equal(t, execution.ErrMissingHeader, r)

	bs.RequestHeader(2)
	hdr2 := block.Header{Nonce: 2}
	bp.AddHeader(2, &hdr2)

	r = bs.SynchBlock()
	assert.Equal(t, execution.ErrMissingBody, r)

	bs.RequestBody(2)
	go bs.ReceivedBody(2)

	r = bs.SynchBlock()
	assert.Equal(t, execution.ErrMissingBody, r)

	bs.RequestBody(2)
	blk2 := block.Block{}
	bp.AddBody(2, &blk2)

	r = bs.SynchBlock()
	assert.Nil(t, r)
}

func TestBootstrap_ShouldSynch(t *testing.T) {
	bs := synchBlock.NewBootstrap(nil, nil, nil, nil, WaitTime)
	r := bs.ShouldSynch()
	assert.Equal(t, false, r)

	rnd := chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond))
	bs = synchBlock.NewBootstrap(nil, nil, rnd, nil, WaitTime)
	r = bs.ShouldSynch()
	assert.Equal(t, false, r)

	rnd = chronology.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), time.Duration(100*time.Millisecond))
	bs = synchBlock.NewBootstrap(nil, nil, rnd, nil, WaitTime)
	r = bs.ShouldSynch()
	assert.Equal(t, true, r)

	hdr := block.Header{Nonce: 0}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	rnd = chronology.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), time.Duration(100*time.Millisecond))
	bs = synchBlock.NewBootstrap(nil, &blkc, rnd, nil, WaitTime)
	r = bs.ShouldSynch()
	assert.Equal(t, false, r)

	rnd = chronology.NewRound(time.Now(), time.Now().Add(200*time.Millisecond), time.Duration(100*time.Millisecond))
	bs = synchBlock.NewBootstrap(nil, &blkc, rnd, nil, WaitTime)
	r = bs.ShouldSynch()
	assert.Equal(t, true, r)
}

func TestBootstrap_GetHeaderFromPool(t *testing.T) {
	bs := synchBlock.NewBootstrap(nil, nil, nil, nil, WaitTime)
	r := bs.GetHeaderFromPool(0)
	assert.Nil(t, r)

	bp := blockPool.NewBlockPool(nil)
	bs = synchBlock.NewBootstrap(bp, nil, nil, nil, WaitTime)
	r = bs.GetHeaderFromPool(0)
	assert.Nil(t, r)

	hdr := block.Header{Nonce: 0}

	bp.AddHeader(0, &hdr)
	r = bs.GetHeaderFromPool(0)
	assert.Equal(t, &hdr, r)
}

func TestBootstrap_GetBlockFromPool(t *testing.T) {
	bs := synchBlock.NewBootstrap(nil, nil, nil, nil, WaitTime)
	r := bs.GetBodyFromPool(0)
	assert.Nil(t, r)

	bp := blockPool.NewBlockPool(nil)
	bs = synchBlock.NewBootstrap(bp, nil, nil, nil, WaitTime)
	r = bs.GetBodyFromPool(0)
	assert.Nil(t, r)

	blk := block.Block{}

	bp.AddBody(0, &blk)
	r = bs.GetBodyFromPool(0)
	assert.Equal(t, &blk, r)
}
