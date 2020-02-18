package block

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/serviceContainer"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/block/processedMb"
	"github.com/ElrondNetwork/elrond-go/process/throttle"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
)

// metaProcessor implements metaProcessor interface and actually it tries to execute block
type metaProcessor struct {
	*baseProcessor
	core                     serviceContainer.Core
	scDataGetter             external.SCQueryService
	scToProtocol             process.SmartContractToProtocolHandler
	peerChanges              process.PeerChangesHandler
	epochStartCreator        process.EpochStartDataCreator
	pendingMiniBlocksHandler process.PendingMiniBlocksHandler

	shardsHeadersNonce *sync.Map
	shardBlockFinality uint32
	chRcvAllHdrs       chan bool
	headersCounter     *headersCounter
}

// NewMetaProcessor creates a new metaProcessor object
func NewMetaProcessor(arguments ArgMetaProcessor) (*metaProcessor, error) {
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
	if check.IfNil(arguments.SCDataGetter) {
		return nil, process.ErrNilSCDataGetter
	}
	if check.IfNil(arguments.PeerChangesHandler) {
		return nil, process.ErrNilPeerChangesHandler
	}
	if check.IfNil(arguments.SCToProtocol) {
		return nil, process.ErrNilSCToProtocol
	}
	if check.IfNil(arguments.PendingMiniBlocksHandler) {
		return nil, process.ErrNilPendingMiniBlocksHandler
	}

	blockSizeThrottler, err := throttle.NewBlockSizeThrottle()
	if err != nil {
		return nil, err
	}

	base := &baseProcessor{
		accounts:                     arguments.Accounts,
		blockSizeThrottler:           blockSizeThrottler,
		forkDetector:                 arguments.ForkDetector,
		hasher:                       arguments.Hasher,
		marshalizer:                  arguments.Marshalizer,
		store:                        arguments.Store,
		shardCoordinator:             arguments.ShardCoordinator,
		nodesCoordinator:             arguments.NodesCoordinator,
		specialAddressHandler:        arguments.SpecialAddressHandler,
		uint64Converter:              arguments.Uint64Converter,
		requestHandler:               arguments.RequestHandler,
		appStatusHandler:             statusHandler.NewNilStatusHandler(),
		blockChainHook:               arguments.BlockChainHook,
		txCoordinator:                arguments.TxCoordinator,
		validatorStatisticsProcessor: arguments.ValidatorStatisticsProcessor,
		epochStartTrigger:            arguments.EpochStartTrigger,
		headerValidator:              arguments.HeaderValidator,
		rounder:                      arguments.Rounder,
		bootStorer:                   arguments.BootStorer,
		blockTracker:                 arguments.BlockTracker,
		dataPool:                     arguments.DataPool,
	}

	mp := metaProcessor{
		core:                     arguments.Core,
		baseProcessor:            base,
		headersCounter:           NewHeaderCounter(),
		scDataGetter:             arguments.SCDataGetter,
		peerChanges:              arguments.PeerChangesHandler,
		scToProtocol:             arguments.SCToProtocol,
		pendingMiniBlocksHandler: arguments.PendingMiniBlocksHandler,
	}

	mp.epochStartCreator, err = createEpochStartDataCreator(arguments)
	if err != nil {
		return nil, err
	}

	mp.baseProcessor.requestBlockBodyHandler = &mp
	mp.blockProcessor = &mp

	mp.hdrsForCurrBlock.hdrHashAndInfo = make(map[string]*hdrInfo)
	mp.hdrsForCurrBlock.highestHdrNonce = make(map[uint32]uint64)

	headerPool := mp.dataPool.Headers()
	headerPool.RegisterHandler(mp.receivedShardHeader)

	mp.chRcvAllHdrs = make(chan bool)

	mp.shardBlockFinality = process.BlockFinality

	mp.shardsHeadersNonce = &sync.Map{}

	return &mp, nil
}

func createEpochStartDataCreator(arguments ArgMetaProcessor) (process.EpochStartDataCreator, error) {
	argsNewEpochStartData := ArgsNewEpochStartData{
		Marshalizer:       arguments.Marshalizer,
		Hasher:            arguments.Hasher,
		Store:             arguments.Store,
		DataPool:          arguments.DataPool,
		BlockTracker:      arguments.BlockTracker,
		ShardCoordinator:  arguments.ShardCoordinator,
		EpochStartTrigger: arguments.EpochStartTrigger,
	}
	epochStartDataObject, err := NewEpochStartData(argsNewEpochStartData)
	if err != nil {
		return nil, err
	}

	return epochStartDataObject, nil
}

// ProcessBlock processes a block. It returns nil if all ok or the specific error
func (mp *metaProcessor) ProcessBlock(
	chainHandler data.ChainHandler,
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
	haveTime func() time.Duration,
) error {

	if haveTime == nil {
		return process.ErrNilHaveTimeHandler
	}

	err := mp.checkBlockValidity(chainHandler, headerHandler, bodyHandler)
	if err != nil {
		if err == process.ErrBlockHashDoesNotMatch {
			log.Debug("requested missing meta header",
				"hash", headerHandler.GetPrevHash(),
				"for shard", headerHandler.GetShardID(),
			)

			go mp.requestHandler.RequestMetaHeader(headerHandler.GetPrevHash())
		}

		return err
	}

	mp.requestHandler.SetEpoch(headerHandler.GetEpoch())

	log.Debug("started processing block",
		"epoch", headerHandler.GetEpoch(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce())

	header, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	err = mp.checkHeaderBodyCorrelation(header.MiniBlockHeaders, body)
	if err != nil {
		return err
	}

	go getMetricsFromMetaHeader(
		header,
		mp.marshalizer,
		mp.appStatusHandler,
		mp.dataPool.Headers().Len(),
		mp.headersCounter.getNumShardMBHeadersTotalProcessed(),
	)

	mp.createBlockStarted()
	mp.blockChainHook.SetCurrentHeader(headerHandler)
	mp.txCoordinator.RequestBlockTransactions(body)

	requestedShardHdrs, requestedFinalityAttestingShardHdrs := mp.requestShardHeaders(header)

	if haveTime() < 0 {
		return process.ErrTimeIsOut
	}

	err = mp.txCoordinator.IsDataPreparedForProcessing(haveTime)
	if err != nil {
		return err
	}

	haveMissingShardHeaders := requestedShardHdrs > 0 || requestedFinalityAttestingShardHdrs > 0
	if haveMissingShardHeaders {
		log.Debug("requested missing shard headers",
			"num headers", requestedShardHdrs,
		)
		log.Debug("requested missing finality attesting shard headers",
			"num finality shard headers", requestedFinalityAttestingShardHdrs,
		)

		err = mp.waitForBlockHeaders(haveTime())

		mp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
		missingShardHdrs := mp.hdrsForCurrBlock.missingHdrs
		mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

		mp.resetMissingHdrs()

		if requestedShardHdrs > 0 {
			log.Debug("received missing shard headers",
				"num headers", requestedShardHdrs-missingShardHdrs,
			)
		}

		if err != nil {
			return err
		}
	}

	if mp.accounts.JournalLen() != 0 {
		return process.ErrAccountStateDirty
	}

	defer func() {
		go mp.checkAndRequestIfShardHeadersMissing(header.Round)
	}()

	mp.epochStartTrigger.Update(header.GetRound())

	err = mp.checkEpochCorrectness(header, chainHandler)
	if err != nil {
		return err
	}

	err = mp.epochStartCreator.VerifyEpochStartDataForMetablock(header)
	if err != nil {
		return err
	}

	highestNonceHdrs, err := mp.checkShardHeadersValidity(header)
	if err != nil {
		return err
	}

	err = mp.checkShardHeadersFinality(highestNonceHdrs)
	if err != nil {
		return err
	}

	err = mp.verifyCrossShardMiniBlockDstMe(header)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			mp.RevertAccountState()
		}
	}()

	err = mp.processBlockHeaders(header, header.Round, haveTime)
	if err != nil {
		return err
	}

	err = mp.txCoordinator.ProcessBlockTransaction(body, haveTime)
	if err != nil {
		return err
	}

	err = mp.txCoordinator.VerifyCreatedBlockTransactions(header, body)
	if err != nil {
		return err
	}

	err = mp.scToProtocol.UpdateProtocol(body, header.Round)
	if err != nil {
		return err
	}

	err = mp.peerChanges.VerifyPeerChanges(header.PeerInfo)
	if err != nil {
		return err
	}

	if !mp.verifyStateRoot(header.GetRootHash()) {
		err = process.ErrRootStateDoesNotMatch
		return err
	}

	err = mp.verifyValidatorStatisticsRootHash(header)
	if err != nil {
		return err
	}

	return nil
}

