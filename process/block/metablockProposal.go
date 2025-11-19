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
)

const numHeadersToRequestInAdvance = 10

// usedShardHeadersInfo holds the used shard headers information
type usedShardHeadersInfo struct {
	headersPerShard          map[uint32][]ShardHeaderInfo
	orderedShardHeaders      []data.HeaderHandler
	orderedShardHeaderHashes [][]byte
}

// CreateNewHeaderProposal creates a new header
func (mp *metaProcessor) CreateNewHeaderProposal(round uint64, nonce uint64) (data.HeaderHandler, error) {
	// TODO: the trigger would need to be changed upon commit of a block with the epoch start results
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
	if mp.epochStartData == nil {
		return nil, process.ErrNilEpochStartData
	}

	// TODO: clean up the epoch start data upon commit
	err = metaHeader.SetEpochStartHandler(mp.epochStartData)
	if err != nil {
		return nil, err
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

	metaHdr.SoftwareVersion = []byte(mp.headerIntegrityVerifier.GetVersion(metaHdr.Epoch, metaHdr.Round))

	mp.epochStartTrigger.Update(metaHdr.Round, metaHdr.Nonce)
	if metaHdr.IsStartOfEpochBlock() || metaHdr.GetEpochChangeProposed() || mp.epochStartTrigger.IsEpochStart() {
		// no new transactions in start of epoch block
		// to simplify bootstrapping
		return metaHdr, &block.Body{}, nil
	}

	mp.gasComputation.Reset()
	mp.miniBlocksSelectionSession.ResetSelectionSession()
	err := mp.createBlockBodyProposal(metaHdr, haveTime)
	if err != nil {
		return nil, nil, err
	}

	mbsToMe := mp.miniBlocksSelectionSession.GetMiniBlocks()
	miniBlocksHeadersToMe := mp.miniBlocksSelectionSession.GetMiniBlockHeaderHandlers()
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

	txsInExecutionResults, err := getTxCountExecutionResults(metaHdr)
	if err != nil {
		return nil, nil, err
	}

	totalProcessedTxs := getTxCount(shardDataHandlers) + txsInExecutionResults
	// TODO: consider if tx count per metablock header is still needed
	// as we still have it in the execution results
	err = metaHdr.SetTxCount(totalProcessedTxs)
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

	mp.epochStartTrigger.Update(header.Round, header.Nonce)
	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	err = mp.checkHeaderBodyCorrelationProposal(header.GetMiniBlockHeaderHandlers(), body)
	if err != nil {
		return err
	}

	// TODO: analyse if it should be enforced that execution results on start of epoch block include only start of epoch execution results
	err = mp.executionResultsVerifier.VerifyHeaderExecutionResults(header)
	if err != nil {
		return err
	}

	err = mp.checkInclusionEstimationForExecutionResults(header)
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

	return mp.verifyGasLimit(header)
}

// ProcessBlockProposal processes the proposed block. It returns nil if all ok or the specific error
func (mp *metaProcessor) ProcessBlockProposal(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) (data.BaseExecutionResultHandler, error) {
	return nil, nil
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

	return mp.createProposalMiniBlocks(haveTime)
}

func (mp *metaProcessor) createProposalMiniBlocks(haveTime func() bool) error {
	if !haveTime() {
		log.Debug("metaProcessor.createProposalMiniBlocks", "error", process.ErrTimeIsOut)
		return nil
	}

	startTime := time.Now()
	err := mp.selectIncomingMiniBlocksForProposal(haveTime)
	if err != nil {
		return err
	}
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to create mbs to me", "time", elapsedTime)

	return nil
}

func (mp *metaProcessor) selectIncomingMiniBlocksForProposal(
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

	lastShardHdr, err := mp.getLastCrossNotarizedShardHeaders()
	if err != nil {
		return err
	}

	maxShardHeadersFromSameShard := core.MaxUint32(
		process.MinShardHeadersFromSameShardInOneMetaBlock,
		process.MaxShardHeadersAllowedInOneMetaBlock/mp.shardCoordinator.NumberOfShards(),
	)
	err = mp.selectIncomingMiniBlocks(lastShardHdr, orderedHdrs, orderedHdrsHashes, maxShardHeadersFromSameShard, haveTime)
	if err != nil {
		return err
	}

	return nil
}

func (mp *metaProcessor) selectIncomingMiniBlocks(
	lastShardHdr map[uint32]ShardHeaderInfo,
	orderedHdrs []data.HeaderHandler,
	orderedHdrsHashes [][]byte,
	maxShardHeadersFromSameShard uint32,
	haveTime func() bool,
) error {
	hdrsAdded := uint32(0)
	maxShardHeadersAllowedInOneMetaBlock := maxShardHeadersFromSameShard * mp.shardCoordinator.NumberOfShards()
	hdrsAddedForShard := make(map[uint32]uint32)
	var err error

	if len(orderedHdrs) != len(orderedHdrsHashes) {
		return process.ErrInconsistentShardHeadersAndHashes
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
		lastShardHeaderInfo, ok := lastShardHdr[currHdr.GetShardID()]
		if !ok {
			return process.ErrMissingHeader
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

		hasProofForHdr := mp.proofsPool.HasProof(currHdr.GetShardID(), currHdrHash)
		if !hasProofForHdr {
			log.Trace("no proof for shard header",
				"shard", currHdr.GetShardID(),
				"hash", logger.DisplayByteSlice(currHdrHash),
			)
			continue
		}

		if len(currHdr.GetMiniBlockHeadersWithDst(mp.shardCoordinator.SelfId())) == 0 {
			mp.miniBlocksSelectionSession.AddReferencedHeader(currHdr, currHdrHash)
			lastShardHdr[currHdr.GetShardID()] = ShardHeaderInfo{
				Header:      currHdr,
				Hash:        currHdrHash,
				UsedInBlock: true,
			}
			hdrsAddedForShard[currHdr.GetShardID()]++
			hdrsAdded++
			continue
		}

		createIncomingMbsResult, errCreated := mp.createMbsCrossShardDstMe(currHdrHash, currHdr, nil)
		if errCreated != nil {
			return errCreated
		}
		if !createIncomingMbsResult.HeaderFinished {
			mp.revertGasForCrossShardDstMeMiniBlocks(createIncomingMbsResult.AddedMiniBlocks, createIncomingMbsResult.PendingMiniBlocks)
			log.Debug("shard header cannot be fully added",
				"round", currHdr.GetRound(),
				"nonce", currHdr.GetNonce(),
				"hash", currHdrHash)
			break
		}

		if len(createIncomingMbsResult.AddedMiniBlocks) > 0 {
			err = mp.miniBlocksSelectionSession.AddMiniBlocksAndHashes(createIncomingMbsResult.AddedMiniBlocks)
			if err != nil {
				return err
			}
		}

		mp.miniBlocksSelectionSession.AddReferencedHeader(currHdr, currHdrHash)
		lastShardHdr[currHdr.GetShardID()] = ShardHeaderInfo{
			Header:      currHdr,
			Hash:        currHdrHash,
			UsedInBlock: true,
		}
		hdrsAddedForShard[currHdr.GetShardID()]++
		hdrsAdded++
	}

	go mp.requestShardHeadersInAdvanceIfNeeded(lastShardHdr)

	return nil
}

func (mp *metaProcessor) requestShardHeadersInAdvanceIfNeeded(
	lastShardHdr map[uint32]ShardHeaderInfo,
) {
	for shardID := uint32(0); shardID < mp.shardCoordinator.NumberOfShards(); shardID++ {
		mp.requestHeadersFromHeaderIfNeeded(lastShardHdr[shardID].Header)
	}
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
	epochStartDataMatches := mp.epochStartData.Equal(headerHandler.GetEpochStartHandler())
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
		header, err = mp.dataPool.Headers().GetHeaderByHash(execResult.GetHeaderHash())
		if err != nil {
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

	ok := mp.hasProofsForHeaders(usedShardHeaders.headersPerShard)
	if !ok {
		return process.ErrMissingHeaderProof
	}

	err = mp.verifyUsedShardHeadersValidity(usedShardHeaders.headersPerShard, lastCrossNotarizedHeader)
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
) error {
	var err error
	for shardID, hdrsForShard := range usedShardHeaders {
		err = mp.checkHeadersSequenceCorrectness(hdrsForShard, lastCrossNotarizedHeader[shardID])
		if err != nil {
			return err
		}
	}
	return nil
}

func (mp *metaProcessor) checkHeadersSequenceCorrectness(hdrsForShard []ShardHeaderInfo, lastNotarizedHeaderInfoForShard ShardHeaderInfo) error {
	var err error
	for _, shardHdrInfo := range hdrsForShard {
		if mp.isGenesisShardBlockAndFirstMeta(shardHdrInfo.Header.GetNonce()) {
			continue
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
		header, err = mp.dataPool.Headers().GetHeaderByHash(shardInfoHandler.GetHeaderHash())
		if err != nil {
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
