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
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var log = logger.DefaultLogger()

// sleepTime defines the time in milliseconds between each iteration made in syncBlocks method
const sleepTime = time.Duration(5 * time.Millisecond)

// Bootstrap implements the boostrsap mechanism
type Bootstrap struct {
	headers          storage.Cacher
	headersNonces    data.Uint64Cacher
	miniBlocks       storage.Cacher
	blkc             data.ChainHandler
	rounder          consensus.Rounder
	blkExecutor      process.BlockProcessor
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	forkDetector     process.ForkDetector
	shardCoordinator sharding.Coordinator
	accounts         state.AccountsAdapter

	mutHeader   sync.RWMutex
	headerNonce *uint64
	chRcvHdr    chan bool

	requestedHashes process.RequiredDataPool
	chRcvMiniBlocks chan bool

	chStopSync chan bool
	waitTime   time.Duration

	resolversFinder   dataRetriever.ResolversFinder
	hdrRes            dataRetriever.HeaderResolver
	miniBlockResolver dataRetriever.MiniBlocksResolver

	isNodeSynchronized    bool
	hasLastBlock          bool
	roundIndex            int32
	syncStateListeners    []func(bool)
	mutSyncStateListeners sync.RWMutex

	highestNonceReceived uint64
	isForkDetected       bool

	BroadcastBlock func(data.BodyHandler, data.HeaderHandler) error
}

// NewBootstrap creates a new Bootstrap object
func NewBootstrap(
	poolsHolder data.PoolsHolder,
	blkc data.ChainHandler,
	rounder consensus.Rounder,
	blkExecutor process.BlockProcessor,
	waitTime time.Duration,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	forkDetector process.ForkDetector,
	resolversFinder dataRetriever.ResolversFinder,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
) (*Bootstrap, error) {

	err := checkBootstrapNilParameters(
		poolsHolder,
		blkc,
		rounder,
		blkExecutor,
		hasher,
		marshalizer,
		forkDetector,
		resolversFinder,
		shardCoordinator,
		accounts,
	)

	if err != nil {
		return nil, err
	}

	boot := Bootstrap{
		headers:          poolsHolder.Headers(),
		headersNonces:    poolsHolder.HeadersNonces(),
		miniBlocks:       poolsHolder.MiniBlocks(),
		blkc:             blkc,
		rounder:          rounder,
		blkExecutor:      blkExecutor,
		waitTime:         waitTime,
		hasher:           hasher,
		marshalizer:      marshalizer,
		forkDetector:     forkDetector,
		shardCoordinator: shardCoordinator,
		accounts:         accounts,
	}

	//there is one header topic so it is ok to save it
	hdrResolver, err := resolversFinder.IntraShardResolver(factory.HeadersTopic)
	if err != nil {
		return nil, err
	}

	//sync should request the missing block body on the intrashard topic
	miniBlocksResolver, err := resolversFinder.IntraShardResolver(factory.MiniBlocksTopic)
	if err != nil {
		return nil, err
	}

	//placed in struct fields for performance reasons
	boot.hdrRes = hdrResolver.(dataRetriever.HeaderResolver)
	boot.miniBlockResolver = miniBlocksResolver.(dataRetriever.MiniBlocksResolver)

	boot.chRcvHdr = make(chan bool)
	boot.chRcvMiniBlocks = make(chan bool)

	boot.setRequestedHeaderNonce(nil)
	boot.setRequestedMiniBlocks(nil)

	boot.headersNonces.RegisterHandler(boot.receivedHeaderNonce)
	boot.miniBlocks.RegisterHandler(boot.receivedBodyHash)
	boot.headers.RegisterHandler(boot.receivedHeaders)

	boot.chStopSync = make(chan bool)

	boot.syncStateListeners = make([]func(bool), 0)
	boot.requestedHashes = process.RequiredDataPool{}

	return &boot, nil
}

