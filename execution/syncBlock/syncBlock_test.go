package syncBlock_test

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution/syncBlock"
	"github.com/stretchr/testify/assert"
)

// WaitTime defines the time in milliseconds until node waits the requested info from the network
const WaitTime = time.Duration(100 * time.Millisecond)

func TestNewBootstrapShouldThrowNilBlockPool(t *testing.T) {
	bs, err := syncBlock.NewBootstrap(nil, nil, nil, nil, WaitTime)

	assert.Nil(t, bs)
	assert.Equal(t, execution.ErrNilBlockPool, err)
}

func TestNewBootstrapShouldThrowNilBlockchain(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	bs, err := syncBlock.NewBootstrap(bp, nil, nil, nil, WaitTime)

	assert.Nil(t, bs)
	assert.Equal(t, execution.ErrNilBlockChain, err)
}

func TestNewBootstrapShouldThrowNilRound(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	blkc := blockchain.BlockChain{}
	bs, err := syncBlock.NewBootstrap(bp, &blkc, nil, nil, WaitTime)

	assert.Nil(t, bs)
	assert.Equal(t, execution.ErrNilRound, err)
}

func TestNewBootstrapShouldThrowNilBlockExecutor(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	blkc := blockchain.BlockChain{}
	rnd := chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond))
	bs, err := syncBlock.NewBootstrap(bp, &blkc, rnd, nil, WaitTime)

	assert.Nil(t, bs)
	assert.Equal(t, execution.ErrNilBlockExecutor, err)
}

func TestNewBootstrapShouldWork(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)
	blkc := blockchain.BlockChain{}
	rnd := chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond))
	ebm := mock.ExecBlockMock{}
	bs, err := syncBlock.NewBootstrap(bp, &blkc, rnd, &ebm, WaitTime)

	assert.NotNil(t, bs)
	assert.Nil(t, err)
}

func TestSyncBlockShouldReturnMissingHeader(t *testing.T) {
	hdr := block.Header{Nonce: 1}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	bp := blockPool.NewBlockPool(nil)

	bs, _ := syncBlock.NewBootstrap(
		bp,
		&blkc,
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.ExecBlockMock{},
		WaitTime)

	bs.OnRequestHeader = func(nonce uint64) {}
	bs.OnRequestBody = func(nonce uint64) {}

	r := bs.SyncBlock()

	assert.Equal(t, execution.ErrMissingHeader, r)
}

func TestSyncBlockShouldReturnMissingBody(t *testing.T) {
	hdr := block.Header{Nonce: 1}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	bp := blockPool.NewBlockPool(nil)

	bs, _ := syncBlock.NewBootstrap(
		bp,
		&blkc,
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.ExecBlockMock{},
		WaitTime)

	bs.RequestHeader(2)
	hdr2 := block.Header{Nonce: 2}
	bp.AddHeader(2, &hdr2)

	r := bs.SyncBlock()

	assert.Equal(t, execution.ErrMissingBody, r)
}

func TestStartSyncShouldNotNeedToSync(t *testing.T) {
	ebm := mock.ExecBlockMock{}
	ebm.ProcessBlockCalled = func(blk *blockchain.BlockChain, hdr *block.Header, bdy *block.Block) error {
		blk.CurrentBlockHeader = hdr
		return nil
	}

	hdr := block.Header{Nonce: 1, Round: 0}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	bp := blockPool.NewBlockPool(nil)

	bs, _ := syncBlock.NewBootstrap(
		bp,
		&blkc,
		chronology.NewRound(time.Now(), time.Now().Add(0*time.Millisecond), time.Duration(100*time.Millisecond)),
		&ebm,
		WaitTime)

	bs.StartSync()
	time.Sleep(200 * time.Millisecond)
	bs.StopSync()
}

func TestStartSyncShouldSyncOneBlock(t *testing.T) {
	ebm := mock.ExecBlockMock{}
	ebm.ProcessBlockCalled = func(blk *blockchain.BlockChain, hdr *block.Header, bdy *block.Block) error {
		blk.CurrentBlockHeader = hdr
		return nil
	}

	hdr := block.Header{Nonce: 1, Round: 0}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	bp := blockPool.NewBlockPool(nil)

	bs, _ := syncBlock.NewBootstrap(
		bp,
		&blkc,
		chronology.NewRound(time.Now(), time.Now().Add(200*time.Millisecond), time.Duration(100*time.Millisecond)),
		&ebm,
		WaitTime)

	bs.StartSync()

	time.Sleep(200 * time.Millisecond)

	hdr2 := block.Header{Nonce: 2, Round: 1}
	bp.AddHeader(2, &hdr2)

	blk2 := block.Block{}
	bp.AddBody(2, &blk2)

	time.Sleep(200 * time.Millisecond)

	bs.StopSync()
}

