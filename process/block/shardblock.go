package block

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/serviceContainer"
	"github.com/ElrondNetwork/elrond-go/core/sliceUtil"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/throttle"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
)

const maxCleanTime = time.Second

// shardProcessor implements shardProcessor interface and actually it tries to execute block
type shardProcessor struct {
	*baseProcessor
	dataPool          dataRetriever.PoolsHolder
	metaBlockFinality uint32
	chRcvAllMetaHdrs  chan bool

	chRcvEpochStart chan bool

	processedMiniBlocks    map[string]map[string]struct{}
	mutProcessedMiniBlocks sync.RWMutex
	core                   serviceContainer.Core
	txCounter              *transactionCounter
	txsPoolsCleaner        process.PoolsCleaner
}

// NewShardProcessor creates a new shardProcessor object
func NewShardProcessor(arguments ArgShardProcessor) (*shardProcessor, error) {
	err := checkProcessorNilParameters(arguments.ArgBaseProcessor)
	if err != nil {
		return nil, err
	}

	if arguments.DataPool == nil || arguments.DataPool.IsInterfaceNil() {
		return nil, process.ErrNilDataPoolHolder
	}

	blockSizeThrottler, err := throttle.NewBlockSizeThrottle()
	if err != nil {
		return nil, err
	}

	base := &baseProcessor{
		accounts:                      arguments.Accounts,
		blockSizeThrottler:            blockSizeThrottler,
		forkDetector:                  arguments.ForkDetector,
		hasher:                        arguments.Hasher,
		marshalizer:                   arguments.Marshalizer,
		store:                         arguments.Store,
		shardCoordinator:              arguments.ShardCoordinator,
		nodesCoordinator:              arguments.NodesCoordinator,
		specialAddressHandler:         arguments.SpecialAddressHandler,
		uint64Converter:               arguments.Uint64Converter,
		onRequestHeaderHandlerByNonce: arguments.RequestHandler.RequestHeaderByNonce,
		appStatusHandler:              statusHandler.NewNilStatusHandler(),
		blockChainHook:                arguments.BlockChainHook,
		txCoordinator:                 arguments.TxCoordinator,
		rounder:                       arguments.Rounder,
		epochStartTrigger:             arguments.EpochStartTrigger,
		headerValidator:               arguments.HeaderValidator,
		bootStorer:                    arguments.BootStorer,
		validatorStatisticsProcessor:  arguments.ValidatorStatisticsProcessor,
	}

	err = base.setLastNotarizedHeadersSlice(arguments.StartHeaders)
	if err != nil {
		return nil, err
	}

	if arguments.TxsPoolsCleaner == nil || arguments.TxsPoolsCleaner.IsInterfaceNil() {
		return nil, process.ErrNilTxsPoolsCleaner
	}

	sp := shardProcessor{
		core:            arguments.Core,
		baseProcessor:   base,
		dataPool:        arguments.DataPool,
		txCounter:       NewTransactionCounter(),
		txsPoolsCleaner: arguments.TxsPoolsCleaner,
	}

	sp.baseProcessor.requestBlockBodyHandler = &sp

	sp.chRcvAllMetaHdrs = make(chan bool)

	transactionPool := sp.dataPool.Transactions()
	if transactionPool == nil {
		return nil, process.ErrNilTransactionPool
	}

	sp.hdrsForCurrBlock.hdrHashAndInfo = make(map[string]*hdrInfo)
	sp.hdrsForCurrBlock.highestHdrNonce = make(map[uint32]uint64)
	sp.processedMiniBlocks = make(map[string]map[string]struct{})

	metaBlockPool := sp.dataPool.MetaBlocks()
	if metaBlockPool == nil {
		return nil, process.ErrNilMetaBlocksPool
	}
	metaBlockPool.RegisterHandler(sp.receivedMetaBlock)
	sp.onRequestHeaderHandler = arguments.RequestHandler.RequestHeader

	sp.metaBlockFinality = process.MetaBlockFinality

	sp.lastHdrs = make(mapShardHeader)

	return &sp, nil
}

// ProcessBlock processes a block. It returns nil if all ok or the specific error
func (sp *shardProcessor) ProcessBlock(
	chainHandler data.ChainHandler,
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
	haveTime func() time.Duration,
) error {

	if haveTime == nil {
		return process.ErrNilHaveTimeHandler
	}

	err := sp.checkBlockValidity(chainHandler, headerHandler, bodyHandler)
	if err != nil {
		if err == process.ErrBlockHashDoesNotMatch {
			log.Debug("requested missing shard header",
				"hash", headerHandler.GetPrevHash(),
				"for shard", headerHandler.GetShardID(),
			)

			go sp.onRequestHeaderHandler(headerHandler.GetShardID(), headerHandler.GetPrevHash())
		}

		return err
	}

	header, ok := headerHandler.(*block.Header)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	go getMetricsFromBlockBody(body, sp.marshalizer, sp.appStatusHandler)

	err = sp.checkHeaderBodyCorrelation(header.MiniBlockHeaders, body)
	if err != nil {
		return err
	}

	numTxWithDst := sp.txCounter.getNumTxsFromPool(header.ShardId, sp.dataPool, sp.shardCoordinator.NumberOfShards())
	go getMetricsFromHeader(header, uint64(numTxWithDst), sp.marshalizer, sp.appStatusHandler)

	log.Debug("total txs in pool",
		"num txs", numTxWithDst,
	)

	err = sp.specialAddressHandler.SetShardConsensusData(
		headerHandler.GetPrevRandSeed(),
		headerHandler.GetRound(),
		headerHandler.GetEpoch(),
		headerHandler.GetShardID(),
	)
	if err != nil {
		return err
	}

	sp.createBlockStarted()
	sp.blockChainHook.SetCurrentHeader(headerHandler)

	sp.txCoordinator.RequestBlockTransactions(body)
	requestedMetaHdrs, requestedFinalityAttestingMetaHdrs := sp.requestMetaHeaders(header)

	if haveTime() < 0 {
		return process.ErrTimeIsOut
	}

	err = sp.txCoordinator.IsDataPreparedForProcessing(haveTime)
	if err != nil {
		return err
	}

	haveMissingMetaHeaders := requestedMetaHdrs > 0 || requestedFinalityAttestingMetaHdrs > 0
	if haveMissingMetaHeaders {
		log.Debug("requested missing meta headers",
			"num headers", requestedMetaHdrs,
		)
		log.Debug("requested missing finality attesting meta headers",
			"num finality shard headers", requestedFinalityAttestingMetaHdrs,
		)

		err = sp.waitForMetaHdrHashes(haveTime())

		sp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
		missingMetaHdrs := sp.hdrsForCurrBlock.missingHdrs
		sp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

		sp.resetMissingHdrs()

		if requestedMetaHdrs > 0 {
			log.Debug("received missing meta headers",
				"num headers", requestedMetaHdrs-missingMetaHdrs,
			)
		}

		if err != nil {
			return err
		}
	}

	err = sp.requestEpochStartInfo(header, haveTime())
	if err != nil {
		return err
	}

	if sp.accounts.JournalLen() != 0 {
		return process.ErrAccountStateDirty
	}

	defer func() {
		go sp.checkAndRequestIfMetaHeadersMissing(header.Round)
	}()

	err = sp.checkEpochCorrectness(header, chainHandler)
	if err != nil {
		return err
	}

	err = sp.checkMetaHeadersValidityAndFinality()
	if err != nil {
		return err
	}

	err = sp.verifyCrossShardMiniBlockDstMe(header)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			sp.RevertAccountState()
		}
	}()

	processedMetaHdrs, err := sp.getOrderedProcessedMetaBlocksFromMiniBlocks(body)
	if err != nil {
		return err
	}

	err = sp.setMetaConsensusData(processedMetaHdrs)
	if err != nil {
		return err
	}

	startTime := time.Now()
	err = sp.txCoordinator.ProcessBlockTransaction(body, haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to process block transaction",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	err = sp.txCoordinator.VerifyCreatedBlockTransactions(body)
	if err != nil {
		return err
	}

	if !sp.verifyStateRoot(header.GetRootHash()) {
		err = process.ErrRootStateDoesNotMatch
		return err
	}

	startTime = time.Now()
	err = sp.checkValidatorStatisticsRootHash(header, processedMetaHdrs)
	elapsedTime = time.Since(startTime)
	log.Debug("elapsed time to check validator statistics root hash",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	return nil
}

func (sp *shardProcessor) requestEpochStartInfo(header *block.Header, waitTime time.Duration) error {
	_ = process.EmptyChannel(sp.chRcvEpochStart)
	haveMissingMetaHeaders := header.IsStartOfEpochBlock() && !sp.epochStartTrigger.IsEpochStart()

	if haveMissingMetaHeaders {
		select {
		case <-sp.chRcvEpochStart:
			return nil
		case <-time.After(waitTime):
			return process.ErrTimeIsOut
		}
	}

	return nil
}

func (sp *shardProcessor) checkEpochCorrectness(
	header *block.Header,
	chainHandler data.ChainHandler,
) error {
	currentBlockHeader := chainHandler.GetCurrentBlockHeader()
	if currentBlockHeader == nil {
		return nil
	}

	isEpochIncorrect := header.GetEpoch() < currentBlockHeader.GetEpoch()
	if isEpochIncorrect {
		return process.ErrEpochDoesNotMatch
	}

	isEpochIncorrect = header.GetEpoch() != currentBlockHeader.GetEpoch() &&
		sp.epochStartTrigger.Epoch() == currentBlockHeader.GetEpoch()
	if isEpochIncorrect {
		return process.ErrEpochDoesNotMatch
	}

	isOldEpochAndShouldBeNew := sp.epochStartTrigger.IsEpochStart() &&
		header.GetRound() > sp.epochStartTrigger.EpochFinalityAttestingRound()+process.EpochChangeGracePeriod &&
		header.GetEpoch() != sp.epochStartTrigger.Epoch()
	if isOldEpochAndShouldBeNew {
		return process.ErrEpochDoesNotMatch
	}

	isEpochStartMetaHashIncorrect := header.IsStartOfEpochBlock() &&
		!bytes.Equal(header.EpochStartMetaHash, sp.epochStartTrigger.EpochStartMetaHdrHash())
	if isEpochStartMetaHashIncorrect {
		go sp.onRequestHeaderHandler(sharding.MetachainShardId, header.EpochStartMetaHash)
		sp.epochStartTrigger.Revert()
		return process.ErrEpochDoesNotMatch
	}

	return nil
}

// SetNumProcessedObj will set the num of processed transactions
func (sp *shardProcessor) SetNumProcessedObj(numObj uint64) {
	sp.txCounter.totalTxs = numObj
}

func (sp *shardProcessor) setMetaConsensusData(finalizedMetaBlocks []data.HeaderHandler) error {
	sp.specialAddressHandler.ClearMetaConsensusData()

	// for every finalized metablock header, reward the metachain consensus group members with accounts in shard
	for _, metaBlock := range finalizedMetaBlocks {
		round := metaBlock.GetRound()
		epoch := metaBlock.GetEpoch()
		err := sp.specialAddressHandler.SetMetaConsensusData(metaBlock.GetPrevRandSeed(), round, epoch)
		if err != nil {
			return err
		}
	}

	return nil
}

// checkMetaHeadersValidity - checks if listed metaheaders are valid as construction
func (sp *shardProcessor) checkMetaHeadersValidityAndFinality() error {
	tmpNotedHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return err
	}

	usedMetaHdrs := sp.sortHeadersForCurrentBlockByNonce(true)
	if len(usedMetaHdrs[sharding.MetachainShardId]) == 0 {
		return nil
	}

	for _, metaHdr := range usedMetaHdrs[sharding.MetachainShardId] {
		err = sp.headerValidator.IsHeaderConstructionValid(metaHdr, tmpNotedHdr)
		if err != nil {
			return err
		}

		tmpNotedHdr = metaHdr
	}

	err = sp.checkMetaHdrFinality(tmpNotedHdr)
	if err != nil {
		return err
	}

	return nil
}

