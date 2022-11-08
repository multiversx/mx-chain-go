package block

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/queue"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/headerVersionData"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/process/block/processedMb"
	"github.com/ElrondNetwork/elrond-go/state"
)

const firstHeaderNonce = uint64(1)

var _ process.BlockProcessor = (*metaProcessor)(nil)

// metaProcessor implements metaProcessor interface, and actually it tries to execute block
type metaProcessor struct {
	*baseProcessor
	scToProtocol                 process.SmartContractToProtocolHandler
	epochStartDataCreator        process.EpochStartDataCreator
	epochEconomics               process.EndOfEpochEconomics
	epochRewardsCreator          process.RewardsCreator
	validatorInfoCreator         process.EpochStartValidatorInfoCreator
	epochSystemSCProcessor       process.EpochStartSystemSCProcessor
	pendingMiniBlocksHandler     process.PendingMiniBlocksHandler
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor
	shardsHeadersNonce           *sync.Map
	shardBlockFinality           uint32
	chRcvAllHdrs                 chan bool
	headersCounter               *headersCounter
	userStatePruningQueue        core.Queue
	peerStatePruningQueue        core.Queue
	processStatusHandler         common.ProcessStatusHandler
}

// NewMetaProcessor creates a new metaProcessor object
func NewMetaProcessor(arguments ArgMetaProcessor) (*metaProcessor, error) {
	err := checkProcessorNilParameters(arguments.ArgBaseProcessor)
	if err != nil {
		return nil, err
	}
	if check.IfNil(arguments.DataComponents.Datapool()) {
		return nil, process.ErrNilDataPoolHolder
	}
	if check.IfNil(arguments.DataComponents.Datapool().Headers()) {
		return nil, process.ErrNilHeadersDataPool
	}
	if check.IfNil(arguments.SCToProtocol) {
		return nil, process.ErrNilSCToProtocol
	}
	if check.IfNil(arguments.PendingMiniBlocksHandler) {
		return nil, process.ErrNilPendingMiniBlocksHandler
	}
	if check.IfNil(arguments.EpochStartDataCreator) {
		return nil, process.ErrNilEpochStartDataCreator
	}
	if check.IfNil(arguments.EpochEconomics) {
		return nil, process.ErrNilEpochEconomics
	}
	if check.IfNil(arguments.EpochRewardsCreator) {
		return nil, process.ErrNilRewardsCreator
	}
	if check.IfNil(arguments.EpochValidatorInfoCreator) {
		return nil, process.ErrNilEpochStartValidatorInfoCreator
	}
	if check.IfNil(arguments.ValidatorStatisticsProcessor) {
		return nil, process.ErrNilValidatorStatistics
	}
	if check.IfNil(arguments.EpochSystemSCProcessor) {
		return nil, process.ErrNilEpochStartSystemSCProcessor
	}
	if check.IfNil(arguments.ReceiptsRepository) {
		return nil, process.ErrNilReceiptsRepository
	}

	pruningQueueSize := arguments.Config.StateTriesConfig.PeerStatePruningQueueSize
	pruningDelay := uint32(pruningQueueSize * pruningDelayMultiplier)
	if pruningDelay < defaultPruningDelay {
		log.Warn("using default pruning delay", "pruning queue size", pruningQueueSize)
		pruningDelay = defaultPruningDelay
	}

	processDebugger, err := createDisabledProcessDebugger()
	if err != nil {
		return nil, err
	}

	genesisHdr := arguments.DataComponents.Blockchain().GetGenesisHeader()
	base := &baseProcessor{
		accountsDB:                    arguments.AccountsDB,
		blockSizeThrottler:            arguments.BlockSizeThrottler,
		forkDetector:                  arguments.ForkDetector,
		hasher:                        arguments.CoreComponents.Hasher(),
		marshalizer:                   arguments.CoreComponents.InternalMarshalizer(),
		store:                         arguments.DataComponents.StorageService(),
		shardCoordinator:              arguments.BootstrapComponents.ShardCoordinator(),
		feeHandler:                    arguments.FeeHandler,
		nodesCoordinator:              arguments.NodesCoordinator,
		uint64Converter:               arguments.CoreComponents.Uint64ByteSliceConverter(),
		requestHandler:                arguments.RequestHandler,
		appStatusHandler:              arguments.StatusCoreComponents.AppStatusHandler(),
		blockChainHook:                arguments.BlockChainHook,
		txCoordinator:                 arguments.TxCoordinator,
		epochStartTrigger:             arguments.EpochStartTrigger,
		headerValidator:               arguments.HeaderValidator,
		roundHandler:                  arguments.CoreComponents.RoundHandler(),
		bootStorer:                    arguments.BootStorer,
		blockTracker:                  arguments.BlockTracker,
		dataPool:                      arguments.DataComponents.Datapool(),
		blockChain:                    arguments.DataComponents.Blockchain(),
		stateCheckpointModulus:        arguments.Config.StateTriesConfig.CheckpointRoundsModulus,
		outportHandler:                arguments.StatusComponents.OutportHandler(),
		genesisNonce:                  genesisHdr.GetNonce(),
		versionedHeaderFactory:        arguments.BootstrapComponents.VersionedHeaderFactory(),
		headerIntegrityVerifier:       arguments.BootstrapComponents.HeaderIntegrityVerifier(),
		historyRepo:                   arguments.HistoryRepository,
		epochNotifier:                 arguments.CoreComponents.EpochNotifier(),
		enableEpochsHandler:           arguments.CoreComponents.EnableEpochsHandler(),
		enableRoundsHandler:           arguments.EnableRoundsHandler,
		vmContainerFactory:            arguments.VMContainersFactory,
		vmContainer:                   arguments.VmContainer,
		processDataTriesOnCommitEpoch: arguments.Config.Debug.EpochStart.ProcessDataTrieOnCommitEpoch,
		gasConsumedProvider:           arguments.GasHandler,
		economicsData:                 arguments.CoreComponents.EconomicsData(),
		scheduledTxsExecutionHandler:  arguments.ScheduledTxsExecutionHandler,
		pruningDelay:                  pruningDelay,
		processedMiniBlocksTracker:    arguments.ProcessedMiniBlocksTracker,
		receiptsRepository:            arguments.ReceiptsRepository,
		processDebugger:               processDebugger,
	}

	mp := metaProcessor{
		baseProcessor:                base,
		headersCounter:               NewHeaderCounter(),
		scToProtocol:                 arguments.SCToProtocol,
		pendingMiniBlocksHandler:     arguments.PendingMiniBlocksHandler,
		epochStartDataCreator:        arguments.EpochStartDataCreator,
		epochEconomics:               arguments.EpochEconomics,
		epochRewardsCreator:          arguments.EpochRewardsCreator,
		validatorStatisticsProcessor: arguments.ValidatorStatisticsProcessor,
		validatorInfoCreator:         arguments.EpochValidatorInfoCreator,
		epochSystemSCProcessor:       arguments.EpochSystemSCProcessor,
		processStatusHandler:         arguments.CoreComponents.ProcessStatusHandler(),
	}

	mp.txCounter, err = NewTransactionCounter(mp.hasher, mp.marshalizer)
	if err != nil {
		return nil, err
	}

	mp.requestBlockBodyHandler = &mp
	mp.blockProcessor = &mp

	mp.hdrsForCurrBlock = newHdrForBlock()

	headersPool := mp.dataPool.Headers()
	headersPool.RegisterHandler(mp.receivedShardHeader)

	mp.chRcvAllHdrs = make(chan bool)

	mp.shardBlockFinality = process.BlockFinality

	mp.shardsHeadersNonce = &sync.Map{}
	mp.userStatePruningQueue = queue.NewSliceQueue(arguments.Config.StateTriesConfig.UserStatePruningQueueSize)
	mp.peerStatePruningQueue = queue.NewSliceQueue(arguments.Config.StateTriesConfig.PeerStatePruningQueueSize)

	return &mp, nil
}

func (mp *metaProcessor) isRewardsV2Enabled(headerHandler data.HeaderHandler) bool {
	return headerHandler.GetEpoch() >= mp.enableEpochsHandler.StakingV2EnableEpoch()
}