// SetNumProcessedObj will set the num of processed headers
func (mp *metaProcessor) SetNumProcessedObj(numObj uint64) {
	mp.headersCounter.shardMBHeadersTotalProcessed = numObj
}

func (mp *metaProcessor) checkEpochCorrectness(
	headerHandler data.HeaderHandler,
	chainHandler data.ChainHandler,
) error {
	currentBlockHeader := chainHandler.GetCurrentBlockHeader()
	if currentBlockHeader == nil {
		return nil
	}

	isEpochIncorrect := headerHandler.GetEpoch() != currentBlockHeader.GetEpoch() &&
		mp.epochStartTrigger.Epoch() == currentBlockHeader.GetEpoch()
	if isEpochIncorrect {
		log.Warn("epoch does not match", "currentHeaderEpoch", currentBlockHeader.GetEpoch(), "receivedHeaderEpoch", headerHandler.GetEpoch(), "epochStartTrigger", mp.epochStartTrigger.Epoch())
		return process.ErrEpochDoesNotMatch
	}

	isEpochIncorrect = mp.epochStartTrigger.IsEpochStart() &&
		mp.epochStartTrigger.EpochStartRound() <= headerHandler.GetRound() &&
		headerHandler.GetEpoch() != currentBlockHeader.GetEpoch()+1

	if isEpochIncorrect {
		log.Warn("is epoch start and epoch does not match", "currentHeaderEpoch", currentBlockHeader.GetEpoch(), "receivedHeaderEpoch", headerHandler.GetEpoch(), "epochStartTrigger", mp.epochStartTrigger.Epoch())
		return process.ErrEpochDoesNotMatch
	}

	return nil
}

func (mp *metaProcessor) verifyCrossShardMiniBlockDstMe(header *block.MetaBlock) error {
	miniBlockShardsHashes, err := mp.getAllMiniBlockDstMeFromShards(header)
	if err != nil {
		return err
	}

	//if all miniblockshards hashes are in header miniblocks as well
	mapMetaMiniBlockHdrs := make(map[string]struct{}, len(header.MiniBlockHeaders))
	for _, metaMiniBlock := range header.MiniBlockHeaders {
		mapMetaMiniBlockHdrs[string(metaMiniBlock.Hash)] = struct{}{}
	}

	for hash := range miniBlockShardsHashes {
		if _, ok := mapMetaMiniBlockHdrs[hash]; !ok {
			return process.ErrCrossShardMBWithoutConfirmationFromMeta
		}
	}

	return nil
}

func (mp *metaProcessor) getAllMiniBlockDstMeFromShards(metaHdr *block.MetaBlock) (map[string][]byte, error) {
	miniBlockShardsHashes := make(map[string][]byte)

	mp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	defer mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	for _, shardInfo := range metaHdr.ShardInfo {
		headerInfo, ok := mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardInfo.HeaderHash)]
		if !ok {
			continue
		}
		shardHeader, ok := headerInfo.hdr.(*block.Header)
		if !ok {
			continue
		}

		lastCrossNotarizedHeader, _, err := mp.blockTracker.GetLastCrossNotarizedHeader(shardInfo.ShardID)
		if err != nil {
			return nil, err
		}

		if shardHeader.GetRound() > metaHdr.Round {
			continue
		}
		if shardHeader.GetRound() <= lastCrossNotarizedHeader.GetRound() {
			continue
		}
		if shardHeader.GetNonce() <= lastCrossNotarizedHeader.GetNonce() {
			continue
		}

		crossMiniBlockHashes := shardHeader.GetMiniBlockHeadersWithDst(mp.shardCoordinator.SelfId())
		for hash := range crossMiniBlockHashes {
			miniBlockShardsHashes[hash] = shardInfo.HeaderHash
		}
	}

	return miniBlockShardsHashes, nil
}

func (mp *metaProcessor) checkAndRequestIfShardHeadersMissing(round uint64) {
	orderedHdrsPerShard := mp.blockTracker.GetTrackedHeadersForAllShards()

	for i := uint32(0); i < mp.shardCoordinator.NumberOfShards(); i++ {
		err := mp.requestHeadersIfMissing(orderedHdrsPerShard[i], i, round)
		if err != nil {
			log.Debug("checkAndRequestIfShardHeadersMissing", "error", err.Error())
			continue
		}
	}
}

func (mp *metaProcessor) indexBlock(
	metaBlock data.HeaderHandler,
	body data.BodyHandler,
	lastMetaBlock data.HeaderHandler,
) {
	if mp.core == nil || mp.core.Indexer() == nil {
		return
	}
	// Update tps benchmarks in the DB
	tpsBenchmark := mp.core.TPSBenchmark()
	if tpsBenchmark != nil {
		go mp.core.Indexer().UpdateTPS(tpsBenchmark)
	}

	txPool := mp.txCoordinator.GetAllCurrentUsedTxs(block.TxBlock)
	scPool := mp.txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock)

	for hash, tx := range scPool {
		txPool[hash] = tx
	}

	publicKeys, err := mp.nodesCoordinator.GetValidatorsPublicKeys(metaBlock.GetPrevRandSeed(), metaBlock.GetRound(), sharding.MetachainShardId)
	if err != nil {
		return
	}

	signersIndexes := mp.nodesCoordinator.GetValidatorsIndexes(publicKeys)
	go mp.core.Indexer().SaveBlock(body, metaBlock, txPool, signersIndexes)

	saveRoundInfoInElastic(mp.core.Indexer(), mp.nodesCoordinator, sharding.MetachainShardId, metaBlock, lastMetaBlock, signersIndexes)
}

