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

// Bootstrap implements the boostrsap mechanism
type Bootstrap struct {
	headers       data.ShardedDataCacherNotifier
	headersNonces data.Uint64Cacher
	txBlockBodies storage.Cacher
	blkc          *blockchain.BlockChain
	round         *chronology.Round
	blkExecutor   process.BlockProcessor
	marshalizer   marshal.Marshalizer
	forkDetector  process.ForkDetector

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

	forkDetected bool
}

// NewBootstrap creates a new Bootstrap object
func NewBootstrap(
	transientDataHolder data.TransientDataHolder,
	blkc *blockchain.BlockChain,
	round *chronology.Round,
	blkExecutor process.BlockProcessor,
	waitTime time.Duration,
	marshalizer marshal.Marshalizer,
	forkDetector process.ForkDetector,
) (*Bootstrap, error) {
	err := checkBootstrapNilParameters(transientDataHolder, blkc, round, blkExecutor, marshalizer, forkDetector)

	if err != nil {
		return nil, err
	}

	boot := Bootstrap{
		headers:       transientDataHolder.Headers(),
		headersNonces: transientDataHolder.HeadersNonces(),
		txBlockBodies: transientDataHolder.TxBlocks(),
		blkc:          blkc,
		round:         round,
		blkExecutor:   blkExecutor,
		waitTime:      waitTime,
		marshalizer:   marshalizer,
		forkDetector:  forkDetector,
	}

	boot.chRcvHdr = make(chan bool)
	boot.chRcvTxBdy = make(chan bool)

	boot.setRequestedHeaderNonce(nil)
	boot.setRequestedTxBodyHash(nil)

	boot.headersNonces.RegisterHandler(boot.receivedHeaderNonce)
	boot.txBlockBodies.RegisterHandler(boot.receivedBodyHash)
	boot.headers.RegisterHandler(boot.receivedHeaders)

	boot.chStopSync = make(chan bool)

	return &boot, nil
}

// checkBootstrapNilParameters will check the imput parameters for nil values
func checkBootstrapNilParameters(
	transientDataHolder data.TransientDataHolder,
	blkc *blockchain.BlockChain,
	round *chronology.Round,
	blkExecutor process.BlockProcessor,
	marshalizer marshal.Marshalizer,
	forkDetector process.ForkDetector,
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

	if marshalizer == nil {
		return process.ErrNilMarshalizer
	}

	if forkDetector == nil {
		return process.ErrNilForkDetector
	}

	return nil
}

// setRequestedHeaderNonce method sets the header nonce requested by the sync mechanism
func (boot *Bootstrap) setRequestedHeaderNonce(nonce *uint64) {
	boot.mutHeader.Lock()
	boot.headerNonce = nonce
	boot.mutHeader.Unlock()
}

// requestedHeaderNonce method gets the header nonce requested by the sync mechanism
func (boot *Bootstrap) requestedHeaderNonce() (nonce *uint64) {
	boot.mutHeader.RLock()
	nonce = boot.headerNonce
	boot.mutHeader.RUnlock()

	return
}

func (boot *Bootstrap) getHeaderFromPoolHavingHash(hash []byte) *block.Header {
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

	headerStore := boot.blkc.GetStorer(blockchain.BlockHeaderUnit)

	if headerStore == nil {
		log.Error(process.ErrNilHeadersStorage.Error())
		return nil
	}

	buffHeader, _ := headerStore.Get(hash)
	header := &block.Header{}
	err := boot.marshalizer.Unmarshal(header, buffHeader)
	if err != nil {
		log.Error(process.ErrNilHeadersStorage.Error())
		return nil
	}

	return header
}

func (boot *Bootstrap) receivedHeaders(key []byte) {
	header := boot.getHeaderFromPoolHavingHash(key)

	log.Info(fmt.Sprintf("received header with nonce %d and hash %s\n", header.Nonce, toB64(key)))

	err := boot.forkDetector.AddHeader(header, key, true)

	if err != nil {
		log.Info(err.Error())
	}
}

