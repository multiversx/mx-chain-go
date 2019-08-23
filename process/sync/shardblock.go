package sync

import (
	"fmt"
	"math"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ShardBootstrap implements the bootstrap mechanism
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
	bootstrapRoundIndex uint64,
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
		blkc:                blkc,
		blkExecutor:         blkExecutor,
		store:               store,
		headers:             poolsHolder.Headers(),
		headersNonces:       poolsHolder.HeadersNonces(),
		rounder:             rounder,
		waitTime:            waitTime,
		hasher:              hasher,
		marshalizer:         marshalizer,
		forkDetector:        forkDetector,
		shardCoordinator:    shardCoordinator,
		accounts:            accounts,
		bootstrapRoundIndex: bootstrapRoundIndex,
	}

	boot := ShardBootstrap{
		baseBootstrap: base,
		miniBlocks:    poolsHolder.MiniBlocks(),
	}

	base.storageBootstrapper = &boot

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

	boot.chRcvHdrNonce = make(chan bool)
	boot.chRcvHdrHash = make(chan bool)
	boot.chRcvMiniBlocks = make(chan bool)

	boot.setRequestedHeaderNonce(nil)
	boot.setRequestedHeaderHash(nil)
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
) error {
	err := boot.loadBlocks(
		blockFinality,
		blockUnit,
		hdrNonceHashDataUnit)
	if err != nil {
		return err
	}

	return nil
}

