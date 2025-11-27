package block

import (
	"errors"
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
)

// TODO: maybe move this to config
const maxBlockProcessingTime = 3 * time.Second

// CreateNewHeaderProposal creates a new header proposal
func (sp *shardProcessor) CreateNewHeaderProposal(round uint64, nonce uint64) (data.HeaderHandler, error) {
	epoch := sp.epochStartTrigger.MetaEpoch()
	header := sp.versionedHeaderFactory.Create(epoch, round)

	shardHeader, ok := header.(data.ShardHeaderHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}
	if !shardHeader.IsHeaderV3() {
		return nil, process.ErrInvalidHeader
	}

	err := shardHeader.SetRound(round)
	if err != nil {
		return nil, err
	}

	err = shardHeader.SetNonce(nonce)
	if err != nil {
		return nil, err
	}

	return header, nil
}

// CreateBlockProposal creates a block proposal without executing any of the transactions
func (sp *shardProcessor) CreateBlockProposal(
	initialHdr data.HeaderHandler,
	haveTime func() bool,
) (data.HeaderHandler, data.BodyHandler, error) {
	if check.IfNil(initialHdr) {
		return nil, nil, process.ErrNilBlockHeader
	}
	if !initialHdr.IsHeaderV3() {
		return nil, nil, process.ErrInvalidHeader
	}
	shardHdr, ok := initialHdr.(data.ShardHeaderHandler)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	err := sp.updateEpochIfNeeded(shardHdr)
	if err != nil {
		return nil, nil, err
	}

	sp.gasComputation.Reset()
	sp.miniBlocksSelectionSession.ResetSelectionSession()
	err = sp.createBlockBodyProposal(shardHdr, haveTime)
	if err != nil {
		return nil, nil, err
	}

	miniBlockHeaderHandlers := sp.miniBlocksSelectionSession.GetMiniBlockHeaderHandlers()
	// todo: check empty mini blocks vs nil. Same for block.Body.MiniBlocks
	err = shardHdr.SetMiniBlockHeaderHandlers(miniBlockHeaderHandlers)
	if err != nil {
		return nil, nil, err
	}

	err = shardHdr.SetMetaBlockHashes(sp.miniBlocksSelectionSession.GetReferencedHeaderHashes())
	if err != nil {
		return nil, nil, err
	}

	totalTxCount := computeTxTotalTxCount(miniBlockHeaderHandlers)
	err = shardHdr.SetTxCount(totalTxCount)
	if err != nil {
		return nil, nil, err
	}

	miniBlocks := sp.miniBlocksSelectionSession.GetMiniBlocks()
	err = checkMiniBlocksAndMiniBlockHeadersConsistency(miniBlocks, miniBlockHeaderHandlers)
	if err != nil {
		return nil, nil, err
	}

	err = sp.addExecutionResultsOnHeader(shardHdr)
	if err != nil {
		return nil, nil, err
	}

	// TODO: sanity check use the verify execution results method

	body := &block.Body{MiniBlocks: miniBlocks}

	err = sp.verifyGasLimit(shardHdr)
	if err != nil {
		return nil, nil, err
	}

	sp.appStatusHandler.SetUInt64Value(common.MetricNumTxInBlock, uint64(totalTxCount))
	sp.appStatusHandler.SetUInt64Value(common.MetricNumMiniBlocks, uint64(len(body.MiniBlocks)))

	marshalledBody, err := sp.marshalizer.Marshal(body)
	if err != nil {
		return nil, nil, err
	}

	sp.blockSizeThrottler.Add(shardHdr.GetRound(), uint32(len(marshalledBody)))

	defer func() {
		go sp.checkAndRequestIfMetaHeadersMissing()
	}()

	return shardHdr, body, nil
}