// checkBootstrapNilParameters will check the imput parameters for nil values
func checkBootstrapNilParameters(
	pools data.PoolsHolder,
	blkc data.ChainHandler,
	rounder consensus.Rounder,
	blkExecutor process.BlockProcessor,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	forkDetector process.ForkDetector,
	resolversFinder dataRetriever.ResolversContainer,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
) error {
	if pools == nil {
		return process.ErrNilPoolsHolder
	}

	if pools.Headers() == nil {
		return process.ErrNilHeadersDataPool
	}

	if pools.HeadersNonces() == nil {
		return process.ErrNilHeadersNoncesDataPool
	}

	if pools.MiniBlocks() == nil {
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

	if resolversFinder == nil {
		return process.ErrNilResolverContainer
	}

	if shardCoordinator == nil {
		return process.ErrNilShardCoordinator
	}

	if accounts == nil {
		return process.ErrNilAccountsAdapter
	}

	return nil
}

// AddSyncStateListener adds a syncStateListener that get notified each time the sync status of the node changes
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
	hdr, ok := boot.headers.Peek(hash)
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
	headerStore := boot.blkc.GetStorer(data.BlockHeaderUnit)

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
	if header != nil && header.Nonce > boot.highestNonceReceived {
		boot.highestNonceReceived = header.Nonce
	}
	if header != nil {
		log.Debug(fmt.Sprintf("receivedHeaders: received header with nonce %d and hash %s from network\n", header.Nonce, toB64(headerHash)))
	}

	err := boot.forkDetector.AddHeader(header, headerHash, false)
	if err != nil {
		log.Info(err.Error())
	}
}

// receivedHeaderNonce method is a call back function which is called when a new header is added
// in the block headers pool
func (boot *Bootstrap) receivedHeaderNonce(nonce uint64) {
	//TODO: make sure that header validation done on interceptors do not add headers with wrong nonces/round numbers
	if nonce > boot.highestNonceReceived {
		boot.highestNonceReceived = nonce
	}

	headerHash, _ := boot.headersNonces.Get(nonce)
	if headerHash != nil {
		log.Debug(fmt.Sprintf("receivedHeaderNonce: received header with nonce %d and hash %s from network\n", nonce, toB64(headerHash)))
	}

	n := boot.requestedHeaderNonce()
	if n == nil {
		return
	}

	if *n == nonce {
		log.Info(fmt.Sprintf("received requested header with nonce %d from network and higher nonce received until now is %d\n", nonce, boot.highestNonceReceived))
		boot.setRequestedHeaderNonce(nil)
		boot.chRcvHdr <- true
	}
}

// setRequestedMiniBlocks method sets the body hash requested by the sync mechanism
func (boot *Bootstrap) setRequestedMiniBlocks(hashes [][]byte) {
	boot.requestedHashes.SetHashes(hashes)
}

// receivedBody method is a call back function which is called when a new body is added
// in the block bodies pool
func (boot *Bootstrap) receivedBodyHash(hash []byte) {
	if len(boot.requestedHashes.ExpectedData()) == 0 {
		return
	}

	boot.requestedHashes.SetReceivedHash(hash)
	if boot.requestedHashes.ReceivedAll() {
		log.Info(fmt.Sprintf("received requested txBlockBody with hash %s from network\n", toB64(hash)))
		boot.setRequestedMiniBlocks(nil)
		boot.chRcvMiniBlocks <- true
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
	if !boot.ShouldSync() {
		return nil
	}

	if boot.isForkDetected {
		log.Info("fork detected\n")
		return boot.forkChoice()
	}

	boot.setRequestedHeaderNonce(nil)
	boot.setRequestedMiniBlocks(nil)

	nonce := boot.getNonceForNextBlock()

	hdr, err := boot.getHeaderRequestingIfMissing(nonce)
	if err != nil {
		if err == process.ErrMissingHeader {
			if boot.shouldCreateEmptyBlock(nonce) {
				log.Info(err.Error())
				err = boot.createAndBroadcastEmptyBlock()
			}
		}

		return err
	}

	//TODO remove after all types of block bodies are implemented
	if hdr.BlockBodyType != block.TxBlock {
		return process.ErrNotImplementedBlockProcessingType
	}

	miniBlocksLen := len(hdr.MiniBlockHeaders)
	miniBlockHashes := make([][]byte, miniBlocksLen)
	for i := 0; i < miniBlocksLen; i++ {
		miniBlockHashes[i] = hdr.MiniBlockHeaders[i].Hash
	}
	blk, err := boot.getMiniBlocksRequestingIfMissing(miniBlockHashes)
	if err != nil {
		return err
	}

	haveTime := func() time.Duration {
		return boot.rounder.TimeDuration()
	}

	//TODO remove type assertions and implement a way for block executor to process
	//TODO all kinds of headers
	blockBody := block.Body(blk.(block.MiniBlockSlice))

	err = boot.blkExecutor.ProcessBlock(boot.blkc, hdr, blockBody, haveTime)

	if err != nil {
		isForkDetected := err == process.ErrInvalidBlockHash || err == process.ErrRootStateMissmatch
		if isForkDetected {
			log.Info(err.Error())
			boot.removeHeaderFromPools(hdr)
			err = boot.forkChoice()
		}

		return err
	}

	err = boot.blkExecutor.CommitBlock(boot.blkc, hdr, blockBody)

	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("block with nonce %d was synced successfully\n", hdr.Nonce))

	return nil
}

