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
)

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

	// TODO: check that the limits for the block are not exceeded
	// err = sp.verifyGasLimit(shardHdr)
	// if err != nil {
	// 	return nil, nil, err
	// }

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
	log.Debug("started verifying proposed block",
		"epoch", headerHandler.GetEpoch(),
		"shard", headerHandler.GetShardID(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
	)

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

	go getMetricsFromBlockBody(body, sp.marshalizer, sp.appStatusHandler)

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

	txCounts, rewardCounts, unsignedCounts := sp.txCounter.getPoolCounts(sp.dataPool)
	log.Debug("total txs in pool", "counts", txCounts.String())
	log.Debug("total txs in rewards pool", "counts", rewardCounts.String())
	log.Debug("total txs in unsigned pool", "counts", unsignedCounts.String())

	go getMetricsFromHeader(header, uint64(txCounts.GetTotal()), sp.marshalizer, sp.appStatusHandler)

	sp.missingDataResolver.Reset()
	sp.missingDataResolver.RequestBlockTransactions(body)
	err = sp.missingDataResolver.RequestMissingMetaHeaders(header)
	if err != nil {
		return err
	}

	err = sp.missingDataResolver.WaitForMissingData(haveTime())
	if err != nil {
		return err
	}

	err = sp.requestEpochStartInfo(header, haveTime)
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

	// TODO: check that the limits for the block are not exceeded
	// err = sp.verifyGasLimit(shardHdr)
	// if err != nil {
	// 	return nil, nil, err
	// }

	return nil
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
) error {
	log.Debug("selectIncomingMiniBlocksForProposal has been started")

	sw := core.NewStopWatch()
	sw.Start("ComputeLongestMetaChainFromLastNotarized")
	orderedMetaBlocks, orderedMetaBlocksHashes, err := sp.blockTracker.ComputeLongestMetaChainFromLastNotarized()
	sw.Stop("ComputeLongestMetaChainFromLastNotarized")
	log.Debug("measurements", sw.GetMeasurements()...)
	if err != nil {
		return err
	}

	log.Debug("meta blocks ordered", "num meta blocks", len(orderedMetaBlocks))

	lastMetaHdr, _, err := sp.blockTracker.GetLastCrossNotarizedHeader(core.MetachainShardId)
	if err != nil {
		return err
	}

	err = sp.selectIncomingMiniBlocks(lastMetaHdr, orderedMetaBlocks, orderedMetaBlocksHashes, haveTime)
	if err != nil {
		return err
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

	return nil
}

func (sp *shardProcessor) selectIncomingMiniBlocks(
	lastCrossNotarizedMetaHdr data.HeaderHandler,
	orderedMetaBlocks []data.HeaderHandler,
	orderedMetaBlocksHashes [][]byte,
	haveTime func() bool,
) error {
	var currentMetaBlock data.HeaderHandler
	var currentMetaBlockHash []byte
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
		shouldContinue, errCreated := sp.createMbsCrossShardDstMe(currentMetaBlockHash, metaBlock, currProcessedMiniBlocksInfo)
		if errCreated != nil {
			return errCreated
		}
		if !shouldContinue {
			break
		}

		sp.miniBlocksSelectionSession.AddReferencedMetaBlock(currentMetaBlock, currentMetaBlockHash)
	}

	return nil
}

func (sp *shardProcessor) createMbsCrossShardDstMe(
	currentMetaBlockHash []byte,
	currentMetaBlock data.MetaHeaderHandler,
	miniBlockProcessingInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) (bool, error) {
	currMiniBlocksAdded, currNumTxsAdded, hdrFinished, errCreate := sp.txCoordinator.CreateMbsCrossShardDstMe(
		currentMetaBlock,
		miniBlockProcessingInfo,
	)
	if errCreate != nil {
		return false, errCreate
	}

	err := sp.miniBlocksSelectionSession.AddMiniBlocksAndHashes(currMiniBlocksAdded)
	if err != nil {
		return false, err
	}

	if !hdrFinished {
		log.Debug("meta block cannot be fully processed",
			"round", currentMetaBlock.GetRound(),
			"nonce", currentMetaBlock.GetNonce(),
			"hash", currentMetaBlockHash,
			"num mbs added", len(currMiniBlocksAdded),
			"num txs added", currNumTxsAdded)

		return false, nil
	}

	return true, nil
}

func (sp *shardProcessor) createProposalMiniBlocks(haveTime func() bool) error {
	if !haveTime() {
		log.Debug("shardProcessor.createProposalMiniBlocks", "error", process.ErrTimeIsOut)
		return nil
	}
	startTime := time.Now()
	err := sp.selectIncomingMiniBlocksForProposal(haveTime)
	if err != nil {
		return err
	}
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to create mbs to me", "time", elapsedTime)

	outgoingTransactions := sp.selectOutgoingTransactions()

	err = sp.miniBlocksSelectionSession.CreateAndAddMiniBlockFromTransactions(outgoingTransactions)
	if err != nil {
		log.Debug("shardProcessor.createProposalMiniBlocks", "error", err.Error())
		return err
	}

	// todo: maybe sanitize, removing empty miniBlocks

	return nil
}

func (sp *shardProcessor) selectOutgoingTransactions() [][]byte {
	log.Debug("selectOutgoingTransactions has been started")

	sw := core.NewStopWatch()
	sw.Start("selectOutgoingTransactions")
	defer func() {
		sw.Stop("selectOutgoingTransactions")
		log.Debug("measurements", sw.GetMeasurements()...)
	}()

	outgoingTransactions := sp.txCoordinator.SelectOutgoingTransactions()
	log.Debug("selectOutgoingTransactions has been finished",
		"num txs", len(outgoingTransactions))

	return outgoingTransactions
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