// VerifyBlockProposal verifies the proposed block. It returns nil if all ok or the specific error
func (sp *shardProcessor) VerifyBlockProposal(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
	haveTime func() time.Duration,
) error {
	err := sp.checkBlockValidity(headerHandler, bodyHandler)
	if err != nil {
		if errors.Is(err, process.ErrBlockHashDoesNotMatch) {
			log.Debug("requested missing shard header",
				"hash", headerHandler.GetPrevHash(),
				"for shard", headerHandler.GetShardID(),
			)

			go sp.requestHandler.RequestShardHeaderForEpoch(headerHandler.GetShardID(), headerHandler.GetPrevHash(), headerHandler.GetEpoch())
		}

		return err
	}

	log.Debug("started verifying proposed block",
		"epoch", headerHandler.GetEpoch(),
		"shard", headerHandler.GetShardID(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
	)

	header, ok := headerHandler.(data.ShardHeaderHandler)
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

	err = sp.checkHeaderBodyCorrelationProposal(header.GetMiniBlockHeaderHandlers(), body)
	if err != nil {
		return err
	}

	err = sp.executionResultsVerifier.VerifyHeaderExecutionResults(header)
	if err != nil {
		return err
	}

	err = sp.checkInclusionEstimationForExecutionResults(header)
	if err != nil {
		return err
	}

	sp.updateMetrics(header, body)

	sp.missingDataResolver.Reset()
	sp.missingDataResolver.RequestBlockTransactions(body)
	// the epoch start meta block and its proof is also requested here if missing
	err = sp.missingDataResolver.RequestMissingMetaHeaders(header)
	if err != nil {
		return err
	}

	err = sp.missingDataResolver.WaitForMissingData(haveTime())
	if err != nil {
		return err
	}

	defer func() {
		go sp.checkAndRequestIfMetaHeadersMissing()
	}()

	err = sp.checkEpochCorrectnessCrossChain()
	if err != nil {
		return err
	}

	err = sp.checkEpochCorrectness(header)
	if err != nil {
		return err
	}

	err = sp.checkMetaHeadersValidityAndFinalityProposal(header)
	if err != nil {
		return err
	}

	err = sp.verifyCrossShardMiniBlockDstMe(header)
	if err != nil {
		return err
	}

	err = sp.verifyGasLimit(header)
	if err != nil {
		return err
	}

	hash, err := sp.getHeaderHash(header)
	if err != nil {
		return err
	}

	return sp.OnProposedBlock(body, header, hash)
}

func (sp *shardProcessor) updateMetrics(header data.HeaderHandler, body *block.Body) {
	go getMetricsFromBlockBody(body, sp.marshalizer, sp.appStatusHandler)

	txCounts, rewardCounts, unsignedCounts := sp.txCounter.getPoolCounts(sp.dataPool)
	log.Debug("total txs in pool", "counts", txCounts.String())
	log.Debug("total txs in rewards pool", "counts", rewardCounts.String())
	log.Debug("total txs in unsigned pool", "counts", unsignedCounts.String())

	go getMetricsFromHeader(header, uint64(txCounts.GetTotal()), sp.marshalizer, sp.appStatusHandler)
}

func getHaveTimeForProposal(startTime time.Time, maxDuration time.Duration) func() time.Duration {
	timeOut := startTime.Add(maxDuration)
	haveTime := func() time.Duration {
		return time.Until(timeOut)
	}
	return haveTime
}

// ProcessBlockProposal processes the proposed block. It returns nil if all ok or the specific error
func (sp *shardProcessor) ProcessBlockProposal(
	headerHandler data.HeaderHandler,
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

	sp.processStatusHandler.SetBusy("shardProcessor.ProcessBlockProposal")
	defer sp.processStatusHandler.SetIdle()

	sp.roundNotifier.CheckRound(headerHandler)
	sp.epochNotifier.CheckEpoch(headerHandler)
	sp.requestHandler.SetEpoch(headerHandler.GetEpoch())

	header, ok := headerHandler.(data.ShardHeaderHandler)
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

	if sp.accountsDB[state.UserAccountsState].JournalLen() != 0 {
		log.Error("shardProcessor.ProcessBlockProposal first entry", "stack", string(sp.accountsDB[state.UserAccountsState].GetStackDebugFirstEntry()))
		return nil, process.ErrAccountStateDirty
	}

	err := sp.checkContextBeforeExecution(header)
	if err != nil {
		return nil, err
	}

	// this is used now to reset the context for processing not creation of blocks
	err = sp.createBlockStarted()
	if err != nil {
		return nil, err
	}

	// should already be available in the pools since it passed the block proposal verification,
	// but kept here to update the internal caches (txsForBlock, hdrsForCurrBlock)
	sp.txCoordinator.RequestBlockTransactions(body)
	sp.hdrsForCurrBlock.RequestMetaHeaders(header)

	// although we can have a long time for processing, it being decoupled from consensus,
	// we still give some reasonable timeout
	proposalStartTime := time.Now()
	haveTime := getHaveTimeForProposal(proposalStartTime, maxBlockProcessingTime)

	err = sp.txCoordinator.IsDataPreparedForProcessing(haveTime)
	if err != nil {
		return nil, err
	}

	// TODO: improvement - add also a request if it is missing as a fallback, although it should not be missing at this point
	err = sp.checkEpochStartInfoAvailableIfNeeded(header)
	if err != nil {
		return nil, err
	}

	err = sp.hdrsForCurrBlock.WaitForHeadersIfNeeded(haveTime)
	if err != nil {
		return nil, err
	}

	// TODO: check again before saving the last executed result
	err = sp.blockChainHook.SetCurrentHeader(header)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			sp.RevertCurrentBlock(header)
		}
	}()

	startTime := time.Now()
	err = sp.txCoordinator.ProcessBlockTransaction(header, body, haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to process block transaction",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return nil, err
	}

	err = sp.txCoordinator.VerifyCreatedBlockTransactions(header, body)
	if err != nil {
		return nil, err
	}

	err = sp.commitState(headerHandler)
	if err != nil {
		return nil, err
	}

	// TODO: should receive the header hash instead of re-computing it
	headerHash, err := core.CalculateHash(sp.marshalizer, sp.hasher, header)
	if err != nil {
		return nil, err
	}

	executionResult, err := sp.collectExecutionResults(headerHash, header, body)
	if err != nil {
		return nil, err
	}

	errCutoff := sp.blockProcessingCutoffHandler.HandleProcessErrorCutoff(header)
	if errCutoff != nil {
		return nil, errCutoff
	}

	return executionResult, nil
}