// check if shard headers are final by checking if newer headers were constructed upon them
func (sp *shardProcessor) checkMetaHdrFinality(header data.HeaderHandler) error {
	if header == nil || header.IsInterfaceNil() {
		return process.ErrNilBlockHeader
	}

	finalityAttestingMetaHdrs := sp.sortHeadersForCurrentBlockByNonce(false)

	lastVerifiedHdr := header
	// verify if there are "K" block after current to make this one final
	nextBlocksVerified := uint32(0)
	for _, metaHdr := range finalityAttestingMetaHdrs[sharding.MetachainShardId] {
		if nextBlocksVerified >= sp.metaBlockFinality {
			break
		}

		// found a header with the next nonce
		if metaHdr.GetNonce() == lastVerifiedHdr.GetNonce()+1 {
			err := sp.headerValidator.IsHeaderConstructionValid(metaHdr, lastVerifiedHdr)
			if err != nil {
				go sp.removeHeaderFromPools(metaHdr, sp.dataPool.MetaBlocks(), sp.dataPool.HeadersNonces())
				log.Trace("isHdrConstructionValid", "error", err.Error())
				continue
			}

			lastVerifiedHdr = metaHdr
			nextBlocksVerified += 1
		}
	}

	if nextBlocksVerified < sp.metaBlockFinality {
		go sp.onRequestHeaderHandlerByNonce(lastVerifiedHdr.GetShardID(), lastVerifiedHdr.GetNonce())
		go sp.onRequestHeaderHandlerByNonce(lastVerifiedHdr.GetShardID(), lastVerifiedHdr.GetNonce()+1)
		return process.ErrHeaderNotFinal
	}

	return nil
}

func (sp *shardProcessor) checkAndRequestIfMetaHeadersMissing(round uint64) {
	orderedMetaBlocks, err := sp.getOrderedMetaBlocks(round)
	if err != nil {
		log.Trace("getOrderedMetaBlocks", "error", err.Error())
		return
	}

	sortedHdrs := make([]data.HeaderHandler, 0, len(orderedMetaBlocks))
	for i := 0; i < len(orderedMetaBlocks); i++ {
		hdr, ok := orderedMetaBlocks[i].hdr.(*block.MetaBlock)
		if !ok {
			continue
		}
		sortedHdrs = append(sortedHdrs, hdr)
	}

	err = sp.requestHeadersIfMissing(sortedHdrs, sharding.MetachainShardId, round, sp.dataPool.MetaBlocks())
	if err != nil {
		log.Debug("requestHeadersIfMissing", "error", err.Error())
	}

	lastNotarizedHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		log.Debug("getLastNotarizedHdr", "error", err.Error())
		return
	}

	for i := 0; i < len(sortedHdrs); i++ {
		isMetaBlockOutOfRange := sortedHdrs[i].GetNonce() > lastNotarizedHdr.GetNonce()+process.MaxHeadersToRequestInAdvance
		if isMetaBlockOutOfRange {
			break
		}

		sp.txCoordinator.RequestMiniBlocks(sortedHdrs[i])
	}

	return
}

func (sp *shardProcessor) indexBlockIfNeeded(
	body data.BodyHandler,
	header data.HeaderHandler,
	lastBlockHeader data.HeaderHandler,
) {
	if sp.core == nil || sp.core.Indexer() == nil {
		return
	}

	txPool := sp.txCoordinator.GetAllCurrentUsedTxs(block.TxBlock)
	scPool := sp.txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock)
	rewardPool := sp.txCoordinator.GetAllCurrentUsedTxs(block.RewardsBlock)

	for hash, tx := range scPool {
		txPool[hash] = tx
	}
	for hash, tx := range rewardPool {
		txPool[hash] = tx
	}

	shardId := sp.shardCoordinator.SelfId()
	pubKeys, err := sp.nodesCoordinator.GetValidatorsPublicKeys(header.GetPrevRandSeed(), header.GetRound(), shardId)
	if err != nil {
		return
	}

	signersIndexes := sp.nodesCoordinator.GetValidatorsIndexes(pubKeys)
	go sp.core.Indexer().SaveBlock(body, header, txPool, signersIndexes)

	saveRoundInfoInElastic(sp.core.Indexer(), sp.nodesCoordinator, shardId, header, lastBlockHeader, signersIndexes)
}

