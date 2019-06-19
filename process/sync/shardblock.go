package sync

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/factory"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// ShardBootstrap implements the boostrsap mechanism
type ShardBootstrap struct {
	*baseBootstrap

	miniBlocks storage.Cacher

	chRcvMiniBlocks chan bool

	resolversFinder   dataRetriever.ResolversFinder
	hdrRes            dataRetriever.HeaderResolver
	miniBlockResolver dataRetriever.MiniBlocksResolver
}

// NewShardBootstrap creates a new Bootstrap object
func NewShardBootstrap(
	poolsHolder dataRetriever.PoolsHolder,
	store dataRetriever.StorageService,
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
) (*ShardBootstrap, error) {

	if poolsHolder == nil {
		return nil, process.ErrNilPoolsHolder
	}
	if poolsHolder.Headers() == nil {
		return nil, process.ErrNilHeadersDataPool
	}
	if poolsHolder.HeadersNonces() == nil {
		return nil, process.ErrNilHeadersNoncesDataPool
	}
	if poolsHolder.MiniBlocks() == nil {
		return nil, process.ErrNilTxBlockBody
	}

	err := checkBootstrapNilParameters(
		blkc,
		rounder,
		blkExecutor,
		hasher,
		marshalizer,
		forkDetector,
		resolversFinder,
		shardCoordinator,
		accounts,
		store,
	)
	if err != nil {
		return nil, err
	}

	base := &baseBootstrap{
		blkc:             blkc,
		blkExecutor:      blkExecutor,
		store:            store,
		headers:          poolsHolder.Headers(),
		headersNonces:    poolsHolder.HeadersNonces(),
		rounder:          rounder,
		waitTime:         waitTime,
		hasher:           hasher,
		marshalizer:      marshalizer,
		forkDetector:     forkDetector,
		shardCoordinator: shardCoordinator,
		accounts:         accounts,
	}

	boot := ShardBootstrap{
		baseBootstrap: base,
		miniBlocks:    poolsHolder.MiniBlocks(),
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

	//TODO: This should be injected when BlockProcessor will be refactored
	boot.uint64Converter = uint64ByteSlice.NewBigEndianConverter()

	return &boot, nil
}

func (boot *ShardBootstrap) syncFromStorer(
	blockFinality uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
	notarizedBlockFinality uint64,
	notarizedHdrNonceHashDataUnit dataRetriever.UnitType,
) error {
	err := boot.loadBlocks(blockFinality,
		blockUnit,
		hdrNonceHashDataUnit,
		boot.getHeaderFromStorage,
		boot.getBlockBody)
	if err != nil {
		return err
	}

	err = boot.loadNotarizedBlocks(notarizedBlockFinality,
		notarizedHdrNonceHashDataUnit,
		boot.applyNotarizedBlock)
	if err != nil {
		return err
	}

	return nil
}

func (boot *ShardBootstrap) applyNotarizedBlock(nonce uint64, notarizedHdrNonceHashDataUnit dataRetriever.UnitType) error {
	nonceToByteSlice := boot.uint64Converter.ToByteSlice(nonce)
	headerHash, err := boot.store.Get(notarizedHdrNonceHashDataUnit, nonceToByteSlice)
	if err != nil {
		return err
	}

	header, err := process.GetMetaHeaderFromStorage(headerHash, boot.marshalizer, boot.store)
	if err != nil {
		return err
	}

	boot.blkExecutor.SetLastNotarizedHdr(sharding.MetachainShardId, header)
	return nil
}

func (boot *ShardBootstrap) getHeaderFromStorage(nonce uint64) (data.HeaderHandler, []byte, error) {
	nonceToByteSlice := boot.uint64Converter.ToByteSlice(nonce)
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(boot.shardCoordinator.SelfId())
	headerHash, err := boot.store.Get(hdrNonceHashDataUnit, nonceToByteSlice)
	if err != nil {
		return nil, nil, err
	}

	header, err := process.GetShardHeaderFromStorage(headerHash, boot.marshalizer, boot.store)

	return header, headerHash, err
}

func (boot *ShardBootstrap) getBlockBody(headerHandler data.HeaderHandler) (data.BodyHandler, error) {
	header, ok := headerHandler.(*block.Header)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	miniBlockHashes := make([][]byte, 0)
	for i := 0; i < len(header.MiniBlockHeaders); i++ {
		miniBlockHashes = append(miniBlockHashes, header.MiniBlockHeaders[i].Hash)
	}

	miniBlockSlice := boot.miniBlockResolver.GetMiniBlocks(miniBlockHashes)

	return block.Body(miniBlockSlice), nil
}

func (boot *ShardBootstrap) receivedHeaders(headerHash []byte) {
	header, err := process.GetShardHeader(headerHash, boot.headers, boot.marshalizer, boot.store)
	if err != nil {
		log.Debug(err.Error())
		return
	}

	boot.processReceivedHeader(header, headerHash)
}

// setRequestedMiniBlocks method sets the body hash requested by the sync mechanism
func (boot *ShardBootstrap) setRequestedMiniBlocks(hashes [][]byte) {
	boot.requestedHashes.SetHashes(hashes)
}

// receivedBody method is a call back function which is called when a new body is added
// in the block bodies pool
func (boot *ShardBootstrap) receivedBodyHash(hash []byte) {
	if len(boot.requestedHashes.ExpectedData()) == 0 {
		return
	}

	boot.requestedHashes.SetReceivedHash(hash)
	if boot.requestedHashes.ReceivedAll() {
		log.Info(fmt.Sprintf("received all the requested mini blocks from network\n"))
		boot.setRequestedMiniBlocks(nil)
		boot.chRcvMiniBlocks <- true
	}
}

// StartSync method will start SyncBlocks as a go routine
func (boot *ShardBootstrap) StartSync() {
	// when a node starts it first tries to boostrap from storage, if there already exist a database saved
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(boot.shardCoordinator.SelfId())
	err := boot.syncFromStorer(process.ShardBlockFinality,
		dataRetriever.BlockHeaderUnit,
		hdrNonceHashDataUnit,
		process.MetaBlockFinality,
		dataRetriever.MetaHdrNonceHashDataUnit)
	if err != nil {
		log.Info(err.Error())
	}

	go boot.syncBlocks()
}

// StopSync method will stop SyncBlocks
func (boot *ShardBootstrap) StopSync() {
	boot.chStopSync <- true
}

// syncBlocks method calls repeatedly synchronization method SyncBlock
func (boot *ShardBootstrap) syncBlocks() {
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
func (boot *ShardBootstrap) SyncBlock() error {
	if !boot.ShouldSync() {
		return nil
	}

	if boot.isForkDetected {
		log.Info(fmt.Sprintf("fork detected at nonce %d\n", boot.forkNonce))
		return boot.forkChoice()
	}

	boot.setRequestedHeaderNonce(nil)
	boot.setRequestedMiniBlocks(nil)

	nonce := boot.getNonceForNextBlock()

	hdr, err := boot.getHeaderRequestingIfMissing(nonce)
	if err != nil {
		return err
	}

	//TODO remove after all types of block bodies are implemented
	if hdr.BlockBodyType != block.TxBlock {
		return process.ErrNotImplementedBlockProcessingType
	}

	miniBlockHashes := make([][]byte, 0)
	for i := 0; i < len(hdr.MiniBlockHeaders); i++ {
		miniBlockHashes = append(miniBlockHashes, hdr.MiniBlockHeaders[i].Hash)
	}

	blk, err := boot.getMiniBlocksRequestingIfMissing(miniBlockHashes)
	if err != nil {
		return err
	}

	haveTime := func() time.Duration {
		return boot.rounder.TimeDuration()
	}

	miniBlockSlice, ok := blk.(block.MiniBlockSlice)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	blockBody := block.Body(miniBlockSlice)
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

	timeBefore := time.Now()
	err = boot.blkExecutor.CommitBlock(boot.blkc, hdr, blockBody)
	if err != nil {
		return err
	}
	timeAfter := time.Now()
	log.Info(fmt.Sprintf("time elapsed to commit block: %v sec\n", timeAfter.Sub(timeBefore).Seconds()))

	log.Info(fmt.Sprintf("block with nonce %d has been synced successfully\n", hdr.Nonce))
	return nil
}

func (boot *ShardBootstrap) getHeaderWithNonce(nonce uint64) (*block.Header, error) {
	hdr, err := boot.getHeaderFromPoolWithNonce(nonce)
	if err != nil {
		hash, err := boot.getHeaderHashFromStorage(nonce)
		if err != nil {
			return nil, err
		}

		hdr, err = process.GetShardHeaderFromStorage(hash, boot.marshalizer, boot.store)
		if err != nil {
			return nil, err
		}
	}

	return hdr, nil
}

// getHeaderFromPoolWithNonce method returns the block header from a given nonce
func (boot *ShardBootstrap) getHeaderFromPoolWithNonce(nonce uint64) (*block.Header, error) {
	value, _ := boot.headersNonces.Get(nonce)
	hash, ok := value.([]byte)

	if hash == nil || !ok {
		return nil, process.ErrMissingHashForHeaderNonce
	}

	hdr, ok := boot.headers.Peek(hash)
	if !ok {
		return nil, process.ErrMissingHeader
	}

	header, ok := hdr.(*block.Header)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return header, nil
}

// getHeaderHashFromStorage method returns the block header hash from a given nonce
func (boot *ShardBootstrap) getHeaderHashFromStorage(nonce uint64) ([]byte, error) {
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(boot.shardCoordinator.SelfId())
	headerStore := boot.store.GetStorer(hdrNonceHashDataUnit)
	if headerStore == nil {
		return nil, process.ErrNilHeadersStorage
	}

	nonceToByteSlice := boot.uint64Converter.ToByteSlice(nonce)
	headerHash, err := headerStore.Get(nonceToByteSlice)
	if err != nil {
		return nil, process.ErrMissingHashForHeaderNonce
	}

	return headerHash, nil
}

// requestHeader method requests a block header from network when it is not found in the pool
func (boot *ShardBootstrap) requestHeader(nonce uint64) {
	boot.setRequestedHeaderNonce(&nonce)
	err := boot.hdrRes.RequestDataFromNonce(nonce)

	log.Info(fmt.Sprintf("requested header with nonce %d from network\n", nonce))

	if err != nil {
		log.Error(err.Error())
	}
}

// getHeaderWithNonce method gets the header with given nonce from pool, if it exist there,
// and if not it will be requested from network
func (boot *ShardBootstrap) getHeaderRequestingIfMissing(nonce uint64) (*block.Header, error) {
	hdr, err := boot.getHeaderWithNonce(nonce)
	if err != nil {
		process.EmptyChannel(boot.chRcvHdr)
		boot.requestHeader(nonce)
		err := boot.waitForHeaderNonce()
		if err != nil {
			return nil, err
		}

		hdr, err = boot.getHeaderWithNonce(nonce)
		if err != nil {
			return nil, err
		}
	}

	return hdr, nil
}

// requestMiniBlocks method requests a block body from network when it is not found in the pool
func (boot *ShardBootstrap) requestMiniBlocks(hashes [][]byte) {
	buff, err := boot.marshalizer.Marshal(hashes)
	if err != nil {
		log.Error("could not marshal MiniBlock hashes: ", err.Error())
		return
	}

	boot.setRequestedMiniBlocks(hashes)
	err = boot.miniBlockResolver.RequestDataFromHashArray(hashes)

	log.Info(fmt.Sprintf("requested tx body with hash %s from network\n", core.ToB64(buff)))
	if err != nil {
		log.Error(err.Error())
	}
}

// getMiniBlocksRequestingIfMissing method gets the body with given nonce from pool, if it exist there,
// and if not it will be requested from network
// the func returns interface{} as to match the next implementations for block body fetchers
// that will be added. The block executor should decide by parsing the header block body type value
// what kind of block body received.
func (boot *ShardBootstrap) getMiniBlocksRequestingIfMissing(hashes [][]byte) (interface{}, error) {
	miniBlocks := boot.miniBlockResolver.GetMiniBlocks(hashes)
	if miniBlocks == nil {
		process.EmptyChannel(boot.chRcvMiniBlocks)
		boot.requestMiniBlocks(hashes)
		err := boot.waitForMiniBlocks()
		if err != nil {
			return nil, err
		}

		miniBlocks = boot.miniBlockResolver.GetMiniBlocks(hashes)
		if miniBlocks == nil {
			return nil, process.ErrMissingBody
		}
	}

	return miniBlocks, nil
}

// waitForMiniBlocks method wait for body with the requested nonce to be received
func (boot *ShardBootstrap) waitForMiniBlocks() error {
	select {
	case <-boot.chRcvMiniBlocks:
		return nil
	case <-time.After(boot.waitTime):
		return process.ErrTimeIsOut
	}
}

// forkChoice decides if rollback must be called
func (boot *ShardBootstrap) forkChoice() error {
	log.Info("starting fork choice\n")
	isForkResolved := false
	for !isForkResolved {
		header, err := boot.getCurrentHeader()
		if err != nil {
			return err
		}

		msg := fmt.Sprintf("roll back to header with nonce %d and hash %s",
			header.GetNonce()-1, core.ToB64(header.GetPrevHash()))

		isSigned := isSigned(header)
		if isSigned {
			msg = fmt.Sprintf("%s from a signed block, as the highest final block nonce is %d",
				msg,
				boot.forkDetector.GetHighestFinalBlockNonce())
			canRevertBlock := header.GetNonce() > boot.forkDetector.GetHighestFinalBlockNonce()
			if !canRevertBlock {
				return &ErrSignedBlock{CurrentNonce: header.GetNonce()}
			}
		}

		log.Info(msg + "\n")

		err = boot.rollback(header)
		if err != nil {
			return err
		}

		if header.GetNonce() <= boot.forkNonce {
			isForkResolved = true
		}
	}

	log.Info("ending fork choice\n")
	return nil
}

func (boot *ShardBootstrap) rollback(header *block.Header) error {
	if header.GetNonce() == 0 {
		return process.ErrRollbackFromGenesis
	}

	headerStore := boot.store.GetStorer(dataRetriever.BlockHeaderUnit)
	if headerStore == nil {
		return process.ErrNilHeadersStorage
	}

	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(boot.shardCoordinator.SelfId())
	headerNonceHashStore := boot.store.GetStorer(hdrNonceHashDataUnit)
	if headerNonceHashStore == nil {
		return process.ErrNilHeadersNonceHashStorage
	}

	var err error
	var newHeader *block.Header
	var newBody block.Body
	var newHeaderHash []byte
	var newRootHash []byte

	if header.GetNonce() > 1 {
		newHeader, err = boot.getPrevHeader(headerStore, header)
		if err != nil {
			return err
		}

		newBody, err = boot.getTxBlockBody(newHeader)
		if err != nil {
			return err
		}

		newHeaderHash = header.PrevHash
		newRootHash = newHeader.RootHash
	} else { // rollback to genesis block
		newRootHash = boot.blkc.GetGenesisHeader().GetRootHash()
	}

	err = boot.blkc.SetCurrentBlockHeader(newHeader)
	if err != nil {
		return err
	}

	err = boot.blkc.SetCurrentBlockBody(newBody)
	if err != nil {
		return err
	}

	boot.blkc.SetCurrentBlockHeaderHash(newHeaderHash)

	err = boot.accounts.RecreateTrie(newRootHash)
	if err != nil {
		return err
	}

	body, err := boot.getTxBlockBody(header)
	if err != nil {
		return err
	}

	boot.cleanCachesOnRollback(header, headerStore, headerNonceHashStore)
	errNotCritical := boot.blkExecutor.RestoreBlockIntoPools(nil, body)
	if errNotCritical != nil {
		log.Info(errNotCritical.Error())
	}

	return nil
}

func (boot *ShardBootstrap) getPrevHeader(headerStore storage.Storer, header *block.Header) (*block.Header, error) {
	prevHash := header.PrevHash
	buffHeader, err := headerStore.Get(prevHash)
	if err != nil {
		return nil, err
	}

	newHeader := &block.Header{}
	err = boot.marshalizer.Unmarshal(newHeader, buffHeader)
	if err != nil {
		return nil, err
	}

	return newHeader, nil
}

func (boot *ShardBootstrap) getTxBlockBody(header *block.Header) (block.Body, error) {
	mbLength := len(header.MiniBlockHeaders)
	hashes := make([][]byte, mbLength)
	for i := 0; i < mbLength; i++ {
		hashes[i] = header.MiniBlockHeaders[i].Hash
	}
	bodyMiniBlocks := boot.miniBlockResolver.GetMiniBlocks(hashes)

	return block.Body(bodyMiniBlocks), nil
}

func (boot *ShardBootstrap) getCurrentHeader() (*block.Header, error) {
	blockHeader := boot.blkc.GetCurrentBlockHeader()
	if blockHeader == nil {
		return nil, process.ErrNilBlockHeader
	}

	header, ok := blockHeader.(*block.Header)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	return header, nil
}
