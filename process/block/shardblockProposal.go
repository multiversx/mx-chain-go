package block

import (
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

// CreateBlockProposal - creates a block proposal without executing any of the transactions
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
	numToInclude := sp.executionResultsInclusionEstimator.Decide(lastNotarizedExecutionResultInfo, pendingExecutionResults, shardHeader.GetTimeStamp())

	executionResultsToInclude := pendingExecutionResults[:numToInclude]
	if len(executionResultsToInclude) > 0 {
		lastExecutionResult := executionResultsToInclude[len(executionResultsToInclude)-1]
		lastExecutionResultForCurrentBlock, err = process.CreateLastExecutionResultInfoFromExecutionResult(shardHeader.GetRound(), lastExecutionResult, sp.shardCoordinator.SelfId())
	} else {
		lastExecutionResultForCurrentBlock = lastExecutionResultHandler
	}

	err = shardHeader.SetLastExecutionResultHandler(lastExecutionResultForCurrentBlock)
	if err != nil {
		return err
	}

	return shardHeader.SetExecutionResultsHandlers(ExecutionHandlersToBaseExecutionHandlers(executionResultsToInclude))
}

// ExecutionHandlersToBaseExecutionHandlers converts a slice of ExecutionResultHandler to a slice of BaseExecutionResultHandler
func ExecutionHandlersToBaseExecutionHandlers(execHandlers []data.ExecutionResultHandler) []data.BaseExecutionResultHandler {
	baseExecHandlers := make([]data.BaseExecutionResultHandler, len(execHandlers))
	for i, execHandler := range execHandlers {
		baseExecHandlers[i] = execHandler
	}

	return baseExecHandlers
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
		if len(currentMetaBlock.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())) == 0 {
			sp.miniBlocksSelectionSession.AddReferencedMetaBlock(orderedMetaBlocks[i], orderedMetaBlocksHashes[i])
			continue
		}

		metaBlock, ok := currentMetaBlock.(*block.MetaBlock)
		if !ok {
			log.Warn("selectIncomingMiniBlocks: wrong type assertion for meta block")
			break
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
	currentMetaBlock *block.MetaBlock,
	miniBlockProcessingInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) (bool, error) {
	// if miniBlock was partially executed before, we can continue processing it
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

	outgoingTransactions, err := sp.selectOutgoingTransactions()
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

func (sp *shardProcessor) selectOutgoingTransactions() ([][]byte, error) {
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

	return outgoingTransactions, nil
}