// RestoreBlockIntoPools restores the TxBlock and MetaBlock into associated pools
func (sp *shardProcessor) RestoreBlockIntoPools(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error {
	if check.IfNil(headerHandler) {
		return process.ErrNilBlockHeader
	}
	if check.IfNil(bodyHandler) {
		return process.ErrNilTxBlockBody
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	header, ok := headerHandler.(*block.Header)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	miniBlockHashes := header.MapMiniBlockHashesToShards()
	err := sp.restoreMetaBlockIntoPool(miniBlockHashes, header.MetaBlockHashes)
	if err != nil {
		return err
	}

	restoredTxNr, errNotCritical := sp.txCoordinator.RestoreBlockDataFromStorage(body)
	if errNotCritical != nil {
		log.Debug("RestoreBlockDataFromStorage", "error", errNotCritical.Error())
	}

	if header.IsStartOfEpochBlock() {
		sp.epochStartTrigger.Revert()
	}

	go sp.txCounter.subtractRestoredTxs(restoredTxNr)

	sp.removeLastNotarized()

	return nil
}

func (sp *shardProcessor) restoreMetaBlockIntoPool(mapMiniBlockHashes map[string]uint32, metaBlockHashes [][]byte) error {
	metaBlockPool := sp.dataPool.MetaBlocks()
	if metaBlockPool == nil {
		return process.ErrNilMetaBlocksPool
	}

	metaHeaderNoncesPool := sp.dataPool.HeadersNonces()
	if metaHeaderNoncesPool == nil {
		return process.ErrNilMetaHeadersNoncesDataPool
	}

	mapMetaHashMiniBlockHashes := make(map[string][][]byte, len(metaBlockHashes))

	for _, metaBlockHash := range metaBlockHashes {
		metaBlock, errNotCritical := process.GetMetaHeaderFromStorage(metaBlockHash, sp.marshalizer, sp.store)
		if errNotCritical != nil {
			log.Debug("meta block is not fully processed yet and not committed in MetaBlockUnit",
				"hash", metaBlockHash)
			continue
		}

		processedMiniBlocks := metaBlock.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
		for mbHash := range processedMiniBlocks {
			mapMetaHashMiniBlockHashes[string(metaBlockHash)] = append(mapMetaHashMiniBlockHashes[string(metaBlockHash)], []byte(mbHash))
		}

		metaBlockPool.Put(metaBlockHash, metaBlock)
		syncMap := &dataPool.ShardIdHashSyncMap{}
		syncMap.Store(metaBlock.GetShardID(), metaBlockHash)
		metaHeaderNoncesPool.Merge(metaBlock.GetNonce(), syncMap)

		err := sp.store.GetStorer(dataRetriever.MetaBlockUnit).Remove(metaBlockHash)
		if err != nil {
			log.Debug("unable to remove hash from MetaBlockUnit",
				"hash", metaBlockHash)
			return err
		}

		nonceToByteSlice := sp.uint64Converter.ToByteSlice(metaBlock.GetNonce())
		errNotCritical = sp.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit).Remove(nonceToByteSlice)
		if errNotCritical != nil {
			log.Debug("error not critical",
				"error", errNotCritical.Error())
		}

		log.Trace("meta block has been restored successfully",
			"round", metaBlock.Round,
			"nonce", metaBlock.Nonce,
			"hash", metaBlockHash)
	}

	for metaBlockHash, miniBlockHashes := range mapMetaHashMiniBlockHashes {
		for _, miniBlockHash := range miniBlockHashes {
			sp.addProcessedMiniBlock([]byte(metaBlockHash), miniBlockHash)
		}
	}

	for miniBlockHash := range mapMiniBlockHashes {
		sp.removeProcessedMiniBlock([]byte(miniBlockHash))
	}

	return nil
}

// CreateBlockBody creates a a list of miniblocks by filling them with transactions out of the transactions pools
// as long as the transactions limit for the block has not been reached and there is still time to add transactions
func (sp *shardProcessor) CreateBlockBody(initialHdrData data.HeaderHandler, haveTime func() bool) (data.BodyHandler, error) {
	sp.createBlockStarted()
	sp.blockSizeThrottler.ComputeMaxItems()

	initialHdrData.SetEpoch(sp.epochStartTrigger.Epoch())
	sp.blockChainHook.SetCurrentHeader(initialHdrData)

	err := sp.specialAddressHandler.SetShardConsensusData(
		initialHdrData.GetPrevRandSeed(),
		initialHdrData.GetRound(),
		initialHdrData.GetEpoch(),
		initialHdrData.GetShardID(),
	)
	if err != nil {
		return nil, err
	}

	log.Trace("started creating block body",
		"round", initialHdrData.GetRound(),
		"nonce", initialHdrData.GetNonce(),
		"epoch", initialHdrData.GetEpoch(),
	)

	miniBlocks, err := sp.createMiniBlocks(sp.blockSizeThrottler.MaxItemsToAdd(), initialHdrData.GetRound(), haveTime)
	if err != nil {
		return nil, err
	}

	return miniBlocks, nil
}

// CommitBlock commits the block in the blockchain if everything was checked successfully
func (sp *shardProcessor) CommitBlock(
	chainHandler data.ChainHandler,
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {

	var err error
	defer func() {
		if err != nil {
			sp.RevertAccountState()
		}
	}()

	err = checkForNils(chainHandler, headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	log.Trace("started committing block",
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
	)

	err = sp.checkBlockValidity(chainHandler, headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	header, ok := headerHandler.(*block.Header)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	buff, err := sp.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	headerHash := sp.hasher.Compute(string(buff))
	nonceToByteSlice := sp.uint64Converter.ToByteSlice(header.Nonce)
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(header.ShardId)

	errNotCritical := sp.store.Put(hdrNonceHashDataUnit, nonceToByteSlice, headerHash)
	if errNotCritical != nil {
		log.Debug(fmt.Sprintf("ShardHdrNonceHashDataUnit_%d store.Put", header.ShardId),
			"error", errNotCritical.Error(),
		)
	}

	errNotCritical = sp.store.Put(dataRetriever.BlockHeaderUnit, headerHash, buff)
	if errNotCritical != nil {
		log.Trace("BlockHeaderUnit store.Put", "error", errNotCritical.Error())
	}

	headersNoncesPool := sp.dataPool.HeadersNonces()
	if headersNoncesPool == nil {
		err = process.ErrNilHeadersNoncesDataPool
		return err
	}

	headersPool := sp.dataPool.Headers()
	if headersPool == nil {
		err = process.ErrNilHeadersDataPool
		return err
	}

	headersNoncesPool.Remove(header.GetNonce(), header.GetShardID())
	headersPool.Remove(headerHash)

	body, ok := bodyHandler.(block.Body)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	err = sp.txCoordinator.SaveBlockDataToStorage(body)
	if err != nil {
		return err
	}

	for i := 0; i < len(body); i++ {
		buff, err = sp.marshalizer.Marshal(body[i])
		if err != nil {
			return err
		}

		miniBlockHash := sp.hasher.Compute(string(buff))
		errNotCritical = sp.store.Put(dataRetriever.MiniBlockUnit, miniBlockHash, buff)
		if errNotCritical != nil {
			log.Trace("MiniBlockUnit store.Put", "error", errNotCritical.Error())
		}
	}

	processedMetaHdrs, err := sp.getOrderedProcessedMetaBlocksFromHeader(header)
	if err != nil {
		return err
	}

	err = sp.addProcessedCrossMiniBlocksFromHeader(header)
	if err != nil {
		return err
	}

	finalHeaders, finalHeadersHashes, err := sp.getHighestHdrForOwnShardFromMetachain(processedMetaHdrs)
	if err != nil {
		return err
	}

	err = sp.saveLastNotarizedHeader(sharding.MetachainShardId, processedMetaHdrs)
	if err != nil {
		return err
	}

	err = sp.commitAll()
	if err != nil {
		return err
	}

	if header.IsStartOfEpochBlock() {
		err = sp.checkEpochCorrectnessCrossChain(chainHandler)
		sp.epochStartTrigger.SetProcessed(header)
	}

	log.Info("shard block has been committed successfully",
		"nonce", header.Nonce,
		"round", header.Round,
		"epoch", header.Epoch,
		"hash", headerHash,
	)

	errNotCritical = sp.txCoordinator.RemoveBlockDataFromPool(body)
	if errNotCritical != nil {
		log.Debug("RemoveBlockDataFromPool", "error", errNotCritical.Error())
	}

	errNotCritical = sp.removeProcessedMetaBlocksFromPool(processedMetaHdrs)
	if errNotCritical != nil {
		log.Debug("removeProcessedMetaBlocksFromPool", "error", errNotCritical.Error())
	}

	isMetachainStuck := sp.isShardStuck(sharding.MetachainShardId)

	errNotCritical = sp.forkDetector.AddHeader(header, headerHash, process.BHProcessed, finalHeaders, finalHeadersHashes, isMetachainStuck)
	if errNotCritical != nil {
		log.Debug("forkDetector.AddHeader", "error", errNotCritical.Error())
	}

	highestFinalBlockNonce := sp.forkDetector.GetHighestFinalBlockNonce()
	log.Debug("highest final shard block",
		"nonce", highestFinalBlockNonce,
		"shard", sp.shardCoordinator.SelfId(),
	)

	hdrsToAttestPreviousFinal := uint32(header.Nonce-highestFinalBlockNonce) + 1
	sp.removeNotarizedHdrsBehindPreviousFinal(hdrsToAttestPreviousFinal)

	lastBlockHeader := chainHandler.GetCurrentBlockHeader()

	err = chainHandler.SetCurrentBlockBody(body)
	if err != nil {
		return err
	}

	err = chainHandler.SetCurrentBlockHeader(header)
	if err != nil {
		return err
	}

	chainHandler.SetCurrentBlockHeaderHash(headerHash)
	sp.indexBlockIfNeeded(bodyHandler, headerHandler, lastBlockHeader)

	headerMeta, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return err
	}
	saveMetricsForACommittedBlock(
		sp.appStatusHandler,
		sp.specialAddressHandler.IsCurrentNodeInConsensus(),
		display.DisplayByteSlice(headerHash),
		highestFinalBlockNonce,
		headerMeta.GetNonce(),
	)

	headerInfo := bootstrapStorage.BootstrapHeaderInfo{
		ShardId: header.GetShardID(),
		Nonce:   header.GetNonce(),
		Hash:    headerHash,
	}

	sp.mutProcessedMiniBlocks.RLock()
	processedMiniBlocks := process.ConvertProcessedMiniBlocksMapToSlice(sp.processedMiniBlocks)
	sp.mutProcessedMiniBlocks.RUnlock()

	sp.prepareDataForBootStorer(headerInfo, header.Round, finalHeaders, finalHeadersHashes, processedMiniBlocks)

	go sp.cleanTxsPools()

	// write data to log
	go sp.txCounter.displayLogInfo(
		header,
		body,
		headerHash,
		sp.shardCoordinator.NumberOfShards(),
		sp.shardCoordinator.SelfId(),
		sp.dataPool,
		sp.appStatusHandler,
	)

	sp.blockSizeThrottler.Succeed(header.Round)

	log.Debug("pools info",
		"headers", sp.dataPool.Headers().Len(),
		"headers capacity", sp.dataPool.Headers().MaxSize(),
		"metablocks", sp.dataPool.MetaBlocks().Len(),
		"metablocks capacity", sp.dataPool.MetaBlocks().MaxSize(),
		"miniblocks", sp.dataPool.MiniBlocks().Len(),
		"miniblocks capacity", sp.dataPool.MiniBlocks().MaxSize(),
	)

	go sp.cleanupPools(headersNoncesPool, headersPool, sp.dataPool.MetaBlocks())

	return nil
}

func (sp *shardProcessor) checkEpochCorrectnessCrossChain(blockChain data.ChainHandler) error {
	currentHeader := blockChain.GetCurrentBlockHeader()
	if check.IfNil(currentHeader) {
		return nil
	}

	shouldRevertChain := false
	nonce := currentHeader.GetNonce()
	shouldEnterNewEpochRound := sp.epochStartTrigger.EpochFinalityAttestingRound() + process.EpochChangeGracePeriod

	for round := currentHeader.GetRound(); round > shouldEnterNewEpochRound && currentHeader.GetEpoch() != sp.epochStartTrigger.Epoch(); round = currentHeader.GetRound() {
		shouldRevertChain = true
		prevHeader, _, err := process.GetHeaderFromStorageWithNonce(
			currentHeader.GetNonce()-1,
			sp.shardCoordinator.SelfId(),
			sp.store,
			sp.uint64Converter,
			sp.marshalizer,
		)
		if err != nil {
			return err
		}

		nonce = currentHeader.GetNonce()
		currentHeader = prevHeader
	}

	if shouldRevertChain {
		log.Debug("blockchain is wrongly constructed",
			"reverted to nonce", nonce)

		sp.forkDetector.SetForkNonce(nonce)
		return process.ErrEpochDoesNotMatch
	}

	return nil
}

// ApplyProcessedMiniBlocks will apply processed mini blocks
func (sp *shardProcessor) ApplyProcessedMiniBlocks(processedMiniBlocks map[string]map[string]struct{}) {
	sp.mutProcessedMiniBlocks.Lock()
	for metaHash, miniBlocksHashes := range processedMiniBlocks {
		sp.processedMiniBlocks[metaHash] = miniBlocksHashes
	}
	sp.mutProcessedMiniBlocks.Unlock()
}

func (sp *shardProcessor) cleanTxsPools() {
	_, err := sp.txsPoolsCleaner.Clean(maxCleanTime)
	if err != nil {
		log.Debug("txsPoolsCleaner.Clean", "error", err.Error())
	}
	log.Debug("cleaned txs pool",
		"num txs removed", sp.txsPoolsCleaner.NumRemovedTxs(),
	)
}

// CreateNewHeader creates a new header
func (sp *shardProcessor) CreateNewHeader() data.HeaderHandler {
	return &block.Header{}
}

// getHighestHdrForOwnShardFromMetachain calculates the highest shard header notarized by metachain
func (sp *shardProcessor) getHighestHdrForOwnShardFromMetachain(
	processedHdrs []data.HeaderHandler,
) ([]data.HeaderHandler, [][]byte, error) {

	process.SortHeadersByNonce(processedHdrs)

	ownShIdHdrs := make([]data.HeaderHandler, 0, len(processedHdrs))

	for i := 0; i < len(processedHdrs); i++ {
		hdr, ok := processedHdrs[i].(*block.MetaBlock)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}

		hdrs, err := sp.getHighestHdrForShardFromMetachain(sp.shardCoordinator.SelfId(), hdr)
		if err != nil {
			return nil, nil, err
		}

		ownShIdHdrs = append(ownShIdHdrs, hdrs...)
	}

	process.SortHeadersByNonce(ownShIdHdrs)

	ownShIdHdrsHashes := make([][]byte, len(ownShIdHdrs))
	for i := 0; i < len(ownShIdHdrs); i++ {
		hash, _ := core.CalculateHash(sp.marshalizer, sp.hasher, ownShIdHdrs[i])
		ownShIdHdrsHashes[i] = hash
	}

	return ownShIdHdrs, ownShIdHdrsHashes, nil
}

func (sp *shardProcessor) getHighestHdrForShardFromMetachain(shardId uint32, hdr *block.MetaBlock) ([]data.HeaderHandler, error) {
	ownShIdHdr := make([]data.HeaderHandler, 0, len(hdr.ShardInfo))

	var errFound error
	// search for own shard id in shardInfo from metaHeaders
	for _, shardInfo := range hdr.ShardInfo {
		if shardInfo.ShardID != shardId {
			continue
		}

		ownHdr, err := process.GetShardHeader(shardInfo.HeaderHash, sp.dataPool.Headers(), sp.marshalizer, sp.store)
		if err != nil {
			go sp.onRequestHeaderHandler(shardInfo.ShardID, shardInfo.HeaderHash)

			log.Debug("requested missing shard header",
				"hash", shardInfo.HeaderHash,
				"shard", shardInfo.ShardID,
			)

			errFound = err
			continue
		}

		ownShIdHdr = append(ownShIdHdr, ownHdr)
	}

	if errFound != nil {
		return nil, errFound
	}

	return data.TrimHeaderHandlerSlice(ownShIdHdr), nil
}

// getOrderedProcessedMetaBlocksFromHeader returns all the meta blocks fully processed
func (sp *shardProcessor) getOrderedProcessedMetaBlocksFromHeader(header *block.Header) ([]data.HeaderHandler, error) {
	if header == nil {
		return nil, process.ErrNilBlockHeader
	}

	miniBlockHashes := make(map[int][]byte, len(header.MiniBlockHeaders))
	for i := 0; i < len(header.MiniBlockHeaders); i++ {
		miniBlockHashes[i] = header.MiniBlockHeaders[i].Hash
	}

	log.Trace("cross mini blocks in body",
		"num miniblocks", len(miniBlockHashes),
	)

	processedMetaBlocks, err := sp.getOrderedProcessedMetaBlocksFromMiniBlockHashes(miniBlockHashes)
	if err != nil {
		return nil, err
	}

	return processedMetaBlocks, nil
}

func (sp *shardProcessor) addProcessedCrossMiniBlocksFromHeader(header *block.Header) error {
	if header == nil {
		return process.ErrNilBlockHeader
	}

	miniBlockHashes := make(map[int][]byte, len(header.MiniBlockHeaders))
	for i := 0; i < len(header.MiniBlockHeaders); i++ {
		miniBlockHashes[i] = header.MiniBlockHeaders[i].Hash
	}

	sp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for _, metaBlockHash := range header.MetaBlockHashes {
		headerInfo, ok := sp.hdrsForCurrBlock.hdrHashAndInfo[string(metaBlockHash)]
		if !ok {
			sp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrMissingHeader
		}

		metaBlock, ok := headerInfo.hdr.(*block.MetaBlock)
		if !ok {
			sp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrWrongTypeAssertion
		}

		crossMiniBlockHashes := metaBlock.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
		for key, miniBlockHash := range miniBlockHashes {
			_, ok = crossMiniBlockHashes[string(miniBlockHash)]
			if !ok {
				continue
			}

			sp.addProcessedMiniBlock(metaBlockHash, miniBlockHash)

			delete(miniBlockHashes, key)
		}
	}
	sp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	return nil
}

// getOrderedProcessedMetaBlocksFromMiniBlocks returns all the meta blocks fully processed ordered
func (sp *shardProcessor) getOrderedProcessedMetaBlocksFromMiniBlocks(
	usedMiniBlocks []*block.MiniBlock,
) ([]data.HeaderHandler, error) {

	miniBlockHashes := make(map[int][]byte, len(usedMiniBlocks))
	for i := 0; i < len(usedMiniBlocks); i++ {
		if usedMiniBlocks[i].SenderShardID == sp.shardCoordinator.SelfId() {
			continue
		}

		miniBlockHash, err := core.CalculateHash(sp.marshalizer, sp.hasher, usedMiniBlocks[i])
		if err != nil {
			log.Debug("CalculateHash", "error", err.Error())
			continue
		}

		miniBlockHashes[i] = miniBlockHash
	}

	log.Trace("cross mini blocks in body",
		"num miniblocks", len(miniBlockHashes),
	)
	processedMetaBlocks, err := sp.getOrderedProcessedMetaBlocksFromMiniBlockHashes(miniBlockHashes)

	return processedMetaBlocks, err
}

func (sp *shardProcessor) getOrderedProcessedMetaBlocksFromMiniBlockHashes(
	miniBlockHashes map[int][]byte,
) ([]data.HeaderHandler, error) {

	processedMetaHdrs := make([]data.HeaderHandler, 0, len(sp.hdrsForCurrBlock.hdrHashAndInfo))
	processedCrossMiniBlocksHashes := make(map[string]bool, len(sp.hdrsForCurrBlock.hdrHashAndInfo))

	sp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for metaBlockHash, headerInfo := range sp.hdrsForCurrBlock.hdrHashAndInfo {
		if !headerInfo.usedInBlock {
			continue
		}

		metaBlock, ok := headerInfo.hdr.(*block.MetaBlock)
		if !ok {
			sp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return nil, process.ErrWrongTypeAssertion
		}

		log.Trace("meta header",
			"nonce", metaBlock.Nonce,
		)

		crossMiniBlockHashes := metaBlock.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
		for hash := range crossMiniBlockHashes {
			processedCrossMiniBlocksHashes[hash] = sp.isMiniBlockProcessed([]byte(metaBlockHash), []byte(hash))
		}

		for key, miniBlockHash := range miniBlockHashes {
			_, ok = crossMiniBlockHashes[string(miniBlockHash)]
			if !ok {
				continue
			}

			processedCrossMiniBlocksHashes[string(miniBlockHash)] = true

			delete(miniBlockHashes, key)
		}

		log.Trace("cross mini blocks in meta header",
			"num miniblocks", len(crossMiniBlockHashes),
		)

		processedAll := true
		for hash := range crossMiniBlockHashes {
			if !processedCrossMiniBlocksHashes[hash] {
				processedAll = false
				break
			}
		}

		if processedAll {
			processedMetaHdrs = append(processedMetaHdrs, metaBlock)
		}
	}
	sp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	process.SortHeadersByNonce(processedMetaHdrs)

	return processedMetaHdrs, nil
}

func (sp *shardProcessor) removeProcessedMetaBlocksFromPool(processedMetaHdrs []data.HeaderHandler) error {
	lastNotarizedMetaHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return err
	}

	processed := 0
	// processedMetaHdrs is also sorted
	for i := 0; i < len(processedMetaHdrs); i++ {
		hdr := processedMetaHdrs[i]

		// remove process finished
		if hdr.GetNonce() > lastNotarizedMetaHdr.GetNonce() {
			continue
		}

		// metablock was processed and finalized
		buff, err := sp.marshalizer.Marshal(hdr)
		if err != nil {
			log.Debug("marshalizer.Marshal", "error", err.Error())
			continue
		}

		headerHash := sp.hasher.Compute(string(buff))
		nonceToByteSlice := sp.uint64Converter.ToByteSlice(hdr.GetNonce())
		err = sp.store.Put(dataRetriever.MetaHdrNonceHashDataUnit, nonceToByteSlice, headerHash)
		if err != nil {
			log.Debug("MetaHdrNonceHashDataUnit store.Put", "error", err.Error())
			continue
		}

		err = sp.store.Put(dataRetriever.MetaBlockUnit, headerHash, buff)
		if err != nil {
			log.Debug("MetaBlockUnit store.Put", "error", err.Error())
			continue
		}

		sp.dataPool.MetaBlocks().Remove(headerHash)
		sp.dataPool.HeadersNonces().Remove(hdr.GetNonce(), sharding.MetachainShardId)
		sp.removeAllProcessedMiniBlocks(headerHash)

		log.Trace("metaBlock has been processed completely and removed from pool",
			"round", hdr.GetRound(),
			"nonce", hdr.GetNonce(),
			"hash", headerHash,
		)

		processed++
	}

	if processed > 0 {
		log.Trace("metablocks completely processed and removed from pool",
			"num metablocks", processed,
		)
	}

	return nil
}

// receivedMetaBlock is a callback function when a new metablock was received
// upon receiving, it parses the new metablock and requests miniblocks and transactions
// which destination is the current shard
func (sp *shardProcessor) receivedMetaBlock(metaBlockHash []byte) {
	metaBlocksPool := sp.dataPool.MetaBlocks()
	if metaBlocksPool == nil {
		return
	}

	obj, ok := metaBlocksPool.Peek(metaBlockHash)
	if !ok {
		return
	}

	metaBlock, ok := obj.(*block.MetaBlock)
	if !ok {
		return
	}

	log.Trace("received meta block from network",
		"round", metaBlock.Round,
		"nonce", metaBlock.Nonce,
		"hash", metaBlockHash,
	)

	sp.hdrsForCurrBlock.mutHdrsForBlock.Lock()

	haveMissingMetaHeaders := sp.hdrsForCurrBlock.missingHdrs > 0 || sp.hdrsForCurrBlock.missingFinalityAttestingHdrs > 0
	if haveMissingMetaHeaders {
		hdrInfoForHash := sp.hdrsForCurrBlock.hdrHashAndInfo[string(metaBlockHash)]
		receivedMissingMetaHeader := hdrInfoForHash != nil && (hdrInfoForHash.hdr == nil || hdrInfoForHash.hdr.IsInterfaceNil())
		if receivedMissingMetaHeader {
			hdrInfoForHash.hdr = metaBlock
			sp.hdrsForCurrBlock.missingHdrs--

			if metaBlock.Nonce > sp.hdrsForCurrBlock.highestHdrNonce[sharding.MetachainShardId] {
				sp.hdrsForCurrBlock.highestHdrNonce[sharding.MetachainShardId] = metaBlock.Nonce
			}
		}

		// attesting something
		if sp.hdrsForCurrBlock.missingHdrs == 0 {
			sp.hdrsForCurrBlock.missingFinalityAttestingHdrs = sp.requestMissingFinalityAttestingHeaders(
				sharding.MetachainShardId,
				sp.metaBlockFinality,
				sp.getMetaHeaderFromPoolWithNonce,
				sp.dataPool.MetaBlocks())
			if sp.hdrsForCurrBlock.missingFinalityAttestingHdrs == 0 {
				log.Debug("received all missing finality attesting meta headers")
			}
		}

		missingMetaHdrs := sp.hdrsForCurrBlock.missingHdrs
		missingFinalityAttestingMetaHdrs := sp.hdrsForCurrBlock.missingFinalityAttestingHdrs
		sp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

		allMissingMetaHeadersReceived := missingMetaHdrs == 0 && missingFinalityAttestingMetaHdrs == 0
		if allMissingMetaHeadersReceived {
			sp.chRcvAllMetaHdrs <- true
		}
	} else {
		sp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()
	}

	sp.setLastHdrForShard(metaBlock.GetShardID(), metaBlock)

	if sp.isHeaderOutOfRange(metaBlock, metaBlocksPool) {
		metaBlocksPool.Remove(metaBlockHash)

		headersNoncesPool := sp.dataPool.HeadersNonces()
		if headersNoncesPool != nil {
			headersNoncesPool.Remove(metaBlock.GetNonce(), metaBlock.GetShardID())
		}

		return
	}

	lastNotarizedHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return
	}
	if metaBlock.GetNonce() <= lastNotarizedHdr.GetNonce() {
		return
	}
	if metaBlock.GetRound() <= lastNotarizedHdr.GetRound() {
		return
	}

	sp.epochStartTrigger.ReceivedHeader(metaBlock)
	if sp.epochStartTrigger.IsEpochStart() {
		sp.chRcvEpochStart <- true
	}

	isMetaBlockOutOfRange := metaBlock.GetNonce() > lastNotarizedHdr.GetNonce()+process.MaxHeadersToRequestInAdvance
	if isMetaBlockOutOfRange {
		return
	}

	sp.txCoordinator.RequestMiniBlocks(metaBlock)
}

