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
	// TODO: the trigger would need to be changed upon commit of a block with the epoch start results

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

	if metaHdr.IsStartOfEpochBlock() {
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
	shardDataHandler, err := mp.shardInfoCreateData.CreateShardInfoV3(metaHdr, referencedShardHeaders, referencedShardHeaderHashes)
	if err != nil {
		return nil, nil, err
	}

	err = metaHdr.SetShardInfoHandlers(shardDataHandler)
	if err != nil {
		return nil, nil, err
	}

	err = metaHdr.SetMiniBlockHeaderHandlers(miniBlocksHeadersToMe)
	if err != nil {
		return nil, nil, err
	}

	totalProcessedTxs := getTxCount(shardDataHandler) + getTxCountExecutionResults(metaHdr)
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

func getTxCountExecutionResults(metaHeader data.MetaHeaderHandler) uint32 {
	totalTxs := uint64(0)
	execResults := metaHeader.GetExecutionResultsHandlers()
	for _, execResult := range execResults {
		execResultsMeta, ok := execResult.(data.MetaExecutionResultHandler)
		if !ok {
			continue
		}
		totalTxs += execResultsMeta.GetExecutedTxCount()
	}
	return uint32(totalTxs)
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
	execResults := metaHeader.GetExecutionResultsHandlers()
	for _, execResult := range execResults {
		mbHeaders, err := mp.extractMiniBlocksHeaderHandlersFromExecResult(execResult, common.MetachainShardId)
		if err != nil {
			return false, err
		}
		if hasRewardMiniBlocks(mbHeaders) {
			return true, nil
		}
	}
	return false, nil
}

func hasRewardMiniBlocks(miniBlockHeaders []data.MiniBlockHeaderHandler) bool {
	for _, mbHeader := range miniBlockHeaders {
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

	err = mp.selectIncomingMiniBlocks(lastShardHdr, orderedHdrs, orderedHdrsHashes, haveTime)
	if err != nil {
		return err
	}

	return nil
}

func (mp *metaProcessor) selectIncomingMiniBlocks(
	lastShardHdr map[uint32]ShardHeaderInfo,
	orderedHdrs []data.HeaderHandler,
	orderedHdrsHashes [][]byte,
	haveTime func() bool,
) error {
	hdrsAdded := uint32(0)
	maxShardHeadersFromSameShard := core.MaxUint32(
		process.MinShardHeadersFromSameShardInOneMetaBlock,
		process.MaxShardHeadersAllowedInOneMetaBlock/mp.shardCoordinator.NumberOfShards(),
	)
	maxShardHeadersAllowedInOneMetaBlock := maxShardHeadersFromSameShard * mp.shardCoordinator.NumberOfShards()
	hdrsAddedForShard := make(map[uint32]uint32)
	for i := 0; i < len(orderedHdrs); i++ {
		if !haveTime() {
			log.Debug("time is up after putting cross txs with destination to current shard",
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

		currShardHdr := orderedHdrs[i]
		if currShardHdr.GetNonce() > lastShardHdr[currShardHdr.GetShardID()].Header.GetNonce()+1 {
			log.Trace("skip searching",
				"shard", currShardHdr.GetShardID(),
				"last shard hdr nonce", lastShardHdr[currShardHdr.GetShardID()].Header.GetNonce(),
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

		hasProofForHdr := mp.proofsPool.HasProof(currShardHdr.GetShardID(), orderedHdrsHashes[i])
		if !hasProofForHdr {
			log.Trace("no proof for shard header",
				"shard", currShardHdr.GetShardID(),
				"hash", logger.DisplayByteSlice(orderedHdrsHashes[i]),
			)
			continue
		}

		if len(currShardHdr.GetMiniBlockHeadersWithDst(mp.shardCoordinator.SelfId())) == 0 {
			mp.miniBlocksSelectionSession.AddReferencedHeader(orderedHdrs[i], orderedHdrsHashes[i])
			continue
		}

		_, pendingMiniBlocks, errCreated := mp.createMbsCrossShardDstMe(orderedHdrsHashes[i], currShardHdr, nil, true)
		if errCreated != nil {
			return errCreated
		}

		// pending miniBlocks were already reverted, but still returned to check if we need to stop adding more shard headers
		if len(pendingMiniBlocks) > 0 {
			log.Debug("shard header cannot be fully added",
				"round", currShardHdr.GetRound(),
				"nonce", currShardHdr.GetNonce(),
				"hash", orderedHdrsHashes[i])
			break
		}

		mp.miniBlocksSelectionSession.AddReferencedHeader(currShardHdr, orderedHdrsHashes[i])
	}

	return nil
}
