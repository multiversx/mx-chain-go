package sync

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var log = logger.NewDefaultLogger()

// sleepTime defines the time in milliseconds between each iteration made in syncBlocks method
const sleepTime = time.Duration(5 * time.Millisecond)

// bootstrap implements the boostrsap mechanism
type bootstrap struct {
	headers       data.ShardedDataCacherNotifier
	headersNonces data.Uint64Cacher
	txBlockBodies storage.Cacher
	blkc          *blockchain.BlockChain
	round         *chronology.Round
	blkExecutor   process.BlockProcessor
	marshalizer   marshal.Marshalizer

	mutHeader   sync.RWMutex
	headerNonce *uint64
	chRcvHdr    chan bool

	mutTxBody  sync.RWMutex
	txBodyHash []byte
	chRcvTxBdy chan bool

	RequestHeaderHandler func(nonce uint64)
	RequestTxBodyHandler func(hash []byte)

	chStopSync chan bool
	waitTime   time.Duration
}

// NewBootstrap creates a new bootstrap object
func NewBootstrap(
	transientDataHolder data.TransientDataHolder,
	blkc *blockchain.BlockChain,
	round *chronology.Round,
	blkExecutor process.BlockProcessor,
	waitTime time.Duration,
	marshalizer marshal.Marshalizer,
) (*bootstrap, error) {
	err := checkBootstrapNilParameters(transientDataHolder, blkc, round, blkExecutor)

	if err != nil {
		return nil, err
	}

	boot := bootstrap{
		headers:       transientDataHolder.Headers(),
		headersNonces: transientDataHolder.HeadersNonces(),
		txBlockBodies: transientDataHolder.TxBlocks(),
		blkc:          blkc,
		round:         round,
		blkExecutor:   blkExecutor,
		waitTime:      waitTime,
		marshalizer:   marshalizer,
	}

	boot.chRcvHdr = make(chan bool)
	boot.chRcvTxBdy = make(chan bool)

	boot.setRequestedHeaderNonce(nil)
	boot.setRequestedTxBodyHash(nil)

	boot.headersNonces.RegisterHandler(boot.receivedHeaderNonce)
	boot.txBlockBodies.RegisterHandler(boot.receivedBodyHash)

	boot.chStopSync = make(chan bool)

	return &boot, nil
}