func (sp *shardProcessor) requestMetaHeaders(shardHeader *block.Header) (uint32, uint32) {
	_ = process.EmptyChannel(sp.chRcvAllMetaHdrs)

	if len(shardHeader.MetaBlockHashes) == 0 {
		return 0, 0
	}

	missingHeadersHashes := sp.computeMissingAndExistingMetaHeaders(shardHeader)

	sp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	for _, hash := range missingHeadersHashes {
		sp.hdrsForCurrBlock.hdrHashAndInfo[string(hash)] = &hdrInfo{hdr: nil, usedInBlock: true}
		go sp.onRequestHeaderHandler(sharding.MetachainShardId, hash)
	}

	if sp.hdrsForCurrBlock.missingHdrs == 0 {
		sp.hdrsForCurrBlock.missingFinalityAttestingHdrs = sp.requestMissingFinalityAttestingHeaders(
			sharding.MetachainShardId,
			sp.metaBlockFinality,
			sp.getMetaHeaderFromPoolWithNonce,
			sp.dataPool.MetaBlocks())
	}

	requestedHdrs := sp.hdrsForCurrBlock.missingHdrs
	requestedFinalityAttestingHdrs := sp.hdrsForCurrBlock.missingFinalityAttestingHdrs
	sp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	return requestedHdrs, requestedFinalityAttestingHdrs
}

