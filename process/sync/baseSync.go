package sync

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var log = logger.DefaultLogger()

// sleepTime defines the time in milliseconds between each iteration made in syncBlocks method
const sleepTime = time.Duration(5 * time.Millisecond)

// maxRoundsToWait defines the maximum rounds to wait, when bootstrapping, after which the node will add an empty
// block through recovery mechanism, if its block request is not resolved and no new block header is received meantime
const maxRoundsToWait = 5

type baseBootstrap struct {
	headers       storage.Cacher
	headersNonces dataRetriever.Uint64Cacher

	blkc        data.ChainHandler
	blkExecutor process.BlockProcessor
	store       dataRetriever.StorageService

	rounder          consensus.Rounder
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	forkDetector     process.ForkDetector
	shardCoordinator sharding.Coordinator
	accounts         state.AccountsAdapter

	mutHeader   sync.RWMutex
	headerNonce *uint64
	chRcvHdr    chan bool

	requestedHashes process.RequiredDataPool

	chStopSync chan bool
	waitTime   time.Duration

	isNodeSynchronized bool
	hasLastBlock       bool
	roundIndex         int32

	isForkDetected bool
	forkNonce      uint64

	mutRcvHdrInfo         sync.RWMutex
	syncStateListeners    []func(bool)
	mutSyncStateListeners sync.RWMutex
}

// setRequestedHeaderNonce method sets the header nonce requested by the sync mechanism
func (boot *baseBootstrap) setRequestedHeaderNonce(nonce *uint64) {
	boot.mutHeader.Lock()
	boot.headerNonce = nonce
	boot.mutHeader.Unlock()
}

// requestedHeaderNonce method gets the header nonce requested by the sync mechanism
func (boot *baseBootstrap) requestedHeaderNonce() *uint64 {
	boot.mutHeader.RLock()
	defer boot.mutHeader.RUnlock()
	return boot.headerNonce
}

func (boot *baseBootstrap) processReceivedHeader(header data.HeaderHandler, headerHash []byte) {
	if header == nil {
		log.Info(ErrNilHeader.Error())
		return
	}

	log.Debug(fmt.Sprintf("receivedHeaders: received header with nonce %d and hash %s from network\n",
		header.GetNonce(),
		toB64(headerHash)))

	err := boot.forkDetector.AddHeader(header, headerHash, false)
	if err != nil {
		log.Info(err.Error())
	}
}

// receivedHeaderNonce method is a call back function which is called when a new header is added
// in the block headers pool
func (boot *baseBootstrap) receivedHeaderNonce(nonce uint64) {
	headerHash, _ := boot.headersNonces.Get(nonce)

	log.Debug(fmt.Sprintf("receivedHeaderNonce: received header with nonce %d and hash %s from network\n",
		nonce,
		toB64(headerHash)))

	n := boot.requestedHeaderNonce()
	if n == nil {
		return
	}

	if *n == nonce {
		log.Info(fmt.Sprintf("received requested header with nonce %d from network and probable highest nonce is %d\n",
			nonce,
			boot.forkDetector.ProbableHighestNonce()))
		boot.setRequestedHeaderNonce(nil)
		boot.chRcvHdr <- true
	}
}

// AddSyncStateListener adds a syncStateListener that get notified each time the sync status of the node changes
func (boot *baseBootstrap) AddSyncStateListener(syncStateListener func(bool)) {
	boot.mutSyncStateListeners.Lock()
	boot.syncStateListeners = append(boot.syncStateListeners, syncStateListener)
	boot.mutSyncStateListeners.Unlock()
}

func (boot *baseBootstrap) notifySyncStateListeners() {
	boot.mutSyncStateListeners.RLock()
	for i := 0; i < len(boot.syncStateListeners); i++ {
		go boot.syncStateListeners[i](boot.isNodeSynchronized)
	}
	boot.mutSyncStateListeners.RUnlock()
}

// getNonceForNextBlock will get the nonce for the next block we should request
func (boot *baseBootstrap) getNonceForNextBlock() uint64 {
	nonce := uint64(1) // first block nonce after genesis block
	if boot.blkc != nil && boot.blkc.GetCurrentBlockHeader() != nil {
		nonce = boot.blkc.GetCurrentBlockHeader().GetNonce() + 1
	}
	return nonce
}

// waitForHeaderNonce method wait for header with the requested nonce to be received
func (boot *baseBootstrap) waitForHeaderNonce() {
	select {
	case <-boot.chRcvHdr:
		return
	case <-time.After(boot.waitTime):
		return
	}
}

// ShouldSync method returns the synch state of the node. If it returns 'true', this means that the node
// is not synchronized yet and it has to continue the bootstrapping mechanism, otherwise the node is already
// synched and it can participate to the consensus, if it is in the jobDone group of this rounder
func (boot *baseBootstrap) ShouldSync() bool {
	isNodeSynchronizedInCurrentRound := boot.roundIndex == boot.rounder.Index() && boot.isNodeSynchronized
	if isNodeSynchronizedInCurrentRound {
		return false
	}

	boot.isForkDetected, boot.forkNonce = boot.forkDetector.CheckFork()

	if boot.blkc.GetCurrentBlockHeader() == nil {
		boot.hasLastBlock = boot.forkDetector.ProbableHighestNonce() <= 0
	} else {
		boot.hasLastBlock = boot.forkDetector.ProbableHighestNonce() <= boot.blkc.GetCurrentBlockHeader().GetNonce()
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

func (boot *baseBootstrap) removeHeaderFromPools(header data.HeaderHandler) (hash []byte) {
	hash, _ = boot.headersNonces.Get(header.GetNonce())
	boot.headersNonces.Remove(header.GetNonce())
	boot.headers.Remove(hash)
	return
}

func (boot *baseBootstrap) cleanCachesOnRollback(header data.HeaderHandler, headerStore storage.Storer) {
	hash := boot.removeHeaderFromPools(header)
	boot.forkDetector.RemoveHeaders(header.GetNonce(), hash)
	_ = headerStore.Remove(hash)
}

// checkBootstrapNilParameters will check the imput parameters for nil values
func checkBootstrapNilParameters(
	blkc data.ChainHandler,
	rounder consensus.Rounder,
	blkExecutor process.BlockProcessor,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	forkDetector process.ForkDetector,
	resolversFinder dataRetriever.ResolversContainer,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
	store dataRetriever.StorageService,
) error {
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
	if store == nil {
		return process.ErrNilStore
	}

	return nil
}

func emptyChannel(ch chan bool) {
	for len(ch) > 0 {
		<-ch
	}
}

// isSigned verifies if a block is signed
func isSigned(header data.HeaderHandler) bool {
	// TODO: Later, here it should be done a more complex verification (signature for this round matches with the bitmap,
	// and validators which signed here, were in this round consensus group)
	bitmap := header.GetPubKeysBitmap()
	isBitmapEmpty := bytes.Equal(bitmap, make([]byte, len(bitmap)))

	return !isBitmapEmpty
}

func toB64(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}

	return base64.StdEncoding.EncodeToString(buff)
}