func (boot *Bootstrap) shouldCreateEmptyBlock(nonce uint64) bool {
	if boot.isForkDetected {
		return false
	}

	if nonce <= boot.highestNonceReceived {
		return false
	}

	return true
}

func (boot *Bootstrap) createAndBroadcastEmptyBlock() error {
	txBlockBody, header, err := boot.CreateAndCommitEmptyBlock(boot.shardCoordinator.SelfId())

	if err == nil {
		log.Info(fmt.Sprintf("body and header with root hash %s and nonce %d were created and commited through the recovery mechanism\n",
			toB64(header.GetRootHash()),
			header.GetNonce()))

		err = boot.broadcastEmptyBlock(txBlockBody.(block.Body), header.(*block.Header))
	}

	return err
}

func (boot *Bootstrap) broadcastEmptyBlock(txBlockBody block.Body, header *block.Header) error {
	log.Info(fmt.Sprintf("broadcasting an empty block\n"))

	// broadcast block body and header
	err := boot.BroadcastBlock(txBlockBody, header)

	if err != nil {
		return err
	}

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

	hdr, ok := boot.headers.Peek(hash)
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
	boot.setRequestedHeaderNonce(&nonce)
	err := boot.hdrRes.RequestDataFromNonce(nonce)

	log.Info(fmt.Sprintf("requested header with nonce %d from network\n", nonce))

	if err != nil {
		log.Error(err.Error())
	}
}

// getHeaderWithNonce method gets the header with given nonce from pool, if it exist there,
// and if not it will be requested from network
func (boot *Bootstrap) getHeaderRequestingIfMissing(nonce uint64) (*block.Header, error) {
	hdr := boot.getHeaderFromPoolHavingNonce(nonce)

	if hdr == nil {
		boot.emptyChannel(boot.chRcvHdr)
		boot.requestHeader(nonce)
		boot.waitForHeaderNonce()
		hdr = boot.getHeaderFromPoolHavingNonce(nonce)
		if hdr == nil {
			return nil, process.ErrMissingHeader
		}
	}

	return hdr, nil
}

// requestMiniBlocks method requests a block body from network when it is not found in the pool
func (boot *Bootstrap) requestMiniBlocks(hashes [][]byte) {
	buff, err := boot.marshalizer.Marshal(hashes)
	if err != nil {
		log.Error("Could not marshal MiniBlock hashes: ", err.Error())
		return
	}
	boot.setRequestedMiniBlocks(hashes)
	err = boot.miniBlockResolver.RequestDataFromHashArray(hashes)

	log.Info(fmt.Sprintf("requested tx body with hash %s from network\n", toB64(buff)))
	if err != nil {
		log.Error(err.Error())
	}
}

// getMiniBlocksRequestingIfMissing method gets the body with given nonce from pool, if it exist there,
// and if not it will be requested from network
// the func returns interface{} as to match the next implementations for block body fetchers
// that will be added. The block executor should decide by parsing the header block body type value
// what kind of block body received.
func (boot *Bootstrap) getMiniBlocksRequestingIfMissing(hashes [][]byte) (interface{}, error) {
	miniBlocks := boot.miniBlockResolver.GetMiniBlocks(hashes)

	if miniBlocks == nil {
		boot.emptyChannel(boot.chRcvMiniBlocks)
		boot.requestMiniBlocks(hashes)
		boot.waitForMiniBlocks()
		miniBlocks = boot.miniBlockResolver.GetMiniBlocks(hashes)
		if miniBlocks == nil {
			return nil, process.ErrMissingBody
		}
	}

	return miniBlocks, nil
}