func (sp *shardProcessor) computeMissingAndExistingMetaHeaders(header *block.Header) [][]byte {
	missingHeadersHashes := make([][]byte, 0, len(header.MetaBlockHashes))

	sp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	for i := 0; i < len(header.MetaBlockHashes); i++ {
		hdr, err := process.GetMetaHeaderFromPool(
			header.MetaBlockHashes[i],
			sp.dataPool.MetaBlocks())

		if err != nil {
			missingHeadersHashes = append(missingHeadersHashes, header.MetaBlockHashes[i])
			sp.hdrsForCurrBlock.missingHdrs++
			continue
		}

		sp.hdrsForCurrBlock.hdrHashAndInfo[string(header.MetaBlockHashes[i])] = &hdrInfo{hdr: hdr, usedInBlock: true}

		if hdr.Nonce > sp.hdrsForCurrBlock.highestHdrNonce[sharding.MetachainShardId] {
			sp.hdrsForCurrBlock.highestHdrNonce[sharding.MetachainShardId] = hdr.Nonce
		}
	}
	sp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	return sliceUtil.TrimSliceSliceByte(missingHeadersHashes)
}

func (sp *shardProcessor) verifyCrossShardMiniBlockDstMe(header *block.Header) error {
	miniBlockMetaHashes, err := sp.getAllMiniBlockDstMeFromMeta(header)
	if err != nil {
		return err
	}

	crossMiniBlockHashes := header.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
	for hash := range crossMiniBlockHashes {
		if _, ok := miniBlockMetaHashes[hash]; !ok {
			return process.ErrCrossShardMBWithoutConfirmationFromMeta
		}
	}

	return nil
}

