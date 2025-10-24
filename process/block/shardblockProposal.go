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
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/state"
)

// TODO: maybe move this to config
const maxBlockProcessingTime = 3 * time.Second

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

	err = shardHdr.SetMetaBlockHashes(sp.miniBlocksSelectionSession.GetReferencedMetaBlockHashes())
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

	hash, err := sp.getHeaderHash(shardHdr)
	if err != nil {
		return nil, nil, err
	}

	err = sp.OnProposedBlock(body, shardHdr, hash)
	if err != nil {
		return nil, nil, err
	}

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

	go getMetricsFromBlockBody(body, sp.marshalizer, sp.appStatusHandler)

	err = sp.executionResultsVerifier.VerifyHeaderExecutionResults(header)
	if err != nil {
		return err
	}

	err = sp.checkInclusionEstimationForExecutionResults(header)
	if err != nil {
		return err
	}

	txCounts, rewardCounts, unsignedCounts := sp.txCounter.getPoolCounts(sp.dataPool)
	log.Debug("total txs in pool", "counts", txCounts.String())
	log.Debug("total txs in rewards pool", "counts", rewardCounts.String())
	log.Debug("total txs in unsigned pool", "counts", unsignedCounts.String())

	go getMetricsFromHeader(header, uint64(txCounts.GetTotal()), sp.marshalizer, sp.appStatusHandler)

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