// getNonceForNextBlock will get the nonce for the next block we should request
func (boot *Bootstrap) getNonceForNextBlock() uint64 {
	nonce := uint64(1) // first block nonce after genesis block
	if boot.blkc != nil && boot.blkc.GetCurrentBlockHeader() != nil {
		nonce = boot.blkc.GetCurrentBlockHeader().GetNonce() + 1
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

// waitForMiniBlocks method wait for body with the requested nonce to be received
func (boot *Bootstrap) waitForMiniBlocks() {
	select {
	case <-boot.chRcvMiniBlocks:
		return
	case <-time.After(boot.waitTime):
		return
	}
}

// forkChoice decides if rollback must be called
func (boot *Bootstrap) forkChoice() error {
	log.Info(fmt.Sprintf("starting fork choice\n"))

	header, err := boot.getCurrentHeader()
	if err != nil {
		return err
	}

	if !isEmpty(header) {
		return &ErrNotEmptyHeader{
			CurrentNonce: header.Nonce}
	}

	log.Info(fmt.Sprintf("roll back to header with hash %s\n",
		toB64(header.PrevHash)))

	return boot.rollback(header)
}

func (boot *Bootstrap) cleanCachesOnRollback(header *block.Header, headerStore storage.Storer) {
	hash, _ := boot.headersNonces.Get(header.Nonce)
	boot.headersNonces.Remove(header.Nonce)
	boot.headers.Remove(hash)
	_ = boot.forkDetector.RemoveProcessedHeader(header.Nonce)
	_ = headerStore.Remove(hash)
}

func (boot *Bootstrap) rollback(header *block.Header) error {
	headerStore := boot.blkc.GetStorer(data.BlockHeaderUnit)
	if headerStore == nil {
		return process.ErrNilHeadersStorage
	}

	// genesis block is treated differently
	if header.Nonce == 1 {
		err := boot.blkc.SetCurrentBlockHeader(nil)
		err = boot.blkc.SetCurrentBlockBody(nil)
		boot.blkc.SetCurrentBlockHeaderHash(nil)
		boot.cleanCachesOnRollback(header, headerStore)

		return err
	}

	newHeader, err := boot.getPrevHeader(headerStore, header)
	if err != nil {
		return err
	}

	newTxBlockBody, err := boot.getTxBlockBody(newHeader)
	if err != nil {
		return err
	}

	err = boot.blkc.SetCurrentBlockHeader(newHeader)
	err = boot.blkc.SetCurrentBlockBody(newTxBlockBody)
	boot.blkc.SetCurrentBlockHeaderHash(header.PrevHash)
	boot.cleanCachesOnRollback(header, headerStore)

	return err
}

func (boot *Bootstrap) removeHeaderFromPools(header *block.Header) {
	currentHeader, err := boot.getCurrentHeader()
	if err != nil {
		log.Info(err.Error())
		return
	}

	if !isEmpty(currentHeader) {
		hash, _ := boot.headersNonces.Get(header.Nonce)
		boot.headersNonces.Remove(header.Nonce)
		boot.headers.Remove(hash)
	}
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

func (boot *Bootstrap) getTxBlockBody(header *block.Header) (block.Body, error) {
	mbLength := len(header.MiniBlockHeaders)
	hashes := make([][]byte, mbLength)
	for i := 0; i < mbLength; i++ {
		hashes[i] = header.MiniBlockHeaders[i].Hash
	}
	bodyMiniBlocks := boot.miniBlockResolver.GetMiniBlocks(hashes)
	return block.Body(bodyMiniBlocks), nil
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
	isNodeSynchronizedInCurrentRound := boot.roundIndex == boot.rounder.Index() && boot.isNodeSynchronized

	if isNodeSynchronizedInCurrentRound {
		return false
	}

	boot.isForkDetected = boot.forkDetector.CheckFork()

	if boot.blkc.GetCurrentBlockHeader() == nil {
		boot.hasLastBlock = boot.rounder.Index() <= 0
	} else {
		boot.hasLastBlock = boot.blkc.GetCurrentBlockHeader().GetRound()+1 >= uint32(boot.rounder.Index())
	}

	isNodeSynchronized := !boot.isForkDetected && boot.hasLastBlock

	if isNodeSynchronized != boot.isNodeSynchronized {
		log.Info(fmt.Sprintf("node has changed its synchronized state to %v\n", isNodeSynchronized))
		boot.isNodeSynchronized = isNodeSynchronized
		boot.notifySyncStateListeners()
	}

	boot.roundIndex = boot.rounder.Index()

	return !isNodeSynchronized
}

func (boot *Bootstrap) getTimeStampForRound(roundIndex uint32) time.Time {
	currentRoundIndex := boot.rounder.Index()
	currentRoundTimeStamp := boot.rounder.TimeStamp()
	roundDuration := boot.rounder.TimeDuration()

	diff := int32(roundIndex) - currentRoundIndex

	roundTimeStamp := currentRoundTimeStamp.Add(roundDuration * time.Duration(diff))

	return roundTimeStamp
}

func (boot *Bootstrap) createHeader() (data.HeaderHandler, error) {
	hdr := &block.Header{}

	var prevHeaderHash []byte

	if boot.blkc.GetCurrentBlockHeader() == nil {
		hdr.Nonce = 1
		hdr.Round = 0
		prevHeaderHash = boot.blkc.GetGenesisHeaderHash()
		hdr.PrevRandSeed = boot.blkc.GetGenesisHeader().GetSignature()
	} else {
		hdr.Nonce = boot.blkc.GetCurrentBlockHeader().GetNonce() + 1
		hdr.Round = boot.blkc.GetCurrentBlockHeader().GetRound() + 1
		prevHeaderHash = boot.blkc.GetCurrentBlockHeaderHash()
		hdr.PrevRandSeed = boot.blkc.GetCurrentBlockHeader().GetSignature()
	}

	hdr.RandSeed = []byte{0}
	hdr.TimeStamp = uint64(boot.getTimeStampForRound(hdr.Round).Unix())
	hdr.PrevHash = prevHeaderHash
	hdr.RootHash = boot.accounts.RootHash()
	hdr.PubKeysBitmap = make([]byte, 0)
	hdr.MiniBlockHeaders = make([]block.MiniBlockHeader, 0)

	return hdr, nil
}

// CreateAndCommitEmptyBlock creates and commits an empty block
func (boot *Bootstrap) CreateAndCommitEmptyBlock(shardForCurrentNode uint32) (data.BodyHandler, data.HeaderHandler, error) {
	log.Info(fmt.Sprintf("creating and commiting an empty block\n"))

	blk := make(block.Body, 0)

	hdr, err := boot.createHeader()
	if err != nil {
		return nil, nil, err
	}

	// TODO: decide the signature for the empty block
	headerStr, err := boot.marshalizer.Marshal(hdr)
	if err != nil {
		return nil, nil, err
	}
	hdrHash := boot.hasher.Compute(string(headerStr))
	hdr.SetSignature(hdrHash)

	// Commit the block (commits also the account state)
	err = boot.blkExecutor.CommitBlock(boot.blkc, hdr, blk)

	if err != nil {
		return nil, nil, err
	}

	msg := fmt.Sprintf("Added empty block with nonce  %d  in blockchain", hdr.GetNonce())
	log.Info(log.Headline(msg, "", "*"))

	return blk, hdr, nil
}

func (boot *Bootstrap) emptyChannel(ch chan bool) {
	for len(ch) > 0 {
		<-ch
	}
}

func (boot *Bootstrap) getCurrentHeader() (*block.Header, error) {
	blockHeader := boot.blkc.GetCurrentBlockHeader()

	if blockHeader == nil {
		return nil, process.ErrNilBlockHeader
	}

	// TODO: Forkchoice and sync should work with interfaces, so all type assertion should be removed
	header, ok := blockHeader.(*block.Header)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return header, nil
}