// removeBlockInfoFromPool removes the block info from associated pools
func (mp *metaProcessor) removeBlockInfoFromPool(header *block.MetaBlock) error {
	if header == nil || header.IsInterfaceNil() {
		return process.ErrNilMetaBlockHeader
	}

	headerPool := mp.dataPool.Headers()
	if headerPool == nil || headerPool.IsInterfaceNil() {
		return process.ErrNilHeadersDataPool
	}

	mp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for i := 0; i < len(header.ShardInfo); i++ {
		shardHeaderHash := header.ShardInfo[i].HeaderHash

		headerPool.RemoveHeaderByHash(shardHeaderHash)
	}
	mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	return nil
}

// RestoreBlockIntoPools restores the block into associated pools
func (mp *metaProcessor) RestoreBlockIntoPools(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error {
	if check.IfNil(headerHandler) {
		return process.ErrNilMetaBlockHeader
	}
	if bodyHandler == nil || bodyHandler.IsInterfaceNil() {
		return process.ErrNilTxBlockBody
	}

	metaBlock, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	headerPool := mp.dataPool.Headers()
	if check.IfNil(headerPool) {
		return process.ErrNilHeadersDataPool
	}

	hdrHashes := make([][]byte, len(metaBlock.ShardInfo))
	for i := 0; i < len(metaBlock.ShardInfo); i++ {
		hdrHashes[i] = metaBlock.ShardInfo[i].HeaderHash
	}

	err := mp.pendingMiniBlocksHandler.RevertHeader(metaBlock)
	if err != nil {
		return err
	}

	if metaBlock.IsStartOfEpochBlock() {
		mp.epochStartTrigger.Revert(metaBlock.GetRound())
	}

	for _, hdrHash := range hdrHashes {
		shardHeader, errNotCritical := process.GetShardHeaderFromStorage(hdrHash, mp.marshalizer, mp.store)
		if errNotCritical != nil {
			log.Debug("shard header not found in BlockHeaderUnit",
				"hash", hdrHash,
			)
			continue
		}

		headerPool.AddHeader(hdrHash, shardHeader)

		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardHeader.GetShardID())
		storer := mp.store.GetStorer(hdrNonceHashDataUnit)
		nonceToByteSlice := mp.uint64Converter.ToByteSlice(shardHeader.GetNonce())
		errNotCritical = storer.Remove(nonceToByteSlice)
		if errNotCritical != nil {
			log.Debug("ShardHdrNonceHashDataUnit.Remove", "error", errNotCritical.Error())
		}

		mp.headersCounter.subtractRestoredMBHeaders(len(shardHeader.MiniBlockHeaders))
	}

	_, errNotCritical := mp.txCoordinator.RestoreBlockDataFromStorage(body)
	if errNotCritical != nil {
		log.Debug("RestoreBlockDataFromStorage", "error", errNotCritical.Error())
	}

	mp.blockTracker.RemoveLastNotarizedHeaders()

	return nil
}

// CreateBlockBody creates block body of metachain
func (mp *metaProcessor) CreateBlockBody(initialHdrData data.HeaderHandler, haveTime func() bool) (data.BodyHandler, error) {
	mp.createBlockStarted()
	mp.blockSizeThrottler.ComputeMaxItems()

	mp.epochStartTrigger.Update(initialHdrData.GetRound())
	initialHdrData.SetEpoch(mp.epochStartTrigger.Epoch())
	mp.blockChainHook.SetCurrentHeader(initialHdrData)

	log.Debug("started creating block body",
		"epoch", initialHdrData.GetEpoch(),
		"round", initialHdrData.GetRound(),
		"nonce", initialHdrData.GetNonce(),
	)

	miniBlocks, err := mp.createMiniBlocks(mp.blockSizeThrottler.MaxItemsToAdd(), haveTime)
	if err != nil {
		return nil, err
	}

	err = mp.scToProtocol.UpdateProtocol(miniBlocks, initialHdrData.GetRound())
	if err != nil {
		return nil, err
	}

	mp.requestHandler.SetEpoch(initialHdrData.GetEpoch())

	return miniBlocks, nil
}

func (mp *metaProcessor) createMiniBlocks(
	maxItemsInBlock uint32,
	haveTime func() bool,
) (*block.Body, error) {

	miniBlocks := make(block.MiniBlockSlice, 0)
	if mp.epochStartTrigger.IsEpochStart() {
		return miniBlocks, nil
	}

	if mp.accounts.JournalLen() != 0 {
		return nil, process.ErrAccountStateDirty
	}

	if !haveTime() {
		log.Debug("time is up after entered in createMiniBlocks method")
		return nil, process.ErrTimeIsOut
	}

	txPool := mp.dataPool.Transactions()
	if txPool == nil {
		return nil, process.ErrNilTransactionPool
	}

	destMeMiniBlocks, nbTxs, nbHdrs, err := mp.createAndProcessCrossMiniBlocksDstMe(maxItemsInBlock, haveTime)
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
	maxMbSpaceRemained := mp.getMaxMiniBlocksSpaceRemained(
		maxItemsInBlock,
		uint32(len(destMeMiniBlocks))+nbHdrs,
		uint32(len(miniBlocks)))

	mbFromMe := mp.txCoordinator.CreateMbsAndProcessTransactionsFromMe(
		uint32(maxTxSpaceRemained),
		uint32(maxMbSpaceRemained),
		haveTime)

	if len(mbFromMe) > 0 {
		miniBlocks = append(miniBlocks, mbFromMe...)
	}

	log.Debug("creating mini blocks has been finished",
		"miniblocks created", len(miniBlocks),
	)

	return &block.Body{MiniBlocks: miniBlocks}, nil
}