func (sp *shardProcessor) verifyGasLimit(header data.ShardHeaderHandler) error {
	incomingMiniBlocks, incomingTransactions, outgoingTransactionHashes, outgoingTransactions, err := sp.splitTransactionsForHeader(header)
	if err != nil {
		return err
	}

	sp.gasComputation.Reset()
	_, numPendingMiniBlocks, err := sp.gasComputation.CheckIncomingMiniBlocks(incomingMiniBlocks, incomingTransactions)
	if err != nil {
		return err
	}

	addedTxHashes, pendingMiniBlocksAdded, err := sp.gasComputation.CheckOutgoingTransactions(outgoingTransactionHashes, outgoingTransactions)
	if err != nil {
		return err
	}
	if len(addedTxHashes) != len(outgoingTransactionHashes) {
		return fmt.Errorf("%w, outgoing transactions exceeded the limit", process.ErrInvalidMaxGasLimitPerMiniBlock)
	}

	if numPendingMiniBlocks != len(pendingMiniBlocksAdded) {
		return fmt.Errorf("%w, incoming mini blocks exceeded the limit", process.ErrInvalidMaxGasLimitPerMiniBlock)
	}

	return nil
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

	// this is used now to reset the context for processing not creation of blocks
	err := sp.createBlockStarted()
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

	if sp.accountsDB[state.UserAccountsState].JournalLen() != 0 {
		log.Error("shardProcessor.ProcessBlockProposal first entry", "stack", string(sp.accountsDB[state.UserAccountsState].GetStackDebugFirstEntry()))
		return nil, process.ErrAccountStateDirty
	}

	err = sp.blockChainHook.SetCurrentHeader(header)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			sp.RevertCurrentBlock()
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

func (sp *shardProcessor) splitTransactionsForHeader(header data.HeaderHandler) (
	incomingMiniBlocks []data.MiniBlockHeaderHandler,
	incomingTransactions map[string][]data.TransactionHandler,
	outgoingTransactionHashes [][]byte,
	outgoingTransactions []data.TransactionHandler,
	err error,
) {
	incomingTransactions = make(map[string][]data.TransactionHandler)
	var txsForMb []data.TransactionHandler
	var txHashes [][]byte
	for _, mb := range header.GetMiniBlockHeaderHandlers() {
		txHashes, txsForMb, err = sp.getTransactionsForMiniBlock(mb)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		if mb.GetSenderShardID() == sp.shardCoordinator.SelfId() {
			outgoingTransactionHashes = append(outgoingTransactionHashes, txHashes...)
			outgoingTransactions = append(outgoingTransactions, txsForMb...)
			continue
		}

		incomingMiniBlocks = append(incomingMiniBlocks, mb)
		incomingTransactions[string(mb.GetHash())] = txsForMb
	}

	return incomingMiniBlocks, incomingTransactions, outgoingTransactionHashes, outgoingTransactions, nil
}

func (sp *shardProcessor) getTransactionsForMiniBlock(
	miniBlock data.MiniBlockHeaderHandler,
) ([][]byte, []data.TransactionHandler, error) {
	obj, hashInPool := sp.dataPool.MiniBlocks().Get(miniBlock.GetHash())
	if !hashInPool {
		return nil, nil, process.ErrMissingMiniBlock
	}

	mbForHeaderPtr, typeOk := obj.(*block.MiniBlock)
	if !typeOk {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	txs := make([]data.TransactionHandler, len(mbForHeaderPtr.TxHashes))
	var err error
	for idx, txHash := range mbForHeaderPtr.TxHashes {
		txs[idx], err = process.GetTransactionHandlerFromPool(
			miniBlock.GetSenderShardID(),
			miniBlock.GetReceiverShardID(),
			txHash,
			sp.dataPool.Transactions(),
			process.SearchMethodSearchFirst,
		)
		if err != nil {
			return nil, nil, err
		}
	}

	return mbForHeaderPtr.TxHashes, txs, nil
}

func (sp *shardProcessor) checkInclusionEstimationForExecutionResults(header data.HeaderHandler) error {
	prevBlockLastExecutionResult, err := process.GetPrevBlockLastExecutionResult(sp.blockChain)
	if err != nil {
		return err
	}

	lastResultData, err := process.CreateDataForInclusionEstimation(prevBlockLastExecutionResult)
	if err != nil {
		return err
	}
	executionResults := header.GetExecutionResultsHandlers()
	allowed := sp.executionResultsInclusionEstimator.Decide(lastResultData, executionResults, header.GetRound())
	if allowed != len(executionResults) {
		log.Warn("number of execution results included in the header is not correct",
			"expected", allowed,
			"actual", len(executionResults),
		)
		return process.ErrInvalidNumberOfExecutionResultsInHeader
	}

	return nil
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

func (sp *shardProcessor) addExecutionResultsOnHeader(shardHeader data.HeaderHandler) error {
	pendingExecutionResults, err := sp.executionResultsTracker.GetPendingExecutionResults()
	if err != nil {
		return err
	}

	lastExecutionResultHandler, err := process.GetPrevBlockLastExecutionResult(sp.blockChain)
	if err != nil {
		return err
	}

	lastNotarizedExecutionResultInfo, err := process.CreateDataForInclusionEstimation(lastExecutionResultHandler)
	if err != nil {
		return err
	}

	var lastExecutionResultForCurrentBlock data.LastExecutionResultHandler
	numToInclude := sp.executionResultsInclusionEstimator.Decide(lastNotarizedExecutionResultInfo, pendingExecutionResults, shardHeader.GetRound())

	executionResultsToInclude := pendingExecutionResults[:numToInclude]
	lastExecutionResultForCurrentBlock = lastExecutionResultHandler
	if len(executionResultsToInclude) > 0 {
		lastExecutionResult := executionResultsToInclude[len(executionResultsToInclude)-1]
		lastExecutionResultForCurrentBlock, err = process.CreateLastExecutionResultInfoFromExecutionResult(shardHeader.GetRound(), lastExecutionResult, sp.shardCoordinator.SelfId())
		if err != nil {
			return err
		}
	}

	err = shardHeader.SetLastExecutionResultHandler(lastExecutionResultForCurrentBlock)
	if err != nil {
		return err
	}

	return shardHeader.SetExecutionResultsHandlers(executionResultsToInclude)
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

	return sp.createProposalMiniBlocks(haveTime)
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

	referencedMetaBlocks := sp.miniBlocksSelectionSession.GetReferencedMetaBlocks()
	numHeadersAdded := uint32(len(referencedMetaBlocks))
	if numHeadersAdded > 0 {
		go sp.requestMetaHeadersIfNeeded(numHeadersAdded, referencedMetaBlocks[numHeadersAdded-1])
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
		"num referenced meta blocks", len(sp.miniBlocksSelectionSession.GetReferencedMetaBlocks()))

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
	var shouldContinue bool
	for i := 0; i < len(orderedMetaBlocks); i++ {
		if !haveTime() {
			log.Debug("time is up after putting cross txs with destination to current shard",
				"num txs added", sp.miniBlocksSelectionSession.GetNumTxsAdded(),
			)
			break
		}

		if len(sp.miniBlocksSelectionSession.GetReferencedMetaBlocks()) >= process.MaxMetaHeadersAllowedInOneShardBlock {
			log.Debug("maximum meta headers allowed to be included in one shard block has been reached",
				"meta headers added", len(sp.miniBlocksSelectionSession.GetReferencedMetaBlocks()),
			)
			break
		}

		currentMetaBlock = orderedMetaBlocks[i]
		if currentMetaBlock.GetNonce() > lastCrossNotarizedMetaHdr.GetNonce()+1 {
			log.Debug("skip searching",
				"last meta hdr nonce", lastCrossNotarizedMetaHdr.GetNonce(),
				"curr meta hdr nonce", currentMetaBlock.GetNonce())
			break
		}

		hasProofForHdr := sp.proofsPool.HasProof(core.MetachainShardId, orderedMetaBlocksHashes[i])
		if !hasProofForHdr {
			log.Trace("no proof for meta header",
				"hash", logger.DisplayByteSlice(orderedMetaBlocksHashes[i]),
			)
			break
		}

		currentMetaBlockHash = orderedMetaBlocksHashes[i]
		metaBlock, ok := currentMetaBlock.(data.MetaHeaderHandler)
		if !ok {
			log.Warn("selectIncomingMiniBlocks: wrong type assertion for meta block")
			break
		}

		if len(currentMetaBlock.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())) == 0 {
			sp.miniBlocksSelectionSession.AddReferencedMetaBlock(orderedMetaBlocks[i], orderedMetaBlocksHashes[i])
			continue
		}

		currProcessedMiniBlocksInfo := sp.processedMiniBlocksTracker.GetProcessedMiniBlocksInfo(currentMetaBlockHash)
		shouldContinue, pendingMiniBlocks, errCreated = sp.createMbsCrossShardDstMe(currentMetaBlockHash, metaBlock, currProcessedMiniBlocksInfo)
		if errCreated != nil {
			return nil, errCreated
		}
		if !shouldContinue {
			break
		}

		sp.miniBlocksSelectionSession.AddReferencedMetaBlock(currentMetaBlock, currentMetaBlockHash)
	}

	return pendingMiniBlocks, nil
}

func (sp *shardProcessor) createMbsCrossShardDstMe(
	currentMetaBlockHash []byte,
	currentMetaBlock data.MetaHeaderHandler,
	miniBlockProcessingInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) (bool, []block.MiniblockAndHash, error) {
	currMiniBlocksAdded, pendingMiniBlocks, currNumTxsAdded, hdrFinished, errCreate := sp.txCoordinator.CreateMbsCrossShardDstMe(
		currentMetaBlock,
		miniBlockProcessingInfo,
	)
	if errCreate != nil {
		return false, nil, errCreate
	}

	err := sp.miniBlocksSelectionSession.AddMiniBlocksAndHashes(currMiniBlocksAdded)
	if err != nil {
		return false, nil, err
	}

	if !hdrFinished {
		log.Debug("meta block cannot be fully processed",
			"round", currentMetaBlock.GetRound(),
			"nonce", currentMetaBlock.GetNonce(),
			"hash", currentMetaBlockHash,
			"num mbs added", len(currMiniBlocksAdded),
			"num txs added", currNumTxsAdded)

		return false, pendingMiniBlocks, nil
	}

	return true, pendingMiniBlocks, nil
}

func (sp *shardProcessor) createProposalMiniBlocks(haveTime func() bool) error {
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

	outgoingTransactions, pendingIncomingMiniBlocksAdded := sp.selectOutgoingTransactions()

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

func (sp *shardProcessor) selectOutgoingTransactions() ([][]byte, []data.MiniBlockHeaderHandler) {
	log.Debug("selectOutgoingTransactions has been started")

	sw := core.NewStopWatch()
	sw.Start("selectOutgoingTransactions")
	defer func() {
		sw.Stop("selectOutgoingTransactions")
		log.Debug("measurements", sw.GetMeasurements()...)
	}()

	outgoingTransactions, pendingIncomingMiniBlocksAdded := sp.txCoordinator.SelectOutgoingTransactions()
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
	crossShardIncomingMiniBlocks := sp.getCrossShardIncomingMiniBlocksFromBody(body)
	miniBlocksFromSelf := sp.txCoordinator.GetCreatedMiniBlocksFromMe()
	postProcessMiniBlocks := sp.txCoordinator.CreatePostProcessMiniBlocks()

	allMiniBlocks := make([]*block.MiniBlock, 0, len(crossShardIncomingMiniBlocks)+len(miniBlocksFromSelf)+len(postProcessMiniBlocks))
	allMiniBlocks = append(allMiniBlocks, crossShardIncomingMiniBlocks...)
	allMiniBlocks = append(allMiniBlocks, miniBlocksFromSelf...)
	allMiniBlocks = append(allMiniBlocks, postProcessMiniBlocks...)

	receiptHash, err := sp.txCoordinator.CreateReceiptsHash()
	if err != nil {
		return nil, err
	}

	gasAndFees := sp.getGasAndFees()
	gasNotUsedForProcessing := gasAndFees.GetGasPenalized() + gasAndFees.GetGasRefunded()
	if gasAndFees.GetGasProvided() < gasNotUsedForProcessing {
		return nil, process.ErrGasUsedExceedsGasProvided
	}

	gasUsed := gasAndFees.GetGasProvided() - gasNotUsedForProcessing // needed for inclusion estimation

	bodyAfterExecution := &block.Body{MiniBlocks: allMiniBlocks}
	// remove the self-receipts and self smart contract results mini blocks - similar to Pre-Supernova
	sanitizedBodyAfterExecution := deleteSelfReceiptsMiniBlocks(bodyAfterExecution)

	// giving an empty processedMiniBlockInfo would cause all miniBlockHeaders to be created as fully processed.
	processedMiniBlockInfo := make(map[string]*processedMb.ProcessedMiniBlockInfo)

	totalTxCount, miniBlockHeaderHandlers, err := sp.createMiniBlockHeaderHandlers(sanitizedBodyAfterExecution, processedMiniBlockInfo)
	if err != nil {
		return nil, err
	}

	intraMiniBlocks := sp.txCoordinator.GetCreatedInShardMiniBlocks()
	err = sp.cacheIntraShardMiniBlocks(headerHash, intraMiniBlocks)
	if err != nil {
		return nil, err
	}

	err = sp.cacheExecutedMiniBlocks(sanitizedBodyAfterExecution, miniBlockHeaderHandlers)
	if err != nil {
		return nil, err
	}

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