func (sp *shardProcessor) getAllMiniBlockDstMeFromMeta(header *block.Header) (map[string][]byte, error) {
	lastHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return nil, err
	}

	miniBlockMetaHashes := make(map[string][]byte)

	sp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for _, metaBlockHash := range header.MetaBlockHashes {
		headerInfo, ok := sp.hdrsForCurrBlock.hdrHashAndInfo[string(metaBlockHash)]
		if !ok {
			continue
		}
		metaBlock, ok := headerInfo.hdr.(*block.MetaBlock)
		if !ok {
			continue
		}
		if metaBlock.GetRound() > header.Round {
			continue
		}
		if metaBlock.GetRound() <= lastHdr.GetRound() {
			continue
		}
		if metaBlock.GetNonce() <= lastHdr.GetNonce() {
			continue
		}

		crossMiniBlockHashes := metaBlock.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
		for hash := range crossMiniBlockHashes {
			miniBlockMetaHashes[hash] = metaBlockHash
		}
	}
	sp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	return miniBlockMetaHashes, nil
}

func (sp *shardProcessor) getOrderedMetaBlocks(round uint64) ([]*hashAndHdr, error) {
	metaBlocksPool := sp.dataPool.MetaBlocks()
	if metaBlocksPool == nil {
		return nil, process.ErrNilMetaBlocksPool
	}

	lastHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return nil, err
	}

	orderedMetaBlocks := make([]*hashAndHdr, 0)
	for _, key := range metaBlocksPool.Keys() {
		val, _ := metaBlocksPool.Peek(key)
		if val == nil {
			continue
		}

		hdr, ok := val.(*block.MetaBlock)
		if !ok {
			continue
		}

		if hdr.GetRound() > round {
			continue
		}
		if hdr.GetRound() <= lastHdr.GetRound() {
			continue
		}
		if hdr.GetNonce() <= lastHdr.GetNonce() {
			continue
		}

		orderedMetaBlocks = append(orderedMetaBlocks, &hashAndHdr{hdr: hdr, hash: key})
	}

	if len(orderedMetaBlocks) > 1 {
		sort.Slice(orderedMetaBlocks, func(i, j int) bool {
			return orderedMetaBlocks[i].hdr.GetNonce() < orderedMetaBlocks[j].hdr.GetNonce()
		})
	}

	return orderedMetaBlocks, nil
}