// receivedHeaderNonce method is a call back function which is called when a new header is added
// in the block headers pool
func (boot *Bootstrap) receivedHeaderNonce(nonce uint64) {
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
func (boot *Bootstrap) requestedTxBodyHash() []byte {
	boot.mutTxBody.RLock()
	hash := boot.txBodyHash
	boot.mutTxBody.RUnlock()

	return hash
}

// setRequestedTxBodyHash method sets the body hash requested by the sync mechanism
func (boot *Bootstrap) setRequestedTxBodyHash(hash []byte) {
	boot.mutTxBody.Lock()
	boot.txBodyHash = hash
	boot.mutTxBody.Unlock()
}

// receivedBody method is a call back function which is called when a new body is added
// in the block bodies pool
func (boot *Bootstrap) receivedBodyHash(hash []byte) {
	if bytes.Equal(boot.requestedTxBodyHash(), hash) {
		boot.setRequestedTxBodyHash(nil)
		boot.chRcvTxBdy <- true
	}
}

// StartSync method will start SyncBlocks as a go routine
func (boot *Bootstrap) StartSync() {
	go boot.syncBlocks()
}

// StopSync method will stop SyncBlocks
func (boot *Bootstrap) StopSync() {
	boot.chStopSync <- true
}

// syncBlocks method calls repeatedly synchronization method SyncBlock
func (boot *Bootstrap) syncBlocks() {
	for {
		time.Sleep(sleepTime)
		select {
		case <-boot.chStopSync:
			return
		default:
			if boot.ShouldSync() {
				if boot.forkDetected {
					log.Info(fmt.Sprintf("\n\n#################### FORK DETECTED ####################\n\n"))
				}
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
func (boot *Bootstrap) SyncBlock() error {
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
	//TODO all kinds of headers
	err = boot.blkExecutor.ProcessAndCommit(boot.blkc, hdr, blk.(*block.TxBlockBody))

	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Block with nonce %d was synced successfully\n", hdr.Nonce))

	return nil
}

// getHeaderFromPoolHavingNonce method returns the block header from a given nonce
func (boot *Bootstrap) getHeaderFromPoolHavingNonce(nonce uint64) *block.Header {
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
func (boot *Bootstrap) requestHeader(nonce uint64) {
	if boot.RequestHeaderHandler != nil {
		boot.setRequestedHeaderNonce(&nonce)
		boot.RequestHeaderHandler(nonce)
	}
}

// getHeaderWithNonce method gets the header with given nonce from pool, if it exist there,
// and if not it will be requested from network
func (boot *Bootstrap) getHeaderWithNonce(nonce uint64) (*block.Header, error) {
	hdr := boot.getHeaderFromPoolHavingNonce(nonce)

	if hdr == nil {
		boot.requestHeader(nonce)
		boot.waitForHeaderNonce()
		hdr = boot.getHeaderFromPoolHavingNonce(nonce)
		if hdr == nil {
			return nil, process.ErrMissingHeader
		}
	}

	return hdr, nil
}

// getBodyFromPool method returns the block body from a given hash
func (boot *Bootstrap) getTxBody(hash []byte) interface{} {
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
func (boot *Bootstrap) requestTxBody(hash []byte) {
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
func (boot *Bootstrap) getTxBodyWithHash(hash []byte) (interface{}, error) {
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
func (boot *Bootstrap) getNonceForNextBlock() uint64 {
	nonce := uint64(1) // first block nonce after genesis block
	if boot.blkc != nil && boot.blkc.CurrentBlockHeader != nil {
		nonce = boot.blkc.CurrentBlockHeader.Nonce + 1
	}

	return nonce
}

// waitForHeaderNonce method wait for header with the requested nonce to be received
func (boot *Bootstrap) waitForHeaderNonce() {
	select {
	case <-boot.chRcvHdr:
		return
	case <-time.After(boot.waitTime):
		return
	}
}

// waitForBodyNonce method wait for body with the requested nonce to be received
func (boot *Bootstrap) waitForTxBodyHash() {
	select {
	case <-boot.chRcvTxBdy:
		return
	case <-time.After(boot.waitTime):
		return
	}
}

func (boot *Bootstrap) ForkChoice() {
	log.Info(fmt.Sprintf("\n#################### STARTING FORK CHOICE ####################\n\n"))

	header := boot.blkc.CurrentBlockHeader

	if header == nil {
		log.Info(fmt.Sprintf("The current header is nil\n"))
		return
	}

	if !isEmpty(header) {
		log.Info(fmt.Sprintf("The current header is not from an empty block: header.PubKeysBitmap = %v\n",
			header.PubKeysBitmap))
		return
	}

	log.Info(fmt.Sprintf("\n#################### ROLL BACK TO HEADER WITH HASH: %s ####################\n\n",
		toB64(header.PrevHash)))

	boot.rollback(header)
}

func (boot *Bootstrap) rollback(header *block.Header) {
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
	boot.forkDetector.RemoveHeader(header.Nonce)

	boot.blkc.CurrentBlockHeader = newHeader
	boot.blkc.CurrentTxBlockBody = newTxBlockBody
	boot.blkc.CurrentBlockHeaderHash = header.PrevHash
}

func (boot *Bootstrap) getPrevHeader(headerStore storage.Storer, header *block.Header) (*block.Header, error) {
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

func (boot *Bootstrap) getTxBlockBody(txBlockBodyStore storage.Storer,
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

func toB64(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}
	return base64.StdEncoding.EncodeToString(buff)
}

// ShouldSync method returns the synch state of the node. If it returns 'true', this means that the node
// is not synchronized yet and it has to continue the bootstrapping mechanism, otherwise the node is already
// synched and it can participate to the consensus, if it is in the jobDone group of this round
func (boot *Bootstrap) ShouldSync() bool {
	if boot.blkc.CurrentBlockHeader == nil {
		return boot.round.Index() > 0
	}

	if boot.blkc.CurrentBlockHeader.Round+1 < uint32(boot.round.Index()) {
		return true
	}

	boot.forkDetected = boot.forkDetector.CheckFork()

	return boot.forkDetected
}
