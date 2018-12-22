package sync

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

const bytesInUint64 = 8

// bootstrap implements the boostrsap mechanism
type bootstrap struct {
	blkPool     *blockPool.BlockPool
	blkc        *blockchain.BlockChain
	round       *chronology.Round
	blkExecutor process.BlockProcessor

	mutHeader   sync.RWMutex
	headerNonce int64
	chRcvHdr    chan bool

	mutBody   sync.RWMutex
	bodyNonce int64
	chRcvBdy  chan bool

	OnRequestHeader func(nonce uint64)
	OnRequestBody   func(nonce uint64)

	chStopSync chan bool

	waitTime time.Duration
}

// NewBootstrap creates a new bootstrap object
func NewBootstrap(
	blkPool *blockPool.BlockPool,
	blkc *blockchain.BlockChain,
	round *chronology.Round,
	blkExecutor process.BlockProcessor,
	waitTime time.Duration,
) (*bootstrap, error) {
	err := checkBootstrapNilParameters(blkPool, blkc, round, blkExecutor)

	if err != nil {
		return nil, err
	}

	boot := bootstrap{
		blkPool:     blkPool,
		blkc:        blkc,
		round:       round,
		blkExecutor: blkExecutor,
		waitTime:    waitTime,
	}

	boot.chRcvHdr = make(chan bool)
	boot.chRcvBdy = make(chan bool)

	if boot.blkPool != nil {
		boot.blkPool.RegisterHeaderHandler(boot.receivedHeader)
		boot.blkPool.RegisterBodyHandler(boot.receivedBody)
	}

	boot.setRequestedHeaderNonce(-1)
	boot.setRequestedBodyNonce(-1)

	boot.chStopSync = make(chan bool)

	return &boot, nil
}

// checkBootstrapNilParameters will check the imput parameters for nil values
func checkBootstrapNilParameters(
	blkPool *blockPool.BlockPool,
	blkc *blockchain.BlockChain,
	round *chronology.Round,
	blkExecutor process.BlockProcessor,
) error {
	if blkPool == nil {
		return process.ErrNilBlockPool
	}

	if blkc == nil {
		return process.ErrNilBlockChain
	}

	if round == nil {
		return process.ErrNilRound
	}

	if blkExecutor == nil {
		return process.ErrNilBlockExecutor
	}

	return nil
}

// requestedHeaderNonce method gets the header nonce requested by the sync mechanism
func (boot *bootstrap) requestedHeaderNonce() int64 {
	boot.mutHeader.RLock()
	nonce := boot.headerNonce
	boot.mutHeader.RUnlock()
	return nonce
}

// requestedBodyNonce method gets the body nonce requested by the sync mechanism
func (boot *bootstrap) requestedBodyNonce() int64 {
	boot.mutBody.RLock()
	nonce := boot.bodyNonce
	boot.mutBody.RUnlock()
	return nonce
}

// setRequestedHeaderNonce method sets the header nonce requested by the sync mechanism
func (boot *bootstrap) setRequestedHeaderNonce(nonce int64) {
	boot.mutHeader.Lock()
	boot.headerNonce = nonce
	boot.mutHeader.Unlock()
}

// setRequestedBodyNonce method sets the body nonce requested by the sync mechanism
func (boot *bootstrap) setRequestedBodyNonce(nonce int64) {
	boot.mutBody.Lock()
	boot.bodyNonce = nonce
	boot.mutBody.Unlock()
}

// StartSync method will start SyncBlocks as a go routine
func (boot *bootstrap) StartSync() {
	go boot.syncBlocks()
}

// StopSync method will stop SyncBlocks
func (boot *bootstrap) StopSync() {
	boot.chStopSync <- true
}

// syncBlocks method calls repeatedly synchronization method SyncBlock
func (boot *bootstrap) syncBlocks() {
	for {
		select {
		case <-boot.chStopSync:
			return
		default:
			if boot.shouldSync() {
				err := boot.SyncBlock()
				log.LogIfError(err)
			}
		}
	}
}

// SyncBlock method actually does the synchronization. It requests the next block header from the pool
// and if it is not found there it will be requested from the network. After the header is received,
// it requests the block body in the same way(pool and than, if it is not found in the pool, from network).
// If either header and body are received the ProcessBlock method will be called. This method will do execute
// the block and its transactions. Finally if everything works, the block will be committed in the blockchain,
// and all this mechanism will be reiterated for the next block.
func (boot *bootstrap) SyncBlock() error {
	boot.setRequestedHeaderNonce(-1)
	boot.setRequestedBodyNonce(-1)

	nonce := boot.getNonceForNextBlock()

	hdr, err := boot.getHeaderWithNonce(nonce)

	if err != nil {
		return err
	}

	blk, err := boot.getBodyWithNonce(nonce)

	if err != nil {
		return err
	}

	err = boot.blkExecutor.ProcessBlock(boot.blkc, hdr, blk)

	if err == nil {
		log.Debug("block synced successfully")
	}

	return err
}