// full verification through metachain header
func (mp *metaProcessor) createAndProcessCrossMiniBlocksDstMe(
	maxItemsInBlock uint32,
	haveTime func() bool,
) (block.MiniBlockSlice, uint32, uint32, error) {

	miniBlocks := make(block.MiniBlockSlice, 0)
	txsAdded := uint32(0)
	hdrsAdded := uint32(0)

	sw := core.NewStopWatch()
	sw.Start("ComputeLongestShardsChainsFromLastNotarized")
	orderedHdrs, orderedHdrsHashes, _, err := mp.blockTracker.ComputeLongestShardsChainsFromLastNotarized()
	sw.Stop("ComputeLongestShardsChainsFromLastNotarized")
	log.Debug("measurements ComputeLongestShardsChainsFromLastNotarized", sw.GetMeasurements()...)
	if err != nil {
		return nil, 0, 0, err
	}

	log.Debug("shard headers ordered",
		"num shard headers", len(orderedHdrs),
	)

	lastShardHdr, err := mp.blockTracker.GetLastCrossNotarizedHeadersForAllShards()
	if err != nil {
		return nil, 0, 0, err
	}

	hdrsAddedForShard := make(map[uint32]uint32)

	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	for i := 0; i < len(orderedHdrs); i++ {
		if !haveTime() {
			log.Debug("time is up after putting cross txs with destination to current shard",
				"num txs", txsAdded,
			)
			break
		}

		if len(miniBlocks) >= core.MaxMiniBlocksInBlock {
			log.Debug("max number of mini blocks allowed to be added in one shard block has been reached",
				"num miniblocks", len(miniBlocks),
			)
			break
		}

		itemsAddedInHeader := uint32(len(mp.hdrsForCurrBlock.hdrHashAndInfo) + len(miniBlocks))
		if itemsAddedInHeader >= maxItemsInBlock {
			log.Debug("max records allowed to be added in shard header has been reached",
				"num max items", maxItemsInBlock,
			)
			break
		}

		currShardHdr := orderedHdrs[i]
		if currShardHdr.GetNonce() > lastShardHdr[currShardHdr.GetShardID()].GetNonce()+1 {
			log.Debug("skip searching",
				"shard", currShardHdr.GetShardID(),
				"last shard hdr nonce", lastShardHdr[currShardHdr.GetShardID()].GetNonce(),
				"curr shard hdr nonce", currShardHdr.GetNonce())
			continue
		}

		if len(currShardHdr.GetMiniBlockHeadersWithDst(mp.shardCoordinator.SelfId())) == 0 {
			mp.hdrsForCurrBlock.hdrHashAndInfo[string(orderedHdrsHashes[i])] = &hdrInfo{hdr: currShardHdr, usedInBlock: true}
			hdrsAdded++
			hdrsAddedForShard[currShardHdr.GetShardID()]++
			lastShardHdr[currShardHdr.GetShardID()] = currShardHdr
			continue
		}

		itemsAddedInBody := txsAdded
		if itemsAddedInBody >= maxItemsInBlock {
			continue
		}

		maxTxSpaceRemained := int32(maxItemsInBlock) - int32(itemsAddedInBody)
		maxMbSpaceRemained := mp.getMaxMiniBlocksSpaceRemained(
			maxItemsInBlock,
			itemsAddedInHeader+1,
			uint32(len(miniBlocks)))

		if maxTxSpaceRemained > 0 && maxMbSpaceRemained > 0 {
			snapshot := mp.accounts.JournalLen()
			currMBProcessed, currTxsAdded, hdrProcessFinished := mp.txCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe(
				currShardHdr,
				nil,
				uint32(maxTxSpaceRemained),
				uint32(maxMbSpaceRemained),
				haveTime)

			if !hdrProcessFinished {
				log.Debug("shard header cannot be fully processed",
					"round", currShardHdr.GetRound(),
					"nonce", currShardHdr.GetNonce(),
					"hash", orderedHdrsHashes[i])

				// shard header must be processed completely
				errAccountState := mp.accounts.RevertToSnapshot(snapshot)
				if errAccountState != nil {
					// TODO: evaluate if reloading the trie from disk will might solve the problem
					log.Warn("accounts.RevertToSnapshot", "error", errAccountState.Error())
				}
				break
			}

			// all txs processed, add to processed miniblocks
			miniBlocks = append(miniBlocks, currMBProcessed...)
			txsAdded += currTxsAdded

			mp.hdrsForCurrBlock.hdrHashAndInfo[string(orderedHdrsHashes[i])] = &hdrInfo{hdr: currShardHdr, usedInBlock: true}
			hdrsAdded++
			hdrsAddedForShard[currShardHdr.GetShardID()]++

			lastShardHdr[currShardHdr.GetShardID()] = currShardHdr
		}
	}
	mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	mp.requestShardHeadersIfNeeded(hdrsAddedForShard, lastShardHdr)

	return miniBlocks, txsAdded, hdrsAdded, nil
}

func (mp *metaProcessor) requestShardHeadersIfNeeded(
	hdrsAddedForShard map[uint32]uint32,
	lastShardHdr map[uint32]data.HeaderHandler,
) {
	for shardID := uint32(0); shardID < mp.shardCoordinator.NumberOfShards(); shardID++ {
		log.Debug("shard hdrs added",
			"shard", shardID,
			"nb", hdrsAddedForShard[shardID],
			"lastShardHdr", lastShardHdr[shardID].GetNonce())

		if hdrsAddedForShard[shardID] == 0 {
			fromNonce := lastShardHdr[shardID].GetNonce() + 1
			toNonce := fromNonce + uint64(mp.shardBlockFinality)
			for nonce := fromNonce; nonce <= toNonce; nonce++ {
				go mp.requestHandler.RequestShardHeaderByNonce(shardID, nonce)
			}
		}
	}
}

func (mp *metaProcessor) processBlockHeaders(header *block.MetaBlock, round uint64, haveTime func() time.Duration) error {
	arguments := make([]interface{}, 0, len(header.ShardInfo))
	for i := 0; i < len(header.ShardInfo); i++ {
		shardData := header.ShardInfo[i]
		for j := 0; j < len(shardData.ShardMiniBlockHeaders); j++ {
			if haveTime() < 0 {
				return process.ErrTimeIsOut
			}

			headerHash := shardData.HeaderHash
			shardMiniBlockHeader := &shardData.ShardMiniBlockHeaders[j]
			err := mp.checkAndProcessShardMiniBlockHeader(
				headerHash,
				shardMiniBlockHeader,
				round,
				shardData.ShardID,
			)

			if err != nil {
				return err
			}

			arguments = append(arguments, "hash", shardMiniBlockHeader.Hash)
		}
	}

	if len(arguments) > 0 {
		log.Trace("the following miniblocks hashes were successfully processed", arguments...)
	}

	return nil
}