// isMetaHeaderFinal verifies if meta is trully final, in order to not do rollbacks
func (sp *shardProcessor) isMetaHeaderFinal(currHdr data.HeaderHandler, sortedHdrs []*hashAndHdr, startPos int) bool {
	if currHdr == nil || currHdr.IsInterfaceNil() {
		return false
	}
	if sortedHdrs == nil {
		return false
	}

	// verify if there are "K" block after current to make this one final
	lastVerifiedHdr := currHdr
	nextBlocksVerified := uint32(0)

	for i := startPos; i < len(sortedHdrs); i++ {
		if nextBlocksVerified >= sp.metaBlockFinality {
			return true
		}

		// found a header with the next nonce
		tmpHdr := sortedHdrs[i].hdr
		if tmpHdr.GetNonce() == lastVerifiedHdr.GetNonce()+1 {
			err := sp.headerValidator.IsHeaderConstructionValid(tmpHdr, lastVerifiedHdr)
			if err != nil {
				continue
			}

			lastVerifiedHdr = tmpHdr
			nextBlocksVerified += 1
		}
	}

	if nextBlocksVerified >= sp.metaBlockFinality {
		return true
	}

	return false
}

// full verification through metachain header
func (sp *shardProcessor) createAndProcessCrossMiniBlocksDstMe(
	maxItemsInBlock uint32,
	round uint64,
	haveTime func() bool,
) (block.MiniBlockSlice, uint32, uint32, error) {

	miniBlocks := make(block.MiniBlockSlice, 0)
	txsAdded := uint32(0)
	hdrsAdded := uint32(0)

	orderedMetaBlocks, err := sp.getOrderedMetaBlocks(round)
	if err != nil {
		return nil, 0, 0, err
	}

	log.Debug("metablocks ordered",
		"num metablocks", len(orderedMetaBlocks),
	)

	lastMetaHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return nil, 0, 0, err
	}

	// do processing in order
	sp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	for i := 0; i < len(orderedMetaBlocks); i++ {
		if !haveTime() {
			log.Debug("time is up after putting cross txs with destination to current shard",
				"num txs", txsAdded,
			)
			break
		}

		if len(miniBlocks) >= core.MaxMiniBlocksInBlock {
			log.Debug("max number of mini blocks allowed to be added in one shard block has been reached",
				"limit", len(miniBlocks),
			)
			break
		}

		itemsAddedInHeader := uint32(len(sp.hdrsForCurrBlock.hdrHashAndInfo) + len(miniBlocks))
		if itemsAddedInHeader >= maxItemsInBlock {
			log.Debug("max records allowed to be added in shard header has been reached",
				"limit", maxItemsInBlock,
			)
			break
		}

		hdr, ok := orderedMetaBlocks[i].hdr.(*block.MetaBlock)
		if !ok {
			continue
		}

		err = sp.headerValidator.IsHeaderConstructionValid(hdr, lastMetaHdr)
		if err != nil {
			continue
		}

		isFinal := sp.isMetaHeaderFinal(hdr, orderedMetaBlocks, i+1)
		if !isFinal {
			continue
		}

		if len(hdr.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())) == 0 {
			sp.hdrsForCurrBlock.hdrHashAndInfo[string(orderedMetaBlocks[i].hash)] = &hdrInfo{hdr: hdr, usedInBlock: true}
			hdrsAdded++
			lastMetaHdr = hdr
			continue
		}

		itemsAddedInBody := txsAdded
		if itemsAddedInBody >= maxItemsInBlock {
			continue
		}

		maxTxSpaceRemained := int32(maxItemsInBlock) - int32(itemsAddedInBody)
		maxMbSpaceRemained := sp.getMaxMiniBlocksSpaceRemained(
			maxItemsInBlock,
			itemsAddedInHeader+1,
			uint32(len(miniBlocks)))

		if maxTxSpaceRemained > 0 && maxMbSpaceRemained > 0 {
			processedMiniBlocksHashes := sp.getProcessedMiniBlocksHashes(orderedMetaBlocks[i].hash)
			currMBProcessed, currTxsAdded, hdrProcessFinished := sp.txCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe(
				hdr,
				processedMiniBlocksHashes,
				uint32(maxTxSpaceRemained),
				uint32(maxMbSpaceRemained),
				haveTime)

			// all txs processed, add to processed miniblocks
			miniBlocks = append(miniBlocks, currMBProcessed...)
			txsAdded = txsAdded + currTxsAdded

			if currTxsAdded > 0 {
				sp.hdrsForCurrBlock.hdrHashAndInfo[string(orderedMetaBlocks[i].hash)] = &hdrInfo{hdr: hdr, usedInBlock: true}
				hdrsAdded++
			}

			if !hdrProcessFinished {
				break
			}

			lastMetaHdr = hdr
		}
	}
	sp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	return miniBlocks, txsAdded, hdrsAdded, nil
}

func (sp *shardProcessor) createMiniBlocks(
	maxItemsInBlock uint32,
	round uint64,
	haveTime func() bool,
) (block.Body, error) {

	miniBlocks := make(block.Body, 0)

	if sp.accounts.JournalLen() != 0 {
		return nil, process.ErrAccountStateDirty
	}

	if !haveTime() {
		log.Debug("time is up after entered in createMiniBlocks method")
		return nil, process.ErrTimeIsOut
	}

	txPool := sp.dataPool.Transactions()
	if txPool == nil {
		return nil, process.ErrNilTransactionPool
	}

	startTime := time.Now()
	destMeMiniBlocks, nbTxs, nbHdrs, err := sp.createAndProcessCrossMiniBlocksDstMe(maxItemsInBlock, round, haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to create mbs to me",
		"time [s]", elapsedTime,
	)
	if err != nil {
		log.Debug("createAndProcessCrossMiniBlocksDstMe", "error", err.Error())
	}

	processedMetaHdrs, errNotCritical := sp.getOrderedProcessedMetaBlocksFromMiniBlocks(destMeMiniBlocks)
	if errNotCritical != nil {
		log.Debug("getOrderedProcessedMetaBlocksFromMiniBlocks", "error", errNotCritical.Error())
	}

	err = sp.setMetaConsensusData(processedMetaHdrs)
	if err != nil {
		return nil, err
	}

	startTime = time.Now()
	err = sp.updatePeerStateForFinalMetaHeaders(processedMetaHdrs)
	elapsedTime = time.Since(startTime)
	log.Debug("elapsed time to update peer state",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return nil, err
	}

	log.Debug("processed miniblocks and txs with destination in self shard",
		"num miniblocks", len(destMeMiniBlocks),
		"num txs", nbTxs,
	)

	if len(destMeMiniBlocks) > 0 {
		miniBlocks = append(miniBlocks, destMeMiniBlocks...)
	}

	maxTxSpaceRemained := int32(maxItemsInBlock) - int32(nbTxs)
	maxMbSpaceRemained := sp.getMaxMiniBlocksSpaceRemained(
		maxItemsInBlock,
		uint32(len(destMeMiniBlocks))+nbHdrs,
		uint32(len(miniBlocks)))

	startTime = time.Now()
	mbFromMe := sp.txCoordinator.CreateMbsAndProcessTransactionsFromMe(
		uint32(maxTxSpaceRemained),
		uint32(maxMbSpaceRemained),
		haveTime)
	elapsedTime = time.Since(startTime)
	log.Debug("elapsed time to create mbs from me",
		"time [s]", elapsedTime,
	)

	if len(mbFromMe) > 0 {
		miniBlocks = append(miniBlocks, mbFromMe...)
	}

	log.Debug("creating mini blocks has been finished",
		"num miniblocks", len(miniBlocks),
	)
	return miniBlocks, nil
}