func computeTxTotalTxCount(miniBlockHeaders []data.MiniBlockHeaderHandler) uint32 {
	totalTxCount := uint32(0)
	for i := range miniBlockHeaders {
		totalTxCount += miniBlockHeaders[i].GetTxCount()
	}
	return totalTxCount
}

func checkMiniBlocksAndMiniBlockHeadersConsistency(miniBlocks block.MiniBlockSlice, miniBlockHeaders []data.MiniBlockHeaderHandler) error {
	if len(miniBlocks) != len(miniBlockHeaders) {
		log.Warn("transactionCoordinator.verifyFees: num of mini blocks and mini blocks headers does not match", "num of mb", len(miniBlocks), "num of mbh", len(miniBlockHeaders))
		return process.ErrNumOfMiniBlocksAndMiniBlocksHeadersMismatch
	}

	// TODO: check if the reserved field or other fields are consistent.
	return nil
}

func (sp *shardProcessor) createBlockBodyProposal(
	shardHdr data.HeaderHandler,
	haveTime func() bool,
) error {
	sp.blockSizeThrottler.ComputeCurrentMaxSize()

	log.Debug("started creating block body",
		"epoch", shardHdr.GetEpoch(),
		"round", shardHdr.GetRound(),
		"nonce", shardHdr.GetNonce(),
	)

	_, err := sp.prepareAccountsForProposal()
	if err != nil {
		return err
	}

	return sp.createProposalMiniBlocks(haveTime, shardHdr.GetNonce())
}