func (boot *ShardBootstrap) getHeader(shardId uint32, nonce uint64) (data.HeaderHandler, []byte, error) {
	return boot.getShardHeaderFromStorage(shardId, nonce)
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

func (boot *ShardBootstrap) removeBlockBody(
	nonce uint64,
	blockUnit dataRetriever.UnitType,
	hdrNonceHashDataUnit dataRetriever.UnitType,
) error {

	blockBodyStore := boot.store.GetStorer(dataRetriever.MiniBlockUnit)
	if blockBodyStore == nil {
		return process.ErrNilBlockBodyStorage
	}

	txStore := boot.store.GetStorer(dataRetriever.TransactionUnit)
	if txStore == nil {
		return process.ErrNilTxStorage
	}

	nonceToByteSlice := boot.uint64Converter.ToByteSlice(nonce)
	headerHash, err := boot.store.Get(hdrNonceHashDataUnit, nonceToByteSlice)
	if err != nil {
		return err
	}

	hdrBuff, err := boot.store.Get(blockUnit, headerHash)
	if err != nil {
		return err
	}

	hdr := block.Header{}
	err = boot.marshalizer.Unmarshal(&hdr, hdrBuff)
	if err != nil {
		return err
	}

	miniBlockHashes := make([][]byte, 0)
	for i := 0; i < len(hdr.MiniBlockHeaders); i++ {
		miniBlockHashes = append(miniBlockHashes, hdr.MiniBlockHeaders[i].Hash)
	}

	miniBlocks, err := boot.store.GetAll(dataRetriever.MiniBlockUnit, miniBlockHashes)
	if err != nil {
		return err
	}

	for miniBlockHash, miniBlockBuff := range miniBlocks {
		miniBlock := block.MiniBlock{}
		err = boot.marshalizer.Unmarshal(&miniBlock, miniBlockBuff)
		if err != nil {
			return err
		}

		for _, txHash := range miniBlock.TxHashes {
			err = txStore.Remove(txHash)
			if err != nil {
				return err
			}
		}

		err = blockBodyStore.Remove([]byte(miniBlockHash))
		if err != nil {
			return err
		}
	}

	return nil
}

func (boot *ShardBootstrap) getNonceWithLastNotarized(nonce uint64) (uint64, map[uint32]uint64, map[uint32]uint64) {
	ni := notarizedInfo{}
	ni.reset()
	shardId := sharding.MetachainShardId
	for currentNonce := nonce; currentNonce > 0; currentNonce-- {
		header, ok := boot.isHeaderValid(currentNonce)
		if !ok {
			ni.reset()
			continue
		}

		if ni.startNonce == 0 {
			ni.startNonce = currentNonce
		}

		if len(header.MetaBlockHashes) == 0 {
			continue
		}

		minNonce, err := boot.getMinNotarizedMetaBlockNonceInHeader(header, &ni)
		if err != nil {
			log.Info(err.Error())
			ni.reset()
			continue
		}

		if boot.areNotarizedMetaBlocksFound(&ni, minNonce) {
			break
		}
	}

	if ni.blockWithLastNotarized[shardId]-ni.blockWithFinalNotarized[shardId] > 1 {
		ni.finalNotarized[shardId] = ni.lastNotarized[shardId]
	}

	if ni.blockWithLastNotarized[shardId] != 0 {
		ni.startNonce = ni.blockWithLastNotarized[shardId]
	}

	log.Info(fmt.Sprintf("bootstrap from shard block with nonce %d which contains last notarized meta block\n"+
		"last notarized meta block is %d and final notarized meta block is %d\n",
		ni.startNonce, ni.lastNotarized[shardId], ni.finalNotarized[shardId]))

	return ni.startNonce, ni.finalNotarized, ni.lastNotarized
}

func (boot *ShardBootstrap) isHeaderValid(nonce uint64) (*block.Header, bool) {
	headerHandler, _, err := boot.getHeader(boot.shardCoordinator.SelfId(), nonce)
	if err != nil {
		log.Info(err.Error())
		return nil, false
	}

	header, ok := headerHandler.(*block.Header)
	if !ok {
		log.Info(process.ErrWrongTypeAssertion.Error())
		return nil, false
	}

	if header.Round > boot.bootstrapRoundIndex {
		log.Info(ErrHigherRoundInBlock.Error())
		return nil, false
	}

	return header, true
}

func (boot *ShardBootstrap) getMinNotarizedMetaBlockNonceInHeader(
	header *block.Header,
	ni *notarizedInfo,
) (uint64, error) {

	minNonce := uint64(math.MaxUint64)
	shardId := sharding.MetachainShardId
	for _, metaBlockHash := range header.MetaBlockHashes {
		metaBlock, err := process.GetMetaHeaderFromStorage(metaBlockHash, boot.marshalizer, boot.store)
		if err != nil {
			return minNonce, err
		}

		if ni.blockWithLastNotarized[shardId] == 0 && metaBlock.Nonce == ni.lastNotarized[shardId] {
			ni.blockWithLastNotarized[shardId] = header.Nonce
		}

		if ni.blockWithFinalNotarized[shardId] == 0 && metaBlock.Nonce == ni.finalNotarized[shardId] {
			ni.blockWithFinalNotarized[shardId] = header.Nonce
		}

		if metaBlock.Nonce < minNonce {
			minNonce = metaBlock.Nonce
		}
	}

	return minNonce, nil
}

func (boot *ShardBootstrap) areNotarizedMetaBlocksFound(ni *notarizedInfo, notarizedNonce uint64) bool {
	shardId := sharding.MetachainShardId
	if notarizedNonce == 0 {
		return false
	}

	if ni.lastNotarized[shardId] == 0 {
		ni.lastNotarized[shardId] = notarizedNonce - 1
		return false
	}

	if ni.blockWithLastNotarized[shardId] == 0 {
		return false
	}

	if ni.finalNotarized[shardId] == 0 {
		ni.finalNotarized[shardId] = notarizedNonce - 1
		return false
	}

	if ni.blockWithFinalNotarized[shardId] == 0 {
		return false
	}

	return true
}

func (boot *ShardBootstrap) applyNotarizedBlocks(
	finalNotarized map[uint32]uint64,
	lastNotarized map[uint32]uint64,
) error {
	nonce := finalNotarized[sharding.MetachainShardId]
	if nonce > 0 {
		headerHandler, _, err := boot.getMetaHeaderFromStorage(sharding.MetachainShardId, nonce)
		if err != nil {
			return err
		}

		boot.blkExecutor.AddLastNotarizedHdr(sharding.MetachainShardId, headerHandler)
	}

	nonce = lastNotarized[sharding.MetachainShardId]
	if nonce > 0 {
		headerHandler, _, err := boot.getMetaHeaderFromStorage(sharding.MetachainShardId, nonce)
		if err != nil {
			return err
		}

		boot.blkExecutor.AddLastNotarizedHdr(sharding.MetachainShardId, headerHandler)
	}

	return nil
}

func (boot *ShardBootstrap) cleanupNotarizedStorage(lastNotarized map[uint32]uint64) {
	highestNonceInStorer := boot.computeHighestNonce(dataRetriever.MetaHdrNonceHashDataUnit)

	for i := lastNotarized[sharding.MetachainShardId] + 1; i <= highestNonceInStorer; i++ {
		errNotCritical := boot.removeBlockHeader(i, dataRetriever.MetaBlockUnit, dataRetriever.MetaHdrNonceHashDataUnit)
		if errNotCritical != nil {
			log.Info(fmt.Sprintf("remove notarized block header with nonce %d: %s\n", i, errNotCritical.Error()))
		}
	}
}

func (boot *ShardBootstrap) receivedHeaders(headerHash []byte) {
	header, err := process.GetShardHeaderFromPool(headerHash, boot.headers)
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
	// when a node starts it first tries to bootstrap from storage, if there already exist a database saved
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(boot.shardCoordinator.SelfId())
	errNotCritical := boot.syncFromStorer(process.ShardBlockFinality,
		dataRetriever.BlockHeaderUnit,
		hdrNonceHashDataUnit,
		process.MetaBlockFinality)
	if errNotCritical != nil {
		log.Info(errNotCritical.Error())
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

func (boot *ShardBootstrap) doJobOnSyncBlockFail(hdr *block.Header, err error) {
	if err == process.ErrTimeIsOut {
		boot.requestsWithTimeout++
	}

	shouldRollBack := err != process.ErrTimeIsOut || boot.requestsWithTimeout >= process.MaxRequestsWithTimeoutAllowed
	if shouldRollBack {
		boot.requestsWithTimeout = 0
		hash := boot.removeHeaderFromPools(hdr)
		boot.forkDetector.RemoveHeaders(hdr.Nonce, hash)
		errNotCritical := boot.forkChoice()
		if errNotCritical != nil {
			log.Info(errNotCritical.Error())
		}
	}

	// The below section of code fixed a situation when all peers would have replaced in their headerNonceHash pool a
	// good/used header in their blockchain construction, with a wrong/unused header on which they didn't construct,
	// but which came after a late broadcast from a valid proposer.
	if err == process.ErrBlockHashDoesNotMatch {
		prevHdr, errNotCritical := boot.getHeaderWithHashRequestingIfMissing(hdr.GetPrevHash())
		if errNotCritical != nil {
			log.Info(errNotCritical.Error())
		} else {
			syncMap := &dataPool.ShardIdHashSyncMap{}
			syncMap.Store(prevHdr.GetShardID(), hdr.GetPrevHash())
			boot.headersNonces.Merge(prevHdr.GetNonce(), syncMap)
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
		log.Info(fmt.Sprintf("fork detected at nonce %d with hash %s\n",
			boot.forkNonce,
			core.ToB64(boot.forkHash)))
		err := boot.forkChoice()
		if err != nil {
			log.Info(err.Error())
		}
	}

	boot.setRequestedHeaderNonce(nil)
	boot.setRequestedHeaderHash(nil)
	boot.setRequestedMiniBlocks(nil)

	nonce := boot.getNonceForNextBlock()

	var hdr *block.Header
	var err error

	if boot.isForkDetected {
		hdr, err = boot.getHeaderWithHashRequestingIfMissing(boot.forkHash)
	} else {
		hdr, err = boot.getHeaderWithNonceRequestingIfMissing(nonce)
	}

	if err != nil {
		boot.forkDetector.ResetProbableHighestNonceIfNeeded()
		return err
	}

	defer func() {
		if err != nil {
			boot.doJobOnSyncBlockFail(hdr, err)
		}
	}()

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
		err = process.ErrWrongTypeAssertion
		return err
	}

	blockBody := block.Body(miniBlockSlice)
	timeBefore := time.Now()
	err = boot.blkExecutor.ProcessBlock(boot.blkc, hdr, blockBody, haveTime)
	if err != nil {
		return err
	}
	timeAfter := time.Now()
	log.Info(fmt.Sprintf("time elapsed to process block: %v sec\n", timeAfter.Sub(timeBefore).Seconds()))

	timeBefore = time.Now()
	err = boot.blkExecutor.CommitBlock(boot.blkc, hdr, blockBody)
	if err != nil {
		return err
	}
	timeAfter = time.Now()
	log.Info(fmt.Sprintf("time elapsed to commit block: %v sec\n", timeAfter.Sub(timeBefore).Seconds()))

	log.Info(fmt.Sprintf("block with nonce %d has been synced successfully\n", hdr.Nonce))
	boot.requestsWithTimeout = 0

	return nil
}

// requestHeaderWithNonce method requests a block header from network when it is not found in the pool
func (boot *ShardBootstrap) requestHeaderWithNonce(nonce uint64) {
	boot.setRequestedHeaderNonce(&nonce)
	err := boot.hdrRes.RequestDataFromNonce(nonce)

	log.Info(fmt.Sprintf("requested header with nonce %d from network\n", nonce))

	if err != nil {
		log.Error(err.Error())
	}
}

// requestHeaderWithHash method requests a block header from network when it is not found in the pool
func (boot *ShardBootstrap) requestHeaderWithHash(hash []byte) {
	boot.setRequestedHeaderHash(hash)
	err := boot.hdrRes.RequestDataFromHash(hash)

	log.Info(fmt.Sprintf("requested header with hash %s from network\n", core.ToB64(hash)))

	if err != nil {
		log.Error(err.Error())
	}
}

// getHeaderWithNonceRequestingIfMissing method gets the header with a given nonce from pool. If it is not found there, it will
// be requested from network
func (boot *ShardBootstrap) getHeaderWithNonceRequestingIfMissing(nonce uint64) (*block.Header, error) {
	hdr, _, err := process.GetShardHeaderFromPoolWithNonce(
		nonce,
		boot.shardCoordinator.SelfId(),
		boot.headers,
		boot.headersNonces)
	if err != nil {
		process.EmptyChannel(boot.chRcvHdrNonce)
		boot.requestHeaderWithNonce(nonce)
		err := boot.waitForHeaderNonce()
		if err != nil {
			return nil, err
		}

		hdr, _, err = process.GetShardHeaderFromPoolWithNonce(
			nonce,
			boot.shardCoordinator.SelfId(),
			boot.headers,
			boot.headersNonces)
		if err != nil {
			return nil, err
		}
	}

	return hdr, nil
}

// getHeaderWithHashRequestingIfMissing method gets the header with a given hash from pool. If it is not found there,
// it will be requested from network
func (boot *ShardBootstrap) getHeaderWithHashRequestingIfMissing(hash []byte) (*block.Header, error) {
	hdr, err := process.GetShardHeaderFromPool(hash, boot.headers)
	if err != nil {
		process.EmptyChannel(boot.chRcvHdrHash)
		boot.requestHeaderWithHash(hash)
		err := boot.waitForHeaderHash()
		if err != nil {
			return nil, err
		}

		hdr, err = process.GetShardHeaderFromPool(hash, boot.headers)
		if err != nil {
			return nil, err
		}
	}

	return hdr, nil
}

// requestMiniBlocks method requests a block body from network when it is not found in the pool
func (boot *ShardBootstrap) requestMiniBlocks(hashes [][]byte) {
	_, err := boot.marshalizer.Marshal(hashes)
	if err != nil {
		log.Error("could not marshal MiniBlock hashes: ", err.Error())
		return
	}

	boot.setRequestedMiniBlocks(hashes)
	err = boot.miniBlockResolver.RequestDataFromHashArray(hashes)

	log.Info(fmt.Sprintf("requested %v miniblocks from network\n", len(hashes)))

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
	errNotCritical := boot.blkExecutor.RestoreBlockIntoPools(header, body)
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

// IsInterfaceNil returns true if there is no value under the interface
func (boot *ShardBootstrap) IsInterfaceNil() bool {
	if boot == nil {
		return true
	}
	return false
}