func TestSyncBlockShouldReturnNilErr(t *testing.T) {
	ebm := mock.ExecBlockMock{}
	ebm.ProcessBlockCalled = func(blockChain *blockchain.BlockChain, header *block.Header, body *block.Block) error {
		return nil
	}

	hdr := block.Header{Nonce: 1}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	bp := blockPool.NewBlockPool(nil)

	bs, _ := syncBlock.NewBootstrap(
		bp,
		&blkc,
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&ebm,
		WaitTime)

	bs.RequestHeader(2)
	hdr2 := block.Header{Nonce: 2}
	bp.AddHeader(2, &hdr2)

	bs.RequestBody(2)
	blk2 := block.Block{}
	bp.AddBody(2, &blk2)

	r := bs.SyncBlock()

	assert.Nil(t, r)
}

func TestShouldSyncShouldReturnFalseWhenCurrentBlockIsNilAndRoundIndexIsZero(t *testing.T) {
	bs, _ := syncBlock.NewBootstrap(
		blockPool.NewBlockPool(nil),
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.ExecBlockMock{},
		WaitTime)

	r := bs.ShouldSync()

	assert.Equal(t, false, r)
}

func TestShouldSyncShouldReturnTrueWhenCurrentBlockIsNilAndRoundIndexIsGreaterThanZero(t *testing.T) {
	bs, _ := syncBlock.NewBootstrap(
		blockPool.NewBlockPool(nil),
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), time.Duration(100*time.Millisecond)),
		&mock.ExecBlockMock{},
		WaitTime)

	r := bs.ShouldSync()

	assert.Equal(t, true, r)
}

func TestShouldSyncShouldReturnFalseWhenNodeIsSynced(t *testing.T) {
	hdr := block.Header{Nonce: 0}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	bs, _ := syncBlock.NewBootstrap(
		blockPool.NewBlockPool(nil),
		&blkc,
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.ExecBlockMock{},
		WaitTime)

	r := bs.ShouldSync()

	assert.Equal(t, false, r)
}

func TestShouldSyncShouldReturnTrueWhenNodeIsNotSynced(t *testing.T) {
	hdr := block.Header{Nonce: 0}
	blkc := blockchain.BlockChain{}
	blkc.CurrentBlockHeader = &hdr

	bs, _ := syncBlock.NewBootstrap(
		blockPool.NewBlockPool(nil),
		&blkc,
		chronology.NewRound(time.Now(), time.Now().Add(100*time.Millisecond), time.Duration(100*time.Millisecond)),
		&mock.ExecBlockMock{},
		WaitTime)

	r := bs.ShouldSync()

	assert.Equal(t, false, r)
}

func TestGetHeaderFromPoolShouldReturnNil(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)

	bs, _ := syncBlock.NewBootstrap(
		bp,
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.ExecBlockMock{},
		WaitTime)

	r := bs.GetDataFromPool(bp.HeaderStore(), 0)

	assert.Nil(t, r)
}

func TestGetHeaderFromPoolShouldReturnHeader(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)

	bs, _ := syncBlock.NewBootstrap(
		bp,
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.ExecBlockMock{},
		WaitTime)

	hdr := block.Header{Nonce: 0}

	bp.AddHeader(0, &hdr)
	r := bs.GetDataFromPool(bp.HeaderStore(), 0)

	assert.Equal(t, &hdr, r)
}

func TestGetBlockFromPoolShouldReturnNil(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)

	bs, _ := syncBlock.NewBootstrap(
		bp,
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.ExecBlockMock{},
		WaitTime)

	r := bs.GetDataFromPool(bp.BodyStore(), 0)

	assert.Nil(t, r)
}

func TestGetBlockFromPoolShouldReturnBlock(t *testing.T) {
	bp := blockPool.NewBlockPool(nil)

	bs, _ := syncBlock.NewBootstrap(
		bp,
		&blockchain.BlockChain{},
		chronology.NewRound(time.Now(), time.Now(), time.Duration(100*time.Millisecond)),
		&mock.ExecBlockMock{},
		WaitTime)

	blk := block.Block{}

	bp.AddBody(0, &blk)
	r := bs.GetDataFromPool(bp.BodyStore(), 0)

	assert.Equal(t, &blk, r)
}