func (sp *shardProcessor) selectIncomingMiniBlocksForProposal(
	haveTime func() bool,
) ([]block.MiniblockAndHash, error) {
	log.Debug("selectIncomingMiniBlocksForProposal has been started")

	sw := core.NewStopWatch()
	sw.Start("ComputeLongestMetaChainFromLastNotarized")
	orderedMetaBlocks, orderedMetaBlocksHashes, err := sp.blockTracker.ComputeLongestMetaChainFromLastNotarized()
	sw.Stop("ComputeLongestMetaChainFromLastNotarized")
	log.Debug("measurements", sw.GetMeasurements()...)
	if err != nil {
		return nil, err
	}

	log.Debug("meta blocks ordered", "num meta blocks", len(orderedMetaBlocks))

	lastMetaHdr, _, err := sp.blockTracker.GetLastCrossNotarizedHeader(core.MetachainShardId)
	if err != nil {
		return nil, err
	}

	pendingMiniBlocks, err := sp.selectIncomingMiniBlocks(lastMetaHdr, orderedMetaBlocks, orderedMetaBlocksHashes, haveTime)
	if err != nil {
		return nil, err
	}

	miniBlockHeaderHandlers := sp.miniBlocksSelectionSession.GetMiniBlockHeaderHandlers()
	for _, miniBlockHeader := range miniBlockHeaderHandlers {
		log.Debug("mini block info",
			"type", miniBlockHeader.GetTypeInt32(),
			"sender shard", miniBlockHeader.GetSenderShardID(),
			"receiver shard", miniBlockHeader.GetReceiverShardID(),
			"txs added", miniBlockHeader.GetTxCount())
	}

	log.Debug("selectIncomingMiniBlocksForProposal has been finished",
		"num txs added", sp.miniBlocksSelectionSession.GetNumTxsAdded(),
		"num referenced meta blocks", len(sp.miniBlocksSelectionSession.GetReferencedHeaders()))

	return pendingMiniBlocks, nil
}

func (sp *shardProcessor) selectIncomingMiniBlocks(
	lastCrossNotarizedMetaHdr data.HeaderHandler,
	orderedMetaBlocks []data.HeaderHandler,
	orderedMetaBlocksHashes [][]byte,
	haveTime func() bool,
) ([]block.MiniblockAndHash, error) {
	var currentMetaBlock data.HeaderHandler
	var currentMetaBlockHash []byte
	var pendingMiniBlocks []block.MiniblockAndHash
	var errCreated error
	var createIncomingMbsResult *CrossShardIncomingMbsCreationResult
	lastMeta := lastCrossNotarizedMetaHdr
	for i := 0; i < len(orderedMetaBlocks); i++ {
		if !haveTime() {
			log.Debug("time is up after putting cross txs with destination to current shard",
				"num txs added", sp.miniBlocksSelectionSession.GetNumTxsAdded(),
			)
			break
		}

		if len(sp.miniBlocksSelectionSession.GetReferencedHeaders()) >= process.MaxMetaHeadersAllowedInOneShardBlock {
			log.Debug("maximum meta headers allowed to be included in one shard block has been reached",
				"meta headers added", len(sp.miniBlocksSelectionSession.GetReferencedHeaders()),
			)
			break
		}

		currentMetaBlock = orderedMetaBlocks[i]
		currentMetaBlockHash = orderedMetaBlocksHashes[i]
		if currentMetaBlock.GetNonce() != lastMeta.GetNonce()+1 {
			log.Debug("skip searching",
				"last meta hdr nonce", lastMeta.GetNonce(),
				"curr meta hdr nonce", currentMetaBlock.GetNonce())
			continue
		}

		hasProofForHdr := sp.proofsPool.HasProof(core.MetachainShardId, currentMetaBlockHash)
		if !hasProofForHdr {
			log.Trace("no proof for meta header",
				"hash", logger.DisplayByteSlice(currentMetaBlockHash),
			)
			continue
		}

		metaBlock, ok := currentMetaBlock.(data.MetaHeaderHandler)
		if !ok {
			log.Warn("selectIncomingMiniBlocks: wrong type assertion for meta block")
			break
		}

		if len(currentMetaBlock.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())) == 0 {
			sp.miniBlocksSelectionSession.AddReferencedHeader(currentMetaBlock, currentMetaBlockHash)
			lastMeta = currentMetaBlock
			continue
		}

		currProcessedMiniBlocksInfo := sp.processedMiniBlocksTracker.GetProcessedMiniBlocksInfo(currentMetaBlockHash)
		createIncomingMbsResult, errCreated = sp.createMbsCrossShardDstMe(currentMetaBlockHash, metaBlock, currProcessedMiniBlocksInfo)
		if errCreated != nil {
			return nil, errCreated
		}

		pendingMiniBlocks = append(pendingMiniBlocks, createIncomingMbsResult.PendingMiniBlocks...)
		if len(createIncomingMbsResult.AddedMiniBlocks) > 0 {
			errAdd := sp.miniBlocksSelectionSession.AddMiniBlocksAndHashes(createIncomingMbsResult.AddedMiniBlocks)
			if errAdd != nil {
				return nil, errAdd
			}
			sp.miniBlocksSelectionSession.AddReferencedHeader(currentMetaBlock, currentMetaBlockHash)
			lastMeta = currentMetaBlock
		}

		if !createIncomingMbsResult.HeaderFinished {
			break
		}
	}

	go sp.requestHeadersFromHeaderIfNeeded(lastMeta)

	return pendingMiniBlocks, nil
}