// ApplyBodyToHeader creates a miniblock header list given a block body
func (sp *shardProcessor) ApplyBodyToHeader(hdr data.HeaderHandler, bodyHandler data.BodyHandler) error {
	log.Trace("started creating block header",
		"round", hdr.GetRound(),
	)
	shardHeader, ok := hdr.(*block.Header)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	shardHeader.MiniBlockHeaders = make([]block.MiniBlockHeader, 0)
	shardHeader.RootHash = sp.getRootHash()

	defer func() {
		go sp.checkAndRequestIfMetaHeadersMissing(hdr.GetRound())
	}()

	if bodyHandler == nil || bodyHandler.IsInterfaceNil() {
		return nil
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	totalTxCount, miniBlockHeaders, err := sp.createMiniBlockHeaders(body)
	if err != nil {
		return err
	}

	shardHeader.MiniBlockHeaders = miniBlockHeaders
	shardHeader.TxCount = uint32(totalTxCount)
	metaBlockHashes := sp.sortHeaderHashesForCurrentBlockByNonce(true)
	shardHeader.MetaBlockHashes = metaBlockHashes[sharding.MetachainShardId]

	if sp.epochStartTrigger.IsEpochStart() {
		shardHeader.EpochStartMetaHash = sp.epochStartTrigger.EpochStartMetaHdrHash()
	}

	sp.appStatusHandler.SetUInt64Value(core.MetricNumTxInBlock, uint64(totalTxCount))
	sp.appStatusHandler.SetUInt64Value(core.MetricNumMiniBlocks, uint64(len(body)))

	rootHash, err := sp.validatorStatisticsProcessor.RootHash()
	if err != nil {
		return err
	}

	shardHeader.ValidatorStatsRootHash = rootHash

	sp.blockSizeThrottler.Add(
		hdr.GetRound(),
		core.MaxUint32(hdr.ItemsInBody(), hdr.ItemsInHeader()))

	return nil
}

func (sp *shardProcessor) waitForMetaHdrHashes(waitTime time.Duration) error {
	select {
	case <-sp.chRcvAllMetaHdrs:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
}

// MarshalizedDataToBroadcast prepares underlying data into a marshalized object according to destination
func (sp *shardProcessor) MarshalizedDataToBroadcast(
	_ data.HeaderHandler,
	bodyHandler data.BodyHandler,
) (map[uint32][]byte, map[string][][]byte, error) {

	if bodyHandler == nil || bodyHandler.IsInterfaceNil() {
		return nil, nil, process.ErrNilMiniBlocks
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	mrsData := make(map[uint32][]byte, sp.shardCoordinator.NumberOfShards()+1)
	bodies, mrsTxs := sp.txCoordinator.CreateMarshalizedData(body)

	for shardId, subsetBlockBody := range bodies {
		buff, err := sp.marshalizer.Marshal(subsetBlockBody)
		if err != nil {
			log.Debug("marshalizer.Marshal", "error", process.ErrMarshalWithoutSuccess.Error())
			continue
		}
		mrsData[shardId] = buff
	}

	return mrsData, mrsTxs, nil
}

// DecodeBlockBody method decodes block body from a given byte array
func (sp *shardProcessor) DecodeBlockBody(dta []byte) data.BodyHandler {
	if dta == nil {
		return nil
	}

	var body block.Body

	err := sp.marshalizer.Unmarshal(&body, dta)
	if err != nil {
		log.Debug("marshalizer.Unmarshal", "error", err.Error())
		return nil
	}

	return body
}

// DecodeBlockHeader method decodes block header from a given byte array
func (sp *shardProcessor) DecodeBlockHeader(dta []byte) data.HeaderHandler {
	if dta == nil {
		return nil
	}

	var header block.Header

	err := sp.marshalizer.Unmarshal(&header, dta)
	if err != nil {
		log.Debug("marshalizer.Unmarshal", "error", err.Error())
		return nil
	}

	return &header
}

// IsInterfaceNil returns true if there is no value under the interface
func (sp *shardProcessor) IsInterfaceNil() bool {
	if sp == nil {
		return true
	}
	return false
}

func (sp *shardProcessor) addProcessedMiniBlock(metaBlockHash []byte, miniBlockHash []byte) {
	sp.mutProcessedMiniBlocks.Lock()
	miniBlocksProcessed, ok := sp.processedMiniBlocks[string(metaBlockHash)]
	if !ok {
		miniBlocksProcessed := make(map[string]struct{})
		miniBlocksProcessed[string(miniBlockHash)] = struct{}{}
		sp.processedMiniBlocks[string(metaBlockHash)] = miniBlocksProcessed
		sp.mutProcessedMiniBlocks.Unlock()
		return
	}

	miniBlocksProcessed[string(miniBlockHash)] = struct{}{}
	sp.mutProcessedMiniBlocks.Unlock()
}

func (sp *shardProcessor) removeProcessedMiniBlock(miniBlockHash []byte) {
	sp.mutProcessedMiniBlocks.Lock()
	for metaHash, miniBlocksProcessed := range sp.processedMiniBlocks {
		_, isProcessed := miniBlocksProcessed[string(miniBlockHash)]
		if isProcessed {
			delete(miniBlocksProcessed, string(miniBlockHash))
		}

		if len(miniBlocksProcessed) == 0 {
			delete(sp.processedMiniBlocks, metaHash)
		}
	}
	sp.mutProcessedMiniBlocks.Unlock()
}

func (sp *shardProcessor) removeAllProcessedMiniBlocks(metaBlockHash []byte) {
	sp.mutProcessedMiniBlocks.Lock()
	delete(sp.processedMiniBlocks, string(metaBlockHash))
	sp.mutProcessedMiniBlocks.Unlock()
}

func (sp *shardProcessor) getProcessedMiniBlocksHashes(metaBlockHash []byte) map[string]struct{} {
	sp.mutProcessedMiniBlocks.RLock()
	processedMiniBlocksHashes := sp.processedMiniBlocks[string(metaBlockHash)]
	sp.mutProcessedMiniBlocks.RUnlock()

	return processedMiniBlocksHashes
}

func (sp *shardProcessor) isMiniBlockProcessed(metaBlockHash []byte, miniBlockHash []byte) bool {
	sp.mutProcessedMiniBlocks.RLock()
	miniBlocksProcessed, ok := sp.processedMiniBlocks[string(metaBlockHash)]
	if !ok {
		sp.mutProcessedMiniBlocks.RUnlock()
		return false
	}

	_, isProcessed := miniBlocksProcessed[string(miniBlockHash)]
	sp.mutProcessedMiniBlocks.RUnlock()

	return isProcessed
}

func (sp *shardProcessor) getMaxMiniBlocksSpaceRemained(
	maxItemsInBlock uint32,
	itemsAddedInBlock uint32,
	miniBlocksAddedInBlock uint32,
) int32 {
	mbSpaceRemainedInBlock := int32(maxItemsInBlock) - int32(itemsAddedInBlock)
	mbSpaceRemainedInCache := int32(core.MaxMiniBlocksInBlock) - int32(miniBlocksAddedInBlock)
	maxMbSpaceRemained := core.MinInt32(mbSpaceRemainedInBlock, mbSpaceRemainedInCache)

	return maxMbSpaceRemained
}

func (sp *shardProcessor) getMetaHeaderFromPoolWithNonce(
	nonce uint64,
	_ uint32,
) (data.HeaderHandler, []byte, error) {

	metaHeader, metaHeaderHash, err := process.GetMetaHeaderFromPoolWithNonce(
		nonce,
		sp.dataPool.MetaBlocks(),
		sp.dataPool.HeadersNonces())

	return metaHeader, metaHeaderHash, err
}

func (sp *shardProcessor) updatePeerStateForFinalMetaHeaders(finalHeaders []data.HeaderHandler) error {
	for _, header := range finalHeaders {
		_, err := sp.validatorStatisticsProcessor.UpdatePeerState(header)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sp *shardProcessor) checkValidatorStatisticsRootHash(currentHeader *block.Header, processedMetaHdrs []data.HeaderHandler) error {
	for _, metaHeader := range processedMetaHdrs {
		rootHash, err := sp.validatorStatisticsProcessor.UpdatePeerState(metaHeader)
		if err != nil {
			return err
		}

		if !bytes.Equal(rootHash, metaHeader.GetValidatorStatsRootHash()) {
			return process.ErrValidatorStatsRootHashDoesNotMatch
		}
	}

	vRootHash, _ := sp.validatorStatisticsProcessor.RootHash()
	if !bytes.Equal(vRootHash, currentHeader.GetValidatorStatsRootHash()) {
		return process.ErrValidatorStatsRootHashDoesNotMatch
	}

	return nil
}

func (sp *shardProcessor) GetBlockBodyFromPool(headerHandler data.HeaderHandler) (data.BodyHandler, error) {
	miniBlockPool := sp.dataPool.MiniBlocks()
	if miniBlockPool == nil {
		return nil, process.ErrNilMiniBlockPool
	}

	header, ok := headerHandler.(*block.Header)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	miniBlocks := make(block.MiniBlockSlice, 0)
	for i := 0; i < len(header.MiniBlockHeaders); i++ {
		obj, ok := miniBlockPool.Get(header.MiniBlockHeaders[i].Hash)
		if !ok {
			continue
		}

		miniBlock, ok := obj.(*block.MiniBlock)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		miniBlocks = append(miniBlocks, miniBlock)
	}

	return block.Body(miniBlocks), nil
}