// CommitBlock commits the block in the blockchain if everything was checked successfully
func (mp *metaProcessor) CommitBlock(
	chainHandler data.ChainHandler,
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {

	var err error
	defer func() {
		if err != nil {
			mp.RevertAccountState()
		}
	}()

	err = checkForNils(chainHandler, headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	log.Debug("started committing block",
		"epoch", headerHandler.GetEpoch(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
	)

	err = mp.checkBlockValidity(chainHandler, headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	header, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	marshalizedHeader, err := mp.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	headerHash := mp.hasher.Compute(string(marshalizedHeader))

	go mp.saveMetaHeader(header, headerHash, marshalizedHeader)

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	go mp.saveBody(body)

	mp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for i := 0; i < len(header.ShardInfo); i++ {
		shardHeaderHash := header.ShardInfo[i].HeaderHash
		headerInfo, hashExists := mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardHeaderHash)]
		if !hashExists {
			mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrMissingHeader
		}

		shardBlock, ok := headerInfo.hdr.(*block.Header)
		if !ok {
			mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrWrongTypeAssertion
		}

		mp.updateShardHeadersNonce(shardBlock.ShardID, shardBlock.Nonce)

		marshalizedHeader, err := mp.marshalizer.Marshal(shardBlock)
		if err != nil {
			mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return err
		}

		go func(header data.HeaderHandler, headerHash []byte, marshalizedHeader []byte) {
			mp.saveShardHeader(header, headerHash, marshalizedHeader)
		}(shardBlock, shardHeaderHash, marshalizedHeader)
	}
	mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	mp.saveMetricCrossCheckBlockHeight()

	err = mp.commitAll()
	if err != nil {
		return err
	}

	mp.commitEpochStart(header, chainHandler)

	mp.cleanupBlockTrackerPools(headerHandler)

	err = mp.saveLastNotarizedHeader(header)
	if err != nil {
		return err
	}

	err = mp.pendingMiniBlocksHandler.AddProcessedHeader(header)
	if err != nil {
		return err
	}

	log.Info("meta block has been committed successfully",
		"epoch", header.Epoch,
		"round", header.Round,
		"nonce", header.Nonce,
		"hash", headerHash)

	errNotCritical := mp.removeBlockInfoFromPool(header)
	if errNotCritical != nil {
		log.Debug("removeBlockInfoFromPool", "error", errNotCritical.Error())
	}

	errNotCritical = mp.txCoordinator.RemoveBlockDataFromPool(body)
	if errNotCritical != nil {
		log.Debug(errNotCritical.Error())
	}

	errNotCritical = mp.forkDetector.AddHeader(header, headerHash, process.BHProcessed, nil, nil)
	if errNotCritical != nil {
		log.Debug("forkDetector.AddHeader", "error", errNotCritical.Error())
	}

	currentHeader, currentHeaderHash := getLastSelfNotarizedHeaderByItself(chainHandler)
	mp.blockTracker.AddSelfNotarizedHeader(mp.shardCoordinator.SelfId(), currentHeader, currentHeaderHash)

	for shardID := uint32(0); shardID < mp.shardCoordinator.NumberOfShards(); shardID++ {
		lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash := mp.getLastSelfNotarizedHeaderByShard(shardID)
		mp.blockTracker.AddSelfNotarizedHeader(shardID, lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash)
	}

	log.Debug("highest final meta block",
		"nonce", mp.forkDetector.GetHighestFinalBlockNonce(),
	)

	lastMetaBlock := chainHandler.GetCurrentBlockHeader()

	err = chainHandler.SetCurrentBlockBody(body)
	if err != nil {
		return err
	}

	err = chainHandler.SetCurrentBlockHeader(header)
	if err != nil {
		return err
	}

	chainHandler.SetCurrentBlockHeaderHash(headerHash)

	if mp.core != nil && mp.core.TPSBenchmark() != nil {
		mp.core.TPSBenchmark().Update(header)
	}

	mp.indexBlock(header, body, lastMetaBlock)

	saveMetachainCommitBlockMetrics(mp.appStatusHandler, header, headerHash, mp.nodesCoordinator)

	go mp.headersCounter.displayLogInfo(
		header,
		body,
		headerHash,
		mp.dataPool.Headers().Len(),
		mp.blockTracker,
	)

	headerInfo := bootstrapStorage.BootstrapHeaderInfo{
		ShardId: header.GetShardID(),
		Nonce:   header.GetNonce(),
		Hash:    headerHash,
	}

	go mp.prepareDataForBootStorer(
		headerInfo,
		header.Round,
		nil,
		mp.getPendingMiniBlocks(),
		mp.forkDetector.GetHighestFinalBlockNonce(),
		nil,
	)

	mp.blockSizeThrottler.Succeed(header.Round)

	log.Debug("pools info",
		"headers pool", mp.dataPool.Headers().Len(),
		"headers pool capacity", mp.dataPool.Headers().MaxSize(),
	)

	go mp.cleanupPools(headerHandler)

	return nil
}

func (mp *metaProcessor) commitAll() error {
	_, err := mp.accounts.Commit()
	if err != nil {
		return err
	}

	_, err = mp.validatorStatisticsProcessor.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (mp *metaProcessor) getLastSelfNotarizedHeaderByShard(_ uint32) (data.HeaderHandler, []byte) {
	//TODO: Implement mechanism to extract last meta header notarized by the given shard if this info will be needed later
	return nil, nil
}

// ApplyProcessedMiniBlocks will do nothing on meta processor
func (mp *metaProcessor) ApplyProcessedMiniBlocks(_ *processedMb.ProcessedMiniBlockTracker) {
}

func (mp *metaProcessor) commitEpochStart(header data.HeaderHandler, chainHandler data.ChainHandler) {
	if header.IsStartOfEpochBlock() {
		mp.epochStartTrigger.SetProcessed(header)
	} else {
		currentHeader := chainHandler.GetCurrentBlockHeader()
		if currentHeader != nil && currentHeader.IsStartOfEpochBlock() {
			mp.epochStartTrigger.SetFinalityAttestingRound(header.GetRound())
		}
	}
}

// RevertAccountState reverts the account state for cleanup failed process
func (mp *metaProcessor) RevertAccountState() {
	err := mp.accounts.RevertToSnapshot(0)
	if err != nil {
		log.Debug("RevertToSnapshot", "error", err.Error())
	}

	err = mp.validatorStatisticsProcessor.RevertPeerStateToSnapshot(0)
	if err != nil {
		log.Debug("RevertPeerStateToSnapshot", "error", err.Error())
	}
}

// RevertStateToBlock recreates the state tries to the root hashes indicated by the provided header
func (mp *metaProcessor) RevertStateToBlock(header data.HeaderHandler) error {
	err := mp.accounts.RecreateTrie(header.GetRootHash())
	if err != nil {
		log.Debug("recreate trie with error for header",
			"nonce", header.GetNonce(),
			"hash", header.GetRootHash(),
		)

		return err
	}

	err = mp.validatorStatisticsProcessor.RevertPeerState(header)
	if err != nil {
		log.Debug("revert peer state with error for header",
			"nonce", header.GetNonce(),
			"validators root hash", header.GetValidatorStatsRootHash(),
		)

		return err
	}

	return nil
}

func (mp *metaProcessor) updateShardHeadersNonce(key uint32, value uint64) {
	valueStoredI, ok := mp.shardsHeadersNonce.Load(key)
	if !ok {
		mp.shardsHeadersNonce.Store(key, value)
		return
	}

	valueStored, ok := valueStoredI.(uint64)
	if !ok {
		mp.shardsHeadersNonce.Store(key, value)
		return
	}

	if valueStored < value {
		mp.shardsHeadersNonce.Store(key, value)
	}
}

func (mp *metaProcessor) saveMetricCrossCheckBlockHeight() {
	crossCheckBlockHeight := ""
	for i := uint32(0); i < mp.shardCoordinator.NumberOfShards(); i++ {
		heightValue := uint64(0)

		valueStoredI, isValueInMap := mp.shardsHeadersNonce.Load(i)
		if isValueInMap {
			valueStored, ok := valueStoredI.(uint64)
			if ok {
				heightValue = valueStored
			}
		}

		crossCheckBlockHeight += fmt.Sprintf("%d: %d, ", i, heightValue)
	}

	mp.appStatusHandler.SetStringValue(core.MetricCrossCheckBlockHeight, crossCheckBlockHeight)
}

func (mp *metaProcessor) saveLastNotarizedHeader(header *block.MetaBlock) error {
	lastCrossNotarizedHeaderForShard := make(map[uint32]*hashAndHdr, mp.shardCoordinator.NumberOfShards())
	for shardID := uint32(0); shardID < mp.shardCoordinator.NumberOfShards(); shardID++ {
		lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash, err := mp.blockTracker.GetLastCrossNotarizedHeader(shardID)
		if err != nil {
			return err
		}

		lastCrossNotarizedHeaderForShard[shardID] = &hashAndHdr{hdr: lastCrossNotarizedHeader, hash: lastCrossNotarizedHeaderHash}
	}

	mp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for i := 0; i < len(header.ShardInfo); i++ {
		shardHeaderHash := header.ShardInfo[i].HeaderHash
		headerInfo, ok := mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardHeaderHash)]
		if !ok {
			mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrMissingHeader
		}

		shardHeader, ok := headerInfo.hdr.(*block.Header)
		if !ok {
			mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrWrongTypeAssertion
		}

		if lastCrossNotarizedHeaderForShard[shardHeader.ShardId].hdr.GetNonce() < shardHeader.Nonce {
			lastCrossNotarizedHeaderForShard[shardHeader.ShardId] = &hashAndHdr{hdr: shardHeader, hash: shardHeaderHash}
		}
	}
	mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	for shardID := uint32(0); shardID < mp.shardCoordinator.NumberOfShards(); shardID++ {
		hdr := lastCrossNotarizedHeaderForShard[shardID].hdr
		hash := lastCrossNotarizedHeaderForShard[shardID].hash
		mp.blockTracker.AddCrossNotarizedHeader(shardID, hdr, hash)
		DisplayLastNotarized(mp.marshalizer, mp.hasher, hdr, shardID)
	}

	return nil
}

// check if shard headers were signed and constructed correctly and returns headers which has to be
// checked for finality
func (mp *metaProcessor) checkShardHeadersValidity(metaHdr *block.MetaBlock) (map[uint32]data.HeaderHandler, error) {
	lastCrossNotarizedHeader := make(map[uint32]data.HeaderHandler, mp.shardCoordinator.NumberOfShards())
	for shardID := uint32(0); shardID < mp.shardCoordinator.NumberOfShards(); shardID++ {
		lastCrossNotarizedHeaderForShard, _, err := mp.blockTracker.GetLastCrossNotarizedHeader(shardID)
		if err != nil {
			return nil, err
		}

		lastCrossNotarizedHeader[shardID] = lastCrossNotarizedHeaderForShard
	}

	usedShardHdrs := mp.sortHeadersForCurrentBlockByNonce(true)
	highestNonceHdrs := make(map[uint32]data.HeaderHandler, len(usedShardHdrs))

	if len(usedShardHdrs) == 0 {
		return highestNonceHdrs, nil
	}

	for shardID, hdrsForShard := range usedShardHdrs {
		for _, shardHdr := range hdrsForShard {
			err := mp.headerValidator.IsHeaderConstructionValid(shardHdr, lastCrossNotarizedHeader[shardID])
			if err != nil {
				return nil, fmt.Errorf("%w : checkShardHeadersValidity -> isHdrConstructionValid", err)
			}

			lastCrossNotarizedHeader[shardID] = shardHdr
			highestNonceHdrs[shardID] = shardHdr
		}
	}

	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	defer mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	for _, shardData := range metaHdr.ShardInfo {
		actualHdr := mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardData.HeaderHash)].hdr
		shardHdr, ok := actualHdr.(*block.Header)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		if len(shardData.ShardMiniBlockHeaders) != len(shardHdr.MiniBlockHeaders) {
			return nil, process.ErrHeaderShardDataMismatch
		}

		mapMiniBlockHeadersInMetaBlock := make(map[string]struct{})
		for _, shardMiniBlockHdr := range shardData.ShardMiniBlockHeaders {
			mapMiniBlockHeadersInMetaBlock[string(shardMiniBlockHdr.Hash)] = struct{}{}
		}

		for _, actualMiniBlockHdr := range shardHdr.MiniBlockHeaders {
			if _, hashExists := mapMiniBlockHeadersInMetaBlock[string(actualMiniBlockHdr.Hash)]; !hashExists {
				return nil, process.ErrHeaderShardDataMismatch
			}
		}
	}

	return highestNonceHdrs, nil
}

