package block

import (
	"bytes"
	"fmt"
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/serviceContainer"
	"github.com/ElrondNetwork/elrond-go/core/sliceUtil"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/block/processedMb"
	"github.com/ElrondNetwork/elrond-go/process/throttle"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
)

const maxCleanTime = time.Second

// shardProcessor implements shardProcessor interface and actually it tries to execute block
type shardProcessor struct {
	*baseProcessor
	metaBlockFinality uint32
	chRcvAllMetaHdrs  chan bool

	chRcvEpochStart chan bool

	processedMiniBlocks *processedMb.ProcessedMiniBlockTracker
	core                serviceContainer.Core
	txCounter           *transactionCounter
	txsPoolsCleaner     process.PoolsCleaner

	stateCheckpointModulus            uint
	lowestNonceInSelfNotarizedHeaders uint64
}

// NewShardProcessor creates a new shardProcessor object
func NewShardProcessor(arguments ArgShardProcessor) (*shardProcessor, error) {
	err := checkProcessorNilParameters(arguments.ArgBaseProcessor)
	if err != nil {
		return nil, err
	}

	if check.IfNil(arguments.DataPool) {
		return nil, process.ErrNilDataPoolHolder
	}
	if check.IfNil(arguments.DataPool.Headers()) {
		return nil, process.ErrNilHeadersDataPool
	}

	blockSizeThrottler, err := throttle.NewBlockSizeThrottle()
	if err != nil {
		return nil, err
	}

	base := &baseProcessor{
		accounts:           arguments.Accounts,
		blockSizeThrottler: blockSizeThrottler,
		forkDetector:       arguments.ForkDetector,
		hasher:             arguments.Hasher,
		marshalizer:        arguments.Marshalizer,
		store:              arguments.Store,
		feeHandler:         arguments.FeeHandler,
		shardCoordinator:   arguments.ShardCoordinator,
		nodesCoordinator:   arguments.NodesCoordinator,
		uint64Converter:    arguments.Uint64Converter,
		requestHandler:     arguments.RequestHandler,
		appStatusHandler:   statusHandler.NewNilStatusHandler(),
		blockChainHook:     arguments.BlockChainHook,
		txCoordinator:      arguments.TxCoordinator,
		rounder:            arguments.Rounder,
		epochStartTrigger:  arguments.EpochStartTrigger,
		headerValidator:    arguments.HeaderValidator,
		bootStorer:         arguments.BootStorer,
		blockTracker:       arguments.BlockTracker,
		dataPool:           arguments.DataPool,
		blockChain:         arguments.BlockChain,
	}

	if check.IfNil(arguments.TxsPoolsCleaner) {
		return nil, process.ErrNilTxsPoolsCleaner
	}

	sp := shardProcessor{
		core:                   arguments.Core,
		baseProcessor:          base,
		txCounter:              NewTransactionCounter(),
		txsPoolsCleaner:        arguments.TxsPoolsCleaner,
		stateCheckpointModulus: arguments.StateCheckpointModulus,
	}

	sp.baseProcessor.requestBlockBodyHandler = &sp
	sp.blockProcessor = &sp

	sp.chRcvAllMetaHdrs = make(chan bool)

	transactionPool := sp.dataPool.Transactions()
	if check.IfNil(transactionPool) {
		return nil, process.ErrNilTransactionPool
	}

	sp.hdrsForCurrBlock.hdrHashAndInfo = make(map[string]*hdrInfo)
	sp.hdrsForCurrBlock.highestHdrNonce = make(map[uint32]uint64)
	sp.processedMiniBlocks = processedMb.NewProcessedMiniBlocks()

	metaBlockPool := sp.dataPool.Headers()
	if check.IfNil(metaBlockPool) {
		return nil, process.ErrNilMetaBlocksPool
	}
	metaBlockPool.RegisterHandler(sp.receivedMetaBlock)

	sp.metaBlockFinality = process.BlockFinality

	return &sp, nil
}