// ProcessBlock processes a block. It returns nil if all ok or the specific error
func (mp *metaProcessor) ProcessBlock(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
	haveTime func() time.Duration,
) error {
	if haveTime == nil {
		return process.ErrNilHaveTimeHandler
	}

	mp.processStatusHandler.SetBusy("metaProcessor.ProcessBlock")
	defer mp.processStatusHandler.SetIdle()

	err := mp.checkBlockValidity(headerHandler, bodyHandler)
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

	mp.enableRoundsHandler.CheckRound(headerHandler.GetRound())
	mp.epochNotifier.CheckEpoch(headerHandler)
	mp.requestHandler.SetEpoch(headerHandler.GetEpoch())

	log.Debug("started processing block",
		"epoch", headerHandler.GetEpoch(),
		"shard", headerHandler.GetShardID(),
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

	err = mp.checkHeaderBodyCorrelation(header.GetMiniBlockHeaderHandlers(), body)
	if err != nil {
		return err
	}

	err = mp.checkScheduledMiniBlocksValidity(headerHandler)
	if err != nil {
		return err
	}

	headersPool := mp.dataPool.Headers()
	numShardHeadersFromPool := 0
	for shardID := uint32(0); shardID < mp.shardCoordinator.NumberOfShards(); shardID++ {
		numShardHeadersFromPool += headersPool.GetNumHeaders(shardID)
	}

	txCounts, rewardCounts, unsignedCounts := mp.txCounter.getPoolCounts(mp.dataPool)
	log.Debug("total txs in pool", "counts", txCounts.String())
	log.Debug("total txs in rewards pool", "counts", rewardCounts.String())
	log.Debug("total txs in unsigned pool", "counts", unsignedCounts.String())

	go getMetricsFromMetaHeader(
		header,
		mp.marshalizer,
		mp.appStatusHandler,
		numShardHeadersFromPool,
		mp.headersCounter.getNumShardMBHeadersTotalProcessed(),
	)

	defer func() {
		if err != nil {
			mp.RevertCurrentBlock()
		}
	}()

	err = mp.createBlockStarted()
	if err != nil {
		return err
	}

	mp.blockChainHook.SetCurrentHeader(header)
	mp.epochStartTrigger.Update(header.GetRound(), header.GetNonce())

	err = mp.checkEpochCorrectness(header)
	if err != nil {
		return err
	}

	if mp.accountsDB[state.UserAccountsState].JournalLen() != 0 {
		log.Error("metaProcessor.ProcessBlock first entry", "stack", string(mp.accountsDB[state.UserAccountsState].GetStackDebugFirstEntry()))
		return process.ErrAccountStateDirty
	}

	err = mp.processIfFirstBlockAfterEpochStart()
	if err != nil {
		return err
	}

	if header.IsStartOfEpochBlock() {
		err = mp.processEpochStartMetaBlock(header, body)
		return err
	}

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
		if requestedShardHdrs > 0 {
			log.Debug("requested missing shard headers",
				"num headers", requestedShardHdrs,
			)
		}
		if requestedFinalityAttestingShardHdrs > 0 {
			log.Debug("requested missing finality attesting shard headers",
				"num finality shard headers", requestedFinalityAttestingShardHdrs,
			)
		}

		err = mp.waitForBlockHeaders(haveTime())

		mp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
		missingShardHdrs := mp.hdrsForCurrBlock.missingHdrs
		mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

		mp.hdrsForCurrBlock.resetMissingHdrs()

		if requestedShardHdrs > 0 {
			log.Debug("received missing shard headers",
				"num headers", requestedShardHdrs-missingShardHdrs,
			)
		}

		if err != nil {
			return err
		}
	}

	defer func() {
		go mp.checkAndRequestIfShardHeadersMissing()
	}()

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

	err = mp.verifyTotalAccumulatedFeesInEpoch(header)
	if err != nil {
		return err
	}

	mbIndex := mp.getIndexOfFirstMiniBlockToBeExecuted(header)
	miniBlocks := body.MiniBlocks[mbIndex:]

	startTime := time.Now()
	err = mp.txCoordinator.ProcessBlockTransaction(header, &block.Body{MiniBlocks: miniBlocks}, haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to process block transaction",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	err = mp.txCoordinator.VerifyCreatedBlockTransactions(header, &block.Body{MiniBlocks: miniBlocks})
	if err != nil {
		return err
	}

	err = mp.scToProtocol.UpdateProtocol(&block.Body{MiniBlocks: miniBlocks}, header.Nonce)
	if err != nil {
		return err
	}

	err = mp.verifyFees(header)
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

func (mp *metaProcessor) processEpochStartMetaBlock(
	header *block.MetaBlock,
	body *block.Body,
) error {
	err := mp.epochStartDataCreator.VerifyEpochStartDataForMetablock(header)
	if err != nil {
		return err
	}

	currentRootHash, err := mp.validatorStatisticsProcessor.RootHash()
	if err != nil {
		return err
	}

	allValidatorsInfo, err := mp.validatorStatisticsProcessor.GetValidatorInfoForRootHash(currentRootHash)
	if err != nil {
		return err
	}

	err = mp.validatorStatisticsProcessor.ProcessRatingsEndOfEpoch(allValidatorsInfo, header.Epoch)
	if err != nil {
		return err
	}

	computedEconomics, err := mp.epochEconomics.ComputeEndOfEpochEconomics(header)
	if err != nil {
		return err
	}

	if mp.isRewardsV2Enabled(header) {
		err = mp.epochSystemSCProcessor.ProcessSystemSmartContract(allValidatorsInfo, header.Nonce, header.Epoch)
		if err != nil {
			return err
		}

		err = mp.epochRewardsCreator.VerifyRewardsMiniBlocks(header, allValidatorsInfo, computedEconomics)
		if err != nil {
			return err
		}
	} else {
		err = mp.epochRewardsCreator.VerifyRewardsMiniBlocks(header, allValidatorsInfo, computedEconomics)
		if err != nil {
			return err
		}

		err = mp.epochSystemSCProcessor.ProcessSystemSmartContract(allValidatorsInfo, header.Nonce, header.Epoch)
		if err != nil {
			return err
		}
	}

	err = mp.epochSystemSCProcessor.ProcessDelegationRewards(body.MiniBlocks, mp.epochRewardsCreator.GetLocalTxCache())
	if err != nil {
		return err
	}

	err = mp.validatorInfoCreator.VerifyValidatorInfoMiniBlocks(body.MiniBlocks, allValidatorsInfo)
	if err != nil {
		return err
	}

	err = mp.validatorStatisticsProcessor.ResetValidatorStatisticsAtNewEpoch(allValidatorsInfo)
	if err != nil {
		return err
	}

	err = mp.epochEconomics.VerifyRewardsPerBlock(header, mp.epochRewardsCreator.GetProtocolSustainabilityRewards(), computedEconomics)
	if err != nil {
		return err
	}

	err = mp.verifyFees(header)
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

	saveEpochStartEconomicsMetrics(mp.appStatusHandler, header)

	return nil
}

// SetNumProcessedObj will set the num of processed headers
func (mp *metaProcessor) SetNumProcessedObj(numObj uint64) {
	mp.headersCounter.shardMBHeadersTotalProcessed = numObj
}

func (mp *metaProcessor) checkEpochCorrectness(
	headerHandler data.HeaderHandler,
) error {
	currentBlockHeader := mp.blockChain.GetCurrentBlockHeader()
	if check.IfNil(currentBlockHeader) {
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

func (mp *metaProcessor) verifyCrossShardMiniBlockDstMe(metaBlock *block.MetaBlock) error {
	miniBlockShardsHashes, err := mp.getAllMiniBlockDstMeFromShards(metaBlock)
	if err != nil {
		return err
	}

	mapMetaMiniBlockHeaders := make(map[string]struct{}, len(metaBlock.MiniBlockHeaders))
	for _, miniBlockHeader := range metaBlock.MiniBlockHeaders {
		mapMetaMiniBlockHeaders[string(miniBlockHeader.Hash)] = struct{}{}
	}

	for hash := range miniBlockShardsHashes {
		if _, ok := mapMetaMiniBlockHeaders[hash]; !ok {
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
		shardHeader, ok := headerInfo.hdr.(data.ShardHeaderHandler)
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

		finalCrossMiniBlockHashes := mp.getFinalCrossMiniBlockHashes(shardHeader)
		for hash := range finalCrossMiniBlockHashes {
			miniBlockShardsHashes[hash] = shardInfo.HeaderHash
		}
	}

	return miniBlockShardsHashes, nil
}

func (mp *metaProcessor) getFinalCrossMiniBlockHashes(headerHandler data.HeaderHandler) map[string]uint32 {
	if !mp.enableEpochsHandler.IsScheduledMiniBlocksFlagEnabled() {
		return headerHandler.GetMiniBlockHeadersWithDst(mp.shardCoordinator.SelfId())
	}
	return process.GetFinalCrossMiniBlockHashes(headerHandler, mp.shardCoordinator.SelfId())
}

func (mp *metaProcessor) checkAndRequestIfShardHeadersMissing() {
	orderedHdrsPerShard := mp.blockTracker.GetTrackedHeadersForAllShards()

	for i := uint32(0); i < mp.shardCoordinator.NumberOfShards(); i++ {
		err := mp.requestHeadersIfMissing(orderedHdrsPerShard[i], i)
		if err != nil {
			log.Debug("checkAndRequestIfShardHeadersMissing", "error", err.Error())
			continue
		}
	}
}

func (mp *metaProcessor) indexBlock(
	metaBlock data.HeaderHandler,
	headerHash []byte,
	body data.BodyHandler,
	lastMetaBlock data.HeaderHandler,
	notarizedHeadersHashes []string,
	rewardsTxs map[string]data.TransactionHandler,
) {
	if !mp.outportHandler.HasDrivers() {
		return
	}

	log.Debug("preparing to index block", "hash", headerHash, "nonce", metaBlock.GetNonce(), "round", metaBlock.GetRound())

	pool := &indexer.Pool{
		Txs:     mp.txCoordinator.GetAllCurrentUsedTxs(block.TxBlock),
		Scrs:    mp.txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock),
		Rewards: rewardsTxs,
		Logs:    mp.txCoordinator.GetAllCurrentLogs(),
	}

	publicKeys, err := mp.nodesCoordinator.GetConsensusValidatorsPublicKeys(
		metaBlock.GetPrevRandSeed(), metaBlock.GetRound(), core.MetachainShardId, metaBlock.GetEpoch(),
	)
	if err != nil {
		log.Debug("indexBlock: GetConsensusValidatorsPublicKeys",
			"hash", headerHash,
			"epoch", metaBlock.GetEpoch(),
			"error", err.Error())
		return
	}

	epoch := metaBlock.GetEpoch()
	shardCoordinatorShardID := mp.shardCoordinator.SelfId()
	nodesCoordinatorShardID, err := mp.nodesCoordinator.ShardIdForEpoch(epoch)
	if err != nil {
		log.Debug("indexBlock",
			"epoch", epoch,
			"error", err.Error())
		return
	}

	if shardCoordinatorShardID != nodesCoordinatorShardID {
		log.Debug("indexBlock",
			"epoch", epoch,
			"shardCoordinator.ShardID", shardCoordinatorShardID,
			"nodesCoordinator.ShardID", nodesCoordinatorShardID)
		return
	}

	signersIndexes, err := mp.nodesCoordinator.GetValidatorsIndexes(publicKeys, epoch)
	if err != nil {
		log.Debug("indexBlock: GetValidatorsIndexes",
			"hash", headerHash,
			"epoch", metaBlock.GetEpoch(),
			"error", err.Error())
		return
	}

	gasProvidedInHeader := mp.baseProcessor.gasConsumedProvider.TotalGasProvidedWithScheduled()
	gasPenalizedInHeader := mp.baseProcessor.gasConsumedProvider.TotalGasPenalized()
	gasRefundedInHeader := mp.baseProcessor.gasConsumedProvider.TotalGasRefunded()
	maxGasInHeader := mp.baseProcessor.economicsData.MaxGasLimitPerBlock(mp.shardCoordinator.SelfId())

	args := &indexer.ArgsSaveBlockData{
		HeaderHash:     headerHash,
		Body:           body,
		Header:         metaBlock,
		SignersIndexes: signersIndexes,
		HeaderGasConsumption: indexer.HeaderGasConsumption{
			GasProvided:    gasProvidedInHeader,
			GasRefunded:    gasRefundedInHeader,
			GasPenalized:   gasPenalizedInHeader,
			MaxGasPerBlock: maxGasInHeader,
		},
		NotarizedHeadersHashes: notarizedHeadersHashes,
		TransactionsPool:       pool,
	}
	mp.outportHandler.SaveBlock(args)
	log.Debug("indexed block", "hash", headerHash, "nonce", metaBlock.GetNonce(), "round", metaBlock.GetRound())

	indexRoundInfo(mp.outportHandler, mp.nodesCoordinator, core.MetachainShardId, metaBlock, lastMetaBlock, signersIndexes)

	if metaBlock.GetNonce() != 1 && !metaBlock.IsStartOfEpochBlock() {
		return
	}

	indexValidatorsRating(mp.outportHandler, mp.validatorStatisticsProcessor, metaBlock)
}

// RestoreBlockIntoPools restores the block into associated pools
func (mp *metaProcessor) RestoreBlockIntoPools(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error {
	if check.IfNil(headerHandler) {
		return process.ErrNilMetaBlockHeader
	}

	metaBlock, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	hdrHashes := make([][]byte, len(metaBlock.ShardInfo))
	for i := 0; i < len(metaBlock.ShardInfo); i++ {
		hdrHashes[i] = metaBlock.ShardInfo[i].HeaderHash
	}

	err := mp.pendingMiniBlocksHandler.RevertHeader(metaBlock)
	if err != nil {
		return err
	}

	headersPool := mp.dataPool.Headers()

	for _, hdrHash := range hdrHashes {
		shardHeader, errNotCritical := process.GetShardHeaderFromStorage(hdrHash, mp.marshalizer, mp.store)
		if errNotCritical != nil {
			log.Debug("shard header not found in BlockHeaderUnit",
				"hash", hdrHash,
			)
			continue
		}

		headersPool.AddHeader(hdrHash, shardHeader)

		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardHeader.GetShardID())
		storer, errNotCritical := mp.store.GetStorer(hdrNonceHashDataUnit)
		if err != nil {
			log.Debug("storage unit not found", "unit", hdrNonceHashDataUnit, "error", errNotCritical.Error())
			continue
		}

		nonceToByteSlice := mp.uint64Converter.ToByteSlice(shardHeader.GetNonce())
		errNotCritical = storer.Remove(nonceToByteSlice)
		if errNotCritical != nil {
			log.Debug("ShardHdrNonceHashDataUnit.Remove", "error", errNotCritical.Error())
		}

		mp.headersCounter.subtractRestoredMBHeaders(len(shardHeader.GetMiniBlockHeaderHandlers()))
	}

	mp.restoreBlockBody(bodyHandler)

	mp.blockTracker.RemoveLastNotarizedHeaders()

	return nil
}

// CreateBlock creates the final block and header for the current round
func (mp *metaProcessor) CreateBlock(
	initialHdr data.HeaderHandler,
	haveTime func() bool,
) (data.HeaderHandler, data.BodyHandler, error) {
	if check.IfNil(initialHdr) {
		return nil, nil, process.ErrNilBlockHeader
	}

	metaHdr, ok := initialHdr.(*block.MetaBlock)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	mp.processStatusHandler.SetBusy("metaProcessor.CreateBlock")
	defer mp.processStatusHandler.SetIdle()

	metaHdr.SoftwareVersion = []byte(mp.headerIntegrityVerifier.GetVersion(metaHdr.Epoch))
	mp.epochNotifier.CheckEpoch(metaHdr)
	mp.blockChainHook.SetCurrentHeader(initialHdr)

	var body data.BodyHandler

	if mp.accountsDB[state.UserAccountsState].JournalLen() != 0 {
		log.Error("metaProcessor.CreateBlock first entry", "stack", string(mp.accountsDB[state.UserAccountsState].GetStackDebugFirstEntry()))
		return nil, nil, process.ErrAccountStateDirty
	}

	err := mp.processIfFirstBlockAfterEpochStart()
	if err != nil {
		return nil, nil, err
	}

	if mp.epochStartTrigger.IsEpochStart() {
		err = mp.updateEpochStartHeader(metaHdr)
		if err != nil {
			return nil, nil, err
		}

		body, err = mp.createEpochStartBody(metaHdr)
		if err != nil {
			return nil, nil, err
		}
	} else {
		body, err = mp.createBlockBody(metaHdr, haveTime)
		if err != nil {
			return nil, nil, err
		}
	}

	body, err = mp.applyBodyToHeader(metaHdr, body)
	if err != nil {
		return nil, nil, err
	}

	mp.requestHandler.SetEpoch(metaHdr.GetEpoch())

	return metaHdr, body, nil
}

func (mp *metaProcessor) isPreviousBlockEpochStart() (uint32, bool) {
	blockHeader := mp.blockChain.GetCurrentBlockHeader()
	if check.IfNil(blockHeader) {
		blockHeader = mp.blockChain.GetGenesisHeader()
	}

	return blockHeader.GetEpoch(), blockHeader.IsStartOfEpochBlock()
}

func (mp *metaProcessor) processIfFirstBlockAfterEpochStart() error {
	epoch, isPreviousEpochStart := mp.isPreviousBlockEpochStart()
	if !isPreviousEpochStart {
		return nil
	}

	nodesForcedToStay, err := mp.validatorStatisticsProcessor.SaveNodesCoordinatorUpdates(epoch)
	if err != nil {
		return err
	}

	err = mp.epochSystemSCProcessor.ToggleUnStakeUnBond(nodesForcedToStay)
	if err != nil {
		return err
	}

	return nil
}

func (mp *metaProcessor) updateEpochStartHeader(metaHdr *block.MetaBlock) error {
	sw := core.NewStopWatch()
	sw.Start("createEpochStartForMetablock")
	defer func() {
		sw.Stop("createEpochStartForMetablock")
		log.Debug("epochStartHeaderDataCreation", sw.GetMeasurements()...)
	}()

	epochStart, err := mp.epochStartDataCreator.CreateEpochStartData()
	if err != nil {
		return err
	}

	metaHdr.EpochStart = *epochStart

	totalAccumulatedFeesInEpoch := big.NewInt(0)
	totalDevFeesInEpoch := big.NewInt(0)
	currentHeader := mp.blockChain.GetCurrentBlockHeader()
	if !check.IfNil(currentHeader) && !currentHeader.IsStartOfEpochBlock() {
		prevMetaHdr, ok := currentHeader.(*block.MetaBlock)
		if !ok {
			return process.ErrWrongTypeAssertion
		}
		totalAccumulatedFeesInEpoch = big.NewInt(0).Set(prevMetaHdr.AccumulatedFeesInEpoch)
		totalDevFeesInEpoch = big.NewInt(0).Set(prevMetaHdr.DevFeesInEpoch)
	}

	metaHdr.AccumulatedFeesInEpoch.Set(totalAccumulatedFeesInEpoch)
	metaHdr.DevFeesInEpoch.Set(totalDevFeesInEpoch)
	economicsData, err := mp.epochEconomics.ComputeEndOfEpochEconomics(metaHdr)
	if err != nil {
		return err
	}

	metaHdr.EpochStart.Economics = *economicsData

	saveEpochStartEconomicsMetrics(mp.appStatusHandler, metaHdr)

	return nil
}

func (mp *metaProcessor) createEpochStartBody(metaBlock *block.MetaBlock) (data.BodyHandler, error) {
	err := mp.createBlockStarted()
	if err != nil {
		return nil, err
	}

	log.Debug("started creating epoch start block body",
		"epoch", metaBlock.GetEpoch(),
		"round", metaBlock.GetRound(),
		"nonce", metaBlock.GetNonce(),
	)

	currentRootHash, err := mp.validatorStatisticsProcessor.RootHash()
	if err != nil {
		return nil, err
	}

	allValidatorsInfo, err := mp.validatorStatisticsProcessor.GetValidatorInfoForRootHash(currentRootHash)
	if err != nil {
		return nil, err
	}

	err = mp.validatorStatisticsProcessor.ProcessRatingsEndOfEpoch(allValidatorsInfo, metaBlock.Epoch)
	if err != nil {
		return nil, err
	}

	var rewardMiniBlocks block.MiniBlockSlice
	if mp.isRewardsV2Enabled(metaBlock) {
		err = mp.epochSystemSCProcessor.ProcessSystemSmartContract(allValidatorsInfo, metaBlock.Nonce, metaBlock.Epoch)
		if err != nil {
			return nil, err
		}

		rewardMiniBlocks, err = mp.epochRewardsCreator.CreateRewardsMiniBlocks(metaBlock, allValidatorsInfo, &metaBlock.EpochStart.Economics)
		if err != nil {
			return nil, err
		}
	} else {
		rewardMiniBlocks, err = mp.epochRewardsCreator.CreateRewardsMiniBlocks(metaBlock, allValidatorsInfo, &metaBlock.EpochStart.Economics)
		if err != nil {
			return nil, err
		}

		err = mp.epochSystemSCProcessor.ProcessSystemSmartContract(allValidatorsInfo, metaBlock.Nonce, metaBlock.Epoch)
		if err != nil {
			return nil, err
		}
	}

	metaBlock.EpochStart.Economics.RewardsForProtocolSustainability.Set(mp.epochRewardsCreator.GetProtocolSustainabilityRewards())

	err = mp.epochSystemSCProcessor.ProcessDelegationRewards(rewardMiniBlocks, mp.epochRewardsCreator.GetLocalTxCache())
	if err != nil {
		return nil, err
	}

	validatorMiniBlocks, err := mp.validatorInfoCreator.CreateValidatorInfoMiniBlocks(allValidatorsInfo)
	if err != nil {
		return nil, err
	}

	err = mp.validatorStatisticsProcessor.ResetValidatorStatisticsAtNewEpoch(allValidatorsInfo)
	if err != nil {
		return nil, err
	}

	finalMiniBlocks := make([]*block.MiniBlock, 0)
	finalMiniBlocks = append(finalMiniBlocks, rewardMiniBlocks...)
	finalMiniBlocks = append(finalMiniBlocks, validatorMiniBlocks...)

	return &block.Body{MiniBlocks: finalMiniBlocks}, nil
}

// createBlockBody creates block body of metachain
func (mp *metaProcessor) createBlockBody(metaBlock data.HeaderHandler, haveTime func() bool) (data.BodyHandler, error) {
	err := mp.createBlockStarted()
	if err != nil {
		return nil, err
	}

	mp.blockSizeThrottler.ComputeCurrentMaxSize()

	log.Debug("started creating meta block body",
		"epoch", metaBlock.GetEpoch(),
		"round", metaBlock.GetRound(),
		"nonce", metaBlock.GetNonce(),
	)

	miniBlocks, err := mp.createMiniBlocks(haveTime, metaBlock.GetPrevRandSeed())
	if err != nil {
		return nil, err
	}

	err = mp.scToProtocol.UpdateProtocol(miniBlocks, metaBlock.GetNonce())
	if err != nil {
		return nil, err
	}

	return miniBlocks, nil
}

func (mp *metaProcessor) createMiniBlocks(
	haveTime func() bool,
	randomness []byte,
) (*block.Body, error) {
	var miniBlocks block.MiniBlockSlice

	if mp.enableEpochsHandler.IsScheduledMiniBlocksFlagEnabled() {
		miniBlocks = mp.scheduledTxsExecutionHandler.GetScheduledMiniBlocks()
		mp.txCoordinator.AddTxsFromMiniBlocks(miniBlocks)
		// TODO: in case we add metachain originating scheduled miniBlocks, we need to add the invalid txs here, same as for shard processor
	}

	if !haveTime() {
		log.Debug("metaProcessor.createMiniBlocks", "error", process.ErrTimeIsOut)

		interMBs := mp.txCoordinator.CreatePostProcessMiniBlocks()
		if len(interMBs) > 0 {
			miniBlocks = append(miniBlocks, interMBs...)
		}

		log.Debug("creating mini blocks has been finished", "num miniblocks", len(miniBlocks))
		return &block.Body{MiniBlocks: miniBlocks}, nil
	}

	mbsToMe, numTxs, numShardHeaders, err := mp.createAndProcessCrossMiniBlocksDstMe(haveTime)
	if err != nil {
		log.Debug("createAndProcessCrossMiniBlocksDstMe", "error", err.Error())
	}

	if len(mbsToMe) > 0 {
		miniBlocks = append(miniBlocks, mbsToMe...)

		log.Debug("processed miniblocks and txs with destination in self shard",
			"num miniblocks", len(mbsToMe),
			"num txs", numTxs,
			"num shard headers", numShardHeaders,
		)
	}

	mbsFromMe := mp.txCoordinator.CreateMbsAndProcessTransactionsFromMe(haveTime, randomness)
	if len(mbsFromMe) > 0 {
		miniBlocks = append(miniBlocks, mbsFromMe...)

		numTxs = 0
		for _, mb := range mbsFromMe {
			numTxs += uint32(len(mb.TxHashes))
		}

		log.Debug("processed miniblocks and txs from self shard",
			"num miniblocks", len(mbsFromMe),
			"num txs", numTxs,
		)
	}

	log.Debug("creating mini blocks has been finished",
		"miniblocks created", len(miniBlocks),
	)

	return &block.Body{MiniBlocks: miniBlocks}, nil
}

func (mp *metaProcessor) isGenesisShardBlockAndFirstMeta(shardHdrNonce uint64) bool {
	return shardHdrNonce == mp.genesisNonce && check.IfNil(mp.blockChain.GetCurrentBlockHeader())
}

// full verification through metachain header
func (mp *metaProcessor) createAndProcessCrossMiniBlocksDstMe(
	haveTime func() bool,
) (block.MiniBlockSlice, uint32, uint32, error) {

	var miniBlocks block.MiniBlockSlice
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

	lastShardHdr, err := mp.getLastCrossNotarizedShardHdrs()
	if err != nil {
		return nil, 0, 0, err
	}

	maxShardHeadersFromSameShard := core.MaxUint32(
		process.MinShardHeadersFromSameShardInOneMetaBlock,
		process.MaxShardHeadersAllowedInOneMetaBlock/mp.shardCoordinator.NumberOfShards(),
	)
	maxShardHeadersAllowedInOneMetaBlock := maxShardHeadersFromSameShard * mp.shardCoordinator.NumberOfShards()
	hdrsAddedForShard := make(map[uint32]uint32)
	haveAdditionalTimeFalse := func() bool {
		return false
	}

	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	for i := 0; i < len(orderedHdrs); i++ {
		if !haveTime() {
			log.Debug("time is up after putting cross txs with destination to current shard",
				"num txs", txsAdded,
			)
			break
		}

		if hdrsAdded >= maxShardHeadersAllowedInOneMetaBlock {
			log.Debug("maximum shard headers allowed to be included in one meta block has been reached",
				"shard headers added", hdrsAdded,
			)
			break
		}

		currShardHdr := orderedHdrs[i]
		if currShardHdr.GetNonce() > lastShardHdr[currShardHdr.GetShardID()].GetNonce()+1 {
			log.Trace("skip searching",
				"shard", currShardHdr.GetShardID(),
				"last shard hdr nonce", lastShardHdr[currShardHdr.GetShardID()].GetNonce(),
				"curr shard hdr nonce", currShardHdr.GetNonce())
			continue
		}

		if hdrsAddedForShard[currShardHdr.GetShardID()] >= maxShardHeadersFromSameShard {
			log.Trace("maximum shard headers from same shard allowed to be included in one meta block has been reached",
				"shard", currShardHdr.GetShardID(),
				"shard headers added", hdrsAddedForShard[currShardHdr.GetShardID()],
			)
			continue
		}

		if len(currShardHdr.GetMiniBlockHeadersWithDst(mp.shardCoordinator.SelfId())) == 0 {
			mp.hdrsForCurrBlock.hdrHashAndInfo[string(orderedHdrsHashes[i])] = &hdrInfo{hdr: currShardHdr, usedInBlock: true}
			hdrsAdded++
			hdrsAddedForShard[currShardHdr.GetShardID()]++
			lastShardHdr[currShardHdr.GetShardID()] = currShardHdr
			continue
		}

		snapshot := mp.accountsDB[state.UserAccountsState].JournalLen()
		currMBProcessed, currTxsAdded, hdrProcessFinished, createErr := mp.txCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe(
			currShardHdr,
			nil,
			haveTime,
			haveAdditionalTimeFalse,
			false)

		if createErr != nil {
			return nil, 0, 0, createErr
		}

		if !hdrProcessFinished {
			log.Debug("shard header cannot be fully processed",
				"round", currShardHdr.GetRound(),
				"nonce", currShardHdr.GetNonce(),
				"hash", orderedHdrsHashes[i])

			// shard header must be processed completely
			errAccountState := mp.accountsDB[state.UserAccountsState].RevertToSnapshot(snapshot)
			if errAccountState != nil {
				// TODO: evaluate if reloading the trie from disk will might solve the problem
				log.Warn("accounts.RevertToSnapshot", "error", errAccountState.Error())
			}
			continue
		}

		// all txs processed, add to processed miniblocks
		miniBlocks = append(miniBlocks, currMBProcessed...)
		txsAdded += currTxsAdded

		mp.hdrsForCurrBlock.hdrHashAndInfo[string(orderedHdrsHashes[i])] = &hdrInfo{hdr: currShardHdr, usedInBlock: true}
		hdrsAdded++
		hdrsAddedForShard[currShardHdr.GetShardID()]++

		lastShardHdr[currShardHdr.GetShardID()] = currShardHdr
	}
	mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	go mp.requestShardHeadersIfNeeded(hdrsAddedForShard, lastShardHdr)

	return miniBlocks, txsAdded, hdrsAdded, nil
}

func (mp *metaProcessor) requestShardHeadersIfNeeded(
	hdrsAddedForShard map[uint32]uint32,
	lastShardHdr map[uint32]data.HeaderHandler,
) {
	for shardID := uint32(0); shardID < mp.shardCoordinator.NumberOfShards(); shardID++ {
		log.Debug("shard headers added",
			"shard", shardID,
			"num", hdrsAddedForShard[shardID],
			"highest nonce", lastShardHdr[shardID].GetNonce())

		roundTooOld := mp.roundHandler.Index() > int64(lastShardHdr[shardID].GetRound()+process.MaxRoundsWithoutNewBlockReceived)
		shouldRequestCrossHeaders := hdrsAddedForShard[shardID] == 0 && roundTooOld
		if shouldRequestCrossHeaders {
			fromNonce := lastShardHdr[shardID].GetNonce() + 1
			toNonce := fromNonce + uint64(mp.shardBlockFinality)
			for nonce := fromNonce; nonce <= toNonce; nonce++ {
				mp.addHeaderIntoTrackerPool(nonce, shardID)
				mp.requestHandler.RequestShardHeaderByNonce(shardID, nonce)
			}
		}
	}
}

// CommitBlock commits the block in the blockchain if everything was checked successfully
func (mp *metaProcessor) CommitBlock(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {
	mp.processStatusHandler.SetBusy("metaProcessor.CommitBlock")
	var err error
	defer func() {
		if err != nil {
			mp.RevertCurrentBlock()
		}
		mp.processStatusHandler.SetIdle()
	}()

	err = checkForNils(headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	log.Debug("started committing block",
		"epoch", headerHandler.GetEpoch(),
		"shard", headerHandler.GetShardID(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
	)

	err = mp.checkBlockValidity(headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	mp.store.SetEpochForPutOperation(headerHandler.GetEpoch())

	header, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	marshalizedHeader, err := mp.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	// must be called before commitEpochStart
	rewardsTxs := mp.getRewardsTxs(header, body)

	mp.commitEpochStart(header, body)
	headerHash := mp.hasher.Compute(string(marshalizedHeader))
	mp.saveMetaHeader(header, headerHash, marshalizedHeader)
	mp.saveBody(body, header, headerHash)

	err = mp.commitAll(headerHandler)
	if err != nil {
		return err
	}

	mp.validatorStatisticsProcessor.DisplayRatings(header.GetEpoch())

	err = mp.saveLastNotarizedHeader(header)
	if err != nil {
		return err
	}

	err = mp.pendingMiniBlocksHandler.AddProcessedHeader(header)
	if err != nil {
		return err
	}

	log.Info("meta block has been committed successfully",
		"epoch", headerHandler.GetEpoch(),
		"shard", headerHandler.GetShardID(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
		"hash", headerHash)

	mp.updateLastCommittedInDebugger(headerHandler.GetRound())

	notarizedHeadersHashes, errNotCritical := mp.updateCrossShardInfo(header)
	if errNotCritical != nil {
		log.Debug("updateCrossShardInfo", "error", errNotCritical.Error())
	}

	errNotCritical = mp.forkDetector.AddHeader(header, headerHash, process.BHProcessed, nil, nil)
	if errNotCritical != nil {
		log.Debug("forkDetector.AddHeader", "error", errNotCritical.Error())
	}

	currentHeader, currentHeaderHash := getLastSelfNotarizedHeaderByItself(mp.blockChain)
	mp.blockTracker.AddSelfNotarizedHeader(mp.shardCoordinator.SelfId(), currentHeader, currentHeaderHash)

	for shardID := uint32(0); shardID < mp.shardCoordinator.NumberOfShards(); shardID++ {
		lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash := mp.getLastSelfNotarizedHeaderByShard(header, shardID)
		mp.blockTracker.AddSelfNotarizedHeader(shardID, lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash)
	}

	go mp.historyRepo.OnNotarizedBlocks(mp.shardCoordinator.SelfId(), []data.HeaderHandler{currentHeader}, [][]byte{currentHeaderHash})

	log.Debug("highest final meta block",
		"nonce", mp.forkDetector.GetHighestFinalBlockNonce(),
	)

	lastHeader := mp.blockChain.GetCurrentBlockHeader()
	lastMetaBlock, ok := lastHeader.(data.MetaHeaderHandler)
	if !ok {
		if headerHandler.GetNonce() == firstHeaderNonce {
			log.Debug("metaBlock.CommitBlock - nil current block header, this is expected at genesis time")
		} else {
			log.Error("metaBlock.CommitBlock - nil current block header, last current header should have not been nil")
		}
	}
	lastMetaBlockHash := mp.blockChain.GetCurrentBlockHeaderHash()
	if mp.lastRestartNonce == 0 {
		mp.lastRestartNonce = header.GetNonce()
	}

	mp.updateState(lastMetaBlock, lastMetaBlockHash)

	committedRootHash, err := mp.accountsDB[state.UserAccountsState].RootHash()
	if err != nil {
		return err
	}

	err = mp.blockChain.SetCurrentBlockHeaderAndRootHash(header, committedRootHash)
	if err != nil {
		return err
	}

	mp.blockChain.SetCurrentBlockHeaderHash(headerHash)

	if !check.IfNil(lastMetaBlock) && lastMetaBlock.IsStartOfEpochBlock() {
		mp.blockTracker.CleanupInvalidCrossHeaders(header.Epoch, header.Round)
	}

	// TODO: Should be sent also validatorInfoTxs alongside rewardsTxs -> mp.validatorInfoCreator.GetValidatorInfoTxs(body) ?
	mp.indexBlock(header, headerHash, body, lastMetaBlock, notarizedHeadersHashes, rewardsTxs)
	mp.recordBlockInHistory(headerHash, headerHandler, bodyHandler)

	highestFinalBlockNonce := mp.forkDetector.GetHighestFinalBlockNonce()
	saveMetricsForCommitMetachainBlock(mp.appStatusHandler, header, headerHash, mp.nodesCoordinator, highestFinalBlockNonce)

	headersPool := mp.dataPool.Headers()
	numShardHeadersFromPool := 0
	for shardID := uint32(0); shardID < mp.shardCoordinator.NumberOfShards(); shardID++ {
		numShardHeadersFromPool += headersPool.GetNumHeaders(shardID)
	}

	go mp.headersCounter.displayLogInfo(
		header,
		body,
		headerHash,
		numShardHeadersFromPool,
		mp.blockTracker,
		uint64(mp.roundHandler.TimeDuration().Seconds()),
	)

	headerInfo := bootstrapStorage.BootstrapHeaderInfo{
		ShardId: header.GetShardID(),
		Epoch:   header.GetEpoch(),
		Nonce:   header.GetNonce(),
		Hash:    headerHash,
	}

	nodesCoordinatorKey := mp.nodesCoordinator.GetSavedStateKey()
	epochStartKey := mp.epochStartTrigger.GetSavedStateKey()

	args := bootStorerDataArgs{
		headerInfo:                 headerInfo,
		round:                      header.Round,
		lastSelfNotarizedHeaders:   mp.getLastSelfNotarizedHeaders(),
		nodesCoordinatorConfigKey:  nodesCoordinatorKey,
		epochStartTriggerConfigKey: epochStartKey,
		pendingMiniBlocks:          mp.getPendingMiniBlocks(),
		highestFinalBlockNonce:     highestFinalBlockNonce,
	}

	mp.prepareDataForBootStorer(args)

	mp.blockSizeThrottler.Succeed(header.Round)

	mp.displayPoolsInfo()

	errNotCritical = mp.removeTxsFromPools(header, body)
	if errNotCritical != nil {
		log.Debug("removeTxsFromPools", "error", errNotCritical.Error())
	}

	mp.cleanupPools(headerHandler)

	return nil
}

func (mp *metaProcessor) updateCrossShardInfo(metaBlock *block.MetaBlock) ([]string, error) {
	mp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	defer mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	notarizedHeadersHashes := make([]string, 0)
	for i := 0; i < len(metaBlock.ShardInfo); i++ {
		shardHeaderHash := metaBlock.ShardInfo[i].HeaderHash
		headerInfo, ok := mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardHeaderHash)]
		if !ok {
			return nil, fmt.Errorf("%w : updateCrossShardInfo shardHeaderHash = %s",
				process.ErrMissingHeader, logger.DisplayByteSlice(shardHeaderHash))
		}

		shardHeader, ok := headerInfo.hdr.(data.ShardHeaderHandler)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		mp.updateShardHeadersNonce(shardHeader.GetShardID(), shardHeader.GetNonce())

		marshalizedShardHeader, err := mp.marshalizer.Marshal(shardHeader)
		if err != nil {
			return nil, err
		}

		notarizedHeadersHashes = append(notarizedHeadersHashes, hex.EncodeToString(shardHeaderHash))

		mp.saveShardHeader(shardHeader, shardHeaderHash, marshalizedShardHeader)
	}

	mp.saveMetricCrossCheckBlockHeight()

	return notarizedHeadersHashes, nil
}

func (mp *metaProcessor) displayPoolsInfo() {
	headersPool := mp.dataPool.Headers()
	miniBlocksPool := mp.dataPool.MiniBlocks()

	for shardID := uint32(0); shardID < mp.shardCoordinator.NumberOfShards(); shardID++ {
		log.Trace("pools info",
			"shard", shardID,
			"num headers", headersPool.GetNumHeaders(shardID))
	}

	log.Trace("pools info",
		"shard", core.MetachainShardId,
		"num headers", headersPool.GetNumHeaders(core.MetachainShardId))

	// numShardsToKeepHeaders represents the total number of shards for which meta node would keep tracking headers
	// (in this case this number is equal with: number of shards + metachain (self shard))
	numShardsToKeepHeaders := int(mp.shardCoordinator.NumberOfShards()) + 1
	capacity := headersPool.MaxSize() * numShardsToKeepHeaders
	log.Debug("pools info",
		"total headers", headersPool.Len(),
		"headers pool capacity", capacity,
		"total miniblocks", miniBlocksPool.Len(),
		"miniblocks pool capacity", miniBlocksPool.MaxSize(),
	)

	mp.displayMiniBlocksPool()
}

func (mp *metaProcessor) updateState(lastMetaBlock data.MetaHeaderHandler, lastMetaBlockHash []byte) {
	if check.IfNil(lastMetaBlock) {
		log.Debug("updateState nil header")
		return
	}

	mp.validatorStatisticsProcessor.SetLastFinalizedRootHash(lastMetaBlock.GetValidatorStatsRootHash())

	prevMetaBlockHash := lastMetaBlock.GetPrevHash()
	prevMetaBlock, errNotCritical := process.GetMetaHeader(
		prevMetaBlockHash,
		mp.dataPool.Headers(),
		mp.marshalizer,
		mp.store,
	)
	if errNotCritical != nil {
		log.Debug("could not get meta header from storage")
		return
	}

	if lastMetaBlock.IsStartOfEpochBlock() {
		log.Debug("trie snapshot",
			"rootHash", lastMetaBlock.GetRootHash(),
			"prevRootHash", prevMetaBlock.GetRootHash(),
			"validatorStatsRootHash", lastMetaBlock.GetValidatorStatsRootHash())
		mp.accountsDB[state.UserAccountsState].SnapshotState(lastMetaBlock.GetRootHash())
		mp.accountsDB[state.PeerAccountsState].SnapshotState(lastMetaBlock.GetValidatorStatsRootHash())
		go func() {
			metaBlock, ok := lastMetaBlock.(*block.MetaBlock)
			if !ok {
				log.Warn("cannot commit Trie Epoch Root Hash: lastMetaBlock is not *block.MetaBlock")
				return
			}
			err := mp.commitTrieEpochRootHashIfNeeded(metaBlock, lastMetaBlock.GetRootHash())
			if err != nil {
				log.Warn("couldn't commit trie checkpoint", "epoch", metaBlock.Epoch, "error", err)
			}
		}()
	}

	mp.updateStateStorage(
		lastMetaBlock,
		lastMetaBlock.GetRootHash(),
		prevMetaBlock.GetRootHash(),
		mp.accountsDB[state.UserAccountsState],
		mp.userStatePruningQueue,
	)

	mp.updateStateStorage(
		lastMetaBlock,
		lastMetaBlock.GetValidatorStatsRootHash(),
		prevMetaBlock.GetValidatorStatsRootHash(),
		mp.accountsDB[state.PeerAccountsState],
		mp.peerStatePruningQueue,
	)

	mp.setFinalizedHeaderHashInIndexer(lastMetaBlock.GetPrevHash())
	mp.blockChain.SetFinalBlockInfo(lastMetaBlock.GetNonce(), lastMetaBlockHash, lastMetaBlock.GetRootHash())
}

func (mp *metaProcessor) getLastSelfNotarizedHeaderByShard(
	metaBlock *block.MetaBlock,
	shardID uint32,
) (data.HeaderHandler, []byte) {

	lastNotarizedMetaHeader, lastNotarizedMetaHeaderHash, err := mp.blockTracker.GetLastSelfNotarizedHeader(shardID)
	if err != nil {
		log.Warn("getLastSelfNotarizedHeaderByShard.GetLastSelfNotarizedHeader",
			"shard", shardID,
			"error", err.Error())
		return nil, nil
	}

	maxNotarizedNonce := lastNotarizedMetaHeader.GetNonce()
	for _, shardData := range metaBlock.ShardInfo {
		if shardData.ShardID != shardID {
			continue
		}

		mp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
		headerInfo, ok := mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardData.HeaderHash)]
		mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
		if !ok {
			log.Debug("getLastSelfNotarizedHeaderByShard",
				"error", process.ErrMissingHeader,
				"hash", shardData.HeaderHash)
			continue
		}

		shardHeader, ok := headerInfo.hdr.(data.ShardHeaderHandler)
		if !ok {
			log.Debug("getLastSelfNotarizedHeaderByShard",
				"error", process.ErrWrongTypeAssertion,
				"hash", shardData.HeaderHash)
			continue
		}

		for _, metaHash := range shardHeader.GetMetaBlockHashes() {
			metaHeader, errGet := process.GetMetaHeader(
				metaHash,
				mp.dataPool.Headers(),
				mp.marshalizer,
				mp.store,
			)
			if errGet != nil {
				log.Trace("getLastSelfNotarizedHeaderByShard.GetMetaHeader", "error", errGet.Error())
				continue
			}

			if metaHeader.Nonce > maxNotarizedNonce {
				maxNotarizedNonce = metaHeader.Nonce
				lastNotarizedMetaHeader = metaHeader
				lastNotarizedMetaHeaderHash = metaHash
			}
		}
	}

	if lastNotarizedMetaHeader != nil {
		log.Debug("last notarized meta header in shard",
			"shard", shardID,
			"epoch", lastNotarizedMetaHeader.GetEpoch(),
			"round", lastNotarizedMetaHeader.GetRound(),
			"nonce", lastNotarizedMetaHeader.GetNonce(),
			"hash", lastNotarizedMetaHeaderHash,
		)
	}

	return lastNotarizedMetaHeader, lastNotarizedMetaHeaderHash
}

// getRewardsTxs must be called before method commitEpoch start because when commit is done rewards txs are removed from pool and saved in storage
func (mp *metaProcessor) getRewardsTxs(header *block.MetaBlock, body *block.Body) (rewardsTx map[string]data.TransactionHandler) {
	if !mp.outportHandler.HasDrivers() {
		return
	}
	if !header.IsStartOfEpochBlock() {
		return
	}

	rewardsTx = mp.epochRewardsCreator.GetRewardsTxs(body)
	return rewardsTx
}

func (mp *metaProcessor) commitEpochStart(header *block.MetaBlock, body *block.Body) {
	if header.IsStartOfEpochBlock() {
		mp.epochStartTrigger.SetProcessed(header, body)
		go mp.epochRewardsCreator.SaveBlockDataToStorage(header, body)
		go mp.validatorInfoCreator.SaveBlockDataToStorage(header, body)
	} else {
		currentHeader := mp.blockChain.GetCurrentBlockHeader()
		if !check.IfNil(currentHeader) && currentHeader.IsStartOfEpochBlock() {
			mp.epochStartTrigger.SetFinalityAttestingRound(header.GetRound())
			mp.nodesCoordinator.ShuffleOutForEpoch(currentHeader.GetEpoch())
		}
	}
}

// RevertStateToBlock recreates the state tries to the root hashes indicated by the provided root hash and header
func (mp *metaProcessor) RevertStateToBlock(header data.HeaderHandler, rootHash []byte) error {
	err := mp.accountsDB[state.UserAccountsState].RecreateTrie(rootHash)
	if err != nil {
		log.Debug("recreate trie with error for header",
			"nonce", header.GetNonce(),
			"header root hash", header.GetRootHash(),
			"given root hash", rootHash,
			"error", err.Error(),
		)

		return err
	}

	metaHeader, ok := header.(data.MetaHeaderHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	err = mp.validatorStatisticsProcessor.RevertPeerState(metaHeader)
	if err != nil {
		log.Debug("revert peer state with error for header",
			"nonce", metaHeader.GetNonce(),
			"validators root hash", metaHeader.GetValidatorStatsRootHash(),
			"error", err.Error(),
		)

		return err
	}

	err = mp.epochStartTrigger.RevertStateToBlock(metaHeader)
	if err != nil {
		log.Debug("revert epoch start trigger for header",
			"nonce", metaHeader.GetNonce(),
			"error", err,
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

	mp.appStatusHandler.SetStringValue(common.MetricCrossCheckBlockHeight, crossCheckBlockHeight)
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
			return fmt.Errorf("%w : saveLastNotarizedHeader shardHeaderHash = %s",
				process.ErrMissingHeader, logger.DisplayByteSlice(shardHeaderHash))
		}

		shardHeader, ok := headerInfo.hdr.(data.ShardHeaderHandler)
		if !ok {
			mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrWrongTypeAssertion
		}

		if lastCrossNotarizedHeaderForShard[shardHeader.GetShardID()].hdr.GetNonce() < shardHeader.GetNonce() {
			lastCrossNotarizedHeaderForShard[shardHeader.GetShardID()] = &hashAndHdr{hdr: shardHeader, hash: shardHeaderHash}
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

func (mp *metaProcessor) getLastCrossNotarizedShardHdrs() (map[uint32]data.HeaderHandler, error) {
	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	defer mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	lastCrossNotarizedHeader := make(map[uint32]data.HeaderHandler, mp.shardCoordinator.NumberOfShards())
	for shardID := uint32(0); shardID < mp.shardCoordinator.NumberOfShards(); shardID++ {
		lastCrossNotarizedHeaderForShard, hash, err := mp.blockTracker.GetLastCrossNotarizedHeader(shardID)
		if err != nil {
			return nil, err
		}

		log.Debug("lastCrossNotarizedHeader for shard", "shardID", shardID, "hash", hash)
		lastCrossNotarizedHeader[shardID] = lastCrossNotarizedHeaderForShard
		usedInBlock := mp.isGenesisShardBlockAndFirstMeta(lastCrossNotarizedHeaderForShard.GetNonce())
		mp.hdrsForCurrBlock.hdrHashAndInfo[string(hash)] = &hdrInfo{hdr: lastCrossNotarizedHeaderForShard, usedInBlock: usedInBlock}
	}

	return lastCrossNotarizedHeader, nil
}

// check if shard headers were signed and constructed correctly and returns headers which has to be
// checked for finality
func (mp *metaProcessor) checkShardHeadersValidity(metaHdr *block.MetaBlock) (map[uint32]data.HeaderHandler, error) {
	lastCrossNotarizedHeader, err := mp.getLastCrossNotarizedShardHdrs()
	if err != nil {
		return nil, err
	}

	usedShardHdrs := mp.sortHeadersForCurrentBlockByNonce(true)
	highestNonceHdrs := make(map[uint32]data.HeaderHandler, len(usedShardHdrs))

	if len(usedShardHdrs) == 0 {
		return highestNonceHdrs, nil
	}

	for shardID, hdrsForShard := range usedShardHdrs {
		for _, shardHdr := range hdrsForShard {
			if !mp.isGenesisShardBlockAndFirstMeta(shardHdr.GetNonce()) {
				err = mp.headerValidator.IsHeaderConstructionValid(shardHdr, lastCrossNotarizedHeader[shardID])
				if err != nil {
					return nil, fmt.Errorf("%w : checkShardHeadersValidity -> isHdrConstructionValid", err)
				}
			}

			lastCrossNotarizedHeader[shardID] = shardHdr
			highestNonceHdrs[shardID] = shardHdr
		}
	}

	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	defer mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	for _, shardData := range metaHdr.ShardInfo {
		headerInfo, ok := mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardData.HeaderHash)]
		if !ok {
			return nil, fmt.Errorf("%w : checkShardHeadersValidity -> hash not found %s ",
				process.ErrHeaderShardDataMismatch, hex.EncodeToString(shardData.HeaderHash))
		}
		actualHdr := headerInfo.hdr
		shardHdr, ok := actualHdr.(data.ShardHeaderHandler)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		finalMiniBlockHeaders := mp.getFinalMiniBlockHeaders(shardHdr.GetMiniBlockHeaderHandlers())

		if len(shardData.ShardMiniBlockHeaders) != len(finalMiniBlockHeaders) {
			return nil, process.ErrHeaderShardDataMismatch
		}
		if shardData.AccumulatedFees.Cmp(shardHdr.GetAccumulatedFees()) != 0 {
			return nil, process.ErrAccumulatedFeesDoNotMatch
		}
		if shardData.DeveloperFees.Cmp(shardHdr.GetDeveloperFees()) != 0 {
			return nil, process.ErrDeveloperFeesDoNotMatch
		}

		mapMiniBlockHeadersInMetaBlock := make(map[string]struct{})
		for _, shardMiniBlockHdr := range shardData.ShardMiniBlockHeaders {
			mapMiniBlockHeadersInMetaBlock[string(shardMiniBlockHdr.Hash)] = struct{}{}
		}

		for _, actualMiniBlockHdr := range finalMiniBlockHeaders {
			if _, hashExists := mapMiniBlockHeadersInMetaBlock[string(actualMiniBlockHdr.GetHash())]; !hashExists {
				return nil, process.ErrHeaderShardDataMismatch
			}
		}
	}

	return highestNonceHdrs, nil
}

func (mp *metaProcessor) getFinalMiniBlockHeaders(miniBlockHeaderHandlers []data.MiniBlockHeaderHandler) []data.MiniBlockHeaderHandler {
	if !mp.enableEpochsHandler.IsScheduledMiniBlocksFlagEnabled() {
		return miniBlockHeaderHandlers
	}

	miniBlockHeaders := make([]data.MiniBlockHeaderHandler, 0)
	for _, miniBlockHeader := range miniBlockHeaderHandlers {
		if !miniBlockHeader.IsFinal() {
			log.Debug("metaProcessor.getFinalMiniBlockHeaders: do not check validity for mini block which is not final", "mb hash", miniBlockHeader.GetHash())
			continue
		}

		miniBlockHeaders = append(miniBlockHeaders, miniBlockHeader)
	}

	return miniBlockHeaders
}

// check if shard headers are final by checking if newer headers were constructed upon them
func (mp *metaProcessor) checkShardHeadersFinality(
	highestNonceHdrs map[uint32]data.HeaderHandler,
) error {
	finalityAttestingShardHdrs := mp.sortHeadersForCurrentBlockByNonce(false)

	var errFinal error

	for shardId, lastVerifiedHdr := range highestNonceHdrs {
		if check.IfNil(lastVerifiedHdr) {
			return process.ErrNilBlockHeader
		}
		if lastVerifiedHdr.GetShardID() != shardId {
			return process.ErrShardIdMissmatch
		}
		isGenesisShardBlockAndFirstMeta := mp.isGenesisShardBlockAndFirstMeta(lastVerifiedHdr.GetNonce())
		if isGenesisShardBlockAndFirstMeta {
			continue
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
	shardHeader, ok := headerHandler.(data.ShardHeaderHandler)
	if !ok {
		return
	}

	log.Trace("received shard header from network",
		"shard", shardHeader.GetShardID(),
		"round", shardHeader.GetRound(),
		"nonce", shardHeader.GetNonce(),
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

			if shardHeader.GetNonce() > mp.hdrsForCurrBlock.highestHdrNonce[shardHeader.GetShardID()] {
				mp.hdrsForCurrBlock.highestHdrNonce[shardHeader.GetShardID()] = shardHeader.GetNonce()
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

	go mp.requestMiniBlocksIfNeeded(headerHandler)
}

// requestMissingFinalityAttestingShardHeaders requests the headers needed to accept the current selected headers for
// processing the current block. It requests the shardBlockFinality headers greater than the highest shard header,
// for each shard, related to the block which should be processed
// this method should be called only under the mutex protection: hdrsForCurrBlock.mutHdrsForBlock
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
	_ = core.EmptyChannel(mp.chRcvAllHdrs)

	if len(metaBlock.ShardInfo) == 0 {
		return 0, 0
	}

	return mp.computeExistingAndRequestMissingShardHeaders(metaBlock)
}

func (mp *metaProcessor) computeExistingAndRequestMissingShardHeaders(metaBlock *block.MetaBlock) (uint32, uint32) {
	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	defer mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	for _, shardData := range metaBlock.ShardInfo {
		if shardData.Nonce == mp.genesisNonce {
			lastCrossNotarizedHeaderForShard, hash, err := mp.blockTracker.GetLastCrossNotarizedHeader(shardData.ShardID)
			if err != nil {
				log.Warn("computeExistingAndRequestMissingShardHeaders.GetLastCrossNotarizedHeader", "error", err.Error())
				continue
			}
			if !bytes.Equal(hash, shardData.HeaderHash) {
				log.Warn("genesis hash missmatch",
					"last notarized nonce", lastCrossNotarizedHeaderForShard.GetNonce(),
					"last notarized hash", hash,
					"genesis nonce", mp.genesisNonce,
					"genesis hash", shardData.HeaderHash)
			}
			continue
		}

		hdr, err := process.GetShardHeaderFromPool(
			shardData.HeaderHash,
			mp.dataPool.Headers())

		if err != nil {
			mp.hdrsForCurrBlock.missingHdrs++
			mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardData.HeaderHash)] = &hdrInfo{
				hdr:         nil,
				usedInBlock: true,
			}
			go mp.requestHandler.RequestShardHeader(shardData.ShardID, shardData.HeaderHash)
			continue
		}

		mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardData.HeaderHash)] = &hdrInfo{
			hdr:         hdr,
			usedInBlock: true,
		}

		if hdr.GetNonce() > mp.hdrsForCurrBlock.highestHdrNonce[shardData.ShardID] {
			mp.hdrsForCurrBlock.highestHdrNonce[shardData.ShardID] = hdr.GetNonce()
		}
	}

	if mp.hdrsForCurrBlock.missingHdrs == 0 {
		mp.hdrsForCurrBlock.missingFinalityAttestingHdrs = mp.requestMissingFinalityAttestingShardHeaders()
	}

	return mp.hdrsForCurrBlock.missingHdrs, mp.hdrsForCurrBlock.missingFinalityAttestingHdrs
}

func (mp *metaProcessor) createShardInfo() ([]data.ShardDataHandler, error) {
	var shardInfo []data.ShardDataHandler
	if mp.epochStartTrigger.IsEpochStart() {
		return shardInfo, nil
	}

	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	for hdrHash, headerInfo := range mp.hdrsForCurrBlock.hdrHashAndInfo {
		if !headerInfo.usedInBlock {
			continue
		}

		shardHdr, ok := headerInfo.hdr.(data.ShardHeaderHandler)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		shardData := block.ShardData{}
		shardData.TxCount = shardHdr.GetTxCount()
		shardData.ShardID = shardHdr.GetShardID()
		shardData.HeaderHash = []byte(hdrHash)
		shardData.Round = shardHdr.GetRound()
		shardData.PrevHash = shardHdr.GetPrevHash()
		shardData.Nonce = shardHdr.GetNonce()
		shardData.PrevRandSeed = shardHdr.GetPrevRandSeed()
		shardData.PubKeysBitmap = shardHdr.GetPubKeysBitmap()
		shardData.NumPendingMiniBlocks = uint32(len(mp.pendingMiniBlocksHandler.GetPendingMiniBlocks(shardData.ShardID)))
		header, _, err := mp.blockTracker.GetLastSelfNotarizedHeader(shardHdr.GetShardID())
		if err != nil {
			return nil, err
		}
		shardData.LastIncludedMetaNonce = header.GetNonce()
		shardData.AccumulatedFees = shardHdr.GetAccumulatedFees()
		shardData.DeveloperFees = shardHdr.GetDeveloperFees()

		for i := 0; i < len(shardHdr.GetMiniBlockHeaderHandlers()); i++ {
			if mp.enableEpochsHandler.IsScheduledMiniBlocksFlagEnabled() {
				miniBlockHeader := shardHdr.GetMiniBlockHeaderHandlers()[i]
				if !miniBlockHeader.IsFinal() {
					log.Debug("metaProcessor.createShardInfo: do not create shard data with mini block which is not final", "mb hash", miniBlockHeader.GetHash())
					continue
				}
			}

			shardMiniBlockHeader := block.MiniBlockHeader{}
			shardMiniBlockHeader.SenderShardID = shardHdr.GetMiniBlockHeaderHandlers()[i].GetSenderShardID()
			shardMiniBlockHeader.ReceiverShardID = shardHdr.GetMiniBlockHeaderHandlers()[i].GetReceiverShardID()
			shardMiniBlockHeader.Hash = shardHdr.GetMiniBlockHeaderHandlers()[i].GetHash()
			shardMiniBlockHeader.TxCount = shardHdr.GetMiniBlockHeaderHandlers()[i].GetTxCount()
			shardMiniBlockHeader.Type = block.Type(shardHdr.GetMiniBlockHeaderHandlers()[i].GetTypeInt32())

			shardData.ShardMiniBlockHeaders = append(shardData.ShardMiniBlockHeaders, shardMiniBlockHeader)
		}

		shardInfo = append(shardInfo, &shardData)
	}
	mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	log.Debug("created shard data",
		"size", len(shardInfo),
	)
	return shardInfo, nil
}

func (mp *metaProcessor) verifyTotalAccumulatedFeesInEpoch(metaHdr *block.MetaBlock) error {
	computedTotalFees, computedTotalDevFees, err := mp.computeAccumulatedFeesInEpoch(metaHdr)
	if err != nil {
		return err
	}

	if computedTotalFees.Cmp(metaHdr.AccumulatedFeesInEpoch) != 0 {
		return fmt.Errorf("%w, got %v, computed %v", process.ErrAccumulatedFeesInEpochDoNotMatch, metaHdr.AccumulatedFeesInEpoch, computedTotalFees)
	}

	if computedTotalDevFees.Cmp(metaHdr.DevFeesInEpoch) != 0 {
		return fmt.Errorf("%w, got %v, computed %v", process.ErrDevFeesInEpochDoNotMatch, metaHdr.DevFeesInEpoch, computedTotalDevFees)
	}

	return nil
}

func (mp *metaProcessor) computeAccumulatedFeesInEpoch(metaHdr data.MetaHeaderHandler) (*big.Int, *big.Int, error) {
	currentlyAccumulatedFeesInEpoch := big.NewInt(0)
	currentDevFeesInEpoch := big.NewInt(0)

	lastHdr := mp.blockChain.GetCurrentBlockHeader()
	if !check.IfNil(lastHdr) {
		lastMeta, ok := lastHdr.(*block.MetaBlock)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}

		if !lastHdr.IsStartOfEpochBlock() {
			currentlyAccumulatedFeesInEpoch = big.NewInt(0).Set(lastMeta.AccumulatedFeesInEpoch)
			currentDevFeesInEpoch = big.NewInt(0).Set(lastMeta.DevFeesInEpoch)
		}
	}

	currentlyAccumulatedFeesInEpoch.Add(currentlyAccumulatedFeesInEpoch, metaHdr.GetAccumulatedFees())
	currentDevFeesInEpoch.Add(currentDevFeesInEpoch, metaHdr.GetDeveloperFees())
	for _, shardData := range metaHdr.GetShardInfoHandlers() {
		currentlyAccumulatedFeesInEpoch.Add(currentlyAccumulatedFeesInEpoch, shardData.GetAccumulatedFees())
		currentDevFeesInEpoch.Add(currentDevFeesInEpoch, shardData.GetDeveloperFees())
	}

	return currentlyAccumulatedFeesInEpoch, currentDevFeesInEpoch, nil
}

// applyBodyToHeader creates a miniblock header list given a block body
func (mp *metaProcessor) applyBodyToHeader(metaHdr data.MetaHeaderHandler, bodyHandler data.BodyHandler) (data.BodyHandler, error) {
	sw := core.NewStopWatch()
	sw.Start("applyBodyToHeader")
	defer func() {
		sw.Stop("applyBodyToHeader")

		log.Debug("measurements", sw.GetMeasurements()...)
	}()

	if check.IfNil(bodyHandler) {
		return nil, process.ErrNilBlockBody
	}

	var err error
	defer func() {
		go mp.checkAndRequestIfShardHeadersMissing()
	}()

	sw.Start("createShardInfo")
	shardInfo, err := mp.createShardInfo()
	sw.Stop("createShardInfo")
	if err != nil {
		return nil, err
	}

	err = metaHdr.SetEpoch(mp.epochStartTrigger.Epoch())
	if err != nil {
		return nil, err
	}

	err = metaHdr.SetShardInfoHandlers(shardInfo)
	if err != nil {
		return nil, err
	}

	err = metaHdr.SetRootHash(mp.getRootHash())
	if err != nil {
		return nil, err
	}

	err = metaHdr.SetTxCount(getTxCount(shardInfo))
	if err != nil {
		return nil, err
	}

	err = metaHdr.SetAccumulatedFees(mp.feeHandler.GetAccumulatedFees())
	if err != nil {
		return nil, err
	}

	err = metaHdr.SetDeveloperFees(mp.feeHandler.GetDeveloperFees())
	if err != nil {
		return nil, err
	}

	accumulatedFees, devFees, err := mp.computeAccumulatedFeesInEpoch(metaHdr)
	if err != nil {
		return nil, err
	}

	err = metaHdr.SetAccumulatedFeesInEpoch(accumulatedFees)
	if err != nil {
		return nil, err
	}

	err = metaHdr.SetDevFeesInEpoch(devFees)
	if err != nil {
		return nil, err
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return nil, err
	}

	sw.Start("CreateReceiptsHash")
	receiptsHash, err := mp.txCoordinator.CreateReceiptsHash()
	sw.Stop("CreateReceiptsHash")
	if err != nil {
		return nil, err
	}

	err = metaHdr.SetReceiptsHash(receiptsHash)
	if err != nil {
		return nil, err
	}

	totalTxCount, miniBlockHeaderHandlers, err := mp.createMiniBlockHeaderHandlers(body, make(map[string]*processedMb.ProcessedMiniBlockInfo))
	if err != nil {
		return nil, err
	}

	err = metaHdr.SetMiniBlockHeaderHandlers(miniBlockHeaderHandlers)
	if err != nil {
		return nil, err
	}

	txCount := metaHdr.GetTxCount() + uint32(totalTxCount)
	err = metaHdr.SetTxCount(txCount)
	if err != nil {
		return nil, err
	}

	sw.Start("UpdatePeerState")
	mp.prepareBlockHeaderInternalMapForValidatorProcessor()
	valStatRootHash, err := mp.validatorStatisticsProcessor.UpdatePeerState(metaHdr, mp.hdrsForCurrBlock.getHdrHashMap())
	sw.Stop("UpdatePeerState")
	if err != nil {
		return nil, err
	}

	err = metaHdr.SetValidatorStatsRootHash(valStatRootHash)
	if err != nil {
		return nil, err
	}

	marshalizedBody, err := mp.marshalizer.Marshal(body)
	if err != nil {
		return nil, err
	}

	mp.blockSizeThrottler.Add(metaHdr.GetRound(), uint32(len(marshalizedBody)))

	return body, nil
}

func (mp *metaProcessor) prepareBlockHeaderInternalMapForValidatorProcessor() {
	currentBlockHeader := mp.blockChain.GetCurrentBlockHeader()
	currentBlockHeaderHash := mp.blockChain.GetCurrentBlockHeaderHash()

	if check.IfNil(currentBlockHeader) {
		currentBlockHeader = mp.blockChain.GetGenesisHeader()
		currentBlockHeaderHash = mp.blockChain.GetGenesisHeaderHash()
	}

	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	mp.hdrsForCurrBlock.hdrHashAndInfo[string(currentBlockHeaderHash)] = &hdrInfo{false, currentBlockHeader}
	mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()
}

func (mp *metaProcessor) verifyValidatorStatisticsRootHash(header *block.MetaBlock) error {
	mp.prepareBlockHeaderInternalMapForValidatorProcessor()
	validatorStatsRH, err := mp.validatorStatisticsProcessor.UpdatePeerState(header, mp.hdrsForCurrBlock.getHdrHashMap())
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
			logger.DisplayByteSlice(validatorStatsRH),
			logger.DisplayByteSlice(header.GetValidatorStatsRootHash()),
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
func (mp *metaProcessor) CreateNewHeader(round uint64, nonce uint64) (data.HeaderHandler, error) {
	mp.enableRoundsHandler.CheckRound(round)

	mp.epochStartTrigger.Update(round, nonce)
	epoch := mp.epochStartTrigger.Epoch()

	header := mp.versionedHeaderFactory.Create(epoch)
	metaHeader, ok := header.(data.MetaHeaderHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	err := metaHeader.SetRound(round)
	if err != nil {
		return nil, err
	}

	err = metaHeader.SetNonce(nonce)
	if err != nil {
		return nil, err
	}

	err = metaHeader.SetAccumulatedFees(big.NewInt(0))
	if err != nil {
		return nil, err
	}

	err = metaHeader.SetAccumulatedFeesInEpoch(big.NewInt(0))
	if err != nil {
		return nil, err
	}

	err = metaHeader.SetDeveloperFees(big.NewInt(0))
	if err != nil {
		return nil, err
	}

	err = metaHeader.SetDevFeesInEpoch(big.NewInt(0))
	if err != nil {
		return nil, err
	}

	err = mp.setHeaderVersionData(metaHeader)
	if err != nil {
		return nil, err
	}

	return metaHeader, nil
}

func (mp *metaProcessor) setHeaderVersionData(metaHeader data.MetaHeaderHandler) error {
	if check.IfNil(metaHeader) {
		return process.ErrNilHeaderHandler
	}

	rootHash, err := mp.accountsDB[state.UserAccountsState].RootHash()
	if err != nil {
		return err
	}

	scheduledGasAndFees := mp.scheduledTxsExecutionHandler.GetScheduledGasAndFees()
	additionalVersionData := &headerVersionData.AdditionalData{
		ScheduledRootHash:        rootHash,
		ScheduledAccumulatedFees: scheduledGasAndFees.GetAccumulatedFees(),
		ScheduledDeveloperFees:   scheduledGasAndFees.GetDeveloperFees(),
		ScheduledGasProvided:     scheduledGasAndFees.GetGasProvided(),
		ScheduledGasPenalized:    scheduledGasAndFees.GetGasPenalized(),
		ScheduledGasRefunded:     scheduledGasAndFees.GetGasRefunded(),
	}

	return metaHeader.SetAdditionalData(additionalVersionData)
}

// MarshalizedDataToBroadcast prepares underlying data into a marshalized object according to destination
func (mp *metaProcessor) MarshalizedDataToBroadcast(
	hdr data.HeaderHandler,
	bodyHandler data.BodyHandler,
) (map[uint32][]byte, map[string][][]byte, error) {
	if check.IfNil(hdr) {
		return nil, nil, process.ErrNilMetaBlockHeader
	}
	if check.IfNil(bodyHandler) {
		return nil, nil, process.ErrNilMiniBlocks
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	var mrsTxs map[string][][]byte
	if hdr.IsStartOfEpochBlock() {
		mrsTxs = mp.getAllMarshalledTxs(body)
	} else {
		mrsTxs = mp.txCoordinator.CreateMarshalizedData(body)
	}

	bodies := make(map[uint32]block.MiniBlockSlice)
	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.SenderShardID != mp.shardCoordinator.SelfId() ||
			miniBlock.ReceiverShardID == mp.shardCoordinator.SelfId() {
			continue
		}
		bodies[miniBlock.ReceiverShardID] = append(bodies[miniBlock.ReceiverShardID], miniBlock)
	}

	mrsData := make(map[uint32][]byte, len(bodies))
	for shardId, subsetBlockBody := range bodies {
		buff, err := mp.marshalizer.Marshal(&block.Body{MiniBlocks: subsetBlockBody})
		if err != nil {
			log.Error("metaProcessor.MarshalizedDataToBroadcast.Marshal", "error", err.Error())
			continue
		}
		mrsData[shardId] = buff
	}

	return mrsData, mrsTxs, nil
}

func (mp *metaProcessor) getAllMarshalledTxs(body *block.Body) map[string][][]byte {
	allMarshalledTxs := make(map[string][][]byte)

	marshalledRewardsTxs := mp.epochRewardsCreator.CreateMarshalledData(body)
	marshalledValidatorInfoTxs := mp.validatorInfoCreator.CreateMarshalledData(body)

	for topic, marshalledTxs := range marshalledRewardsTxs {
		allMarshalledTxs[topic] = append(allMarshalledTxs[topic], marshalledTxs...)
		log.Trace("metaProcessor.getAllMarshalledTxs", "topic", topic, "num rewards txs", len(marshalledTxs))
	}

	for topic, marshalledTxs := range marshalledValidatorInfoTxs {
		allMarshalledTxs[topic] = append(allMarshalledTxs[topic], marshalledTxs...)
		log.Trace("metaProcessor.getAllMarshalledTxs", "topic", topic, "num validator info txs", len(marshalledTxs))
	}

	return allMarshalledTxs
}

func getTxCount(shardInfo []data.ShardDataHandler) uint32 {
	txs := uint32(0)
	for i := 0; i < len(shardInfo); i++ {
		shardDataHandlers := shardInfo[i].GetShardMiniBlockHeaderHandlers()
		for j := 0; j < len(shardDataHandlers); j++ {
			txs += shardDataHandlers[j].GetTxCount()
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
	metaBlock, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	miniBlocksPool := mp.dataPool.MiniBlocks()
	var miniBlocks block.MiniBlockSlice

	for _, mbHeader := range metaBlock.MiniBlockHeaders {
		obj, hashInPool := miniBlocksPool.Get(mbHeader.Hash)
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

func (mp *metaProcessor) getPendingMiniBlocks() []bootstrapStorage.PendingMiniBlocksInfo {
	pendingMiniBlocksInfo := make([]bootstrapStorage.PendingMiniBlocksInfo, mp.shardCoordinator.NumberOfShards())

	for shardID := uint32(0); shardID < mp.shardCoordinator.NumberOfShards(); shardID++ {
		pendingMiniBlocksInfo[shardID] = bootstrapStorage.PendingMiniBlocksInfo{
			MiniBlocksHashes: mp.pendingMiniBlocksHandler.GetPendingMiniBlocks(shardID),
			ShardID:          shardID,
		}
	}

	return pendingMiniBlocksInfo
}

func (mp *metaProcessor) removeStartOfEpochBlockDataFromPools(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {

	if !headerHandler.IsStartOfEpochBlock() {
		return nil
	}

	metaBlock, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	mp.epochRewardsCreator.RemoveBlockDataFromPools(metaBlock, body)
	mp.validatorInfoCreator.RemoveBlockDataFromPools(metaBlock, body)

	return nil
}

// Close - closes all underlying components
func (mp *metaProcessor) Close() error {
	return mp.baseProcessor.Close()
}

// DecodeBlockHeader method decodes block header from a given byte array
func (mp *metaProcessor) DecodeBlockHeader(dta []byte) data.HeaderHandler {
	if dta == nil {
		return nil
	}

	metaBlock := &block.MetaBlock{}
	err := mp.marshalizer.Unmarshal(metaBlock, dta)
	if err != nil {
		log.Debug("DecodeBlockHeader.Unmarshal", "error", err.Error())
		return nil
	}

	return metaBlock
}