// check if shard headers are final by checking if newer headers were constructed upon them
func (mp *metaProcessor) checkShardHeadersFinality(highestNonceHdrs map[uint32]data.HeaderHandler) error {
	finalityAttestingShardHdrs := mp.sortHeadersForCurrentBlockByNonce(false)

	var errFinal error

	for shardId, lastVerifiedHdr := range highestNonceHdrs {
		if lastVerifiedHdr == nil || lastVerifiedHdr.IsInterfaceNil() {
			return process.ErrNilBlockHeader
		}
		if lastVerifiedHdr.GetShardID() != shardId {
			return process.ErrShardIdMissmatch
		}

		// verify if there are "K" block after current to make this one final
		nextBlocksVerified := uint32(0)
		for _, shardHdr := range finalityAttestingShardHdrs[shardId] {
			if nextBlocksVerified >= mp.shardBlockFinality {
				break
			}

			// found a header with the next nonce
			if shardHdr.GetNonce() == lastVerifiedHdr.GetNonce()+1 {
				err := mp.headerValidator.IsHeaderConstructionValid(shardHdr, lastVerifiedHdr)
				if err != nil {
					log.Debug("checkShardHeadersFinality -> isHdrConstructionValid",
						"error", err.Error())
					continue
				}

				lastVerifiedHdr = shardHdr
				nextBlocksVerified += 1
			}
		}

		if nextBlocksVerified < mp.shardBlockFinality {
			go mp.requestHandler.RequestShardHeaderByNonce(lastVerifiedHdr.GetShardID(), lastVerifiedHdr.GetNonce())
			go mp.requestHandler.RequestShardHeaderByNonce(lastVerifiedHdr.GetShardID(), lastVerifiedHdr.GetNonce()+1)
			errFinal = process.ErrHeaderNotFinal
		}
	}

	return errFinal
}

// receivedShardHeader is a call back function which is called when a new header
// is added in the headers pool
func (mp *metaProcessor) receivedShardHeader(headerHandler data.HeaderHandler, shardHeaderHash []byte) {
	shardHeadersPool := mp.dataPool.Headers()
	if shardHeadersPool == nil {
		return
	}

	shardHeader, ok := headerHandler.(*block.Header)
	if !ok {
		return
	}

	log.Trace("received shard header from network",
		"shard", shardHeader.ShardId,
		"round", shardHeader.Round,
		"nonce", shardHeader.Nonce,
		"hash", shardHeaderHash,
	)

	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()

	haveMissingShardHeaders := mp.hdrsForCurrBlock.missingHdrs > 0 || mp.hdrsForCurrBlock.missingFinalityAttestingHdrs > 0
	if haveMissingShardHeaders {
		hdrInfoForHash := mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardHeaderHash)]
		headerInfoIsNotNil := hdrInfoForHash != nil
		headerIsMissing := headerInfoIsNotNil && check.IfNil(hdrInfoForHash.hdr)
		if headerIsMissing {
			hdrInfoForHash.hdr = shardHeader
			mp.hdrsForCurrBlock.missingHdrs--

			if shardHeader.Nonce > mp.hdrsForCurrBlock.highestHdrNonce[shardHeader.ShardID] {
				mp.hdrsForCurrBlock.highestHdrNonce[shardHeader.ShardID] = shardHeader.Nonce
			}
		}

		if mp.hdrsForCurrBlock.missingHdrs == 0 {
			mp.hdrsForCurrBlock.missingFinalityAttestingHdrs = mp.requestMissingFinalityAttestingShardHeaders()
			if mp.hdrsForCurrBlock.missingFinalityAttestingHdrs == 0 {
				log.Debug("received all missing finality attesting shard headers")
			}
		}

		missingShardHdrs := mp.hdrsForCurrBlock.missingHdrs
		missingFinalityAttestingShardHdrs := mp.hdrsForCurrBlock.missingFinalityAttestingHdrs
		mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

		allMissingShardHeadersReceived := missingShardHdrs == 0 && missingFinalityAttestingShardHdrs == 0
		if allMissingShardHeadersReceived {
			mp.chRcvAllHdrs <- true
		}
	} else {
		mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()
	}

	if mp.isHeaderOutOfRange(shardHeader) {
		shardHeadersPool.RemoveHeaderByHash(shardHeaderHash)
		return
	}

	lastCrossNotarizedHeader, _, err := mp.blockTracker.GetLastCrossNotarizedHeader(shardHeader.GetShardID())
	if err != nil {
		log.Debug("receivedShardHeader.GetLastCrossNotarizedHeader",
			"shard", shardHeader.GetShardID(),
			"error", err.Error())
		return
	}

	if shardHeader.GetNonce() <= lastCrossNotarizedHeader.GetNonce() {
		return
	}
	if shardHeader.GetRound() <= lastCrossNotarizedHeader.GetRound() {
		return
	}

	isShardHeaderOutOfRequestRange := shardHeader.GetNonce() > lastCrossNotarizedHeader.GetNonce()+process.MaxHeadersToRequestInAdvance
	if isShardHeaderOutOfRequestRange {
		return
	}

	go mp.txCoordinator.RequestMiniBlocks(shardHeader)
}

