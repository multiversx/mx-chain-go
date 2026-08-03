package block

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
)

// metaArbitrationWindowRounds is the arbitration discovery window: rounds after a contended shard
// header's round during which meta holds it so competing proofs can surface
const metaArbitrationWindowRounds = 3

// usedShardHeadersInfo holds the used shard headers information
type usedShardHeadersInfo struct {
	headersPerShard          map[uint32][]ShardHeaderInfo
	orderedShardHeaders      []data.HeaderHandler
	orderedShardHeaderHashes [][]byte
}

// CreateNewHeaderProposal creates a new header
func (mp *metaProcessor) CreateNewHeaderProposal(round uint64, nonce uint64) (data.HeaderHandler, error) {
	epoch := mp.epochStartTrigger.Epoch()

	header := mp.versionedHeaderFactory.Create(epoch, round)
	metaHeader, ok := header.(data.MetaHeaderHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	if !metaHeader.IsHeaderV3() {
		return nil, process.ErrInvalidHeader
	}

	epochChangeProposed := mp.epochStartTrigger.ShouldProposeEpochChange(round, nonce)
	metaHeader.SetEpochChangeProposed(epochChangeProposed)
	err := metaHeader.SetRound(round)
	if err != nil {
		return nil, err
	}

	err = metaHeader.SetNonce(nonce)
	if err != nil {
		return nil, err
	}

	err = mp.addExecutionResultsOnHeader(metaHeader)
	if err != nil {
		return nil, err
	}

	err = mp.checkNonceGaps(metaHeader)
	if err != nil {
		return nil, err
	}

	err = metaHeader.SetEpochStartHandler(&block.EpochStart{})
	if err != nil {
		return nil, err
	}
	hasEpochStartResults, err := mp.hasStartOfEpochExecutionResults(metaHeader)
	if err != nil {
		return nil, err
	}
	if !hasEpochStartResults {
		return metaHeader, nil
	}

	err = metaHeader.SetEpoch(epoch + 1)
	if err != nil {
		return nil, err
	}

	epochStartData, err := mp.getComputedEpochStartData(epoch + 1)
	if err != nil {
		return nil, err
	}

	err = metaHeader.SetEpochStartHandler(epochStartData)
	if err != nil {
		return nil, err
	}

	err = mp.checkEpochCorrectnessV3(metaHeader)
	if err != nil {
		return nil, fmt.Errorf("created meta header with invalid epoch start data: %w", err)
	}

	return metaHeader, nil
}

// CreateBlockProposal creates a block proposal without executing any of the transactions
func (mp *metaProcessor) CreateBlockProposal(
	initialHdr data.HeaderHandler,
	haveTime func() bool,
) (data.HeaderHandler, data.BodyHandler, error) {
	if check.IfNil(initialHdr) {
		return nil, nil, process.ErrNilBlockHeader
	}
	if !initialHdr.IsHeaderV3() {
		return nil, nil, process.ErrInvalidHeader
	}

	metaHdr, ok := initialHdr.(*block.MetaBlockV3)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	err := mp.checkLegacyPredecessorReadyForV3(metaHdr)
	if err != nil {
		return nil, nil, err
	}

	metaHdr.SoftwareVersion = []byte(mp.headerIntegrityVerifier.GetVersion(metaHdr.Epoch, metaHdr.Round))

	if metaHdr.IsStartOfEpochBlock() || metaHdr.GetEpochChangeProposed() || mp.epochStartTrigger.GetEpochChangeProposed() {
		// no new transactions in start of epoch block
		// to simplify bootstrapping
		return metaHdr, &block.Body{}, nil
	}

	mp.gasComputation.Reset()
	mp.miniBlocksSelectionSession.ResetSelectionSession()
	err = mp.createBlockBodyProposal(metaHdr, haveTime)
	if err != nil {
		return nil, nil, err
	}

	mbsToMe := mp.miniBlocksSelectionSession.GetMiniBlocks()
	miniBlocksHeadersToMe := mp.miniBlocksSelectionSession.GetMiniBlockHeaderHandlers()
	err = checkProposalMiniBlocksConsistency(miniBlocksHeadersToMe, mbsToMe, metaHdr.GetShardID())
	if err != nil {
		return nil, nil, err
	}

	numTxs := mp.miniBlocksSelectionSession.GetNumTxsAdded()
	referencedShardHeaderHashes := mp.miniBlocksSelectionSession.GetReferencedHeaderHashes()
	referencedShardHeaders := mp.miniBlocksSelectionSession.GetReferencedHeaders()
	body := &block.Body{
		MiniBlocks: mbsToMe,
	}

	if len(mbsToMe) > 0 {
		log.Debug("created miniblocks with txs with destination in self shard",
			"num miniblocks", len(mbsToMe),
			"num txs proposed", numTxs,
			"num shard headers", len(referencedShardHeaderHashes),
		)

	}

	defer func() {
		go mp.checkAndRequestIfShardHeadersMissing()
	}()

	shardDataProposalHandlers, shardDataHandlers, err := mp.shardInfoCreateData.CreateShardInfoV3(metaHdr, referencedShardHeaders, referencedShardHeaderHashes)
	if err != nil {
		return nil, nil, err
	}

	err = metaHdr.SetShardInfoHandlers(shardDataHandlers)
	if err != nil {
		return nil, nil, err
	}

	err = metaHdr.SetShardInfoProposalHandlers(shardDataProposalHandlers)
	if err != nil {
		return nil, nil, err
	}

	err = metaHdr.SetMiniBlockHeaderHandlers(miniBlocksHeadersToMe)
	if err != nil {
		return nil, nil, err
	}

	marshalledBody, err := mp.marshalizer.Marshal(body)
	if err != nil {
		return nil, nil, err
	}
	mp.blockSizeThrottler.Add(metaHdr.GetRound(), uint32(len(marshalledBody)))

	return metaHdr, body, nil
}

// VerifyBlockProposal verifies the proposed block. It returns nil if all ok or the specific error
func (mp *metaProcessor) VerifyBlockProposal(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
	haveTime func() time.Duration,
) error {
	err := mp.checkBlockValidity(headerHandler, bodyHandler)
	if err != nil {
		if errors.Is(err, process.ErrBlockHashDoesNotMatch) {
			log.Debug("requested missing meta header",
				"hash", headerHandler.GetPrevHash(),
				"for shard", headerHandler.GetShardID(),
			)

			go mp.requestHandler.RequestMetaHeaderForEpoch(headerHandler.GetPrevHash(), headerHandler.GetEpoch())
		}

		return err
	}

	log.Debug("started verifying proposed meta block",
		"epoch", headerHandler.GetEpoch(),
		"shard", headerHandler.GetShardID(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce())

	header, ok := headerHandler.(*block.MetaBlockV3)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	if !header.IsHeaderV3() {
		return process.ErrInvalidHeader
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	err = mp.checkLegacyPredecessorReadyForV3(header)
	if err != nil {
		return err
	}

	shouldProposeEpochChange := mp.epochStartTrigger.ShouldProposeEpochChange(headerHandler.GetRound(), headerHandler.GetNonce())
	isEpochChangeProposed := header.IsEpochChangeProposed()
	// The header flag must match the trigger state in both directions:
	// it is invalid if the header proposes an epoch change too early or misses one when required.
	if isEpochChangeProposed != shouldProposeEpochChange {
		log.Warn("epoch change proposal flag does not match trigger state",
			"round", headerHandler.GetRound(),
			"nonce", headerHandler.GetNonce(),
			"flag from header", isEpochChangeProposed,
			"flag from trigger", shouldProposeEpochChange,
			"epochStartTrigger", mp.epochStartTrigger.Epoch())
		return process.ErrEpochChangeProposedOutsideTriggerWindow
	}

	if header.IsEpochChangeProposed() && len(body.MiniBlocks) != 0 {
		return process.ErrEpochStartProposeBlockHasMiniBlocks
	}

	if header.IsStartOfEpochBlock() {
		if len(header.GetShardInfoHandlers()) > 0 {
			return process.ErrShardInfoOnEpochStartBlock
		}

		err := mp.verifyEpochStartMiniBlocks(header)
		if err != nil {
			return err
		}
	}

	err = mp.checkHeaderBodyCorrelation(header.GetMiniBlockHeaderHandlers(), body, header.GetShardID(), header.GetEpoch(), true)
	if err != nil {
		return err
	}

	err = mp.waitForExecutionResultsVerification(header, haveTime)
	if err != nil {
		return err
	}

	err = mp.checkInclusionEstimationForExecutionResults(header)
	if err != nil {
		return err
	}

	err = mp.checkNonceGaps(header)
	if err != nil {
		return err
	}

	mp.updateMetrics(header)

	mp.missingDataResolver.Reset()
	mp.missingDataResolver.RequestBlockTransactions(body)
	err = mp.missingDataResolver.RequestMissingShardHeaders(header)
	if err != nil {
		return err
	}

	err = mp.missingDataResolver.WaitForMissingData(haveTime())
	if err != nil {
		return err
	}

	defer func() {
		go mp.checkAndRequestIfShardHeadersMissing()
	}()

	err = mp.checkEpochCorrectnessV3(header)
	if err != nil {
		return err
	}

	err = mp.checkShardHeadersValidityAndFinalityProposal(header)
	if err != nil {
		return err
	}

	err = mp.verifyCrossShardMiniBlockDstMe(header)
	if err != nil {
		return err
	}

	return mp.verifyGasLimit(header, body.MiniBlocks, false)
}

// ProcessBlockProposal processes the proposed block. It returns nil if all ok or the specific error
func (mp *metaProcessor) ProcessBlockProposal(
	headerHandler data.HeaderHandler,
	headerHash []byte,
	bodyHandler data.BodyHandler,
) (data.BaseExecutionResultHandler, error) {
	if check.IfNil(headerHandler) {
		return nil, process.ErrNilBlockHeader
	}
	if check.IfNil(bodyHandler) {
		return nil, process.ErrNilBlockBody
	}
	if !headerHandler.IsHeaderV3() {
		return nil, process.ErrInvalidHeader
	}

	if !mp.processStatusHandler.TrySetBusy("metaProcessor.ProcessBlockProposal") {
		return nil, process.ErrBlockProcessorBusy
	}
	defer mp.processStatusHandler.SetIdle()

	mp.roundNotifier.CheckRound(headerHandler)
	mp.epochNotifier.CheckEpoch(headerHandler)
	mp.requestHandler.SetEpoch(headerHandler.GetEpoch())

	header, ok := headerHandler.(data.MetaHeaderHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	log.Debug("started processing block",
		"epoch", headerHandler.GetEpoch(),
		"shard", headerHandler.GetShardID(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
	)

	if mp.accountsDB[state.UserAccountsState].JournalLen() != 0 {
		log.Error("metaProcessor.ProcessBlockProposal first entry", "stack", string(mp.accountsDB[state.UserAccountsState].GetStackDebugFirstEntry()))
		return nil, fmt.Errorf("%w for user accounts", process.ErrAccountStateDirty)
	}
	if mp.accountsDB[state.PeerAccountsState].JournalLen() != 0 {
		log.Error("metaProcessor.ProcessBlockProposal peer accounts first entry", "stack", string(mp.accountsDB[state.PeerAccountsState].GetStackDebugFirstEntry()))
		return nil, fmt.Errorf("%w for peer accounts", process.ErrAccountStateDirty)
	}

	var err error
	defer func() {
		if err != nil {
			mp.RevertCurrentBlock()
		}
	}()

	err = mp.checkAndUpdateContextBeforeExecution(header, headerHash)
	if err != nil {
		return nil, err
	}

	err = mp.createBlockStarted()
	if err != nil {
		return nil, err
	}

	err = mp.blockChainHook.SetCurrentHeader(header)
	if err != nil {
		return nil, err
	}

	err = mp.processIfFirstBlockAfterEpochStartBlockV3()
	if err != nil {
		return nil, err
	}

	var execResult data.BaseExecutionResultHandler
	if header.IsEpochChangeProposed() {
		// in case of error, will be picked up by the deferred revert
		execResult, err = mp.processEpochStartProposeBlock(header, body)
		return execResult, err
	}

	mp.txCoordinator.RequestBlockTransactions(body)
	mp.hdrsForCurrBlock.RequestShardHeaders(header)

	// although we can have a long time for processing, it being decoupled from consensus,
	// we still give some reasonable timeout
	proposalStartTime := time.Now()
	haveTime := getHaveTimeForProposal(proposalStartTime, mp.processConfigsHandler.GetMaxBlockProcessingTime(headerHandler.GetRound()))

	err = mp.txCoordinator.IsDataPreparedForProcessing(haveTime)
	if err != nil {
		return nil, err
	}

	err = mp.hdrsForCurrBlock.WaitForHeadersIfNeeded(haveTime)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	err = mp.txCoordinator.ProcessBlockTransaction(header, body, haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to process block transaction",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return nil, err
	}

	err = mp.txCoordinator.VerifyCreatedBlockTransactions(header, body)
	if err != nil {
		return nil, err
	}

	constructedBody := mp.createBlockBodyAfterExecution(body)
	err = mp.scToProtocol.UpdateProtocol(constructedBody, header.GetNonce())
	if err != nil {
		return nil, err
	}

	var valStatRootHash []byte
	valStatRootHash, err = mp.updateValidatorStatistics(header)
	if err != nil {
		return nil, err
	}

	// in case of error, will be picked up by the deferred revert
	execResult, err = mp.collectExecutionResults(headerHash, header, body, valStatRootHash)
	if err != nil {
		return nil, err
	}

	err = mp.blockProcessingCutoffHandler.HandleProcessErrorCutoff(header)
	if err != nil {
		return nil, err
	}

	return execResult, nil
}

// CommitBlockProposalState commits the accounts state after processing a block proposal
// and performs any post-commit operations (e.g. saving epoch start economics metrics).
func (mp *metaProcessor) CommitBlockProposalState(headerHandler data.HeaderHandler) error {
	if check.IfNil(headerHandler) {
		return process.ErrNilBlockHeader
	}

	// runs on the execution goroutine outside CommitBlock; the snapshot yields for the duration so it
	// cannot steal CPU/disk from this span - a latency concern only, trie nodes are immutable
	mp.processStatusHandler.BlockBackgroundJobs("metaProcessor.CommitBlockProposalState")
	defer mp.processStatusHandler.UnblockBackgroundJobs()

	mp.cleanupDismissedEWLEntries()

	err := mp.commitState(headerHandler)
	if err != nil {
		return err
	}

	metaHeader, ok := headerHandler.(data.MetaHeaderHandler)
	if ok {
		mp.saveEpochStartEconomicsMetricsV3IfNeeded(metaHeader)
	}

	return nil
}

// RevertBlockProposalState reverts the uncommitted accounts state after a block proposal processing failure
func (mp *metaProcessor) RevertBlockProposalState() {
	mp.RevertCurrentBlock()
}

func (mp *metaProcessor) checkNonceGaps(metaHeader data.MetaHeaderHandler) error {
	err := mp.checkHeaderExecutionResultNonceGap(metaHeader)
	if err != nil {
		return err
	}

	shardDataFinalizedNonces := make(map[uint32]uint64)

	// Initialize shardDataFinalizedNonces with data from block tracker
	lastCrossNotarizedForAllShards, err := mp.blockTracker.GetLastCrossNotarizedHeadersForAllShards()
	if err != nil {
		return err
	}

	// Get highest finalized nonce per shard from ShardInfoHandlers
	for _, shardData := range metaHeader.GetShardInfoHandlers() {
		shardID := shardData.GetShardID()
		nonce := shardData.GetNonce()

		existing, found := shardDataFinalizedNonces[shardID]
		if !found || nonce > existing {
			lastCrossNotarizedInBlockTracker, foundInTracker := lastCrossNotarizedForAllShards[shardID]
			if !foundInTracker {
				log.Warn("missing cross notarized header for shard in block tracker", "shard", shardID)
				return process.ErrMissingCrossNotarizedHeader
			}

			lastExecResultNonceOfLastCrossNotarized := common.GetLastExecutionResultNonce(lastCrossNotarizedInBlockTracker)
			if nonce < lastExecResultNonceOfLastCrossNotarized {
				log.Warn("found proposed nonce lower than last exec result of cross notarized",
					"shard", shardID,
					"shardInfoNonce", nonce,
					"lastExecResultNonceOfLastCrossNotarized", lastExecResultNonceOfLastCrossNotarized,
				)
				return process.ErrInvalidShardInfo
			}

			shardDataFinalizedNonces[shardID] = nonce
		}
	}

	// fill missing data from block tracker
	for shardID, lastCrossNotarizedInBlockTracker := range lastCrossNotarizedForAllShards {
		_, found := shardDataFinalizedNonces[shardID]
		if found {
			continue
		}

		shardDataFinalizedNonces[shardID] = common.GetLastExecutionResultNonce(lastCrossNotarizedInBlockTracker)
	}

	// Get highest proposed nonce per shard from ShardInfoProposalHandlers
	shardDataProposedNonces := make(map[uint32]uint64)
	for _, shardProposalData := range metaHeader.GetShardInfoProposalHandlers() {
		shardID := shardProposalData.GetShardID()
		nonce := shardProposalData.GetNonce()

		if existing, found := shardDataProposedNonces[shardID]; !found || nonce > existing {
			shardDataProposedNonces[shardID] = nonce
		}
	}

	// Check nonce gaps for each shard
	for shardID, maxProposedNonce := range shardDataProposedNonces {
		lastFinalizedNonce, found := shardDataFinalizedNonces[shardID]
		if !found {
			log.Warn("missing last notarized header for shard", "shard", shardID)
			return process.ErrMissingCrossNotarizedHeader
		}

		if maxProposedNonce < lastFinalizedNonce {
			return fmt.Errorf("%w: shard %d, last finalized nonce %d, proposed nonce %d",
				process.ErrInvalidProposedNonce,
				shardID,
				lastFinalizedNonce,
				maxProposedNonce)
		}

		nonceGap := maxProposedNonce - lastFinalizedNonce
		if nonceGap > mp.maxProposalNonceGap {
			return fmt.Errorf("%w: shard %d has nonce gap of %d between finalized nonce %d and proposed nonce %d, max allowed gap is %d",
				process.ErrNonceGapTooLarge,
				shardID,
				nonceGap,
				lastFinalizedNonce,
				maxProposedNonce,
				mp.maxProposalNonceGap)
		}
	}

	return nil
}

func (mp *metaProcessor) processEpochStartProposeBlock(
	metaHeader data.MetaHeaderHandler,
	body *block.Body,
) (data.BaseExecutionResultHandler, error) {
	if check.IfNil(metaHeader) {
		return nil, process.ErrNilBlockHeader
	}
	if body == nil {
		return nil, process.ErrNilBlockBody
	}
	if len(body.MiniBlocks) != 0 {
		return nil, process.ErrEpochStartProposeBlockHasMiniBlocks
	}

	log.Debug("processing epoch start propose block",
		"block epoch", metaHeader.GetEpoch(),
		"for epoch", metaHeader.GetEpoch()+1,
		"round", metaHeader.GetRound(),
		"nonce", metaHeader.GetNonce(),
	)

	err := mp.processEconomicsDataForEpochStartProposeBlock(metaHeader)
	if err != nil {
		return nil, err
	}

	var computedEconomics *block.Economics
	computedEconomics, err = mp.getComputedEconomics(metaHeader.GetEpoch() + 1)
	if err != nil {
		return nil, err
	}

	constructedBody, err := mp.processEpochStartMiniBlocks(metaHeader, computedEconomics)
	if err != nil {
		return nil, err
	}

	valStatRootHash, err := mp.updateValidatorStatistics(metaHeader)
	if err != nil {
		return nil, err
	}

	headerHash, err := core.CalculateHash(mp.marshalizer, mp.hasher, metaHeader)
	if err != nil {
		return nil, err
	}

	execResult, err := mp.collectExecutionResultsEpochStartProposal(headerHash, metaHeader, constructedBody, valStatRootHash)
	if err != nil {
		return nil, err
	}

	err = mp.blockProcessingCutoffHandler.HandleProcessErrorCutoff(metaHeader)
	if err != nil {
		return nil, err
	}

	return execResult, nil
}

func (mp *metaProcessor) saveEpochStartEconomicsMetricsV3IfNeeded(metaBlock data.MetaHeaderHandler) {
	if !metaBlock.IsHeaderV3() {
		// fee metrics for meta block will be handled on commit
		return
	}

	if !metaBlock.IsEpochChangeProposed() {
		return
	}

	lastExecutionResult := mp.blockChain.GetLastExecutionResult()
	if !bytes.Equal(lastExecutionResult.GetHeaderHash(), metaBlock.GetPrevHash()) {
		// should never happen, as this is called while processing proposeEpochChangeMetaBlock
		return
	}

	lastMetaExecutionResult, ok := lastExecutionResult.(data.BaseMetaExecutionResultHandler)
	if !ok {
		// should never happen
		return
	}

	mp.appStatusHandler.SetStringValue(common.MetricTotalFees, lastMetaExecutionResult.GetAccumulatedFeesInEpoch().String())
	mp.appStatusHandler.SetStringValue(common.MetricDevRewardsInEpoch, lastMetaExecutionResult.GetDevFeesInEpoch().String())
}

func (mp *metaProcessor) updateValidatorStatistics(header data.MetaHeaderHandler) ([]byte, error) {
	sw := core.NewStopWatch()
	sw.Start("UpdatePeerState")
	mp.prepareBlockHeaderInternalMapForValidatorProcessor(header)
	valStatRootHash, err := mp.updatePeerState(header, mp.hdrsForCurrBlock.GetHeadersMap())
	sw.Stop("UpdatePeerState")
	return valStatRootHash, err
}

func (mp *metaProcessor) collectExecutionResultsEpochStartProposal(
	headerHash []byte,
	header data.MetaHeaderHandler,
	constructedBody *block.Body,
	valStatRootHash []byte,
) (data.BaseExecutionResultHandler, error) {
	totalTxCount, miniBlockHeaderHandlers, err := mp.createMiniBlockHeaderHandlersForExecutionResults(constructedBody)
	if err != nil {
		return nil, err
	}

	receiptHash, err := mp.txCoordinator.CreateReceiptsHash()
	if err != nil {
		return nil, err
	}

	// we consider the rewards and peer mini blocks as post process mbs (post execution of start of epoch proposed block)
	err = mp.cacheIntraShardMiniBlocks(headerHash, constructedBody.MiniBlocks)
	if err != nil {
		return nil, err
	}

	err = mp.cacheExecutedMiniBlocks(&block.Body{MiniBlocks: constructedBody.MiniBlocks}, miniBlockHeaderHandlers)
	if err != nil {
		return nil, err
	}

	return mp.createExecutionResult(miniBlockHeaderHandlers, header, headerHash, receiptHash, valStatRootHash, totalTxCount)
}

// collectExecutionResults collects the execution results after processing the block
func (mp *metaProcessor) collectExecutionResults(
	headerHash []byte,
	header data.MetaHeaderHandler,
	body *block.Body,
	valStatRootHash []byte,
) (data.BaseExecutionResultHandler, error) {
	miniBlockHeaderHandlers, totalTxCount, receiptHash, err := mp.collectMiniBlocks(headerHash, body)
	if err != nil {
		return nil, err
	}

	return mp.createExecutionResult(miniBlockHeaderHandlers, header, headerHash, receiptHash, valStatRootHash, totalTxCount)
}

func (mp *metaProcessor) createExecutionResult(
	miniBlockHeaderHandlers []data.MiniBlockHeaderHandler,
	header data.MetaHeaderHandler,
	headerHash []byte,
	receiptHash []byte,
	valStatRootHash []byte,
	totalTxCount int,
) (data.BaseExecutionResultHandler, error) {
	gasAndFees := mp.getGasAndFees()
	gasNotUsedForProcessing := gasAndFees.GetGasPenalized() + gasAndFees.GetGasRefunded()
	if gasAndFees.GetGasProvided() < gasNotUsedForProcessing {
		return nil, process.ErrGasUsedExceedsGasProvided
	}

	gasUsed := gasAndFees.GetGasProvided() - gasNotUsedForProcessing // needed for inclusion estimation

	accumulatedFeesInEpoch, devFeesInEpoch, err := mp.computeAccumulatedFeesInEpoch(header)
	if err != nil {
		return nil, err
	}

	executionResult := &block.MetaExecutionResult{
		ExecutionResult: &block.BaseMetaExecutionResult{
			BaseExecutionResult: &block.BaseExecutionResult{
				HeaderHash:  headerHash,
				HeaderNonce: header.GetNonce(),
				HeaderRound: header.GetRound(),
				HeaderEpoch: header.GetEpoch(),
				RootHash:    mp.getRootHash(),
				GasUsed:     gasUsed,
			},
			ValidatorStatsRootHash: valStatRootHash,
			AccumulatedFeesInEpoch: accumulatedFeesInEpoch,
			DevFeesInEpoch:         devFeesInEpoch,
		},
		ReceiptsHash:    receiptHash,
		DeveloperFees:   gasAndFees.GetDeveloperFees(),
		AccumulatedFees: gasAndFees.GetAccumulatedFees(),
		ExecutedTxCount: uint64(totalTxCount),
	}

	err = executionResult.SetMiniBlockHeadersHandlers(miniBlockHeaderHandlers)
	if err != nil {
		return nil, err
	}

	logs := mp.txCoordinator.GetAllCurrentLogs()
	err = mp.cacheLogEvents(headerHash, logs)
	if err != nil {
		return nil, err
	}

	err = mp.cacheIntermediateTxsForHeader(headerHash)
	if err != nil {
		return nil, err
	}

	mp.cacheOrderedTxHashes(headerHash)
	mp.cacheUnexecutableTxHashes(headerHash)
	mp.cacheHeaderGasData(headerHash)

	return executionResult, nil
}

func getTxCountExecutionResults(metaHeader data.MetaHeaderHandler) (uint32, error) {
	if check.IfNil(metaHeader) {
		return 0, nil
	}

	totalTxs := uint64(0)
	execResults := metaHeader.GetExecutionResultsHandlers()
	for _, execResult := range execResults {
		execResultsMeta, ok := execResult.(data.MetaExecutionResultHandler)
		if !ok {
			return 0, process.ErrWrongTypeAssertion
		}
		totalTxs += execResultsMeta.GetExecutedTxCount()
	}
	return uint32(totalTxs), nil
}

func (mp *metaProcessor) hasStartOfEpochExecutionResults(metaHeader data.MetaHeaderHandler) (bool, error) {
	if check.IfNil(metaHeader) {
		return false, process.ErrNilHeaderHandler
	}
	execResults := metaHeader.GetExecutionResultsHandlers()
	for _, execResult := range execResults {
		ok, err := mp.hasRewardOrPeerMiniBlocksOnExecResult(execResult)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func (mp *metaProcessor) hasRewardOrPeerMiniBlocksOnExecResult(execResult data.BaseExecutionResultHandler) (bool, error) {
	mbHeaders, err := common.GetMiniBlocksHeaderHandlersFromExecResult(execResult)
	if err != nil {
		return false, err
	}

	return hasRewardOrPeerMiniBlocksFromMeta(mbHeaders), nil
}

func hasRewardOrPeerMiniBlocksFromMeta(miniBlockHeaders []data.MiniBlockHeaderHandler) bool {
	for _, mbHeader := range miniBlockHeaders {
		if mbHeader.GetSenderShardID() != common.MetachainShardId {
			continue
		}
		if mbHeader.GetTypeInt32() == int32(block.RewardsBlock) ||
			mbHeader.GetTypeInt32() == int32(block.PeerBlock) {
			return true
		}
	}
	return false
}

func (mp *metaProcessor) createBlockBodyProposal(
	metaHdr data.MetaHeaderHandler,
	haveTime func() bool,
) error {
	mp.blockSizeThrottler.ComputeCurrentMaxSize()

	log.Debug("started creating block body",
		"epoch", metaHdr.GetEpoch(),
		"round", metaHdr.GetRound(),
		"nonce", metaHdr.GetNonce(),
	)

	return mp.createProposalMiniBlocks(metaHdr.GetRound(), haveTime)
}

func (mp *metaProcessor) createProposalMiniBlocks(round uint64, haveTime func() bool) error {
	if !haveTime() {
		log.Debug("metaProcessor.createProposalMiniBlocks", "error", process.ErrTimeIsOut)
		return nil
	}

	startTime := time.Now()
	err := mp.selectIncomingMiniBlocksForProposal(round, haveTime)
	if err != nil {
		return err
	}
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to create mbs to me", "time", elapsedTime)

	return nil
}

func (mp *metaProcessor) selectIncomingMiniBlocksForProposal(
	round uint64,
	haveTime func() bool,
) error {
	sw := core.NewStopWatch()
	sw.Start("ComputeLongestShardsChainsFromLastNotarized")
	orderedHdrs, orderedHdrsHashes, _, err := mp.blockTracker.ComputeLongestShardsChainsFromLastNotarized()
	sw.Stop("ComputeLongestShardsChainsFromLastNotarized")
	log.Debug("measurements ComputeLongestShardsChainsFromLastNotarized", sw.GetMeasurements()...)
	if err != nil {
		return err
	}

	log.Debug("shard headers ordered",
		"num shard headers", len(orderedHdrs),
	)

	lastShardHdrs, err := mp.getLastCrossNotarizedShardHeaders()
	if err != nil {
		return err
	}

	maxShardHeadersFromSameShard := core.MaxUint32(
		process.MinShardHeadersFromSameShardInOneMetaBlock,
		process.MaxShardHeadersAllowedInOneMetaBlock/mp.shardCoordinator.NumberOfShards(),
	)
	ancestryView := mp.newProposalAncestryView()
	hdrsAddedForShard, err := mp.selectIncomingMiniBlocks(lastShardHdrs, orderedHdrs, orderedHdrsHashes, maxShardHeadersFromSameShard, ancestryView, haveTime)
	if err != nil {
		return err
	}

	err = mp.selectContendedShardHeaders(round, lastShardHdrs, hdrsAddedForShard, ancestryView, haveTime)
	if err != nil {
		return err
	}

	// spawned only after every selection step stopped mutating lastShardHdrs
	go mp.requestShardHeadersInAdvanceIfNeeded(lastShardHdrs)

	return nil
}

func (mp *metaProcessor) selectIncomingMiniBlocks(
	lastShardHdrs map[uint32]ShardHeaderInfo,
	orderedHdrs []data.HeaderHandler,
	orderedHdrsHashes [][]byte,
	maxShardHeadersFromSameShard uint32,
	ancestryView *metaAncestryView,
	haveTime func() bool,
) (map[uint32]uint32, error) {
	hdrsAdded := uint32(0)
	maxShardHeadersAllowedInOneMetaBlock := maxShardHeadersFromSameShard * mp.shardCoordinator.NumberOfShards()
	hdrsAddedForShard := make(map[uint32]uint32)

	if len(orderedHdrs) != len(orderedHdrsHashes) {
		return nil, process.ErrInconsistentShardHeadersAndHashes
	}

	for i := 0; i < len(orderedHdrs); i++ {
		if !haveTime() {
			log.Debug("time is up after putting cross txs with destination to  metachain",
				"num txs", mp.miniBlocksSelectionSession.GetNumTxsAdded(),
			)
			break
		}

		if hdrsAdded >= maxShardHeadersAllowedInOneMetaBlock {
			log.Debug("maximum shard headers allowed to be included in one meta block has been reached",
				"shard headers added", hdrsAdded,
			)
			break
		}

		currHdr := orderedHdrs[i]
		currHdrHash := orderedHdrsHashes[i]
		lastShardHeaderInfo, ok := lastShardHdrs[currHdr.GetShardID()]
		if !ok {
			return nil, process.ErrMissingHeader
		}
		if currHdr.GetNonce() != lastShardHeaderInfo.Header.GetNonce()+1 {
			log.Trace("skip searching",
				"shard", currHdr.GetShardID(),
				"last shard hdr nonce", lastShardHeaderInfo.Header.GetNonce(),
				"curr shard hdr nonce", currHdr.GetNonce())
			continue
		}

		if hdrsAddedForShard[currHdr.GetShardID()] >= maxShardHeadersFromSameShard {
			log.Trace("maximum shard headers from same shard allowed to be included in one meta block has been reached",
				"shard", currHdr.GetShardID(),
				"shard headers added", hdrsAddedForShard[currHdr.GetShardID()],
			)
			continue
		}

		needsProof := common.IsProofsFlagEnabledForHeader(mp.enableEpochsHandler, currHdr)
		if needsProof && !mp.proofsPool.HasProof(currHdr.GetShardID(), currHdrHash) {
			log.Trace("no proof for shard header",
				"shard", currHdr.GetShardID(),
				"hash", logger.DisplayByteSlice(currHdrHash),
			)
			continue
		}

		errAncestry := mp.checkReferencedMetaAncestry(currHdr, ancestryView)
		if errAncestry != nil {
			log.Trace("shard header skipped on referenced meta ancestry",
				"shard", currHdr.GetShardID(),
				"hash", logger.DisplayByteSlice(currHdrHash),
				"error", errAncestry,
			)
			continue
		}

		added, err := mp.addShardHeaderToSelection(currHdr, currHdrHash, lastShardHdrs)
		if err != nil {
			return nil, err
		}
		if !added {
			break
		}

		hdrsAddedForShard[currHdr.GetShardID()]++
		hdrsAdded++
	}

	return hdrsAddedForShard, nil
}

// addShardHeaderToSelection includes the shard header in the current selection session; it returns
// false when the header cannot be fully added and the selection should stop
func (mp *metaProcessor) addShardHeaderToSelection(
	currHdr data.HeaderHandler,
	currHdrHash []byte,
	lastShardHdrs map[uint32]ShardHeaderInfo,
) (bool, error) {
	if len(currHdr.GetMiniBlockHeadersWithDst(mp.shardCoordinator.SelfId())) > 0 {
		createIncomingMbsResult, errCreated := mp.createMbsCrossShardDstMe(currHdrHash, currHdr, nil)
		if errCreated != nil {
			return false, errCreated
		}
		if !createIncomingMbsResult.HeaderFinished {
			mp.revertGasForCrossShardDstMeMiniBlocks(createIncomingMbsResult.AddedMiniBlocks, createIncomingMbsResult.PendingMiniBlocks)
			log.Debug("shard header cannot be fully added",
				"round", currHdr.GetRound(),
				"nonce", currHdr.GetNonce(),
				"hash", currHdrHash)
			return false, nil
		}

		if len(createIncomingMbsResult.AddedMiniBlocks) > 0 {
			err := mp.miniBlocksSelectionSession.AddMiniBlocksAndHashes(createIncomingMbsResult.AddedMiniBlocks)
			if err != nil {
				return false, err
			}
		}
	}

	mp.miniBlocksSelectionSession.AddReferencedHeader(currHdr, currHdrHash)
	lastShardHdrs[currHdr.GetShardID()] = ShardHeaderInfo{
		Header:      currHdr,
		Hash:        currHdrHash,
		UsedInBlock: true,
	}

	return true, nil
}

// selectContendedShardHeaders arbitrates shards stalled on a contended nonce: past the discovery
// window from the candidate's round, the lowest-(round,hash) proofed extender is included
func (mp *metaProcessor) selectContendedShardHeaders(
	round uint64,
	lastShardHdrs map[uint32]ShardHeaderInfo,
	hdrsAddedForShard map[uint32]uint32,
	ancestryView *metaAncestryView,
	haveTime func() bool,
) error {
	if !mp.enableEpochsHandler.IsFlagEnabled(common.SupernovaFlag) {
		return nil
	}

	for shardID := uint32(0); shardID < mp.shardCoordinator.NumberOfShards(); shardID++ {
		if !haveTime() {
			return nil
		}
		if hdrsAddedForShard[shardID] > 0 {
			continue
		}

		lastShardHeaderInfo, ok := lastShardHdrs[shardID]
		if !ok {
			continue
		}

		candidate, candidateHash := mp.getArbitrationCandidate(lastShardHeaderInfo, ancestryView)
		if check.IfNil(candidate) {
			continue
		}
		if !common.IsContendedHeader(candidate, lastShardHeaderInfo.Header) {
			continue
		}
		if round < candidate.GetRound()+metaArbitrationWindowRounds {
			log.Debug("selectContendedShardHeaders: holding contended shard header for proof discovery",
				"shard", shardID,
				"nonce", candidate.GetNonce(),
				"candidate round", candidate.GetRound(),
				"current round", round)
			continue
		}

		added, err := mp.addShardHeaderToSelection(candidate, candidateHash, lastShardHdrs)
		if err != nil {
			return err
		}
		if !added {
			continue
		}

		log.Debug("selectContendedShardHeaders: included contended shard header after discovery window",
			"shard", shardID,
			"nonce", candidate.GetNonce(),
			"round", candidate.GetRound(),
			"hash", candidateHash)
	}

	return nil
}

// contentionContext is the enclosing meta block's committed round and its lazily resolved network
// verdict: once the block carries its own proof, the subjective competitor check is superseded
type contentionContext struct {
	metaRound    uint64
	isOwnProofed func() bool
}

func (mp *metaProcessor) newContentionContext(metaHeader data.HeaderHandler) contentionContext {
	return contentionContext{
		metaRound:    metaHeader.GetRound(),
		isOwnProofed: mp.ownProofResolver(metaHeader),
	}
}

// checkShardHeaderContention gates a contended unsettled shard header by regime: a proofed meta
// block is the network verdict and passes; an unproofed one must honor the discovery window
// (committed rounds, deterministic) and beats any locally actionable better competitor
func (mp *metaProcessor) checkShardHeaderContention(header data.HeaderHandler, headerHash []byte, parentInfo ShardHeaderInfo, ancestryView *metaAncestryView, contentionCtx contentionContext) error {
	if !mp.isContendedUnsettledCrossHeader(header, parentInfo.Header, headerHash) {
		return nil
	}

	if contentionCtx.isOwnProofed() {
		return nil
	}

	if contentionCtx.metaRound < header.GetRound()+metaArbitrationWindowRounds {
		return fmt.Errorf("%w with hash %x", errContendedHeaderInsideArbitrationWindow, headerHash)
	}

	candidate, candidateHash := mp.getArbitrationCandidate(parentInfo, ancestryView)
	if check.IfNil(candidate) || bytes.Equal(candidateHash, headerHash) {
		return nil
	}

	isBetter := candidate.GetRound() < header.GetRound() ||
		(candidate.GetRound() == header.GetRound() && bytes.Compare(candidateHash, headerHash) < 0)
	if isBetter {
		return fmt.Errorf("%w with hash %x, competitor hash %x", errContendedHeaderWithBetterCompetitor, headerHash, candidateHash)
	}

	return nil
}

// checkShardHeaderContentionComputingHash computes the hashes only on the contended path
func (mp *metaProcessor) checkShardHeaderContentionComputingHash(header data.HeaderHandler, parentHeader data.HeaderHandler, ancestryView *metaAncestryView, contentionCtx contentionContext) error {
	if !mp.enableEpochsHandler.IsFlagEnabledInEpoch(common.SupernovaFlag, header.GetEpoch()) {
		return nil
	}
	if !common.IsContendedHeader(header, parentHeader) {
		return nil
	}

	headerHash, err := mp.getHeaderHash(header)
	if err != nil {
		return err
	}
	parentHash, err := mp.getHeaderHash(parentHeader)
	if err != nil {
		return err
	}

	return mp.checkShardHeaderContention(header, headerHash, ShardHeaderInfo{Header: parentHeader, Hash: parentHash}, ancestryView, contentionCtx)
}

// getArbitrationCandidate returns the lowest-(round,hash) proofed, construction-valid header at the
// nonce right after the given one and extending it, or nil when none is actionable locally
func (mp *metaProcessor) getArbitrationCandidate(parentInfo ShardHeaderInfo, ancestryView *metaAncestryView) (data.HeaderHandler, []byte) {
	shardID := parentInfo.Header.GetShardID()
	nonce := parentInfo.Header.GetNonce() + 1

	headers, hashes, err := mp.dataPool.Headers().GetHeadersByNonceAndShardId(nonce, shardID)
	if err != nil {
		return nil, nil
	}

	var best data.HeaderHandler
	var bestHash []byte
	for i, header := range headers {
		if check.IfNil(header) || !bytes.Equal(header.GetPrevHash(), parentInfo.Hash) {
			continue
		}
		needsProof := common.IsProofsFlagEnabledForHeader(mp.enableEpochsHandler, header)
		if needsProof && !mp.proofsPool.HasProof(shardID, hashes[i]) {
			continue
		}
		errValidity := mp.headerValidator.IsHeaderConstructionValid(header, parentInfo.Header)
		if errValidity != nil {
			continue
		}
		if mp.checkReferencedMetaAncestry(header, ancestryView) != nil {
			continue
		}

		isBetter := check.IfNil(best) ||
			header.GetRound() < best.GetRound() ||
			(header.GetRound() == best.GetRound() && bytes.Compare(hashes[i], bestHash) < 0)
		if isBetter {
			best = header
			bestHash = hashes[i]
		}
	}

	return best, bestHash
}

// metaAncestryView is a per proposal hash index over the ancestors of the meta block being built:
// a lazy pool walk above the pool horizon and a canonical storer hash cache below it, single threaded
type metaAncestryView struct {
	walked          map[uint64][]byte
	walkedHashes    map[string]struct{}
	lowestWalked    uint64
	parentNonce     uint64
	cursor          []byte
	walkFrozen      bool
	canonicalHashes map[string]struct{}
	coveredNonces   map[uint64]struct{}
}

func newMetaAncestryView(parentHeader data.HeaderHandler, parentHash []byte) *metaAncestryView {
	return &metaAncestryView{
		walked:          map[uint64][]byte{parentHeader.GetNonce(): parentHash},
		walkedHashes:    map[string]struct{}{string(parentHash): {}},
		lowestWalked:    parentHeader.GetNonce(),
		parentNonce:     parentHeader.GetNonce(),
		cursor:          parentHeader.GetPrevHash(),
		canonicalHashes: make(map[string]struct{}),
		coveredNonces:   make(map[uint64]struct{}),
	}
}

// newProposalAncestryView anchors the ancestor test at the head the proposal extends, matching the
// prev hash chain a validator sees on the received proposal
func (mp *metaProcessor) newProposalAncestryView() *metaAncestryView {
	if !mp.enableEpochsHandler.IsFlagEnabled(common.SupernovaFlag) {
		return nil
	}
	if check.IfNil(mp.blockChain) {
		return nil
	}

	parentHeader := mp.blockChain.GetCurrentBlockHeader()
	parentHash := mp.blockChain.GetCurrentBlockHeaderHash()
	if check.IfNil(parentHeader) {
		parentHeader = mp.blockChain.GetGenesisHeader()
		parentHash = mp.blockChain.GetGenesisHeaderHash()
	}
	if check.IfNil(parentHeader) {
		return nil
	}

	return newMetaAncestryView(parentHeader, parentHash)
}

func (mp *metaProcessor) newVerifyAncestryView(metaHeaderHandler data.MetaHeaderHandler) (*metaAncestryView, error) {
	if !mp.enableEpochsHandler.IsFlagEnabled(common.SupernovaFlag) {
		return nil, nil
	}

	prevHash := metaHeaderHandler.GetPrevHash()
	if !check.IfNil(mp.blockChain) {
		genesisHeader := mp.blockChain.GetGenesisHeader()
		if !check.IfNil(genesisHeader) && bytes.Equal(prevHash, mp.blockChain.GetGenesisHeaderHash()) {
			return newMetaAncestryView(genesisHeader, prevHash), nil
		}
	}

	parentHeader, err := process.GetMetaHeader(prevHash, mp.dataPool.Headers(), mp.marshalizer, mp.store)
	if err != nil {
		return nil, fmt.Errorf("%w : newVerifyAncestryView", err)
	}

	return newMetaAncestryView(parentHeader, prevHash), nil
}

// checkReferencedMetaAncestry rejects, fail-closed, a shard header referencing any meta block that
// is not an ancestor of the block being built; canonical references always resolve on the builder
func (mp *metaProcessor) checkReferencedMetaAncestry(header data.HeaderHandler, view *metaAncestryView) error {
	if !mp.enableEpochsHandler.IsFlagEnabled(common.SupernovaFlag) {
		return nil
	}

	shardHeader, ok := header.(data.ShardHeaderHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	metaHashes := shardHeader.GetMetaBlockHashes()
	if len(metaHashes) == 0 {
		return nil
	}
	if view == nil {
		return errNilMetaAncestryView
	}

	for idx, metaHash := range metaHashes {
		if !mp.isAncestorMetaBlock(view, metaHash, len(metaHashes)-idx) {
			return fmt.Errorf("%w with hash %x", errReferencedNonAncestorMetaHeader, metaHash)
		}
	}

	return nil
}

// isAncestorMetaBlock answers by hash set membership when possible; a miss resolves the reference
// once and extends the matching region, so following references of the same run answer with no reads
func (mp *metaProcessor) isAncestorMetaBlock(view *metaAncestryView, refHash []byte, refsLeft int) bool {
	if _, isWalked := view.walkedHashes[string(refHash)]; isWalked {
		return true
	}
	if _, isCanonical := view.canonicalHashes[string(refHash)]; isCanonical {
		return true
	}

	refHeader, err := process.GetMetaHeader(refHash, mp.dataPool.Headers(), mp.marshalizer, mp.store)
	if err != nil {
		return false
	}

	refNonce := refHeader.GetNonce()
	if refNonce > view.parentNonce {
		return false
	}

	mp.extendAncestryWalk(view, refNonce)
	if refNonce >= view.lowestWalked {
		return bytes.Equal(view.walked[refNonce], refHash)
	}

	// below the ancestry chain created through the pool walk, the canonical nonce -> hash storer, is the ancestor test
	mp.extendCanonicalHashes(view, refNonce, refsLeft)
	_, isCanonical := view.canonicalHashes[string(refHash)]

	return isCanonical
}

// extendAncestryWalk steps down the prev hash chain through the pool only; below the pool horizon
// the canonical storer cache takes over
func (mp *metaProcessor) extendAncestryWalk(view *metaAncestryView, downToNonce uint64) {
	if view.walkFrozen {
		return
	}

	for view.lowestWalked > downToNonce && view.lowestWalked > 0 {
		header, err := mp.dataPool.Headers().GetHeaderByHash(view.cursor)
		if err != nil || check.IfNil(header) || header.GetNonce() >= view.lowestWalked {
			return
		}

		view.walked[header.GetNonce()] = view.cursor
		view.walkedHashes[string(view.cursor)] = struct{}{}
		view.lowestWalked = header.GetNonce()
		view.cursor = header.GetPrevHash()
	}
}

// extendCanonicalHashes sweeps the canonical storer over the run the current shard header still
// claims, each nonce read at most once; the walk freezes so the two regions cannot overlap
func (mp *metaProcessor) extendCanonicalHashes(view *metaAncestryView, fromNonce uint64, refsLeft int) {
	view.walkFrozen = true

	if refsLeft < 1 {
		return
	}
	if refsLeft > process.MaxMetaHeadersAllowedInOneShardBlock {
		refsLeft = process.MaxMetaHeadersAllowedInOneShardBlock
	}
	for nonce := fromNonce; nonce < fromNonce+uint64(refsLeft) && nonce < view.lowestWalked; nonce++ {
		if _, isCovered := view.coveredNonces[nonce]; isCovered {
			continue
		}
		view.coveredNonces[nonce] = struct{}{}

		storedHash, err := process.GetHeaderHashFromStorageWithNonce(nonce, mp.store, mp.uint64Converter, mp.marshalizer, dataRetriever.MetaHdrNonceHashDataUnit)
		if err != nil {
			continue
		}
		view.canonicalHashes[string(storedHash)] = struct{}{}
	}
}

func (mp *metaProcessor) requestShardHeadersInAdvanceIfNeeded(
	lastShardHdr map[uint32]ShardHeaderInfo,
) {
	for shardID := uint32(0); shardID < mp.shardCoordinator.NumberOfShards(); shardID++ {
		mp.requestHeadersFromHeaderIfNeeded(lastShardHdr[shardID].Header)
	}
}

func (mp *metaProcessor) verifyEpochStartData(
	headerHandler data.MetaHeaderHandler,
) bool {
	epochStartData, err := mp.getComputedEpochStartData(headerHandler.GetEpoch())
	if err != nil {
		// only an epoch start header needs the data; for any other header the result is discarded
		if headerHandler.IsStartOfEpochBlock() {
			log.Error("verifyEpochStartData: failed to get epoch start data", "error", err)
		} else {
			log.Debug("verifyEpochStartData: no epoch start data for header epoch", "error", err)
		}
		return false
	}

	return epochStartData.Equal(headerHandler.GetEpochStartHandler())
}

func (mp *metaProcessor) checkEpochCorrectnessV3(
	headerHandler data.MetaHeaderHandler,
) error {
	currentBlockHeader := mp.blockChain.GetCurrentBlockHeader()
	if check.IfNil(currentBlockHeader) {
		return nil
	}

	hasEpochStartExecutionResults, err := mp.hasStartOfEpochExecutionResults(headerHandler)
	if err != nil {
		return err
	}

	wasEpochStartProposed, err := mp.hasExecutionResultsForProposedEpochChange(headerHandler)
	if err != nil {
		return err
	}

	isEpochStartBlock := headerHandler.IsStartOfEpochBlock()

	epochStartDataMatches := mp.verifyEpochStartData(headerHandler)
	hasAllEpochStartData := hasEpochStartExecutionResults && isEpochStartBlock && wasEpochStartProposed && epochStartDataMatches
	hasAnyEpochStartData := hasEpochStartExecutionResults || isEpochStartBlock || wasEpochStartProposed
	hasIncompleteEpochStartData := hasAnyEpochStartData && !hasAllEpochStartData

	if hasIncompleteEpochStartData {
		log.Warn("block has incomplete epoch start data",
			"hasEpochStartExecutionResults", hasEpochStartExecutionResults,
			"isEpochStartBlock", isEpochStartBlock,
			"wasEpochStartProposed", wasEpochStartProposed,
			"epochStartTrigger", mp.epochStartTrigger.Epoch())
		return process.ErrEpochDoesNotMatch
	}

	isEpochIncorrect := headerHandler.GetEpoch() != currentBlockHeader.GetEpoch() && !hasAllEpochStartData
	if isEpochIncorrect {
		log.Warn("block does not have epoch start results but epoch has changed",
			"currentHeaderEpoch", currentBlockHeader.GetEpoch(),
			"receivedHeaderEpoch", headerHandler.GetEpoch(),
			"epochStartTrigger", mp.epochStartTrigger.Epoch())
		return process.ErrEpochDoesNotMatch
	}

	isEpochIncorrect = headerHandler.GetEpoch() == currentBlockHeader.GetEpoch() && hasAllEpochStartData
	if isEpochIncorrect {
		log.Warn("block has epoch start results but epoch did not change",
			"currentHeaderEpoch", currentBlockHeader.GetEpoch(),
			"receivedHeaderEpoch", headerHandler.GetEpoch(),
			"epochStartTrigger", mp.epochStartTrigger.Epoch())
		return process.ErrEpochDoesNotMatch
	}

	isEpochIncorrect = headerHandler.GetEpoch() != currentBlockHeader.GetEpoch()+1 && hasAllEpochStartData
	if isEpochIncorrect {
		log.Warn("block did not correctly change epoch, with proposed epoch change",
			"currentHeaderEpoch", currentBlockHeader.GetEpoch(),
			"receivedHeaderEpoch", headerHandler.GetEpoch(),
			"epochStartTrigger", mp.epochStartTrigger.Epoch())
		return process.ErrEpochDoesNotMatch
	}

	return nil
}

func (mp *metaProcessor) hasExecutionResultsForProposedEpochChange(headerHandler data.MetaHeaderHandler) (bool, error) {
	executionResults := headerHandler.GetExecutionResultsHandlers()
	var header data.HeaderHandler
	var err error

	for _, execResult := range executionResults {
		header, err = process.GetHeader(
			execResult.GetHeaderHash(),
			mp.dataPool.Headers(),
			mp.store,
			mp.marshalizer,
			headerHandler.GetShardID(),
		)
		if err != nil {
			log.Debug("hasExecutionResultsForProposedEpochChange: could not find header",
				"hash", execResult.GetHeaderHash(),
			)
			return false, err
		}
		metaHeaderHandler, ok := header.(data.MetaHeaderHandler)
		if !ok {
			return false, process.ErrWrongTypeAssertion
		}

		isEpochChangeProposed := metaHeaderHandler.IsEpochChangeProposed()
		hasStartOfEpochOnExecutionResult, err := mp.hasRewardOrPeerMiniBlocksOnExecResult(execResult)
		if err != nil {
			return false, err
		}

		if isEpochChangeProposed && !hasStartOfEpochOnExecutionResult {
			return false, process.ErrStartOfEpochExecutionResultsDoNotExist
		}

		if isEpochChangeProposed {
			return true, nil
		}
	}

	return false, nil
}

func (mp *metaProcessor) checkShardHeadersValidityAndFinalityProposal(
	metaHeaderHandler data.MetaHeaderHandler,
) error {
	lastCrossNotarizedHeader, err := mp.getLastCrossNotarizedShardHeaders()
	if err != nil {
		return err
	}

	usedShardHeaders, err := mp.getShardHeadersFromMetaHeader(metaHeaderHandler)
	if err != nil {
		return fmt.Errorf("%w : checkShardHeadersValidityAndFinalityProposal -> getShardHeadersFromMetaHeader", err)
	}

	shouldNotHaveShardHeaders := metaHeaderHandler.IsStartOfEpochBlock() || metaHeaderHandler.IsEpochChangeProposed() || mp.epochStartTrigger.GetEpochChangeProposed()
	if len(usedShardHeaders.orderedShardHeaders) > 0 && shouldNotHaveShardHeaders {
		return fmt.Errorf("%w : between epoch change proposed and epoch start block", process.ErrShardHeadersShouldNotBeNotarized)
	}

	ok := mp.hasProofsForHeaders(usedShardHeaders.headersPerShard)
	if !ok {
		return process.ErrMissingHeaderProof
	}

	ancestryView, err := mp.newVerifyAncestryView(metaHeaderHandler)
	if err != nil {
		return fmt.Errorf("%w : checkShardHeadersValidityAndFinalityProposal", err)
	}

	err = mp.verifyUsedShardHeadersValidity(usedShardHeaders.headersPerShard, lastCrossNotarizedHeader, ancestryView, mp.newContentionContext(metaHeaderHandler))
	if err != nil {
		return fmt.Errorf("%w : checkShardHeadersValidityAndFinalityProposal -> verifyUsedShardHeadersValidity", err)
	}

	return mp.checkShardInfoValidity(metaHeaderHandler, usedShardHeaders)
}

func (mp *metaProcessor) checkShardInfoValidity(metaHeaderHandler data.MetaHeaderHandler, usedShardHeadersInfo *usedShardHeadersInfo) error {
	createdShardInfoProposal, createdShardInfo, err := mp.shardInfoCreateData.CreateShardInfoV3(metaHeaderHandler, usedShardHeadersInfo.orderedShardHeaders, usedShardHeadersInfo.orderedShardHeaderHashes)
	if err != nil {
		return fmt.Errorf("%w : checkShardInfoValidity -> CreateShardInfoV3", err)
	}

	headerShardInfo := metaHeaderHandler.GetShardInfoHandlers()
	headerShardInfoProposal := metaHeaderHandler.GetShardInfoProposalHandlers()
	if len(createdShardInfo) != len(headerShardInfo) || len(createdShardInfoProposal) != len(headerShardInfoProposal) {
		return process.ErrHeaderShardDataMismatch
	}

	for i := 0; i < len(headerShardInfo); i++ {
		if !headerShardInfo[i].Equal(createdShardInfo[i]) {
			return fmt.Errorf("%w for shardInfo item %d", process.ErrHeaderShardDataMismatch, i)
		}
	}
	for i := 0; i < len(headerShardInfoProposal); i++ {
		if !headerShardInfoProposal[i].Equal(createdShardInfoProposal[i]) {
			return fmt.Errorf("%w for shardInfoProposal item %d", process.ErrHeaderShardDataMismatch, i)
		}
	}

	return nil
}

func (mp *metaProcessor) verifyUsedShardHeadersValidity(
	usedShardHeaders map[uint32][]ShardHeaderInfo,
	lastCrossNotarizedHeader map[uint32]ShardHeaderInfo,
	ancestryView *metaAncestryView,
	contentionCtx contentionContext,
) error {
	var err error
	for shardID, hdrsForShard := range usedShardHeaders {
		err = mp.checkHeadersSequenceCorrectness(hdrsForShard, lastCrossNotarizedHeader[shardID], ancestryView, contentionCtx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mp *metaProcessor) checkHeadersSequenceCorrectness(
	hdrsForShard []ShardHeaderInfo,
	lastNotarizedHeaderInfoForShard ShardHeaderInfo,
	ancestryView *metaAncestryView,
	contentionCtx contentionContext,
) error {
	var err error
	for _, shardHdrInfo := range hdrsForShard {
		if mp.isGenesisShardBlockAndFirstMeta(shardHdrInfo.Header.GetNonce()) {
			continue
		}

		err = mp.checkShardHeaderContention(shardHdrInfo.Header, shardHdrInfo.Hash, lastNotarizedHeaderInfoForShard, ancestryView, contentionCtx)
		if err != nil {
			return err
		}

		err = mp.checkReferencedMetaAncestry(shardHdrInfo.Header, ancestryView)
		if err != nil {
			return err
		}

		err = mp.headerValidator.IsHeaderConstructionValid(shardHdrInfo.Header, lastNotarizedHeaderInfoForShard.Header)
		if err != nil {
			return err
		}

		lastNotarizedHeaderInfoForShard = shardHdrInfo
	}

	return nil
}

func (mp *metaProcessor) hasProofsForHeaders(headersPerShard map[uint32][]ShardHeaderInfo) bool {
	for _, headersForShard := range headersPerShard {
		for _, headerInfo := range headersForShard {
			if !common.IsProofsFlagEnabledForHeader(mp.enableEpochsHandler, headerInfo.Header) {
				continue
			}
			if !mp.proofsPool.HasProof(headerInfo.Header.GetShardID(), headerInfo.Hash) {
				log.Debug("missing proof for shard header", "shard", headerInfo.Header.GetShardID(), "headerHash", headerInfo.Hash)
				return false
			}
		}
	}
	return true
}

func (mp *metaProcessor) getShardHeadersFromMetaHeader(
	metaHeaderHandler data.MetaHeaderHandler,
) (*usedShardHeadersInfo, error) {
	shardInfoProposalHandlers := metaHeaderHandler.GetShardInfoProposalHandlers()
	usedShardHeaders := make(map[uint32][]ShardHeaderInfo)
	var err error
	var header data.HeaderHandler
	orderedShardHeaders := make([]data.HeaderHandler, 0, len(shardInfoProposalHandlers))
	orderedShardHeaderHashes := make([][]byte, 0, len(shardInfoProposalHandlers))
	for _, shardInfoHandler := range shardInfoProposalHandlers {
		header, err = process.GetHeader(
			shardInfoHandler.GetHeaderHash(),
			mp.dataPool.Headers(),
			mp.store,
			mp.marshalizer,
			shardInfoHandler.GetShardID(),
		)
		if err != nil {
			log.Debug("getShardHeadersFromMetaHeader: could not find header",
				"hash", shardInfoHandler.GetHeaderHash(),
				"error", err,
			)
			return nil, process.ErrMissingHeader
		}

		usedShardHeaders[header.GetShardID()] = append(usedShardHeaders[header.GetShardID()], ShardHeaderInfo{
			Header:      header,
			Hash:        shardInfoHandler.GetHeaderHash(),
			UsedInBlock: true,
		})
		orderedShardHeaders = append(orderedShardHeaders, header)
		orderedShardHeaderHashes = append(orderedShardHeaderHashes, shardInfoHandler.GetHeaderHash())
	}

	return &usedShardHeadersInfo{
		headersPerShard:          usedShardHeaders,
		orderedShardHeaders:      orderedShardHeaders,
		orderedShardHeaderHashes: orderedShardHeaderHashes,
	}, nil
}

func (mp *metaProcessor) processIfFirstBlockAfterEpochStartBlockV3() error {
	prevExecutedBlock := mp.getPreviousExecutedBlock()
	prevExecutedMetaHeader, ok := prevExecutedBlock.(data.MetaHeaderHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	if !prevExecutedMetaHeader.IsStartOfEpochBlock() {
		return nil
	}

	nodesForcedToStay, err := mp.validatorStatisticsProcessor.SaveNodesCoordinatorUpdates(prevExecutedMetaHeader.GetEpoch())
	if err != nil {
		return err
	}

	err = mp.epochSystemSCProcessor.ToggleUnStakeUnBond(nodesForcedToStay)
	if err != nil {
		return err
	}

	return nil
}

func (mp *metaProcessor) getPreviousExecutedBlock() data.HeaderHandler {
	blockHeader := mp.blockChain.GetLastExecutedBlockHeader()
	if check.IfNil(blockHeader) {
		return mp.blockChain.GetGenesisHeader()
	}
	return blockHeader
}

// getComputedEpochStartData returns the epoch start data computed at propose block processing; the
// epoch guard keeps it valid across an epoch boundary rollback without leaking into other epochs
func (mp *metaProcessor) getComputedEpochStartData(epoch uint32) (*block.EpochStart, error) {
	mp.mutEpochStartData.RLock()
	defer mp.mutEpochStartData.RUnlock()

	if mp.epochStartDataWrapper == nil ||
		mp.epochStartDataWrapper.EpochStartData == nil ||
		mp.epochStartDataWrapper.Epoch != epoch {
		return nil, process.ErrNilEpochStartData
	}

	epochStartData := *mp.epochStartDataWrapper.EpochStartData

	return &epochStartData, nil
}

func (mp *metaProcessor) processEconomicsDataForEpochStartProposeBlock(metaHeader data.MetaHeaderHandler) error {
	baseExecutionResult := mp.blockChain.GetLastExecutionResult()
	if check.IfNil(baseExecutionResult) {
		return fmt.Errorf("%w for blockchain.GetLastExecutionResult", process.ErrNilBaseExecutionResult)
	}
	prevExecutionResult, ok := baseExecutionResult.(data.MetaExecutionResultHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	// since there are no shard headers finalized between the epoch start proposal and the epoch start block,
	// the last finalized data is the same as the one created at epoch start block proposal time
	lastFinalizedData, err := mp.epochStartDataCreator.CreateEpochStartShardDataMetablockV3(metaHeader)
	if err != nil {
		return err
	}
	lastShardData := &block.EpochStart{
		LastFinalizedHeaders: lastFinalizedData,
	}

	economicsData, err := mp.epochEconomics.ComputeEndOfEpochEconomicsV3(metaHeader, prevExecutionResult, lastShardData)
	if err != nil {
		return err
	}

	lastShardData.Economics = *economicsData

	mp.mutEpochStartData.Lock()
	defer mp.mutEpochStartData.Unlock()
	mp.epochStartDataWrapper.Epoch = metaHeader.GetEpoch() + 1
	mp.epochStartDataWrapper.EpochStartData = lastShardData

	return nil
}

func (mp *metaProcessor) getComputedEconomics(epoch uint32) (*block.Economics, error) {
	mp.mutEpochStartData.RLock()
	defer mp.mutEpochStartData.RUnlock()
	if mp.epochStartDataWrapper == nil ||
		mp.epochStartDataWrapper.EpochStartData == nil ||
		mp.epochStartDataWrapper.Epoch != epoch {
		return nil, process.ErrNilEpochStartData
	}
	computedEconomics := &mp.epochStartDataWrapper.EpochStartData.Economics

	return computedEconomics, nil
}
