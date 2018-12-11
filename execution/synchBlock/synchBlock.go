package synchBlock

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/execution"
)

// bootstrap implements the boostrsap mechanism
type bootstrap struct {
	bp          *blockPool.BlockPool
	blkc        *blockchain.BlockChain
	round       *chronology.Round
	blkExecutor execution.BlockExecutor

	requestedHeaderNonce int64
	chRcvHdr             chan bool

	requestedBodyNonce int64
	chRcvBdy           chan bool

	OnRequestHeader func(nonce uint64)
	OnRequestBody   func(nonce uint64)

	waitTime time.Duration
}

// NewBootstrap creates a new bootstrap object
func NewBootstrap(
	bp *blockPool.BlockPool,
	blkc *blockchain.BlockChain,
	round *chronology.Round,
	blkExecutor execution.BlockExecutor,
	waitTime time.Duration,
) *bootstrap {

	bs := bootstrap{
		bp:          bp,
		blkc:        blkc,
		round:       round,
		blkExecutor: blkExecutor,
		waitTime:    waitTime,
	}

	bs.chRcvHdr = make(chan bool)
	bs.chRcvBdy = make(chan bool)

	if bs.bp != nil {
		bs.bp.RegisterHeaderHandler(bs.receivedHeader)
		bs.bp.RegisterBodyHandler(bs.receivedBody)
	}

	return &bs
}

// SynchBlocks method calls repeatedly synchronization method SynchBlock
func (bs *bootstrap) SynchBlocks() {
	for {
		if bs.shouldSynch() {
			err := bs.SynchBlock()
			if err == nil {
				fmt.Println("Body processed successfully")
			}
		}
	}
}

// SynchBlock method actually does the synchronization. It requests the next block header from the pool and if it is not found
// there it will be requested from the network. After the header is received, it requests the block body in the same
// way(pool and than, if it is not found in the pool, from network). If either header and body are received the
// ProcessBlock method will be called. This method will do execute the block and its transactions. Finally if everything
// works, the block will be committed in the blockchain, and all this mechanism will be reiterated for the next block.
func (bs *bootstrap) SynchBlock() error {
	bs.requestedHeaderNonce = -1
	bs.requestedBodyNonce = -1

	nonce := uint64(1) // first block nonce after genesis block
	if bs.blkc != nil && bs.blkc.CurrentBlockHeader != nil {
		nonce = bs.blkc.CurrentBlockHeader.Nonce + 1
	}

	hdr := bs.getHeaderFromPool(nonce)

	if hdr == nil {
		bs.requestHeader(nonce)
		bs.waitForHeaderNonce()
		hdr = bs.getHeaderFromPool(nonce)
		if hdr == nil {
			return execution.ErrMissingHeader
		}
	}

	blk := bs.getBodyFromPool(nonce)

	if blk == nil {
		bs.requestBody(nonce)
		bs.waitForBodyNonce()
		blk = bs.getBodyFromPool(nonce)
		if blk == nil {
			return execution.ErrMissingBody
		}
	}

	err := bs.blkExecutor.ProcessBlock(bs.blkc, hdr, blk)

	return err
}

// shouldSynch method returns the synch state of the node. If it returns true that means that the node should continue
// the synching mechanism, otherwise the node should stop synching because it is already synched or the chronology
// is not running yet
func (bs *bootstrap) shouldSynch() bool {
	if bs.round == nil {
		return false // chronology is not running yet
	}

	if bs.blkc == nil ||
		bs.blkc.CurrentBlockHeader == nil {
		return bs.round.Index() > 0
	}

	return bs.blkc.CurrentBlockHeader.Round+1 < uint32(bs.round.Index())
}

// getHeaderFromPool method returns the block header from a given nonce
func (bs *bootstrap) getHeaderFromPool(nonce uint64) *block.Header {
	if bs.bp == nil {
		return nil
	}

	headerStore := bs.bp.HeaderStore()

	if headerStore == nil {
		return nil
	}

	key := make([]byte, 8)
	binary.PutUvarint(key, nonce)

	val, ok := headerStore.Get(key)

	if !ok {
		return nil
	}

	return val.(*block.Header)
}

// requestHeader method requests a block header from network when it is not found in the pool
func (bs *bootstrap) requestHeader(nonce uint64) {
	if bs.OnRequestHeader != nil {
		bs.requestedHeaderNonce = int64(nonce)
		bs.OnRequestHeader(nonce)
	}
}

// waitForHeaderNonce method wait for header with the requested nonce to be received
func (bs *bootstrap) waitForHeaderNonce() {
	select {
	case <-bs.chRcvHdr:
		return
	case <-time.After(bs.waitTime):
		return
	}
}

// receivedHeader method is a call back function which is called when a new header is added in the block headers pool
func (bs *bootstrap) receivedHeader(nonce uint64) {
	if bs.requestedHeaderNonce == int64(nonce) {
		bs.requestedHeaderNonce = -1
		bs.chRcvHdr <- true
	}
}

// getBodyFromPool method returns the block body from a given nonce
func (bs *bootstrap) getBodyFromPool(nonce uint64) *block.Block {
	if bs.bp == nil {
		return nil
	}

	blockStore := bs.bp.BodyStore()

	if blockStore == nil {
		return nil
	}

	key := make([]byte, 8)
	binary.PutUvarint(key, nonce)

	val, ok := blockStore.Get(key)

	if !ok {
		return nil
	}

	return val.(*block.Block)
}

// requestBody method requests a block body from network when it is not found in the pool
func (bs *bootstrap) requestBody(nonce uint64) {
	if bs.OnRequestBody != nil {
		bs.requestedBodyNonce = int64(nonce)
		bs.OnRequestBody(nonce)
	}
}

// waitForBodyNonce method wait for body with the requested nonce to be received
func (bs *bootstrap) waitForBodyNonce() {
	select {
	case <-bs.chRcvBdy:
		return
	case <-time.After(bs.waitTime):
		return
	}
}

// receivedBody method is a call back function which is called when a new body is added in the block bodies pool
func (bs *bootstrap) receivedBody(nonce uint64) {
	if bs.requestedBodyNonce == int64(nonce) {
		bs.requestedBodyNonce = -1
		bs.chRcvBdy <- true
	}
}