// ProcessBlock processes a block. It returns nil if all ok or the specific error
func (sp *shardProcessor) ProcessBlock(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
	haveTime func() time.Duration,
) error {

	if haveTime == nil {
		return process.ErrNilHaveTimeHandler
	}

	err := sp.checkBlockValidity(headerHandler, bodyHandler)
	if err != nil {
		if err == process.ErrBlockHashDoesNotMatch {
			log.Debug("requested missing shard header",
				"hash", headerHandler.GetPrevHash(),
				"for shard", headerHandler.GetShardID(),
			)

			go sp.requestHandler.RequestShardHeader(headerHandler.GetShardID(), headerHandler.GetPrevHash())
		}

		return err
	}

	sp.requestHandler.SetEpoch(headerHandler.GetEpoch())

	log.Debug("started processing block",
		"epoch", headerHandler.GetEpoch(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
	)

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

	err = sp.checkEpochCorrectness(header)
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

	startTime := time.Now()
	err = sp.txCoordinator.ProcessBlockTransaction(body, haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to process block transaction",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	err = sp.txCoordinator.VerifyCreatedBlockTransactions(header, body)
	if err != nil {
		return err
	}

	err = sp.verifyAccumulatedFees(header)
	if err != nil {
		return err
	}

	if !sp.verifyStateRoot(header.GetRootHash()) {
		err = process.ErrRootStateDoesNotMatch
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

// RevertAccountState reverts the account state for cleanup failed process
func (sp *shardProcessor) RevertAccountState() {
	err := sp.accounts.RevertToSnapshot(0)
	if err != nil {
		log.Debug("RevertToSnapshot", "error", err.Error())
	}
}

// RevertStateToBlock recreates the state tries to the root hashes indicated by the provided header
func (sp *shardProcessor) RevertStateToBlock(header data.HeaderHandler) error {
	err := sp.accounts.RecreateTrie(header.GetRootHash())
	if err != nil {
		log.Debug("recreate trie with error for header",
			"nonce", header.GetNonce(),
			"hash", header.GetRootHash(),
		)

		return err
	}

	return nil
}

func (sp *shardProcessor) checkEpochCorrectness(
	header *block.Header,
) error {
	currentBlockHeader := sp.blockChain.GetCurrentBlockHeader()
	if check.IfNil(currentBlockHeader) {
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
		go sp.requestHandler.RequestMetaHeader(header.EpochStartMetaHash)
		sp.epochStartTrigger.Revert(currentBlockHeader.GetRound())
		return process.ErrEpochDoesNotMatch
	}

	return nil
}

// SetNumProcessedObj will set the num of processed transactions
func (sp *shardProcessor) SetNumProcessedObj(numObj uint64) {
	sp.txCounter.totalTxs = numObj
}

// checkMetaHeadersValidity - checks if listed metaheaders are valid as construction
func (sp *shardProcessor) checkMetaHeadersValidityAndFinality() error {
	lastCrossNotarizedHeader, _, err := sp.blockTracker.GetLastCrossNotarizedHeader(sharding.MetachainShardId)
	if err != nil {
		return err
	}

	usedMetaHdrs := sp.sortHeadersForCurrentBlockByNonce(true)
	if len(usedMetaHdrs[sharding.MetachainShardId]) == 0 {
		return nil
	}

	for _, metaHdr := range usedMetaHdrs[sharding.MetachainShardId] {
		err = sp.headerValidator.IsHeaderConstructionValid(metaHdr, lastCrossNotarizedHeader)
		if err != nil {
			return fmt.Errorf("%w : checkMetaHeadersValidityAndFinality -> isHdrConstructionValid", err)
		}

		lastCrossNotarizedHeader = metaHdr
	}

	err = sp.checkMetaHdrFinality(lastCrossNotarizedHeader)
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
				log.Debug("checkMetaHdrFinality -> isHdrConstructionValid",
					"error", err.Error())
				continue
			}

			lastVerifiedHdr = metaHdr
			nextBlocksVerified += 1
		}
	}

	if nextBlocksVerified < sp.metaBlockFinality {
		go sp.requestHandler.RequestMetaHeaderByNonce(lastVerifiedHdr.GetNonce())
		go sp.requestHandler.RequestMetaHeaderByNonce(lastVerifiedHdr.GetNonce() + 1)
		return process.ErrHeaderNotFinal
	}

	return nil
}

func (sp *shardProcessor) checkAndRequestIfMetaHeadersMissing(round uint64) {
	orderedMetaBlocks, _ := sp.blockTracker.GetTrackedHeaders(sharding.MetachainShardId)

	err := sp.requestHeadersIfMissing(orderedMetaBlocks, sharding.MetachainShardId, round)
	if err != nil {
		log.Debug("checkAndRequestIfMetaHeadersMissing", "error", err.Error())
	}
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
	invalidPool := sp.txCoordinator.GetAllCurrentUsedTxs(block.InvalidBlock)
	receiptPool := sp.txCoordinator.GetAllCurrentUsedTxs(block.ReceiptBlock)

	for hash, tx := range scPool {
		txPool[hash] = tx
	}
	for hash, tx := range rewardPool {
		txPool[hash] = tx
	}
	for hash, tx := range invalidPool {
		txPool[hash] = tx
	}
	for hash, tx := range receiptPool {
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
		sp.epochStartTrigger.Revert(header.GetRound())
	}

	go sp.txCounter.subtractRestoredTxs(restoredTxNr)

	sp.blockTracker.RemoveLastNotarizedHeaders()

	return nil
}

func (sp *shardProcessor) restoreMetaBlockIntoPool(mapMiniBlockHashes map[string]uint32, metaBlockHashes [][]byte) error {
	metaBlockPool := sp.dataPool.Headers()
	if metaBlockPool == nil {
		return process.ErrNilMetaBlocksPool
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

		metaBlockPool.AddHeader(metaBlockHash, metaBlock)

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
			sp.processedMiniBlocks.AddMiniBlockHash(metaBlockHash, string(miniBlockHash))
		}
	}

	for miniBlockHash := range mapMiniBlockHashes {
		sp.processedMiniBlocks.RemoveMiniBlockHash(miniBlockHash)
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

	log.Debug("started creating block body",
		"epoch", initialHdrData.GetEpoch(),
		"round", initialHdrData.GetRound(),
		"nonce", initialHdrData.GetNonce(),
	)

	miniBlocks, err := sp.createMiniBlocks(sp.blockSizeThrottler.MaxItemsToAdd(), haveTime)
	if err != nil {
		return nil, err
	}

	sp.requestHandler.SetEpoch(initialHdrData.GetEpoch())

	return miniBlocks, nil
}

// CommitBlock commits the block in the blockchain if everything was checked successfully
func (sp *shardProcessor) CommitBlock(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {

	var err error
	defer func() {
		if err != nil {
			sp.RevertAccountState()
		}
	}()

	err = checkForNils(headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	log.Debug("started committing block",
		"epoch", headerHandler.GetEpoch(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
	)

	err = sp.checkBlockValidity(headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	header, ok := headerHandler.(*block.Header)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	marshalizedHeader, err := sp.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	headerHash := sp.hasher.Compute(string(marshalizedHeader))

	go sp.saveShardHeader(header, headerHash, marshalizedHeader)

	headersPool := sp.dataPool.Headers()
	if headersPool == nil {
		err = process.ErrNilHeadersDataPool
		return err
	}

	headersPool.RemoveHeaderByHash(headerHash)

	body, ok := bodyHandler.(block.Body)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	go sp.saveBody(body)

	processedMetaHdrs, err := sp.getOrderedProcessedMetaBlocksFromHeader(header)
	if err != nil {
		return err
	}

	err = sp.addProcessedCrossMiniBlocksFromHeader(header)
	if err != nil {
		return err
	}

	selfNotarizedHeaders, selfNotarizedHeadersHashes, err := sp.getHighestHdrForOwnShardFromMetachain(processedMetaHdrs)
	if err != nil {
		return err
	}

	sp.cleanupBlockTrackerPools(headerHandler)

	err = sp.saveLastNotarizedHeader(sharding.MetachainShardId, processedMetaHdrs)
	if err != nil {
		return err
	}

	err = sp.commitAll()
	if err != nil {
		return err
	}

	if header.IsStartOfEpochBlock() {
		err = sp.checkEpochCorrectnessCrossChain()
		sp.epochStartTrigger.SetProcessed(header)
	}

	log.Info("shard block has been committed successfully",
		"epoch", header.Epoch,
		"round", header.Round,
		"nonce", header.Nonce,
		"hash", headerHash,
	)

	errNotCritical := sp.txCoordinator.RemoveBlockDataFromPool(body)
	if errNotCritical != nil {
		log.Debug("RemoveBlockDataFromPool", "error", errNotCritical.Error())
	}

	errNotCritical = sp.removeProcessedMetaBlocksFromPool(processedMetaHdrs)
	if errNotCritical != nil {
		log.Debug("removeProcessedMetaBlocksFromPool", "error", errNotCritical.Error())
	}

	errNotCritical = sp.forkDetector.AddHeader(header, headerHash, process.BHProcessed, selfNotarizedHeaders, selfNotarizedHeadersHashes)
	if errNotCritical != nil {
		log.Debug("forkDetector.AddHeader", "error", errNotCritical.Error())
	}

	sp.blockTracker.AddSelfNotarizedHeader(sp.shardCoordinator.SelfId(), sp.blockChain.GetCurrentBlockHeader(), sp.blockChain.GetCurrentBlockHeaderHash())

	selfNotarizedHeader, selfNotarizedHeaderHash := sp.getLastSelfNotarizedHeader()
	sp.blockTracker.AddSelfNotarizedHeader(sharding.MetachainShardId, selfNotarizedHeader, selfNotarizedHeaderHash)

	sp.updateStateStorage(selfNotarizedHeaders)

	highestFinalBlockNonce := sp.forkDetector.GetHighestFinalBlockNonce()
	log.Debug("highest final shard block",
		"nonce", highestFinalBlockNonce,
		"shard", sp.shardCoordinator.SelfId(),
	)

	lastBlockHeader := sp.blockChain.GetCurrentBlockHeader()

	err = sp.blockChain.SetCurrentBlockBody(body)
	if err != nil {
		return err
	}

	err = sp.blockChain.SetCurrentBlockHeader(header)
	if err != nil {
		return err
	}

	sp.blockChain.SetCurrentBlockHeaderHash(headerHash)
	sp.indexBlockIfNeeded(bodyHandler, headerHandler, lastBlockHeader)

	lastCrossNotarizedHeader, _, err := sp.blockTracker.GetLastCrossNotarizedHeader(sharding.MetachainShardId)
	if err != nil {
		return err
	}

	saveMetricsForACommittedBlock(
		sp.appStatusHandler,
		display.DisplayByteSlice(headerHash),
		highestFinalBlockNonce,
		lastCrossNotarizedHeader,
	)

	headerInfo := bootstrapStorage.BootstrapHeaderInfo{
		ShardId: header.GetShardID(),
		Nonce:   header.GetNonce(),
		Hash:    headerHash,
	}

	if len(selfNotarizedHeaders) > 0 {
		sp.lowestNonceInSelfNotarizedHeaders = selfNotarizedHeaders[0].GetNonce()
	}

	go sp.prepareDataForBootStorer(
		headerInfo,
		header.Round,
		sp.getBootstrapHeadersInfo(selfNotarizedHeaders, selfNotarizedHeadersHashes),
		nil,
		sp.lowestNonceInSelfNotarizedHeaders,
		sp.processedMiniBlocks.ConvertProcessedMiniBlocksMapToSlice(),
	)

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
		sp.blockTracker,
	)

	sp.blockSizeThrottler.Succeed(header.Round)

	log.Debug("pools info",
		"headers", sp.dataPool.Headers().Len(),
		"headers capacity", sp.dataPool.Headers().MaxSize(),
		"miniblocks", sp.dataPool.MiniBlocks().Len(),
		"miniblocks capacity", sp.dataPool.MiniBlocks().MaxSize(),
	)

	go sp.cleanupPools(headerHandler, headersPool)

	return nil
}

func (sp *shardProcessor) commitAll() error {
	_, err := sp.accounts.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (sp *shardProcessor) updateStateStorage(finalHeaders []data.HeaderHandler) {
	if !sp.accounts.IsPruningEnabled() {
		return
	}

	for i := range finalHeaders {
		sp.saveState(finalHeaders[i])

		prevHeader, errNotCritical := process.GetShardHeaderFromStorage(finalHeaders[i].GetPrevHash(), sp.marshalizer, sp.store)
		if errNotCritical != nil {
			log.Debug(errNotCritical.Error())
			continue
		}

		prevRootHash := prevHeader.GetRootHash()
		if prevRootHash == nil {
			continue
		}

		rootHash := finalHeaders[i].GetRootHash()
		if bytes.Equal(prevRootHash, rootHash) {
			continue
		}

		errNotCritical = sp.accounts.PruneTrie(prevRootHash)
		if errNotCritical != nil {
			log.Debug(errNotCritical.Error())
		}

		sp.accounts.CancelPrune(rootHash)
	}
}

func (sp *shardProcessor) saveState(finalHeader data.HeaderHandler) {
	if finalHeader.IsStartOfEpochBlock() {
		log.Debug("trie snapshot", "rootHash", finalHeader.GetRootHash())
		sp.accounts.SnapshotState(finalHeader.GetRootHash())
		return
	}

	// TODO generate checkpoint on a trigger
	if finalHeader.GetRound()%uint64(sp.stateCheckpointModulus) == 0 {
		log.Debug("trie checkpoint", "rootHash", finalHeader.GetRootHash())
		sp.accounts.SetStateCheckpoint(finalHeader.GetRootHash())
	}
}

func (sp *shardProcessor) checkEpochCorrectnessCrossChain() error {
	currentHeader := sp.blockChain.GetCurrentBlockHeader()
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

		sp.forkDetector.SetRollBackNonce(nonce)
		return process.ErrEpochDoesNotMatch
	}

	return nil
}

func (sp *shardProcessor) getLastSelfNotarizedHeader() (data.HeaderHandler, []byte) {
	hash := sp.forkDetector.GetHighestFinalBlockHash()
	header, err := process.GetShardHeader(hash, sp.dataPool.Headers(), sp.marshalizer, sp.store)
	if err != nil {
		log.Warn("getLastSelfNotarizedHeader.GetShardHeader", "error", err.Error(), "hash", hash, "nonce", sp.forkDetector.GetHighestFinalBlockNonce())
		return nil, nil
	}

	return header, hash
}

func (sp *shardProcessor) saveLastNotarizedHeader(shardId uint32, processedHdrs []data.HeaderHandler) error {
	lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash, err := sp.blockTracker.GetLastCrossNotarizedHeader(shardId)
	if err != nil {
		return err
	}

	lenProcessedHdrs := len(processedHdrs)
	if lenProcessedHdrs > 0 {
		if lastCrossNotarizedHeader.GetNonce() < processedHdrs[lenProcessedHdrs-1].GetNonce() {
			lastCrossNotarizedHeader = processedHdrs[lenProcessedHdrs-1]
			lastCrossNotarizedHeaderHash, err = core.CalculateHash(sp.marshalizer, sp.hasher, lastCrossNotarizedHeader)
			if err != nil {
				return err
			}
		}
	}

	sp.blockTracker.AddCrossNotarizedHeader(shardId, lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash)
	DisplayLastNotarized(sp.marshalizer, sp.hasher, lastCrossNotarizedHeader, shardId)

	return nil
}

// ApplyProcessedMiniBlocks will apply processed mini blocks
func (sp *shardProcessor) ApplyProcessedMiniBlocks(processedMiniBlocks *processedMb.ProcessedMiniBlockTracker) {
	sp.processedMiniBlocks = processedMiniBlocks
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
	return &block.Header{AccumulatedFees: big.NewInt(0)}
}

// getHighestHdrForOwnShardFromMetachain calculates the highest shard header notarized by metachain
func (sp *shardProcessor) getHighestHdrForOwnShardFromMetachain(
	processedHdrs []data.HeaderHandler,
) ([]data.HeaderHandler, [][]byte, error) {

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
			go sp.requestHandler.RequestShardHeader(shardInfo.ShardID, shardInfo.HeaderHash)

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

			sp.processedMiniBlocks.AddMiniBlockHash(string(metaBlockHash), string(miniBlockHash))

			delete(miniBlockHashes, key)
		}
	}
	sp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	return nil
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
			processedCrossMiniBlocksHashes[hash] = sp.processedMiniBlocks.IsMiniBlockProcessed(metaBlockHash, hash)
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
	lastCrossNotarizedHeader, _, err := sp.blockTracker.GetLastCrossNotarizedHeader(sharding.MetachainShardId)
	if err != nil {
		return err
	}

	processed := 0
	// processedMetaHdrs is also sorted
	for i := 0; i < len(processedMetaHdrs); i++ {
		hdr := processedMetaHdrs[i]

		// remove process finished
		if hdr.GetNonce() > lastCrossNotarizedHeader.GetNonce() {
			continue
		}

		// metablock was processed and finalized
		marshalizedHeader, err := sp.marshalizer.Marshal(hdr)
		if err != nil {
			log.Debug("marshalizer.Marshal", "error", err.Error())
			continue
		}

		headerHash := sp.hasher.Compute(string(marshalizedHeader))

		go func(header data.HeaderHandler, headerHash []byte, marshalizedHeader []byte) {
			sp.saveMetaHeader(header, headerHash, marshalizedHeader)
		}(hdr, headerHash, marshalizedHeader)

		sp.dataPool.Headers().RemoveHeaderByHash(headerHash)
		sp.processedMiniBlocks.RemoveMetaBlockHash(string(headerHash))

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
func (sp *shardProcessor) receivedMetaBlock(headerHandler data.HeaderHandler, metaBlockHash []byte) {
	metaBlocksPool := sp.dataPool.Headers()
	if metaBlocksPool == nil {
		return
	}

	metaBlock, ok := headerHandler.(*block.MetaBlock)
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
		headerInfoIsNotNil := hdrInfoForHash != nil
		headerIsMissing := headerInfoIsNotNil && check.IfNil(hdrInfoForHash.hdr)
		if headerIsMissing {
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
			)
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

	if sp.isHeaderOutOfRange(metaBlock) {
		metaBlocksPool.RemoveHeaderByHash(metaBlockHash)
		return
	}

	lastCrossNotarizedHeader, _, err := sp.blockTracker.GetLastCrossNotarizedHeader(metaBlock.GetShardID())
	if err != nil {
		log.Debug("receivedMetaBlock.GetLastCrossNotarizedHeader",
			"shard", metaBlock.GetShardID(),
			"error", err.Error())
		return
	}

	if metaBlock.GetNonce() <= lastCrossNotarizedHeader.GetNonce() {
		return
	}
	if metaBlock.GetRound() <= lastCrossNotarizedHeader.GetRound() {
		return
	}

	sp.epochStartTrigger.ReceivedHeader(metaBlock)
	if sp.epochStartTrigger.IsEpochStart() {
		sp.chRcvEpochStart <- true
	}

	isMetaBlockOutOfRequestRange := metaBlock.GetNonce() > lastCrossNotarizedHeader.GetNonce()+process.MaxHeadersToRequestInAdvance
	if isMetaBlockOutOfRequestRange {
		return
	}

	go sp.txCoordinator.RequestMiniBlocks(metaBlock)
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
		go sp.requestHandler.RequestMetaHeader(hash)
	}

	if sp.hdrsForCurrBlock.missingHdrs == 0 {
		sp.hdrsForCurrBlock.missingFinalityAttestingHdrs = sp.requestMissingFinalityAttestingHeaders(
			sharding.MetachainShardId,
			sp.metaBlockFinality,
		)
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
			sp.dataPool.Headers())

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
	lastCrossNotarizedHeader, _, err := sp.blockTracker.GetLastCrossNotarizedHeader(sharding.MetachainShardId)
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
		if metaBlock.GetRound() <= lastCrossNotarizedHeader.GetRound() {
			continue
		}
		if metaBlock.GetNonce() <= lastCrossNotarizedHeader.GetNonce() {
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

// full verification through metachain header
func (sp *shardProcessor) createAndProcessCrossMiniBlocksDstMe(
	maxItemsInBlock uint32,
	haveTime func() bool,
) (block.MiniBlockSlice, uint32, uint32, error) {

	miniBlocks := make(block.MiniBlockSlice, 0)
	txsAdded := uint32(0)
	hdrsAdded := uint32(0)

	sw := core.NewStopWatch()
	sw.Start("ComputeLongestMetaChainFromLastNotarized")
	orderedMetaBlocks, orderedMetaBlocksHashes, err := sp.blockTracker.ComputeLongestMetaChainFromLastNotarized()
	sw.Stop("ComputeLongestMetaChainFromLastNotarized")
	log.Debug("measurements", sw.GetMeasurements()...)
	if err != nil {
		return nil, 0, 0, err
	}

	log.Debug("metablocks ordered",
		"num metablocks", len(orderedMetaBlocks),
	)

	lastMetaHdr, _, err := sp.blockTracker.GetLastCrossNotarizedHeader(sharding.MetachainShardId)
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

		currMetaHdr := orderedMetaBlocks[i]
		if currMetaHdr.GetNonce() > lastMetaHdr.GetNonce()+1 {
			log.Debug("skip searching",
				"last meta hdr nonce", lastMetaHdr.GetNonce(),
				"curr meta hdr nonce", currMetaHdr.GetNonce())
			break
		}

		if len(currMetaHdr.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())) == 0 {
			sp.hdrsForCurrBlock.hdrHashAndInfo[string(orderedMetaBlocksHashes[i])] = &hdrInfo{hdr: currMetaHdr, usedInBlock: true}
			hdrsAdded++
			lastMetaHdr = currMetaHdr
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
			processedMiniBlocksHashes := sp.processedMiniBlocks.GetProcessedMiniBlocksHashes(string(orderedMetaBlocksHashes[i]))
			currMBProcessed, currTxsAdded, hdrProcessFinished := sp.txCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe(
				currMetaHdr,
				processedMiniBlocksHashes,
				uint32(maxTxSpaceRemained),
				uint32(maxMbSpaceRemained),
				haveTime)

			// all txs processed, add to processed miniblocks
			miniBlocks = append(miniBlocks, currMBProcessed...)
			txsAdded += currTxsAdded

			if currTxsAdded > 0 {
				sp.hdrsForCurrBlock.hdrHashAndInfo[string(orderedMetaBlocksHashes[i])] = &hdrInfo{hdr: currMetaHdr, usedInBlock: true}
				hdrsAdded++
			}

			if !hdrProcessFinished {
				log.Debug("meta block cannot be fully processed",
					"round", currMetaHdr.GetRound(),
					"nonce", currMetaHdr.GetNonce(),
					"hash", orderedMetaBlocksHashes[i])

				break
			}

			lastMetaHdr = currMetaHdr
		}
	}
	sp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	sp.requestMetaHeadersIfNeeded(hdrsAdded, lastMetaHdr)

	return miniBlocks, txsAdded, hdrsAdded, nil
}

func (sp *shardProcessor) requestMetaHeadersIfNeeded(hdrsAdded uint32, lastMetaHdr data.HeaderHandler) {
	log.Debug("meta hdrs added",
		"nb", hdrsAdded,
		"lastMetaHdr", lastMetaHdr.GetNonce(),
	)

	if hdrsAdded == 0 {
		fromNonce := lastMetaHdr.GetNonce() + 1
		toNonce := fromNonce + uint64(sp.metaBlockFinality)
		for nonce := fromNonce; nonce <= toNonce; nonce++ {
			go sp.requestHandler.RequestMetaHeaderByNonce(nonce)
		}
	}
}

func (sp *shardProcessor) createMiniBlocks(
	maxItemsInBlock uint32,
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
	destMeMiniBlocks, nbTxs, nbHdrs, err := sp.createAndProcessCrossMiniBlocksDstMe(maxItemsInBlock, haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to create mbs to me",
		"time [s]", elapsedTime,
	)
	if err != nil {
		log.Debug("createAndProcessCrossMiniBlocksDstMe", "error", err.Error())
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
func (sp *shardProcessor) ApplyBodyToHeader(hdr data.HeaderHandler, bodyHandler data.BodyHandler) (data.BodyHandler, error) {
	sw := core.NewStopWatch()
	sw.Start("ApplyBodyToHeader")
	defer func() {
		sw.Stop("ApplyBodyToHeader")
		log.Debug("measurements", sw.GetMeasurements()...)
	}()

	shardHeader, ok := hdr.(*block.Header)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	shardHeader.MiniBlockHeaders = make([]block.MiniBlockHeader, 0)
	shardHeader.RootHash = sp.getRootHash()

	defer func() {
		go sp.checkAndRequestIfMetaHeadersMissing(hdr.GetRound())
	}()

	if check.IfNil(bodyHandler) {
		return nil, process.ErrNilBlockBody
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	var err error
	sw.Start("CreateReceiptsHash")
	shardHeader.ReceiptsHash, err = sp.txCoordinator.CreateReceiptsHash()
	sw.Stop("CreateReceiptsHash")
	if err != nil {
		return nil, err
	}

	newBody := deleteSelfReceiptsMiniBlocks(body)

	sw.Start("createMiniBlockHeaders")
	totalTxCount, miniBlockHeaders, err := sp.createMiniBlockHeaders(newBody)
	sw.Stop("createMiniBlockHeaders")
	if err != nil {
		return nil, err
	}

	shardHeader.MiniBlockHeaders = miniBlockHeaders
	shardHeader.TxCount = uint32(totalTxCount)
	shardHeader.AccumulatedFees = sp.feeHandler.GetAccumulatedFees()

	sw.Start("sortHeaderHashesForCurrentBlockByNonce")
	metaBlockHashes := sp.sortHeaderHashesForCurrentBlockByNonce(true)
	sw.Stop("sortHeaderHashesForCurrentBlockByNonce")
	shardHeader.MetaBlockHashes = metaBlockHashes[sharding.MetachainShardId]

	if sp.epochStartTrigger.IsEpochStart() {
		shardHeader.EpochStartMetaHash = sp.epochStartTrigger.EpochStartMetaHdrHash()
	}

	sp.appStatusHandler.SetUInt64Value(core.MetricNumTxInBlock, uint64(totalTxCount))
	sp.appStatusHandler.SetUInt64Value(core.MetricNumMiniBlocks, uint64(len(body)))

	sp.blockSizeThrottler.Add(
		hdr.GetRound(),
		core.MaxUint32(hdr.ItemsInBody(), hdr.ItemsInHeader()))

	return newBody, nil
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

// IsInterfaceNil returns true if there is no value under the interface
func (sp *shardProcessor) IsInterfaceNil() bool {
	return sp == nil
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

// GetBlockBodyFromPool returns block body from pool for a given header
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
		obj, hashInPool := miniBlockPool.Get(header.MiniBlockHeaders[i].Hash)
		if !hashInPool {
			continue
		}

		miniBlock, typeOk := obj.(*block.MiniBlock)
		if !typeOk {
			return nil, process.ErrWrongTypeAssertion
		}

		miniBlocks = append(miniBlocks, miniBlock)
	}

	return block.Body(miniBlocks), nil
}

func (sp *shardProcessor) getBootstrapHeadersInfo(
	selfNotarizedHeaders []data.HeaderHandler,
	selfNotarizedHeadersHashes [][]byte,
) []bootstrapStorage.BootstrapHeaderInfo {

	lastSelfNotarizedHeaders := make([]bootstrapStorage.BootstrapHeaderInfo, 0, len(selfNotarizedHeaders))

	for index := range selfNotarizedHeaders {
		headerInfo := bootstrapStorage.BootstrapHeaderInfo{
			ShardId: selfNotarizedHeaders[index].GetShardID(),
			Nonce:   selfNotarizedHeaders[index].GetNonce(),
			Hash:    selfNotarizedHeadersHashes[index],
		}

		lastSelfNotarizedHeaders = append(lastSelfNotarizedHeaders, headerInfo)
	}

	return lastSelfNotarizedHeaders
}