func (sp *shardProcessor) createProposalMiniBlocks(
	haveTime func() bool,
	nonce uint64,
) error {
	if !haveTime() {
		log.Debug("shardProcessor.createProposalMiniBlocks", "error", process.ErrTimeIsOut)
		return nil
	}
	startTime := time.Now()
	pendingMiniBlocksLeft, err := sp.selectIncomingMiniBlocksForProposal(haveTime)
	if err != nil {
		return err
	}
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to create mbs to me", "time", elapsedTime)

	outgoingTransactions, pendingIncomingMiniBlocksAdded := sp.selectOutgoingTransactions(nonce)

	err = sp.appendPendingMiniBlocksAddedAfterSelectingOutgoingTransactions(pendingMiniBlocksLeft, pendingIncomingMiniBlocksAdded)
	if err != nil {
		return err
	}

	err = sp.miniBlocksSelectionSession.CreateAndAddMiniBlockFromTransactions(outgoingTransactions)
	if err != nil {
		log.Debug("shardProcessor.createProposalMiniBlocks", "error", err.Error())
		return err
	}

	// todo: maybe sanitize, removing empty miniBlocks

	return nil
}

func (sp *shardProcessor) appendPendingMiniBlocksAddedAfterSelectingOutgoingTransactions(
	pendingMiniBlocksLeft []block.MiniblockAndHash,
	pendingIncomingMiniBlocksAdded []data.MiniBlockHeaderHandler,
) error {
	if len(pendingIncomingMiniBlocksAdded) == 0 {
		return nil
	}

	pendingMiniBlocksLeftMap := miniBlocksAndHashesSliceToMap(pendingMiniBlocksLeft)
	extraMiniBlocksAdded := make([]block.MiniblockAndHash, len(pendingIncomingMiniBlocksAdded))
	for i, pendingMbAdded := range pendingIncomingMiniBlocksAdded {
		miniBlockAndHash, ok := pendingMiniBlocksLeftMap[string(pendingMbAdded.GetHash())]
		if !ok {
			log.Error("pending mini block added does not exists in the remaining pending list")
			return process.ErrInvalidHash
		}

		extraMiniBlocksAdded[i] = miniBlockAndHash
	}

	return sp.miniBlocksSelectionSession.AddMiniBlocksAndHashes(extraMiniBlocksAdded)
}

func miniBlocksAndHashesSliceToMap(providedSlice []block.MiniblockAndHash) map[string]block.MiniblockAndHash {
	result := make(map[string]block.MiniblockAndHash)
	for _, miniBlockAndHash := range providedSlice {
		result[string(miniBlockAndHash.Hash)] = miniBlockAndHash
	}

	return result
}