// requestMissingFinalityAttestingShardHeaders requests the headers needed to accept the current selected headers for
// processing the current block. It requests the shardBlockFinality headers greater than the highest shard header,
// for each shard, related to the block which should be processed
func (mp *metaProcessor) requestMissingFinalityAttestingShardHeaders() uint32 {
	missingFinalityAttestingShardHeaders := uint32(0)

	for shardId := uint32(0); shardId < mp.shardCoordinator.NumberOfShards(); shardId++ {
		missingFinalityAttestingHeaders := mp.requestMissingFinalityAttestingHeaders(
			shardId,
			mp.shardBlockFinality,
		)

		missingFinalityAttestingShardHeaders += missingFinalityAttestingHeaders
	}

	return missingFinalityAttestingShardHeaders
}

func (mp *metaProcessor) requestShardHeaders(metaBlock *block.MetaBlock) (uint32, uint32) {
	_ = process.EmptyChannel(mp.chRcvAllHdrs)

	if len(metaBlock.ShardInfo) == 0 {
		return 0, 0
	}

	missingHeaderHashes := mp.computeMissingAndExistingShardHeaders(metaBlock)

	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	for shardId, shardHeaderHashes := range missingHeaderHashes {
		for _, hash := range shardHeaderHashes {
			mp.hdrsForCurrBlock.hdrHashAndInfo[string(hash)] = &hdrInfo{hdr: nil, usedInBlock: true}
			go mp.requestHandler.RequestShardHeader(shardId, hash)
		}
	}

	if mp.hdrsForCurrBlock.missingHdrs == 0 {
		mp.hdrsForCurrBlock.missingFinalityAttestingHdrs = mp.requestMissingFinalityAttestingShardHeaders()
	}

	requestedHdrs := mp.hdrsForCurrBlock.missingHdrs
	requestedFinalityAttestingHdrs := mp.hdrsForCurrBlock.missingFinalityAttestingHdrs
	mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	return requestedHdrs, requestedFinalityAttestingHdrs
}

func (mp *metaProcessor) computeMissingAndExistingShardHeaders(metaBlock *block.MetaBlock) map[uint32][][]byte {
	missingHeadersHashes := make(map[uint32][][]byte)

	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	for i := 0; i < len(metaBlock.ShardInfo); i++ {
		shardData := metaBlock.ShardInfo[i]
		hdr, err := process.GetShardHeaderFromPool(
			shardData.HeaderHash,
			mp.dataPool.Headers())

		if err != nil {
			missingHeadersHashes[shardData.ShardID] = append(missingHeadersHashes[shardData.ShardID], shardData.HeaderHash)
			mp.hdrsForCurrBlock.missingHdrs++
			continue
		}

		mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardData.HeaderHash)] = &hdrInfo{hdr: hdr, usedInBlock: true}

		if hdr.Nonce > mp.hdrsForCurrBlock.highestHdrNonce[shardData.ShardID] {
			mp.hdrsForCurrBlock.highestHdrNonce[shardData.ShardID] = hdr.Nonce
		}
	}
	mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	return missingHeadersHashes
}

func (mp *metaProcessor) checkAndProcessShardMiniBlockHeader(
	_ []byte,
	_ *block.ShardMiniBlockHeader,
	_ uint64,
	_ uint32,
) error {
	// TODO: real processing has to be done here, using metachain state
	return nil
}

func (mp *metaProcessor) createShardInfo(
	round uint64,
) ([]block.ShardData, error) {

	shardInfo := make([]block.ShardData, 0)
	if mp.epochStartTrigger.IsEpochStart() {
		return shardInfo, nil
	}

	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	for hdrHash, headerInfo := range mp.hdrsForCurrBlock.hdrHashAndInfo {
		shardHdr, ok := headerInfo.hdr.(*block.Header)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		shardData := block.ShardData{}
		shardData.ShardMiniBlockHeaders = make([]block.ShardMiniBlockHeader, 0, len(shardHdr.MiniBlockHeaders))
		shardData.TxCount = shardHdr.TxCount
		shardData.ShardID = shardHdr.ShardID
		shardData.HeaderHash = []byte(hdrHash)
		shardData.Round = shardHdr.Round
		shardData.PrevHash = shardHdr.PrevHash
		shardData.Nonce = shardHdr.Nonce
		shardData.PrevRandSeed = shardHdr.PrevRandSeed
		shardData.PubKeysBitmap = shardHdr.PubKeysBitmap
		shardData.NumPendingMiniBlocks = mp.pendingMiniBlocksHandler.GetNumPendingMiniBlocks(shardData.ShardID)

		for i := 0; i < len(shardHdr.MiniBlockHeaders); i++ {
			shardMiniBlockHeader := block.ShardMiniBlockHeader{}
			shardMiniBlockHeader.SenderShardID = shardHdr.MiniBlockHeaders[i].SenderShardID
			shardMiniBlockHeader.ReceiverShardID = shardHdr.MiniBlockHeaders[i].ReceiverShardID
			shardMiniBlockHeader.Hash = shardHdr.MiniBlockHeaders[i].Hash
			shardMiniBlockHeader.TxCount = shardHdr.MiniBlockHeaders[i].TxCount

			// execute shard miniblock to change the trie root hash
			err := mp.checkAndProcessShardMiniBlockHeader(
				[]byte(hdrHash),
				&shardMiniBlockHeader,
				round,
				shardData.ShardID,
			)
			if err != nil {
				return nil, err
			}

			shardData.ShardMiniBlockHeaders = append(shardData.ShardMiniBlockHeaders, shardMiniBlockHeader)
		}

		shardInfo = append(shardInfo, shardData)
	}
	mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	log.Debug("created shard data",
		"size", len(shardInfo),
	)
	return shardInfo, nil
}