// getHeaderWithNonce method gets the header with given nonce from pool, if it exist there,
// and if not it will be requested from network
func (boot *bootstrap) getHeaderWithNonce(nonce uint64) (*block.Header, error) {
	hdr := boot.getDataFromPool(boot.blkPool.HeaderStore(), nonce)

	if hdr == nil {
		boot.requestHeader(nonce)
		boot.waitForHeaderNonce()
		hdr = boot.getDataFromPool(boot.blkPool.HeaderStore(), nonce)
		if hdr == nil {
			return nil, process.ErrMissingHeader
		}
	}

	return hdr.(*block.Header), nil
}

// getBodyWithNonce method gets the body with given nonce from pool, if it exist there,
// and if not it will be requested from network
func (boot *bootstrap) getBodyWithNonce(nonce uint64) (*block.TxBlockBody, error) {
	blk := boot.getDataFromPool(boot.blkPool.BodyStore(), nonce)

	if blk == nil {
		boot.requestBody(nonce)
		boot.waitForBodyNonce()
		blk = boot.getDataFromPool(boot.blkPool.BodyStore(), nonce)
		if blk == nil {
			return nil, process.ErrMissingBody
		}
	}

	return blk.(*block.TxBlockBody), nil
}

// getNonceForNextBlock will get the nonce for the next block we should request
func (boot *bootstrap) getNonceForNextBlock() uint64 {
	nonce := uint64(1) // first block nonce after genesis block
	if boot.blkc != nil && boot.blkc.CurrentBlockHeader != nil {
		nonce = boot.blkc.CurrentBlockHeader.Nonce + 1
	}

	return nonce
}

// shouldSync method returns the sync state of the node. If it returns true that means that the node should
// continue the syncing mechanism, otherwise the node should stop syncing because it is already synced
func (boot *bootstrap) shouldSync() bool {
	if boot.blkc.CurrentBlockHeader == nil {
		return boot.round.Index() > 0
	}

	return boot.blkc.CurrentBlockHeader.Round+1 < uint32(boot.round.Index())
}

// getDataFromPool method returns the block header or block body from a given nonce
func (boot *bootstrap) getDataFromPool(store storage.Cacher, nonce uint64) interface{} {
	if store == nil {
		return nil
	}

	key := make([]byte, bytesInUint64)
	binary.PutUvarint(key, nonce)

	val, ok := store.Get(key)

	if !ok {
		return nil
	}

	return val
}

// requestHeader method requests a block header from network when it is not found in the pool
func (boot *bootstrap) requestHeader(nonce uint64) {
	if boot.OnRequestHeader != nil {
		boot.setRequestedHeaderNonce(int64(nonce))
		boot.OnRequestHeader(nonce)
	}
}

// waitForHeaderNonce method wait for header with the requested nonce to be received
func (boot *bootstrap) waitForHeaderNonce() {
	select {
	case <-boot.chRcvHdr:
		return
	case <-time.After(boot.waitTime):
		return
	}
}

// receivedHeader method is a call back function which is called when a new header is added
// in the block headers pool
func (boot *bootstrap) receivedHeader(nonce uint64) {
	if boot.requestedHeaderNonce() == int64(nonce) {
		boot.setRequestedHeaderNonce(-1)
		boot.chRcvHdr <- true
	}
}

// requestBody method requests a block body from network when it is not found in the pool
func (boot *bootstrap) requestBody(nonce uint64) {
	if boot.OnRequestBody != nil {
		boot.setRequestedBodyNonce(int64(nonce))
		boot.OnRequestBody(nonce)
	}
}

// waitForBodyNonce method wait for body with the requested nonce to be received
func (boot *bootstrap) waitForBodyNonce() {
	select {
	case <-boot.chRcvBdy:
		return
	case <-time.After(boot.waitTime):
		return
	}
}

// receivedBody method is a call back function which is called when a new body is added
// in the block bodies pool
func (boot *bootstrap) receivedBody(nonce uint64) {
	if boot.requestedBodyNonce() == int64(nonce) {
		boot.setRequestedBodyNonce(-1)
		boot.chRcvBdy <- true
	}
}