func (sp *shardProcessor) selectOutgoingTransactions(
	nonce uint64,
) ([][]byte, []data.MiniBlockHeaderHandler) {
	log.Debug("selectOutgoingTransactions has been started")

	sw := core.NewStopWatch()
	sw.Start("selectOutgoingTransactions")
	defer func() {
		sw.Stop("selectOutgoingTransactions")
		log.Debug("measurements", sw.GetMeasurements()...)
	}()

	outgoingTransactions, pendingIncomingMiniBlocksAdded := sp.txCoordinator.SelectOutgoingTransactions(nonce)
	log.Debug("selectOutgoingTransactions has been finished",
		"num txs", len(outgoingTransactions),
		"num pending mini blocks added", len(pendingIncomingMiniBlocksAdded))

	return outgoingTransactions, pendingIncomingMiniBlocksAdded
}

func (sp *shardProcessor) checkMetaHeadersValidityAndFinalityProposal(header data.ShardHeaderHandler) error {
	lastCrossNotarizedHeader, _, err := sp.blockTracker.GetLastCrossNotarizedHeader(core.MetachainShardId)
	if err != nil {
		return err
	}
	usedMetaHdrHashes := header.GetMetaBlockHashes()
	usedMetaHeaders := make([]data.HeaderHandler, 0, len(usedMetaHdrHashes))
	var metaHdr data.HeaderHandler
	for _, metaHdrHash := range usedMetaHdrHashes {
		metaHdr, err = sp.dataPool.Headers().GetHeaderByHash(metaHdrHash)
		if err != nil {
			return fmt.Errorf("%w : checkMetaHeadersValidityAndFinalityProposal -> getHeaderByHash", err)
		}
		usedMetaHeaders = append(usedMetaHeaders, metaHdr)
	}

	process.SortHeadersByNonce(usedMetaHeaders)

	for _, metaHeader := range usedMetaHeaders {
		err = sp.headerValidator.IsHeaderConstructionValid(metaHeader, lastCrossNotarizedHeader)
		if err != nil {
			return fmt.Errorf("%w : checkMetaHeadersValidityAndFinalityProposal -> isHdrConstructionValid", err)
		}

		err = sp.checkHeaderHasProof(metaHeader)
		if err != nil {
			return fmt.Errorf("%w : checkMetaHeadersValidityAndFinalityProposal -> checkHeaderHasProof", err)
		}
		lastCrossNotarizedHeader = metaHeader
	}

	return nil
}

// collectExecutionResults collects the execution results after processing the block
func (sp *shardProcessor) collectExecutionResults(headerHash []byte, header data.HeaderHandler, body *block.Body) (data.BaseExecutionResultHandler, error) {
	miniBlockHeaderHandlers, totalTxCount, receiptHash, err := sp.collectMiniBlocks(headerHash, body)
	if err != nil {
		return nil, err
	}

	gasAndFees := sp.getGasAndFees()
	gasNotUsedForProcessing := gasAndFees.GetGasPenalized() + gasAndFees.GetGasRefunded()
	if gasAndFees.GetGasProvided() < gasNotUsedForProcessing {
		return nil, process.ErrGasUsedExceedsGasProvided
	}

	gasUsed := gasAndFees.GetGasProvided() - gasNotUsedForProcessing // needed for inclusion estimation

	executionResult := &block.ExecutionResult{
		BaseExecutionResult: &block.BaseExecutionResult{
			HeaderHash:  headerHash,
			HeaderNonce: header.GetNonce(),
			HeaderRound: header.GetRound(),
			HeaderEpoch: header.GetEpoch(),
			RootHash:    sp.getRootHash(),
			GasUsed:     gasUsed,
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

	logs := sp.txCoordinator.GetAllCurrentLogs()
	err = sp.cacheLogEvents(headerHash, logs)
	if err != nil {
		return nil, err
	}

	err = sp.cacheIntermediateTxsForHeader(headerHash)
	if err != nil {
		return nil, err
	}

	return executionResult, nil
}

func (sp *shardProcessor) getCrossShardIncomingMiniBlocksFromBody(body *block.Body) []*block.MiniBlock {
	miniBlocks := make([]*block.MiniBlock, 0)
	for _, mb := range body.MiniBlocks {
		if mb.ReceiverShardID == sp.shardCoordinator.SelfId() && mb.SenderShardID != sp.shardCoordinator.SelfId() {
			miniBlocks = append(miniBlocks, mb)
		}
	}
	return miniBlocks
}
