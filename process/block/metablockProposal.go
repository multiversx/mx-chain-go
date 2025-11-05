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
)

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

	if metaHdr.IsStartOfEpochBlock() || metaHdr.GetEpochChangeProposed() {
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

	// TODO: referenced shard headers should be also notarized, not only the headers corresponding to the execution results
	// this will be needed for metachain to keep track of the already processed headers.
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

	marshalizedBody, err := mp.marshalizer.Marshal(body)
	if err != nil {
		return nil, nil, err
	}
	mp.blockSizeThrottler.Add(metaHdr.GetRound(), uint32(len(marshalizedBody)))

	return metaHdr, body, nil
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

// VerifyBlockProposal will be implemented in a further PR
func (mp *metaProcessor) VerifyBlockProposal(
	_ data.HeaderHandler,
	_ data.BodyHandler,
	_ func() time.Duration,
) error {
	return nil
}

// ProcessBlockProposal processes the proposed block. It returns nil if all ok or the specific error
func (mp *metaProcessor) ProcessBlockProposal(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) (data.BaseExecutionResultHandler, error) {
	return nil, nil
}

func (mp *metaProcessor) hasStartOfEpochExecutionResults(metaHeader data.MetaHeaderHandler) (bool, error) {
	if check.IfNil(metaHeader) {
		return false, process.ErrNilHeaderHandler
	}
	execResults := metaHeader.GetExecutionResultsHandlers()
	for _, execResult := range execResults {
		mbHeaders, err := common.GetMiniBlocksHeaderHandlersFromExecResult(execResult)
		if err != nil {
			return false, err
		}
		if hasRewardOrPeerMiniBlocksFromMeta(mbHeaders) {
			return true, nil
		}
	}
	return false, nil
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

	lastShardHdr, err := mp.getLastCrossNotarizedShardHdrs()
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

	return nil
}