func (mp *metaProcessor) createPeerInfo() ([]block.PeerData, error) {
	peerInfo := mp.peerChanges.PeerChanges()

	return peerInfo, nil
}

// ApplyBodyToHeader creates a miniblock header list given a block body
func (mp *metaProcessor) ApplyBodyToHeader(hdr data.HeaderHandler, bodyHandler data.BodyHandler) (data.BodyHandler, error) {
	sw := core.NewStopWatch()
	sw.Start("ApplyBodyToHeader")
	defer func() {
		sw.Stop("ApplyBodyToHeader")

		log.Debug("measurements", sw.GetMeasurements()...)
	}()

	metaHdr, ok := hdr.(*block.MetaBlock)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	var err error
	defer func() {
		go mp.checkAndRequestIfShardHeadersMissing(hdr.GetRound())

		if err == nil {
			mp.blockSizeThrottler.Add(
				hdr.GetRound(),
				core.MaxUint32(hdr.ItemsInBody(), hdr.ItemsInHeader()))
		}
	}()

	sw.Start("createShardInfo")
	shardInfo, err := mp.createShardInfo(hdr.GetRound())
	sw.Stop("createShardInfo")
	if err != nil {
		return nil, err
	}

	sw.Start("createPeerInfo")
	peerInfo, err := mp.createPeerInfo()
	sw.Stop("createPeerInfo")
	if err != nil {
		return nil, err
	}

	metaHdr.Epoch = mp.epochStartTrigger.Epoch()
	metaHdr.ShardInfo = shardInfo
	metaHdr.PeerInfo = peerInfo
	metaHdr.RootHash = mp.getRootHash()
	metaHdr.TxCount = getTxCount(shardInfo)

	if check.IfNil(bodyHandler) {
		return nil, process.ErrNilBlockBody
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return nil, err
	}

	sw.Start("CreateReceiptsHash")
	metaHdr.ReceiptsHash, err = mp.txCoordinator.CreateReceiptsHash()
	sw.Stop("CreateReceiptsHash")
	if err != nil {
		return nil, err
	}

	totalTxCount, miniBlockHeaders, err := mp.createMiniBlockHeaders(body)
	if err != nil {
		return nil, err
	}

	metaHdr.MiniBlockHeaders = miniBlockHeaders
	metaHdr.TxCount += uint32(totalTxCount)

	sw.Start("UpdatePeerState")
	metaHdr.ValidatorStatsRootHash, err = mp.validatorStatisticsProcessor.UpdatePeerState(metaHdr)
	sw.Stop("UpdatePeerState")
	if err != nil {
		return nil, err
	}

	sw.Start("createEpochStartForMetablock")
	epochStart, err := mp.epochStartCreator.CreateEpochStartData()
	sw.Stop("createEpochStartForMetablock")
	if err != nil {
		return nil, err
	}
	metaHdr.EpochStart = *epochStart

	mp.blockSizeThrottler.Add(
		metaHdr.GetRound(),
		core.MaxUint32(metaHdr.ItemsInBody(), metaHdr.ItemsInHeader()))

	return body, nil
}

func (mp *metaProcessor) verifyValidatorStatisticsRootHash(header *block.MetaBlock) error {
	validatorStatsRH, err := mp.validatorStatisticsProcessor.UpdatePeerState(header)
	if err != nil {
		return err
	}

	if !bytes.Equal(validatorStatsRH, header.GetValidatorStatsRootHash()) {
		log.Debug("validator stats root hash mismatch",
			"computed", validatorStatsRH,
			"received", header.GetValidatorStatsRootHash(),
		)
		return fmt.Errorf("%s, metachain, computed: %s, received: %s, meta header nonce: %d",
			process.ErrValidatorStatsRootHashDoesNotMatch,
			display.DisplayByteSlice(validatorStatsRH),
			display.DisplayByteSlice(header.GetValidatorStatsRootHash()),
			header.Nonce,
		)
	}

	return nil
}

func (mp *metaProcessor) waitForBlockHeaders(waitTime time.Duration) error {
	select {
	case <-mp.chRcvAllHdrs:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
}

// CreateNewHeader creates a new header
func (mp *metaProcessor) CreateNewHeader() data.HeaderHandler {
	return &block.MetaBlock{}
}

// MarshalizedDataToBroadcast prepares underlying data into a marshalized object according to destination
func (mp *metaProcessor) MarshalizedDataToBroadcast(
	_ data.HeaderHandler,
	bodyHandler data.BodyHandler,
) (map[uint32][]byte, map[string][][]byte, error) {

	if bodyHandler == nil || bodyHandler.IsInterfaceNil() {
		return nil, nil, process.ErrNilMiniBlocks
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	bodies, mrsTxs := mp.txCoordinator.CreateMarshalizedData(body)
	mrsData := make(map[uint32][]byte, len(bodies))

	for shardId, subsetBlockBody := range bodies {
		buff, err := mp.marshalizer.Marshal(&block.Body{MiniBlocks: subsetBlockBody})
		if err != nil {
			log.Debug(process.ErrMarshalWithoutSuccess.Error())
			continue
		}
		mrsData[shardId] = buff
	}

	return mrsData, mrsTxs, nil
}

func getTxCount(shardInfo []block.ShardData) uint32 {
	txs := uint32(0)
	for i := 0; i < len(shardInfo); i++ {
		for j := 0; j < len(shardInfo[i].ShardMiniBlockHeaders); j++ {
			txs += shardInfo[i].ShardMiniBlockHeaders[j].TxCount
		}
	}

	return txs
}

// IsInterfaceNil returns true if there is no value under the interface
func (mp *metaProcessor) IsInterfaceNil() bool {
	return mp == nil
}

// GetBlockBodyFromPool returns block body from pool for a given header
func (mp *metaProcessor) GetBlockBodyFromPool(headerHandler data.HeaderHandler) (data.BodyHandler, error) {
	miniBlockPool := mp.dataPool.MiniBlocks()
	if miniBlockPool == nil {
		return nil, process.ErrNilMiniBlockPool
	}

	metaBlock, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	miniBlocks := make(block.MiniBlockSlice, 0)
	for i := 0; i < len(metaBlock.MiniBlockHeaders); i++ {
		obj, hashInPool := miniBlockPool.Get(metaBlock.MiniBlockHeaders[i].Hash)
		if !hashInPool {
			continue
		}

		miniBlock, typeOk := obj.(*block.MiniBlock)
		if !typeOk {
			return nil, process.ErrWrongTypeAssertion
		}

		miniBlocks = append(miniBlocks, miniBlock)
	}

	return &block.Body{MiniBlocks: miniBlocks}, nil
}

func (mp *metaProcessor) getPendingMiniBlocks() []bootstrapStorage.PendingMiniBlockInfo {
	pendingMiniBlocks := make([]bootstrapStorage.PendingMiniBlockInfo, mp.shardCoordinator.NumberOfShards())

	for shardID := uint32(0); shardID < mp.shardCoordinator.NumberOfShards(); shardID++ {
		pendingMiniBlocks[shardID] = bootstrapStorage.PendingMiniBlockInfo{
			NumPendingMiniBlocks: mp.pendingMiniBlocksHandler.GetNumPendingMiniBlocks(shardID),
			ShardID:              shardID,
		}
	}

	return pendingMiniBlocks
}
