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
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
)

type pendingBlocksAfterSelection struct {
	headerHash        []byte
	header            data.HeaderHandler
	pendingMiniBlocks map[string]*block.MiniBlock
}

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

	err = sp.verifyGasLimit(shardHdr, miniBlocks)
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

	err = sp.verifyGasLimit(header, body.MiniBlocks)
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

	// TODO: check again before saving the last executed result
	err = sp.blockChainHook.SetCurrentHeader(header)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			sp.RevertCurrentBlock()
		}
	}()

	err = sp.checkAndUpdateContextBeforeExecution(header, headerHash)
	if err != nil {
		return nil, err
	}

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

	var executionResult data.BaseExecutionResultHandler
	executionResult, err = sp.collectExecutionResults(headerHash, header, body)
	if err != nil {
		return nil, err
	}

	err = sp.blockProcessingCutoffHandler.HandleProcessErrorCutoff(header)
	if err != nil {
		return nil, err
	}

	err = sp.commitState(headerHandler)
	if err != nil {
		return nil, err
	}

	sp.saveEpochStartEconomicsIfNeeded(header)

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
) ([]*pendingBlocksAfterSelection, error) {
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

	pendingBlocks, err := sp.selectIncomingMiniBlocks(lastMetaHdr, orderedMetaBlocks, orderedMetaBlocksHashes, haveTime)
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

	return pendingBlocks, nil
}

func (sp *shardProcessor) selectIncomingMiniBlocks(
	lastCrossNotarizedMetaHdr data.HeaderHandler,
	orderedMetaBlocks []data.HeaderHandler,
	orderedMetaBlocksHashes [][]byte,
	haveTime func() bool,
) ([]*pendingBlocksAfterSelection, error) {
	var currentMetaBlock data.HeaderHandler
	var currentMetaBlockHash []byte
	var pendingBlocks []*pendingBlocksAfterSelection
	var errCreated error
	var createIncomingMbsResult *CrossShardIncomingMbsCreationResult
	lastMeta := lastCrossNotarizedMetaHdr
	lastMetaAdded := lastCrossNotarizedMetaHdr

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
			// if all headers were completely included so far, include this one as well
			if len(pendingBlocks) == 0 {
				sp.miniBlocksSelectionSession.AddReferencedHeader(currentMetaBlock, currentMetaBlockHash)
				lastMeta = currentMetaBlock
				lastMetaAdded = currentMetaBlock
				continue
			}

			// if at least one previous header was not completely included, save the current one as pending but with no mini blocks
			pendingBlocks = append(pendingBlocks, &pendingBlocksAfterSelection{
				headerHash:        currentMetaBlockHash,
				header:            currentMetaBlock,
				pendingMiniBlocks: nil,
			})
			lastMeta = currentMetaBlock
			continue
		}

		currProcessedMiniBlocksInfo := sp.processedMiniBlocksTracker.GetProcessedMiniBlocksInfo(currentMetaBlockHash)
		createIncomingMbsResult, errCreated = sp.createMbsCrossShardDstMe(currentMetaBlockHash, metaBlock, currProcessedMiniBlocksInfo)
		if errCreated != nil {
			return nil, errCreated
		}

		if len(createIncomingMbsResult.AddedMiniBlocks) > 0 {
			errAdd := sp.miniBlocksSelectionSession.AddMiniBlocksAndHashes(createIncomingMbsResult.AddedMiniBlocks)
			if errAdd != nil {
				return nil, errAdd
			}
			sp.miniBlocksSelectionSession.AddReferencedHeader(currentMetaBlock, currentMetaBlockHash)
			lastMeta = currentMetaBlock
			lastMetaAdded = currentMetaBlock
		}

		if createIncomingMbsResult.HeaderFinished {
			continue
		}

		pendingBlocks = append(pendingBlocks, &pendingBlocksAfterSelection{
			headerHash:        currentMetaBlockHash,
			header:            currentMetaBlock,
			pendingMiniBlocks: miniBlocksSliceToMap(createIncomingMbsResult.PendingMiniBlocks),
		})

		canAddMorePendingMiniBlocks := sp.gasComputation.CanAddPendingIncomingMiniBlocks()
		if !canAddMorePendingMiniBlocks {
			break
		}

		// continue saving pending mini blocks until they are done or there is no possible space left in the block
		lastMeta = currentMetaBlock
	}

	go sp.requestHeadersFromHeaderIfNeeded(lastMetaAdded)

	return pendingBlocks, nil
}

