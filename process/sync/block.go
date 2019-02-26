package sync

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/blockchain"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
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
	rounder       consensus.Rounder
	blkExecutor   process.BlockProcessor
	hasher        hashing.Hasher
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

	isNodeSynchronized    bool
	syncStateListeners    []func(bool)
	mutSyncStateListeners sync.RWMutex
}

// NewBootstrap creates a new Bootstrap object
func NewBootstrap(
	transientDataHolder data.TransientDataHolder,
	blkc *blockchain.BlockChain,
	rounder consensus.Rounder,
	blkExecutor process.BlockProcessor,
	waitTime time.Duration,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	forkDetector process.ForkDetector,
) (*Bootstrap, error) {
	err := checkBootstrapNilParameters(transientDataHolder, blkc, rounder, blkExecutor, hasher, marshalizer, forkDetector)

	if err != nil {
		return nil, err
	}

	boot := Bootstrap{
		headers:       transientDataHolder.Headers(),
		headersNonces: transientDataHolder.HeadersNonces(),
		txBlockBodies: transientDataHolder.TxBlocks(),
		blkc:          blkc,
		rounder:       rounder,
		blkExecutor:   blkExecutor,
		waitTime:      waitTime,
		hasher:        hasher,
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

	boot.syncStateListeners = make([]func(bool), 0)

	return &boot, nil
}

// checkBootstrapNilParameters will check the imput parameters for nil values
func checkBootstrapNilParameters(
	transientDataHolder data.TransientDataHolder,
	blkc *blockchain.BlockChain,
	rounder consensus.Rounder,
	blkExecutor process.BlockProcessor,
	hasher hashing.Hasher,
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

	if rounder == nil {
		return process.ErrNilRounder
	}

	if blkExecutor == nil {
		return process.ErrNilBlockExecutor
	}

	if hasher == nil {
		return process.ErrNilHasher
	}

	if marshalizer == nil {
		return process.ErrNilMarshalizer
	}

	if forkDetector == nil {
		return process.ErrNilForkDetector
	}

	return nil
}

func (boot *Bootstrap) AddSyncStateListener(syncStateListener func(bool)) {
	boot.mutSyncStateListeners.Lock()
	boot.syncStateListeners = append(boot.syncStateListeners, syncStateListener)
	boot.mutSyncStateListeners.Unlock()
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

func (boot *Bootstrap) getHeader(hash []byte) *block.Header {
	hdr := boot.getHeaderFromPool(hash)

	if hdr != nil {
		return hdr
	}

	return boot.getHeaderFromStorage(hash)
}

func (boot *Bootstrap) getHeaderFromPool(hash []byte) *block.Header {
	hdr, ok := boot.headers.SearchFirstData(hash)

	if !ok {
		log.Debug(fmt.Sprintf("header with hash %v not found in headers cache\n", hash))
		return nil
	}

	header, ok := hdr.(*block.Header)

	if !ok {
		log.Debug(fmt.Sprintf("header with hash %v not found in headers cache\n", hash))
		return nil
	}

	return header
}

func (boot *Bootstrap) getHeaderFromStorage(hash []byte) *block.Header {
	headerStore := boot.blkc.GetStorer(blockchain.BlockHeaderUnit)

	if headerStore == nil {
		log.Error(process.ErrNilHeadersStorage.Error())
		return nil
	}

	buffHeader, _ := headerStore.Get(hash)
	header := &block.Header{}
	err := boot.marshalizer.Unmarshal(header, buffHeader)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return header
}

func (boot *Bootstrap) receivedHeaders(headerHash []byte) {
	header := boot.getHeader(headerHash)

	err := boot.forkDetector.AddHeader(header, headerHash, true)

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
		log.Info(fmt.Sprintf("received header with nonce %d from network\n", nonce))
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
		log.Info(fmt.Sprintf("received tx body with hash %s from network\n", toB64(hash)))
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
			err := boot.SyncBlock()

			if err != nil {
				log.Info(err.Error())
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
	isNodeSynchronized := !boot.ShouldSync()

	if isNodeSynchronized != boot.isNodeSynchronized {
		log.Info(fmt.Sprintf("node has changed its synchronized state to %v\n", isNodeSynchronized))
		boot.isNodeSynchronized = isNodeSynchronized
		boot.notifySyncStateListeners()
	}

	if boot.isNodeSynchronized {
		return nil
	}

	boot.setRequestedHeaderNonce(nil)
	boot.setRequestedTxBodyHash(nil)

	nonce := boot.getNonceForNextBlock()

	hdr, err := boot.getHeaderRequestingIfMissing(nonce)
	if err != nil {
		return err
	}

	//TODO remove after all types of block bodies are implemented
	if hdr.BlockBodyType != block.TxBlock {
		return process.ErrNotImplementedBlockProcessingType
	}

	blk, err := boot.getTxBodyRequestingIfMissing(hdr.BlockBodyHash)
	if err != nil {
		return err
	}

	haveTime := func() time.Duration {
		return boot.rounder.TimeDuration()
	}

	//TODO remove type assertions and implement a way for block executor to process
	//TODO all kinds of headers
	err = boot.blkExecutor.ProcessAndCommit(boot.blkc, hdr, blk.(*block.TxBlockBody), haveTime)

	if err != nil {
		if err == process.ErrInvalidBlockHash {
			err = boot.forkChoice(hdr)
		}

		return err
	}

	log.Info(fmt.Sprintf("block with nonce %d was synced successfully\n", hdr.Nonce))

	return nil
}

func (boot *Bootstrap) notifySyncStateListeners() {
	boot.mutSyncStateListeners.RLock()

	for i := 0; i < len(boot.syncStateListeners); i++ {
		go boot.syncStateListeners[i](boot.isNodeSynchronized)
	}

	boot.mutSyncStateListeners.RUnlock()
}

// getHeaderFromPoolHavingNonce method returns the block header from a given nonce
func (boot *Bootstrap) getHeaderFromPoolHavingNonce(nonce uint64) *block.Header {
	hash, _ := boot.headersNonces.Get(nonce)
	if hash == nil {
		log.Debug(fmt.Sprintf("nonce %d not found in headers-nonces cache\n", nonce))
		return nil
	}

	hdr, ok := boot.headers.SearchFirstData(hash)
	if !ok {
		log.Debug(fmt.Sprintf("header with hash %v not found in headers cache\n", hash))
		return nil
	}

	header, ok := hdr.(*block.Header)
	if !ok {
		log.Debug(fmt.Sprintf("header with hash %v not found in headers cache\n", hash))
		return nil
	}

	return header
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
func (boot *Bootstrap) getHeaderRequestingIfMissing(nonce uint64) (*block.Header, error) {
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

// getTxBody method returns the block body from a given hash either from data pool or from storage
func (boot *Bootstrap) getTxBody(hash []byte) interface{} {
	txBody, _ := boot.txBlockBodies.Get(hash)

	if txBody != nil {
		return txBody
	}

	txBodyStorer := boot.blkc.GetStorer(blockchain.TxBlockBodyUnit)

	if txBodyStorer == nil {
		return nil
	}

	buff, err := txBodyStorer.Get(hash)
	if buff == nil {
		log.LogIfError(err)
		return nil
	}

	txBody = &block.TxBlockBody{}

	err = boot.marshalizer.Unmarshal(txBody, buff)
	log.LogIfError(err)
	if err != nil {
		err = txBodyStorer.Remove(hash)
		log.LogIfError(err)
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
func (boot *Bootstrap) getTxBodyRequestingIfMissing(hash []byte) (interface{}, error) {
	blk := boot.getTxBody(hash)

	if blk == nil {
		boot.requestTxBody(hash)
		boot.waitForTxBodyHash()
		blk = boot.getTxBody(hash)
		if blk == nil {
			return nil, process.ErrMissingBody
		}
	}

	intercepted, ok := blk.(*block.TxBlockBody)
	if !ok {
		return nil, ErrTxBlockBodyMismatch
	}

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

// forkChoice decides if rollback must be called
func (boot *Bootstrap) forkChoice(hdr *block.Header) error {
	log.Info(fmt.Sprintf("starting fork choice\n"))

	header := boot.blkc.CurrentBlockHeader

	if header == nil {
		return ErrNilCurrentHeader
	}

	if hdr == nil {
		return ErrNilHeader
	}

	if !isEmpty(header) {
		boot.removeHeaderFromPools(hdr)
		return &ErrNotEmptyHeader{
			CurrentNonce: header.Nonce,
			PoolNonce:    hdr.Nonce}
	}

	log.Info(fmt.Sprintf("roll back to header with hash %s\n",
		toB64(header.PrevHash)))

	return boot.rollback(header)
}

func (boot *Bootstrap) cleanCachesOnRollback(header *block.Header, headerStore storage.Storer) {
	hash, _ := boot.headersNonces.Get(header.Nonce)
	boot.headersNonces.Remove(header.Nonce)
	boot.headers.RemoveData(hash, header.ShardId)
	boot.forkDetector.RemoveHeaders(header.Nonce)
	//TODO uncomment this when badBlocks will be implemented
	//_ = headerStore.Remove(hash)
}

func (boot *Bootstrap) rollback(header *block.Header) error {
	headerStore := boot.blkc.GetStorer(blockchain.BlockHeaderUnit)
	if headerStore == nil {
		return process.ErrNilHeadersStorage
	}

	txBlockBodyStore := boot.blkc.GetStorer(blockchain.TxBlockBodyUnit)
	if txBlockBodyStore == nil {
		return process.ErrNilBlockBodyStorage
	}

	// genesis block is treated differently
	if header.Nonce == 1 {
		boot.blkc.CurrentBlockHeader = nil
		boot.blkc.CurrentTxBlockBody = nil
		boot.blkc.CurrentBlockHeaderHash = nil
		boot.cleanCachesOnRollback(header, headerStore)

		return nil
	}

	newHeader, err := boot.getPrevHeader(headerStore, header)
	if err != nil {
		return err
	}

	newTxBlockBody, err := boot.getTxBlockBody(txBlockBodyStore, newHeader)
	if err != nil {
		return err
	}

	boot.blkc.CurrentBlockHeader = newHeader
	boot.blkc.CurrentTxBlockBody = newTxBlockBody
	boot.blkc.CurrentBlockHeaderHash = header.PrevHash
	boot.cleanCachesOnRollback(header, headerStore)

	return nil
}

func (boot *Bootstrap) removeHeaderFromPools(header *block.Header) {
	hash, _ := boot.headersNonces.Get(header.Nonce)

	boot.headersNonces.Remove(header.Nonce)
	boot.headers.RemoveData(hash, header.ShardId)
	boot.forkDetector.RemoveHeaders(header.Nonce)
}

func (boot *Bootstrap) getPrevHeader(headerStore storage.Storer, header *block.Header) (*block.Header, error) {
	prevHash := header.PrevHash
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
// synched and it can participate to the consensus, if it is in the jobDone group of this rounder
func (boot *Bootstrap) ShouldSync() bool {
	if boot.blkc.CurrentBlockHeader == nil {
		isNotSynchronized := boot.rounder.Index() > 0
		return isNotSynchronized
	}

	isNotSynchronized := boot.blkc.CurrentBlockHeader.Round+1 < uint32(boot.rounder.Index())

	if isNotSynchronized {
		return true
	}

	isForkDetected := boot.forkDetector.CheckFork()

	if isForkDetected {
		log.Info(fmt.Sprintf("fork detected\n"))
		return true
	}

	return false
}

// CreateAndCommitEmptyBlock creates and commits an empty block
func (boot *Bootstrap) CreateAndCommitEmptyBlock(shardForCurrentNode uint32) (*block.TxBlockBody, *block.Header) {
	log.Info(fmt.Sprintf("creating and broadcasting an empty block\n"))

	boot.blkExecutor.RevertAccountState()

	blk := boot.blkExecutor.CreateEmptyBlockBody(
		shardForCurrentNode,
		boot.rounder.Index())

	hdr := &block.Header{}
	hdr.Round = uint32(boot.rounder.Index())
	hdr.TimeStamp = uint64(boot.rounder.TimeStamp().Unix())

	var prevHeaderHash []byte

	if boot.blkc.CurrentBlockHeader == nil {
		hdr.Nonce = 1
		prevHeaderHash = boot.blkc.GenesisHeaderHash
	} else {
		hdr.Nonce = boot.blkc.CurrentBlockHeader.Nonce + 1
		prevHeaderHash = boot.blkc.CurrentBlockHeaderHash
	}

	hdr.PrevHash = prevHeaderHash
	blkStr, err := boot.marshalizer.Marshal(blk)

	if err != nil {
		log.Info(err.Error())
		return nil, nil
	}

	hdr.BlockBodyHash = boot.hasher.Compute(string(blkStr))

	hdr.PubKeysBitmap = make([]byte, 0)

	// TODO: decide the signature for the empty block
	headerStr, err := boot.marshalizer.Marshal(hdr)
	hdrHash := boot.hasher.Compute(string(headerStr))
	hdr.Signature = hdrHash
	hdr.Commitment = hdrHash

	// Commit the block (commits also the account state)
	err = boot.blkExecutor.CommitBlock(boot.blkc, hdr, blk)

	if err != nil {
		log.Info(err.Error())
		return nil, nil
	}

	log.Info(fmt.Sprintf("\n******************** ADDED EMPTY BLOCK WITH NONCE  %d  IN BLOCKCHAIN ********************\n\n", hdr.Nonce))

	return blk, hdr
}