// checkBootstrapNilParameters will check the imput parameters for nil values
func checkBootstrapNilParameters(
	transientDataHolder data.TransientDataHolder,
	blkc *blockchain.BlockChain,
	round *chronology.Round,
	blkExecutor process.BlockProcessor,
) error {
	if transientDataHolder == nil {
		return process.ErrNilTransientDataHolder
	}

	if transientDataHolder.Headers() == nil {
		return process.ErrNilHeadersDataPool
	}

	if transientDataHolder.HeadersNonces() == nil {
		return process.ErrNilHeadersNoncesDataPool
	}

	if transientDataHolder.TxBlocks() == nil {
		return process.ErrNilTxBlockBody
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

// setRequestedHeaderNonce method sets the header nonce requested by the sync mechanism
func (boot *bootstrap) setRequestedHeaderNonce(nonce *uint64) {
	boot.mutHeader.Lock()
	boot.headerNonce = nonce
	boot.mutHeader.Unlock()
}

// requestedHeaderNonce method gets the header nonce requested by the sync mechanism
func (boot *bootstrap) requestedHeaderNonce() (nonce *uint64) {
	boot.mutHeader.RLock()
	nonce = boot.headerNonce
	boot.mutHeader.RUnlock()

	return
}

// receivedHeaderNonce method is a call back function which is called when a new header is added
// in the block headers pool
func (boot *bootstrap) receivedHeaderNonce(nonce uint64) {
	n := boot.requestedHeaderNonce()

	if n == nil {
		return
	}

	if *n == nonce {
		boot.setRequestedHeaderNonce(nil)
		boot.chRcvHdr <- true
	}
}

// requestedTxBodyHash method gets the body hash requested by the sync mechanism
func (boot *bootstrap) requestedTxBodyHash() []byte {
	boot.mutTxBody.RLock()
	hash := boot.txBodyHash
	boot.mutTxBody.RUnlock()

	return hash
}

// setRequestedTxBodyHash method sets the body hash requested by the sync mechanism
func (boot *bootstrap) setRequestedTxBodyHash(hash []byte) {
	boot.mutTxBody.Lock()
	boot.txBodyHash = hash
	boot.mutTxBody.Unlock()
}

// receivedBody method is a call back function which is called when a new body is added
// in the block bodies pool
func (boot *bootstrap) receivedBodyHash(hash []byte) {
	if bytes.Equal(boot.requestedTxBodyHash(), hash) {
		boot.setRequestedTxBodyHash(nil)
		boot.chRcvTxBdy <- true
	}
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
		time.Sleep(sleepTime)
		select {
		case <-boot.chStopSync:
			return
		default:
			if boot.shouldSync() {
				err := boot.SyncBlock()
				if err != nil {
					if err == process.ErrInvalidBlockHash {
						boot.ForkChoice()
					}
					//log.Error(fmt.Sprintf("%s\n", err.Error()))
				}
			}
		}
	}
}

// SyncBlock method actually does the synchronization. It requests the next block header from the pool
// and if it is not found there it will be requested from the network. After the header is received,
// it requests the block body in the same way(pool and than, if it is not found in the pool, from network).
// If either header and body are received the ProcessAndCommit method will be called. This method will execute
// the block and its transactions. Finally if everything works, the block will be committed in the blockchain,
// and all this mechanism will be reiterated for the next block.
func (boot *bootstrap) SyncBlock() error {
	boot.setRequestedHeaderNonce(nil)
	boot.setRequestedTxBodyHash(nil)

	nonce := boot.getNonceForNextBlock()

	hdr, err := boot.getHeaderWithNonce(nonce)
	if err != nil {
		return err
	}

	//TODO remove after all types of block bodies are implemented
	if hdr.BlockBodyType != block.TxBlock {
		return process.ErrNotImplementedBlockProcessingType
	}

	blk, err := boot.getTxBodyWithHash(hdr.BlockBodyHash)
	if err != nil {
		return err
	}

	//TODO remove type assertions and implement a way for block executor to process
	//TODO all kinds of blocks
	err = boot.blkExecutor.ProcessAndCommit(boot.blkc, hdr, blk.(*block.TxBlockBody))

	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Block with nonce %d was synced successfully\n", hdr.Nonce))

	return nil
}

// getHeaderFromPool method returns the block header from a given nonce
func (boot *bootstrap) getHeaderFromPool(nonce uint64) *block.Header {
	hash, _ := boot.headersNonces.Get(nonce)
	if hash == nil {
		log.Debug(fmt.Sprintf("nonce %d not found in headers-nonces cache\n", nonce))
		return nil
	}

	hdr := boot.headers.SearchData(hash)
	if len(hdr) == 0 {
		log.Debug(fmt.Sprintf("header with hash %v not found in headers cache\n", hash))
		return nil
	}

	for _, v := range hdr {
		//just get the first header that is ok
		header, ok := v.(*block.Header)

		if ok {
			return header
		}
	}

	return nil
}

// requestHeader method requests a block header from network when it is not found in the pool
func (boot *bootstrap) requestHeader(nonce uint64) {
	if boot.RequestHeaderHandler != nil {
		boot.setRequestedHeaderNonce(&nonce)
		boot.RequestHeaderHandler(nonce)
	}
}

// getHeaderWithNonce method gets the header with given nonce from pool, if it exist there,
// and if not it will be requested from network
func (boot *bootstrap) getHeaderWithNonce(nonce uint64) (*block.Header, error) {
	hdr := boot.getHeaderFromPool(nonce)

	if hdr == nil {
		boot.requestHeader(nonce)
		boot.waitForHeaderNonce()
		hdr = boot.getHeaderFromPool(nonce)
		if hdr == nil {
			return nil, process.ErrMissingHeader
		}
	}

	return hdr, nil
}

// getBodyFromPool method returns the block body from a given hash
func (boot *bootstrap) getTxBody(hash []byte) interface{} {
	txBody, _ := boot.txBlockBodies.Get(hash)

	if txBody != nil {
		return txBody
	}

	txBodyStorer := boot.blkc.GetStorer(blockchain.TxBlockBodyUnit)

	if txBodyStorer == nil {
		return nil
	}

	buff, _ := txBodyStorer.Get(hash)

	if buff == nil {
		return nil
	}

	txBody = &block.TxBlockBody{}

	err := boot.marshalizer.Unmarshal(txBody, buff)

	if err != nil {
		_ = txBodyStorer.Remove(hash)
		txBody = nil
	}

	return txBody
}

// requestBody method requests a block body from network when it is not found in the pool
func (boot *bootstrap) requestTxBody(hash []byte) {
	if boot.RequestTxBodyHandler != nil {
		boot.setRequestedTxBodyHash(hash)
		boot.RequestTxBodyHandler(hash)
	}
}

// getTxBodyWithHash method gets the body with given nonce from pool, if it exist there,
// and if not it will be requested from network
// the func returns interface{} as to match the next implementations for block body fetchers
// that will be added. The block executor should decide by parsing the header block body type value
// what kind of block body received.
func (boot *bootstrap) getTxBodyWithHash(hash []byte) (interface{}, error) {
	blk := boot.getTxBody(hash)

	if blk == nil {
		boot.requestTxBody(hash)
		boot.waitForTxBodyHash()
		blk = boot.getTxBody(hash)
		if blk == nil {
			return nil, process.ErrMissingBody
		}
	}

	intercepted, _ := blk.(*block.TxBlockBody)

	return intercepted, nil
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

// waitForHeaderNonce method wait for header with the requested nonce to be received
func (boot *bootstrap) waitForHeaderNonce() {
	select {
	case <-boot.chRcvHdr:
		return
	case <-time.After(boot.waitTime):
		return
	}
}

// waitForBodyNonce method wait for body with the requested nonce to be received
func (boot *bootstrap) waitForTxBodyHash() {
	select {
	case <-boot.chRcvTxBdy:
		return
	case <-time.After(boot.waitTime):
		return
	}
}

func (boot *bootstrap) ForkChoice() {
	header := boot.blkc.CurrentBlockHeader

	log.Info(fmt.Sprintf("#################### FORK CHOICE FOR HEADER PREVHASH: %s ####################",
		base64.StdEncoding.EncodeToString(header.PrevHash)))

	if header == nil {
		return
	}

	if !isEmpty(header) {
		log.Info(fmt.Sprintf("IS EMPTY HEADER : %v", header.PubKeysBitmap))
		return
	}

	boot.rollback(header)
}

func (boot *bootstrap) rollback(header *block.Header) {
	headerStore := boot.blkc.GetStorer(blockchain.BlockHeaderUnit)
	if headerStore == nil {
		log.Info(process.ErrNilHeadersStorage.Error())
		return
	}

	txBlockBodyStore := boot.blkc.GetStorer(blockchain.TxBlockBodyUnit)
	if txBlockBodyStore == nil {
		log.Info(process.ErrNilBlockBodyStorage.Error())
		return
	}

	newHeader, err := boot.getPrevHeader(headerStore, header)
	if err != nil {
		log.Info(err.Error())
		return
	}

	newTxBlockBody, err := boot.getTxBlockBody(txBlockBodyStore, newHeader)
	if err != nil {
		log.Info(err.Error())
		return
	}

	hash, _ := boot.headersNonces.Get(header.Nonce)

	boot.headersNonces.Remove(header.Nonce)
	boot.headers.RemoveData(hash, header.ShardId)
	_ = headerStore.Remove(hash)

	boot.blkc.CurrentBlockHeader = newHeader
	boot.blkc.CurrentTxBlockBody = newTxBlockBody
	boot.blkc.CurrentBlockHeaderHash = header.PrevHash
}

func (boot *bootstrap) getPrevHeader(headerStore storage.Storer, header *block.Header) (*block.Header, error) {
	prevHash := header.PrevHash
	boot.blkc.CurrentBlockHeaderHash = header.PrevHash
	buffHeader, _ := headerStore.Get(prevHash)
	newHeader := &block.Header{}
	err := boot.marshalizer.Unmarshal(newHeader, buffHeader)
	if err != nil {
		return nil, err
	}

	return newHeader, nil
}

func (boot *bootstrap) getTxBlockBody(txBlockBodyStore storage.Storer,
	header *block.Header) (*block.TxBlockBody, error) {

	buffTxBlockBody, _ := txBlockBodyStore.Get(header.BlockBodyHash)
	txBlockBody := &block.TxBlockBody{}
	err := boot.marshalizer.Unmarshal(txBlockBody, buffTxBlockBody)
	if err != nil {
		return nil, err
	}

	return txBlockBody, nil
}

// IsEmpty verifies if a block is empty
func isEmpty(header *block.Header) bool {
	bitmap := header.PubKeysBitmap
	areEqual := bytes.Equal(bitmap, make([]byte, len(bitmap)))
	return areEqual
}