func miniBlocksSliceToMap(miniBlocksAndHashes []block.MiniblockAndHash) map[string]*block.MiniBlock {
	result := make(map[string]*block.MiniBlock, len(miniBlocksAndHashes))
	for _, miniBlockAndHash := range miniBlocksAndHashes {
		result[string(miniBlockAndHash.Hash)] = miniBlockAndHash.Miniblock
	}
	return result
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
	pendingBlocks, err := sp.selectIncomingMiniBlocksForProposal(haveTime)
	if err != nil {
		return err
	}
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to create mbs to me", "time", elapsedTime)

	outgoingTransactions, pendingIncomingMiniBlocksAdded := sp.selectOutgoingTransactions(nonce, haveTime)

	err = sp.appendPendingMiniBlocksAfterSelectingOutgoingTransactions(pendingBlocks, pendingIncomingMiniBlocksAdded)
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

func (sp *shardProcessor) appendPendingMiniBlocksAfterSelectingOutgoingTransactions(
	pendingBlocksLeft []*pendingBlocksAfterSelection,
	pendingIncomingMiniBlocksAdded []data.MiniBlockHeaderHandler,
) error {
	if len(pendingIncomingMiniBlocksAdded) == 0 {
		// nothing else added, won't append more
		return nil
	}

	referencedHeaders := sp.miniBlocksSelectionSession.GetReferencedHeaders()
	if len(referencedHeaders) == 0 {
		log.Error("appendPendingMiniBlocksAfterSelectingOutgoingTransactions: no header referenced yet")
		return process.ErrNoReferencedHeader
	}

	extraMiniBlocksAdded := make([]block.MiniblockAndHash, len(pendingIncomingMiniBlocksAdded))
	lastNonceReferenced := referencedHeaders[len(referencedHeaders)-1].GetNonce()
	lastHeaderReferencedFinished := false
	for i, pendingMbAdded := range pendingIncomingMiniBlocksAdded {
		miniBlockAndHash, headerHash, header, found, isHeaderFinished := findPendingMiniBlock(pendingBlocksLeft, pendingMbAdded)
		if !found {
			log.Error("pending mini block added does not exists in the remaining pending list")
			return process.ErrInvalidHash
		}

		extraMiniBlocksAdded[i] = miniBlockAndHash

		// if this is still the last one referenced, continue adding its mini blocks
		// possible gaps should have been filled already and the current one already referenced
		if header.GetNonce() == lastNonceReferenced {
			continue
		}

		// if the header is consecutive to the previous one, reference it
		if header.GetNonce() == lastNonceReferenced+1 {
			sp.miniBlocksSelectionSession.AddReferencedHeader(header, headerHash)
			lastNonceReferenced = header.GetNonce()
			lastHeaderReferencedFinished = isHeaderFinished
			continue
		}

		// if the header is not consecutive, check for all headers that should be in between
		// return error if missing, all of them should be available and without mini blocks with dest me
		for missingNonce := lastNonceReferenced + 1; missingNonce < header.GetNonce(); missingNonce++ {
			hash, hdr, err := findPendingHeaderWithNonceAndNoMiniBlocksDstMe(missingNonce, pendingBlocksLeft)
			if err != nil {
				return err
			}

			sp.miniBlocksSelectionSession.AddReferencedHeader(hdr, hash)
		}

		sp.miniBlocksSelectionSession.AddReferencedHeader(header, headerHash)
		lastNonceReferenced = header.GetNonce()
		lastHeaderReferencedFinished = isHeaderFinished
	}

	// if the last header referenced was finished, continue referencing headers that do not have mini blocks dest me
	if lastHeaderReferencedFinished {
		referencedHeaders = sp.miniBlocksSelectionSession.GetReferencedHeaders()
		lastNonceReferenced = referencedHeaders[len(referencedHeaders)-1].GetNonce()
		for {
			hash, hdr, err := findPendingHeaderWithNonceAndNoMiniBlocksDstMe(lastNonceReferenced+1, pendingBlocksLeft)
			if err != nil {
				break
			}

			sp.miniBlocksSelectionSession.AddReferencedHeader(hdr, hash)
			lastNonceReferenced = hdr.GetNonce()
		}
	}

	return sp.miniBlocksSelectionSession.AddMiniBlocksAndHashes(extraMiniBlocksAdded)
}

func findPendingMiniBlock(
	pendingBlocksLeft []*pendingBlocksAfterSelection,
	pendingMbAdded data.MiniBlockHeaderHandler,
) (block.MiniblockAndHash, []byte, data.HeaderHandler, bool, bool) {
	for _, pendingMiniBlocksForHeader := range pendingBlocksLeft {
		miniBlock, ok := pendingMiniBlocksForHeader.pendingMiniBlocks[string(pendingMbAdded.GetHash())]
		if ok {
			// update the map
			delete(pendingMiniBlocksForHeader.pendingMiniBlocks, string(pendingMbAdded.GetHash()))
			isHeaderFinished := len(pendingMiniBlocksForHeader.pendingMiniBlocks) == 0

			return block.MiniblockAndHash{
					Miniblock: miniBlock,
					Hash:      pendingMbAdded.GetHash(),
				},
				pendingMiniBlocksForHeader.headerHash,
				pendingMiniBlocksForHeader.header,
				ok,
				isHeaderFinished
		}
	}

	return block.MiniblockAndHash{}, nil, nil, false, false
}

func findPendingHeaderWithNonceAndNoMiniBlocksDstMe(
	nonce uint64,
	pendingBlocksLeft []*pendingBlocksAfterSelection,
) ([]byte, data.HeaderHandler, error) {
	for _, pendingBlock := range pendingBlocksLeft {
		if pendingBlock.header.GetNonce() == nonce {
			hasPendingMiniBlocks := len(pendingBlock.pendingMiniBlocks) > 0
			if !hasPendingMiniBlocks {
				return pendingBlock.headerHash, pendingBlock.header, nil
			}

			log.Error("findPendingHeaderWithNonceAndNoMiniBlocksDstMe: pending block should not have pending mini blocks with destination me", "nonce", nonce, "pending mini blocks", len(pendingBlock.pendingMiniBlocks))

			return nil, nil, process.ErrInvalidHeader
		}
	}

	return nil, nil, process.ErrMissingHeader
}

func (sp *shardProcessor) selectOutgoingTransactions(
	nonce uint64,
	haveTimeForSelection func() bool,
) ([][]byte, []data.MiniBlockHeaderHandler) {
	log.Debug("selectOutgoingTransactions has been started")

	sw := core.NewStopWatch()
	sw.Start("selectOutgoingTransactions")
	defer func() {
		sw.Stop("selectOutgoingTransactions")
		log.Debug("measurements", sw.GetMeasurements()...)
	}()

	outgoingTransactions, pendingIncomingMiniBlocksAdded := sp.txCoordinator.SelectOutgoingTransactions(nonce, haveTimeForSelection)
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
	_, usedMetaHeaders, err := sp.getReferencedMetaHeadersFromPool(header)
	if err != nil {
		return fmt.Errorf("%w : checkMetaHeadersValidityAndFinalityProposal -> getReferencedMetaHeadersFromPool", err)
	}

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

func (sp *shardProcessor) getReferencedMetaHeadersFromPool(header data.ShardHeaderHandler) ([][]byte, []data.HeaderHandler, error) {
	usedMetaHdrHashes := header.GetMetaBlockHashes()
	usedMetaHeaders := make([]data.HeaderHandler, 0, len(usedMetaHdrHashes))
	var metaHdr data.HeaderHandler
	var err error
	for _, metaHdrHash := range usedMetaHdrHashes {
		metaHdr, err = sp.dataPool.Headers().GetHeaderByHash(metaHdrHash)
		if err != nil {
			return nil, nil, err
		}
		usedMetaHeaders = append(usedMetaHeaders, metaHdr)
	}

	return usedMetaHdrHashes, usedMetaHeaders, nil
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

	sp.cacheOrderedTxHashes(headerHash)
	sp.cacheUnexecutableTxHashes(headerHash)

	return executionResult, nil
}

func (sp *shardProcessor) getOrderedProcessedMetaBlocksFromMiniBlockHashesV3(
	header data.HeaderHandler,
	miniBlockHashes map[int][]byte,
) ([]data.HeaderHandler, error) {
	shardHeader, ok := header.(data.ShardHeaderHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}
	metaHeaderHashes, metaHeaders, err := sp.getReferencedMetaHeadersFromPool(shardHeader)
	if err != nil {
		return nil, err
	}

	hashSet := make(map[string]struct{}, len(miniBlockHashes))
	for _, b := range miniBlockHashes {
		hashSet[string(b)] = struct{}{}
	}

	fullyReferencedMetaBlocks := make([]data.HeaderHandler, 0, len(metaHeaders))
	var remaining int
	var metaHeaderHash []byte
	for i, metaHeader := range metaHeaders {
		metaHeaderHash = metaHeaderHashes[i]
		crossMiniBlockHashes := metaHeader.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
		if len(crossMiniBlockHashes) == 0 {
			fullyReferencedMetaBlocks = append(fullyReferencedMetaBlocks, metaHeader)
			continue
		}

		for hash := range crossMiniBlockHashes {
			if sp.processedMiniBlocksTracker.IsMiniBlockFullyProcessed(metaHeaderHash, []byte(hash)) {
				hashSet[hash] = struct{}{}
			}
		}

		remaining = len(crossMiniBlockHashes)
		for k := range crossMiniBlockHashes {
			_, found := hashSet[k]
			if found {
				remaining--
			}
		}
		if remaining == 0 {
			fullyReferencedMetaBlocks = append(fullyReferencedMetaBlocks, metaHeader)
		}
	}

	process.SortHeadersByNonce(fullyReferencedMetaBlocks)

	return fullyReferencedMetaBlocks, nil
}

func (sp *shardProcessor) saveEpochStartEconomicsIfNeeded(header data.ShardHeaderHandler) {
	var prevHashOfEpochChangeProposed []byte
	var metaEpochChangeProposedHeader data.MetaHeaderHandler
	var metaEpochChangeHeader data.MetaHeaderHandler
	for _, metaHash := range header.GetMetaBlockHashes() {
		hdr, err := sp.getHeaderFromHash(header.IsHeaderV3(), metaHash, core.MetachainShardId)
		if err != nil {
			continue
		}
		metaHdr, ok := hdr.(data.MetaHeaderHandler)
		if !ok {
			continue
		}

		// save epoch change proposed header if found
		if metaHdr.IsEpochChangeProposed() {
			prevHashOfEpochChangeProposed = metaHdr.GetPrevHash()
			metaEpochChangeProposedHeader = metaHdr
			continue
		}

		// save epoch change header if found
		if metaHdr.IsStartOfEpochBlock() {
			metaEpochChangeHeader = metaHdr
			continue
		}
	}

	// if both nil, no metric to be saved
	if check.IfNil(metaEpochChangeHeader) && check.IfNil(metaEpochChangeProposedHeader) {
		return
	}

	// if none of them are nil, iterate through all execution results
	if !check.IfNil(metaEpochChangeHeader) && !check.IfNil(metaEpochChangeProposedHeader) {
		execResults := metaEpochChangeProposedHeader.GetExecutionResultsHandlers()
		execResults = append(execResults, metaEpochChangeHeader.GetExecutionResultsHandlers()...)
		// having both headers implies having prevHashOfEpochChangeProposed as well
		sp.setEpochStartMetricsV3FromExecutionResults(prevHashOfEpochChangeProposed, execResults)
		return
	}

	// if only metaEpochChangeProposedHeader is available, try to set the metrics from it
	// this implies it holds the execution result of its previous header
	if !check.IfNil(metaEpochChangeProposedHeader) {
		sp.setEpochStartMetricsV3FromExecutionResults(prevHashOfEpochChangeProposed, metaEpochChangeProposedHeader.GetExecutionResultsHandlers())
		return
	}

	// if only metaEpochChangeHeader is available, try to set the metrics from it
	// this implies it holds the execution result of epoch start proposed header and its previous
	if !check.IfNil(metaEpochChangeHeader) {
		// first extract epoch change proposed header to get its prev hash
		for _, execResult := range metaEpochChangeHeader.GetExecutionResultsHandlers() {
			headerFromExecResult, err := sp.getHeaderFromHash(true, execResult.GetHeaderHash(), core.MetachainShardId)
			if err != nil {
				// saving the metric should not be blocking
				continue
			}

			metaHeaderFromExecResult, ok := headerFromExecResult.(data.MetaHeaderHandler)
			if !ok {
				// saving the metric should not be blocking
				continue
			}

			if !metaHeaderFromExecResult.IsEpochChangeProposed() {
				continue
			}

			prevHashOfEpochChangeProposed = metaHeaderFromExecResult.GetPrevHash()
		}

		sp.setEpochStartMetricsV3FromExecutionResults(prevHashOfEpochChangeProposed, metaEpochChangeHeader.GetExecutionResultsHandlers())
	}
}

func (sp *shardProcessor) setEpochStartMetricsV3FromExecutionResults(
	prevHashOfEpochChangeProposed []byte,
	executionResults []data.BaseExecutionResultHandler,
) {
	for _, prevExecResult := range executionResults {
		if !bytes.Equal(prevHashOfEpochChangeProposed, prevExecResult.GetHeaderHash()) {
			continue
		}

		metaExecResult, okMetaExecResultCast := prevExecResult.(data.BaseMetaExecutionResultHandler)
		if !okMetaExecResultCast {
			continue
		}

		sp.appStatusHandler.SetStringValue(common.MetricTotalFees, metaExecResult.GetAccumulatedFeesInEpoch().String())
		sp.appStatusHandler.SetStringValue(common.MetricDevRewardsInEpoch, metaExecResult.GetDevFeesInEpoch().String())
		return
	}
}
