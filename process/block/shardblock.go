package block

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/headerVersionData"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	processOutport "github.com/multiversx/mx-chain-go/outport/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/bootstrapStorage"
	"github.com/multiversx/mx-chain-go/process/block/helpers"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/state"
)

var _ process.BlockProcessor = (*shardProcessor)(nil)

const (
	timeBetweenCheckForEpochStart = 100 * time.Millisecond
	pruningDelay                  = 10
)

type createAndProcessMiniBlocksDestMeInfo struct {
	currMetaHdr                 data.HeaderHandler
	currMetaHdrHash             []byte
	currProcessedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo
	allProcessedMiniBlocksInfo  map[string]*processedMb.ProcessedMiniBlockInfo
	haveTime                    func() bool
	haveAdditionalTime          func() bool
	miniBlocks                  block.MiniBlockSlice
	hdrAdded                    bool
	numTxsAdded                 uint32
	numHdrsAdded                uint32
	scheduledMode               bool
}

// shardProcessor implements shardProcessor interface, and actually it tries to execute block
type shardProcessor struct {
	*baseProcessor
	metaBlockFinality uint32
}

// NewShardProcessor creates a new shardProcessor object
func NewShardProcessor(arguments ArgShardProcessor) (*shardProcessor, error) {
	base, err := NewBaseProcessor(arguments.ArgBaseProcessor)
	if err != nil {
		return nil, err
	}

	if check.IfNil(arguments.DataComponents.Datapool().Transactions()) {
		return nil, process.ErrNilTransactionPool
	}

	sp := shardProcessor{
		baseProcessor: base,
	}

	argsTransactionCounter := ArgsTransactionCounter{
		AppStatusHandler: sp.appStatusHandler,
		Hasher:           sp.hasher,
		Marshalizer:      sp.marshalizer,
		ShardID:          sp.shardCoordinator.SelfId(),
	}
	sp.txCounter, err = NewTransactionCounter(argsTransactionCounter)
	if err != nil {
		return nil, err
	}

	sp.requestBlockBodyHandler = &sp
	sp.blockProcessor = &sp

	sp.metaBlockFinality = process.BlockFinality

	return &sp, nil
}

// VerifyProposedBlock verifies the proposed block. It returns nil if all ok or the specific error
func (sp *shardProcessor) VerifyProposedBlock(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
	headerHash []byte,
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

	// todo: check if the following checks are still required
	// sp.roundNotifier.CheckRound(headerHandler)
	// sp.epochNotifier.CheckEpoch(headerHandler)

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

	// todo: need to change reserved field verification for Supernova
	err = sp.checkHeaderBodyCorrelation(header.GetMiniBlockHeaderHandlers(), body)
	if err != nil {
		return err
	}

	err = sp.checkMiniBlocksConstructionState(header)
	if err != nil {
		return err
	}

	err = sp.executionResultsVerifier.VerifyHeaderExecutionResults(headerHash, header)
	if err != nil {
		return err
	}

	txCounts, rewardCounts, unsignedCounts := sp.txCounter.getPoolCounts(sp.dataPool)
	log.Debug("total txs in pool", "counts", txCounts.String())
	log.Debug("total txs in rewards pool", "counts", rewardCounts.String())
	log.Debug("total txs in unsigned pool", "counts", unsignedCounts.String())

	go getMetricsFromHeader(header, uint64(txCounts.GetTotal()), sp.marshalizer, sp.appStatusHandler)

	sp.txCoordinator.RequestBlockTransactions(body)
	// TODO: this needs to be initialized properly for proposed blocks
	sp.missingDataResolver.Reset()
	err = sp.missingDataResolver.RequestMissingMetaHeaders(header)
	if err != nil {
		return err
	}

	err = sp.txCoordinator.IsDataPreparedForProcessing(haveTime)
	if err != nil {
		return err
	}

	// err = sp.hdrsForCurrBlock.WaitForHeadersIfNeeded(haveTime)
	// if err != nil {
	// 	return err
	// }
	//
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

	err = sp.checkMetaHeadersValidityAndFinality()
	if err != nil {
		return err
	}

	err = sp.verifyCrossShardMiniBlockDstMe(header)
	if err != nil {
		return err
	}
	//
	// err = sp.blockChainHook.SetCurrentHeader(header)
	// if err != nil {
	// 	return err
	// }
	//
	// defer func() {
	// 	if err != nil {
	// 		sp.RevertCurrentBlock()
	// 	}
	// }()
	//
	// mbIndex := sp.getIndexOfFirstMiniBlockToBeExecuted(header)
	// miniBlocks := body.MiniBlocks[mbIndex:]
	//
	// startTime := time.Now()
	// err = sp.txCoordinator.ProcessBlockTransaction(header, &block.Body{MiniBlocks: miniBlocks}, haveTime)
	// elapsedTime := time.Since(startTime)
	// log.Debug("elapsed time to process block transaction",
	// 	"time [s]", elapsedTime,
	// )
	// if err != nil {
	// 	return err
	// }
	//
	// err = sp.txCoordinator.VerifyCreatedBlockTransactions(header, &block.Body{MiniBlocks: miniBlocks})
	// if err != nil {
	// 	return err
	// }
	//
	// err = sp.txCoordinator.VerifyCreatedMiniBlocks(header, body)
	// if err != nil {
	// 	return err
	// }
	//
	// err = sp.verifyFees(header)
	// if err != nil {
	// 	return err
	// }
	//
	// if !sp.verifyStateRoot(header.GetRootHash()) {
	// 	err = process.ErrRootStateDoesNotMatch
	// 	return err
	// }
	//
	// err = sp.blockProcessingCutoffHandler.HandleProcessErrorCutoff(header)
	// if err != nil {
	// 	return err
	// }
	//
	return nil
}

func (sp *shardProcessor) checkMiniBlocksConstructionState(header data.HeaderHandler) error {
	miniBlockHeaders := header.GetMiniBlockHeaderHandlers()
	for i := 0; i < len(miniBlockHeaders); i++ {
		// for Supernova all miniBlocks not part of an execution result need to have construction state Proposed
		if miniBlockHeaders[i].GetConstructionState() != int32(block.Proposed) {
			return process.ErrWrongMiniBlockConstructionState
		}
	}
	return nil
}

// ProcessProposedBlock processes a proposed block. It returns nil if all ok or the specific error
func (sp *shardProcessor) ProcessProposedBlock(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
	haveTime func() time.Duration,
) error {

	return nil
}

// ProcessBlock processes a block. It returns nil if all ok or the specific error
func (sp *shardProcessor) ProcessBlock(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
	haveTime func() time.Duration,
) error {
	if haveTime == nil {
		return process.ErrNilHaveTimeHandler
	}

	sp.processStatusHandler.SetBusy("shardProcessor.ProcessBlock")
	defer sp.processStatusHandler.SetIdle()

	err := sp.checkBlockValidity(headerHandler, bodyHandler)
	if err != nil {
		if errors.Is(err, process.ErrBlockHashDoesNotMatch) {
			log.Debug("requested missing shard header",
				"hash", headerHandler.GetPrevHash(),
				"for shard", headerHandler.GetShardID(),
			)

			go sp.requestHandler.RequestShardHeader(headerHandler.GetShardID(), headerHandler.GetPrevHash())
		}

		return err
	}

	sp.roundNotifier.CheckRound(headerHandler)
	sp.epochNotifier.CheckEpoch(headerHandler)
	sp.requestHandler.SetEpoch(headerHandler.GetEpoch())

	err = sp.checkScheduledRootHash(headerHandler)
	if err != nil {
		return err
	}

	log.Debug("started processing block",
		"epoch", headerHandler.GetEpoch(),
		"shard", headerHandler.GetShardID(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
	)

	header, ok := headerHandler.(data.ShardHeaderHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	go getMetricsFromBlockBody(body, sp.marshalizer, sp.appStatusHandler)

	err = sp.checkHeaderBodyCorrelation(header.GetMiniBlockHeaderHandlers(), body)
	if err != nil {
		return err
	}

	err = sp.checkScheduledMiniBlocksValidity(headerHandler)
	if err != nil {
		return err
	}

	txCounts, rewardCounts, unsignedCounts := sp.txCounter.getPoolCounts(sp.dataPool)
	log.Debug("total txs in pool", "counts", txCounts.String())
	log.Debug("total txs in rewards pool", "counts", rewardCounts.String())
	log.Debug("total txs in unsigned pool", "counts", unsignedCounts.String())

	go getMetricsFromHeader(header, uint64(txCounts.GetTotal()), sp.marshalizer, sp.appStatusHandler)

	err = sp.createBlockStarted()
	if err != nil {
		return err
	}

	sp.txCoordinator.RequestBlockTransactions(body)
	sp.hdrsForCurrBlock.RequestMetaHeaders(header)

	if haveTime() < 0 {
		return process.ErrTimeIsOut
	}

	err = sp.txCoordinator.IsDataPreparedForProcessing(haveTime)
	if err != nil {
		return err
	}

	err = sp.hdrsForCurrBlock.WaitForHeadersIfNeeded(haveTime)
	if err != nil {
		return err
	}

	err = sp.requestEpochStartInfo(header, haveTime)
	if err != nil {
		return err
	}

	if sp.accountsDB[state.UserAccountsState].JournalLen() != 0 {
		log.Error("shardProcessor.ProcessBlock first entry", "stack", string(sp.accountsDB[state.UserAccountsState].GetStackDebugFirstEntry()))
		return process.ErrAccountStateDirty
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

	err = sp.checkMetaHeadersValidityAndFinality()
	if err != nil {
		return err
	}

	err = sp.verifyCrossShardMiniBlockDstMe(header)
	if err != nil {
		return err
	}

	err = sp.blockChainHook.SetCurrentHeader(header)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			sp.RevertCurrentBlock()
		}
	}()

	mbIndex := sp.getIndexOfFirstMiniBlockToBeExecuted(header)
	miniBlocks := body.MiniBlocks[mbIndex:]

	startTime := time.Now()
	err = sp.txCoordinator.ProcessBlockTransaction(header, &block.Body{MiniBlocks: miniBlocks}, haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to process block transaction",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	err = sp.txCoordinator.VerifyCreatedBlockTransactions(header, &block.Body{MiniBlocks: miniBlocks})
	if err != nil {
		return err
	}

	err = sp.txCoordinator.VerifyCreatedMiniBlocks(header, body)
	if err != nil {
		return err
	}

	err = sp.verifyFees(header)
	if err != nil {
		return err
	}

	if !sp.verifyStateRoot(header.GetRootHash()) {
		err = process.ErrRootStateDoesNotMatch
		return err
	}

	err = sp.blockProcessingCutoffHandler.HandleProcessErrorCutoff(header)
	if err != nil {
		return err
	}

	return nil
}

func (sp *shardProcessor) requestEpochStartInfo(header data.ShardHeaderHandler, haveTime func() time.Duration) error {
	if !header.IsStartOfEpochBlock() {
		return nil
	}
	if sp.epochStartTrigger.MetaEpoch() >= header.GetEpoch() {
		return nil
	}
	if sp.epochStartTrigger.IsEpochStart() {
		return nil
	}

	// force header cleanup from pool so that the receiving of the epoch start meta block will reach the trigger
	sp.dataPool.Headers().RemoveHeaderByHash(header.GetEpochStartMetaHash())
	go sp.requestHandler.RequestMetaHeader(header.GetEpochStartMetaHash())

	sp.requestEpochStartProofIfNeeded(header.GetEpochStartMetaHash(), header.GetEpoch())

	headersPool := sp.dataPool.Headers()
	for {
		time.Sleep(timeBetweenCheckForEpochStart)
		if haveTime() < 0 {
			break
		}

		if sp.epochStartTrigger.IsEpochStart() {
			return nil
		}

		epochStartMetaHdr, err := headersPool.GetHeaderByHash(header.GetEpochStartMetaHash())
		if err != nil {
			go sp.requestHandler.RequestMetaHeader(header.GetEpochStartMetaHash())
			sp.requestEpochStartProofIfNeeded(header.GetEpochStartMetaHash(), header.GetEpoch())
			continue
		}

		shouldConsiderProofsForNotarization := sp.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, epochStartMetaHdr.GetEpoch())
		if !shouldConsiderProofsForNotarization {
			_, _, err = headersPool.GetHeadersByNonceAndShardId(epochStartMetaHdr.GetNonce()+1, core.MetachainShardId)
			if err != nil {
				go sp.requestHandler.RequestMetaHeaderByNonce(epochStartMetaHdr.GetNonce() + 1)
				continue
			}
		} else {
			hasProof := sp.requestEpochStartProofIfNeeded(header.GetEpochStartMetaHash(), header.GetEpoch())
			if !hasProof {
				continue
			}
		}

		return nil
	}

	return process.ErrTimeIsOut
}

func (sp *shardProcessor) requestEpochStartProofIfNeeded(hash []byte, epoch uint32) bool {
	if !sp.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, epoch) {
		return true // no proof needed
	}

	hasProof := sp.proofsPool.HasProof(core.MetachainShardId, hash)
	if hasProof {
		return true
	}

	go sp.requestHandler.RequestEquivalentProofByHash(core.MetachainShardId, hash)

	return false
}

// RevertStateToBlock recreates the state tries to the root hashes indicated by the provided root hash and header
func (sp *shardProcessor) RevertStateToBlock(header data.HeaderHandler, rootHash []byte) error {
	rootHashHolder := holders.NewDefaultRootHashesHolder(rootHash)
	err := sp.accountsDB[state.UserAccountsState].RecreateTrie(rootHashHolder)
	if err != nil {
		log.Debug("recreate trie with error for header",
			"nonce", header.GetNonce(),
			"header root hash", header.GetRootHash(),
			"given root hash", rootHash,
			"error", err,
		)

		return err
	}

	err = sp.epochStartTrigger.RevertStateToBlock(header)
	if err != nil {
		log.Debug("revert epoch start trigger for header",
			"nonce", header.GetNonce(),
			"error", err,
		)
		return err
	}

	return nil
}

func (sp *shardProcessor) checkEpochCorrectness(
	header data.ShardHeaderHandler,
) error {
	currentBlockHeader := sp.blockChain.GetCurrentBlockHeader()
	if check.IfNil(currentBlockHeader) {
		return nil
	}

	headerEpochBehindCurrentHeader := header.GetEpoch() < currentBlockHeader.GetEpoch()
	if headerEpochBehindCurrentHeader {
		return fmt.Errorf("%w proposed header with older epoch %d than blockchain epoch %d",
			process.ErrEpochDoesNotMatch, header.GetEpoch(), currentBlockHeader.GetEpoch())
	}

	isStartOfEpochButShouldNotBe := header.GetEpoch() == currentBlockHeader.GetEpoch() && header.IsStartOfEpochBlock()
	if isStartOfEpochButShouldNotBe {
		return fmt.Errorf("%w proposed header with same epoch %d as blockchain and it is of epoch start",
			process.ErrEpochDoesNotMatch, currentBlockHeader.GetEpoch())
	}

	incorrectStartOfEpochBlock := header.GetEpoch() != currentBlockHeader.GetEpoch() &&
		sp.epochStartTrigger.MetaEpoch() == currentBlockHeader.GetEpoch()
	if incorrectStartOfEpochBlock {
		if header.IsStartOfEpochBlock() {
			sp.dataPool.Headers().RemoveHeaderByHash(header.GetEpochStartMetaHash())
			go sp.requestHandler.RequestMetaHeader(header.GetEpochStartMetaHash())
		}
		return fmt.Errorf("%w proposed header with new epoch %d with trigger still in last epoch %d",
			process.ErrEpochDoesNotMatch, header.GetEpoch(), sp.epochStartTrigger.MetaEpoch())
	}

	isHeaderOfInvalidEpoch := header.GetEpoch() > sp.epochStartTrigger.MetaEpoch()
	if isHeaderOfInvalidEpoch {
		return fmt.Errorf("%w proposed header with epoch too high %d with trigger in epoch %d",
			process.ErrEpochDoesNotMatch, header.GetEpoch(), sp.epochStartTrigger.MetaEpoch())
	}

	epochChangeConfirmed := sp.epochStartTrigger.EpochStartRound() < sp.epochStartTrigger.EpochFinalityAttestingRound()
	if sp.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, header.GetEpoch()) {
		epochChangeConfirmed = sp.epochStartTrigger.EpochStartRound() <= sp.epochStartTrigger.EpochFinalityAttestingRound()
	}

	gracePeriod, err := sp.epochChangeGracePeriodHandler.GetGracePeriodForEpoch(header.GetEpoch())
	if err != nil {
		return fmt.Errorf("%w could not get grace period for epoch %d", err, header.GetEpoch())
	}

	isOldEpochAndShouldBeNew := sp.epochStartTrigger.IsEpochStart() &&
		header.GetRound() > sp.epochStartTrigger.EpochFinalityAttestingRound()+uint64(gracePeriod) &&
		header.GetEpoch() < sp.epochStartTrigger.MetaEpoch() &&
		epochChangeConfirmed
	if isOldEpochAndShouldBeNew {
		return fmt.Errorf("%w proposed header with epoch %d should be in epoch %d",
			process.ErrEpochDoesNotMatch, header.GetEpoch(), sp.epochStartTrigger.MetaEpoch())
	}

	isEpochStartMetaHashIncorrect := header.IsStartOfEpochBlock() &&
		!bytes.Equal(header.GetEpochStartMetaHash(), sp.epochStartTrigger.EpochStartMetaHdrHash()) &&
		header.GetEpoch() == sp.epochStartTrigger.MetaEpoch()
	if isEpochStartMetaHashIncorrect {
		go sp.requestHandler.RequestMetaHeader(header.GetEpochStartMetaHash())
		log.Warn("epoch start meta hash mismatch", "proposed", header.GetEpochStartMetaHash(), "calculated", sp.epochStartTrigger.EpochStartMetaHdrHash())
		return fmt.Errorf("%w proposed header with epoch %d has invalid epochStartMetaHash",
			process.ErrEpochDoesNotMatch, header.GetEpoch())
	}

	isNotEpochStartButShouldBe := header.GetEpoch() != currentBlockHeader.GetEpoch() &&
		!header.IsStartOfEpochBlock()
	if isNotEpochStartButShouldBe {
		return fmt.Errorf("%w proposed header with new epoch %d is not of type epoch start",
			process.ErrEpochDoesNotMatch, header.GetEpoch())
	}

	isOldEpochStart := header.IsStartOfEpochBlock() && header.GetEpoch() < sp.epochStartTrigger.MetaEpoch()
	if isOldEpochStart {
		metaBlock, err := process.GetMetaHeader(header.GetEpochStartMetaHash(), sp.dataPool.Headers(), sp.marshalizer, sp.store)
		if err != nil {
			go sp.requestHandler.RequestStartOfEpochMetaBlock(header.GetEpoch())
			return fmt.Errorf("%w could not find epoch start metablock for epoch %d",
				err, header.GetEpoch())
		}

		isMetaBlockCorrect := metaBlock.IsStartOfEpochBlock() && metaBlock.GetEpoch() == header.GetEpoch()
		if !isMetaBlockCorrect {
			return fmt.Errorf("%w proposed header with epoch %d does not include correct start of epoch metaBlock %s",
				process.ErrEpochDoesNotMatch, header.GetEpoch(), header.GetEpochStartMetaHash())
		}
	}

	return nil
}

// SetNumProcessedObj will set the num of processed transactions
func (sp *shardProcessor) SetNumProcessedObj(numObj uint64) {
	sp.txCounter.totalTxs = numObj
}

// checkMetaHeadersValidityAndFinality - checks if listed metaheaders are valid as construction
func (sp *shardProcessor) checkMetaHeadersValidityAndFinality() error {
	lastCrossNotarizedHeader, _, err := sp.blockTracker.GetLastCrossNotarizedHeader(core.MetachainShardId)
	if err != nil {
		return err
	}

	log.Trace("checkMetaHeadersValidityAndFinality", "lastCrossNotarizedHeader nonce", lastCrossNotarizedHeader.GetNonce())
	usedMetaHdrs, err := sp.sortHeadersForCurrentBlockByNonce(true)
	if err != nil {
		return err
	}
	if len(usedMetaHdrs[core.MetachainShardId]) == 0 {
		return nil
	}

	for _, metaHdr := range usedMetaHdrs[core.MetachainShardId] {
		log.Trace("checkMetaHeadersValidityAndFinality", "metaHeader nonce", metaHdr.GetNonce())
		err = sp.headerValidator.IsHeaderConstructionValid(metaHdr, lastCrossNotarizedHeader)
		if err != nil {
			return fmt.Errorf("%w : checkMetaHeadersValidityAndFinality -> isHdrConstructionValid", err)
		}

		lastCrossNotarizedHeader = metaHdr
	}

	err = sp.checkMetaHdrFinality(lastCrossNotarizedHeader)
	if err != nil {
		return err
	}

	return nil
}

func (sp *shardProcessor) checkHeaderHasProof(header data.HeaderHandler) error {
	hash, errHash := sp.getHeaderHash(header)
	if errHash != nil {
		return errHash
	}

	if !sp.proofsPool.HasProof(header.GetShardID(), hash) {
		return fmt.Errorf("%w, missing proof for header %s", process.ErrHeaderNotFinal, hex.EncodeToString(hash))
	}

	return nil
}

// check if shard headers are final by checking if newer headers were constructed upon them
func (sp *shardProcessor) checkMetaHdrFinality(header data.HeaderHandler) error {
	if check.IfNil(header) {
		return process.ErrNilBlockHeader
	}

	if sp.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, header.GetEpoch()) {
		return sp.checkHeaderHasProof(header)
	}

	finalityAttestingMetaHdrs, err := sp.sortHeadersForCurrentBlockByNonce(false)
	if err != nil {
		return err
	}

	lastVerifiedHdr := header
	// verify if there are "K" block after current to make this one final
	nextBlocksVerified := uint32(0)
	for _, metaHdr := range finalityAttestingMetaHdrs[core.MetachainShardId] {
		if nextBlocksVerified >= sp.metaBlockFinality {
			break
		}

		// found a header with the next nonce
		if metaHdr.GetNonce() == lastVerifiedHdr.GetNonce()+1 {
			err := sp.headerValidator.IsHeaderConstructionValid(metaHdr, lastVerifiedHdr)
			if err != nil {
				log.Debug("checkMetaHdrFinality -> isHdrConstructionValid",
					"error", err.Error())
				continue
			}

			if sp.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, metaHdr.GetEpoch()) {
				return sp.checkHeaderHasProof(metaHdr)
			}

			lastVerifiedHdr = metaHdr
			nextBlocksVerified += 1
		}
	}

	if nextBlocksVerified < sp.metaBlockFinality {
		go sp.requestHandler.RequestMetaHeaderByNonce(lastVerifiedHdr.GetNonce())
		go sp.requestHandler.RequestMetaHeaderByNonce(lastVerifiedHdr.GetNonce() + 1)
		return process.ErrHeaderNotFinal
	}

	return nil
}

func (sp *shardProcessor) checkAndRequestIfMetaHeadersMissing() {
	orderedMetaBlocks, _ := sp.blockTracker.GetTrackedHeaders(core.MetachainShardId)

	err := sp.requestHeadersIfMissing(orderedMetaBlocks, core.MetachainShardId)
	if err != nil {
		log.Debug("checkAndRequestIfMetaHeadersMissing", "error", err.Error())
	}
}

func (sp *shardProcessor) indexBlockIfNeeded(
	body data.BodyHandler,
	headerHash []byte,
	header data.HeaderHandler,
	lastBlockHeader data.HeaderHandler,
) {
	if !sp.outportHandler.HasDrivers() {
		return
	}

	log.Debug("preparing to index block", "hash", headerHash, "nonce", header.GetNonce(), "round", header.GetRound())
	argSaveBlock, err := sp.outportDataProvider.PrepareOutportSaveBlockData(processOutport.ArgPrepareOutportSaveBlockData{
		HeaderHash:             headerHash,
		Header:                 header,
		Body:                   body,
		PreviousHeader:         lastBlockHeader,
		HighestFinalBlockNonce: sp.forkDetector.GetHighestFinalBlockNonce(),
		HighestFinalBlockHash:  sp.forkDetector.GetHighestFinalBlockHash(),
	})
	if err != nil {
		log.Error("shardProcessor.indexBlockIfNeeded cannot prepare argSaveBlock", "error", err.Error(),
			"hash", headerHash, "nonce", header.GetNonce(), "round", header.GetRound())
		return
	}
	err = sp.outportHandler.SaveBlock(argSaveBlock)
	if err != nil {
		log.Error("shardProcessor.outportHandler.SaveBlock cannot save block", "error", err,
			"hash", headerHash, "nonce", header.GetNonce(), "round", header.GetRound())
		return
	}

	log.Debug("indexed block", "hash", headerHash, "nonce", header.GetNonce(), "round", header.GetRound())

	shardID := sp.shardCoordinator.SelfId()
	indexRoundInfo(
		sp.outportHandler,
		sp.nodesCoordinator,
		shardID, header,
		lastBlockHeader,
		argSaveBlock.SignersIndexes,
		sp.enableEpochsHandler,
	)
}

// RestoreBlockIntoPools restores the TxBlock and MetaBlock into associated pools
func (sp *shardProcessor) RestoreBlockIntoPools(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error {
	if check.IfNil(headerHandler) {
		return process.ErrNilBlockHeader
	}

	header, ok := headerHandler.(data.ShardHeaderHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	miniBlockHashes := header.MapMiniBlockHashesToShards()
	err := sp.restoreMetaBlockIntoPool(headerHandler, miniBlockHashes, header.GetMetaBlockHashes())
	if err != nil {
		return err
	}

	sp.restoreBlockBody(headerHandler, bodyHandler)

	sp.blockTracker.RemoveLastNotarizedHeaders()

	return nil
}

func (sp *shardProcessor) restoreMetaBlockIntoPool(
	headerHandler data.HeaderHandler,
	mapMiniBlockHashes map[string]uint32,
	metaBlockHashes [][]byte,
) error {
	headersPool := sp.dataPool.Headers()

	mapMetaHashMiniBlockHashes := make(map[string][][]byte)
	mapMetaHashMetaBlock := make(map[string]*block.MetaBlock)

	for _, metaBlockHash := range metaBlockHashes {
		metaBlock, errNotCritical := process.GetMetaHeaderFromStorage(metaBlockHash, sp.marshalizer, sp.store)
		if errNotCritical != nil {
			log.Debug("meta block is not fully processed yet and not committed in MetaBlockUnit",
				"hash", metaBlockHash)
			continue
		}

		mapMetaHashMetaBlock[string(metaBlockHash)] = metaBlock
		processedMiniBlocks := metaBlock.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
		for mbHash := range processedMiniBlocks {
			mapMetaHashMiniBlockHashes[string(metaBlockHash)] = append(mapMetaHashMiniBlockHashes[string(metaBlockHash)], []byte(mbHash))
		}

		headersPool.AddHeader(metaBlockHash, metaBlock)

		metablockStorer, err := sp.store.GetStorer(dataRetriever.MetaBlockUnit)
		if err != nil {
			log.Debug("unable to get storage unit",
				"unit", dataRetriever.MetaBlockUnit.String())
			return err
		}

		err = metablockStorer.Remove(metaBlockHash)
		if err != nil {
			log.Debug("unable to remove hash from MetaBlockUnit",
				"hash", metaBlockHash)
			return err
		}

		nonceToByteSlice := sp.uint64Converter.ToByteSlice(metaBlock.GetNonce())

		metaHdrNonceHashStorer, err := sp.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
		if err != nil {
			log.Debug("unable to get storage unit",
				"unit", dataRetriever.MetaHdrNonceHashDataUnit.String())
			return err
		}

		errNotCritical = metaHdrNonceHashStorer.Remove(nonceToByteSlice)
		if errNotCritical != nil {
			log.Debug("error not critical",
				"error", errNotCritical.Error())
		}

		log.Trace("meta block has been restored successfully",
			"round", metaBlock.Round,
			"nonce", metaBlock.Nonce,
			"hash", metaBlockHash)
	}

	for metaBlockHash, miniBlockHashes := range mapMetaHashMiniBlockHashes {
		sp.setProcessedMiniBlocksInfo(miniBlockHashes, metaBlockHash, mapMetaHashMetaBlock[metaBlockHash])
	}

	sp.rollBackProcessedMiniBlocksInfo(headerHandler, mapMiniBlockHashes)

	return nil
}

func (sp *shardProcessor) setProcessedMiniBlocksInfo(miniBlockHashes [][]byte, metaBlockHash string, metaBlock *block.MetaBlock) {
	for _, miniBlockHash := range miniBlockHashes {
		indexOfLastTxProcessed := getIndexOfLastTxProcessedInMiniBlock(miniBlockHash, metaBlock)
		sp.processedMiniBlocksTracker.SetProcessedMiniBlockInfo([]byte(metaBlockHash), miniBlockHash, &processedMb.ProcessedMiniBlockInfo{
			FullyProcessed:         true,
			IndexOfLastTxProcessed: indexOfLastTxProcessed,
		})
	}
}

func getIndexOfLastTxProcessedInMiniBlock(miniBlockHash []byte, metaBlock *block.MetaBlock) int32 {
	for _, mbh := range metaBlock.MiniBlockHeaders {
		if bytes.Equal(mbh.Hash, miniBlockHash) {
			return int32(mbh.TxCount) - 1
		}
	}

	for _, shardData := range metaBlock.ShardInfo {
		for _, mbh := range shardData.ShardMiniBlockHeaders {
			if bytes.Equal(mbh.Hash, miniBlockHash) {
				return int32(mbh.TxCount) - 1
			}
		}
	}

	log.Warn("shardProcessor.getIndexOfLastTxProcessedInMiniBlock",
		"miniBlock hash", miniBlockHash,
		"metaBlock round", metaBlock.Round,
		"metaBlock nonce", metaBlock.Nonce,
		"error", process.ErrMissingMiniBlock)

	return common.MaxIndexOfTxInMiniBlock
}

func (sp *shardProcessor) rollBackProcessedMiniBlocksInfo(headerHandler data.HeaderHandler, mapMiniBlockHashes map[string]uint32) {
	for miniBlockHash := range mapMiniBlockHashes {
		miniBlockHeader := process.GetMiniBlockHeaderWithHash(headerHandler, []byte(miniBlockHash))
		if miniBlockHeader == nil {
			log.Warn("shardProcessor.rollBackProcessedMiniBlocksInfo: GetMiniBlockHeaderWithHash",
				"mb hash", miniBlockHash,
				"error", process.ErrMissingMiniBlockHeader)
			continue
		}

		if miniBlockHeader.GetSenderShardID() == sp.shardCoordinator.SelfId() {
			continue
		}

		sp.rollBackProcessedMiniBlockInfo(miniBlockHeader, []byte(miniBlockHash))
	}
}

func (sp *shardProcessor) rollBackProcessedMiniBlockInfo(miniBlockHeader data.MiniBlockHeaderHandler, miniBlockHash []byte) {
	indexOfFirstTxProcessed := miniBlockHeader.GetIndexOfFirstTxProcessed()
	if indexOfFirstTxProcessed == 0 {
		sp.processedMiniBlocksTracker.RemoveMiniBlockHash(miniBlockHash)
		return
	}

	_, metaBlockHash := sp.processedMiniBlocksTracker.GetProcessedMiniBlockInfo(miniBlockHash)
	if metaBlockHash == nil {
		log.Warn("shardProcessor.rollBackProcessedMiniBlockInfo: mini block was not found in ProcessedMiniBlockTracker component",
			"sender shard", miniBlockHeader.GetSenderShardID(),
			"receiver shard", miniBlockHeader.GetReceiverShardID(),
			"tx count", miniBlockHeader.GetTxCount(),
			"mb hash", miniBlockHash)
		return
	}

	sp.processedMiniBlocksTracker.SetProcessedMiniBlockInfo(metaBlockHash, miniBlockHash, &processedMb.ProcessedMiniBlockInfo{
		FullyProcessed:         false,
		IndexOfLastTxProcessed: indexOfFirstTxProcessed - 1,
	})
}

// CreateBlockProposal - creates a block proposal without executing any of the transactions
func (sp *shardProcessor) CreateBlockProposal(
	initialHdr data.HeaderHandler,
	haveTime func() bool,
) (data.HeaderHandler, data.BodyHandler, error) {
	if check.IfNil(initialHdr) {
		return nil, nil, process.ErrNilBlockHeader
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

	// todo: check empty mini blocks vs nil. Same for block.Body.MiniBlocks
	err = shardHdr.SetMiniBlockHeaderHandlers(sp.miniBlocksSelectionSession.GetMiniBlockHeaderHandlers())
	if err != nil {
		return nil, nil, err
	}

	// todo: check also the rest of the header fields (e.g what was previously done on applyBodyToHeader)

	return shardHdr, &block.Body{MiniBlocks: sp.miniBlocksSelectionSession.GetMiniBlocks()}, nil
}

// CreateBlock creates the final block and header for the current round
func (sp *shardProcessor) CreateBlock(
	initialHdr data.HeaderHandler,
	haveTime func() bool,
) (data.HeaderHandler, data.BodyHandler, error) {
	if check.IfNil(initialHdr) {
		return nil, nil, process.ErrNilBlockHeader
	}
	shardHdr, ok := initialHdr.(data.ShardHeaderHandler)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	sp.processStatusHandler.SetBusy("shardProcessor.CreateBlock")
	defer sp.processStatusHandler.SetIdle()

	err := sp.createBlockStarted()
	if err != nil {
		return nil, nil, err
	}

	// placeholder for shardProcessor.CreateBlock script 1

	// placeholder for shardProcessor.CreateBlock script 2

	err = sp.updateEpochIfNeeded(shardHdr)
	if err != nil {
		return nil, nil, err
	}

	sp.epochNotifier.CheckEpoch(shardHdr)

	err = sp.blockChainHook.SetCurrentHeader(shardHdr)
	if err != nil {
		return nil, nil, err
	}

	body, processedMiniBlocksDestMeInfo, err := sp.createBlockBody(shardHdr, haveTime)
	if err != nil {
		return nil, nil, err
	}

	finalBody, err := sp.applyBodyToHeader(shardHdr, body, processedMiniBlocksDestMeInfo)
	if err != nil {
		return nil, nil, err
	}

	for _, miniBlock := range finalBody.MiniBlocks {
		log.Trace("CreateBlock: miniblock",
			"sender shard", miniBlock.SenderShardID,
			"receiver shard", miniBlock.ReceiverShardID,
			"type", miniBlock.Type,
			"num txs", len(miniBlock.TxHashes))
	}

	return shardHdr, finalBody, nil
}

func (sp *shardProcessor) updateEpochIfNeeded(shardHeader data.ShardHeaderHandler) error {
	if !sp.epochStartTrigger.IsEpochStart() {
		return nil
	}

	log.Debug("CreateBlock", "IsEpochStart", sp.epochStartTrigger.IsEpochStart(),
		"epoch start meta header hash", sp.epochStartTrigger.EpochStartMetaHdrHash())
	err := shardHeader.SetEpochStartMetaHash(sp.epochStartTrigger.EpochStartMetaHdrHash())
	if err != nil {
		return err
	}

	epoch := sp.epochStartTrigger.MetaEpoch()
	if shardHeader.GetEpoch() != epoch {
		log.Debug("shardProcessor.CreateBlock: epoch from header is not the same as epoch from epoch start trigger, overwriting",
			"epoch from header", shardHeader.GetEpoch(), "epoch from epoch start trigger", epoch)
		err = shardHeader.SetEpoch(epoch)
		if err != nil {
			return err
		}
	}

	return nil
}

// createBlockBody creates a list of miniblocks by filling them with transactions out of the transactions pools
// as long as the transactions limit for the block has not been reached and there is still time to add transactions
func (sp *shardProcessor) createBlockBody(shardHdr data.HeaderHandler, haveTime func() bool) (*block.Body, map[string]*processedMb.ProcessedMiniBlockInfo, error) {
	sp.blockSizeThrottler.ComputeCurrentMaxSize()

	log.Debug("started creating block body",
		"epoch", shardHdr.GetEpoch(),
		"round", shardHdr.GetRound(),
		"nonce", shardHdr.GetNonce(),
	)

	randomness := helpers.ComputeRandomnessForTxSorting(shardHdr, sp.enableEpochsHandler)
	miniBlocks, processedMiniBlocksDestMeInfo, err := sp.createMiniBlocks(haveTime, randomness)
	if err != nil {
		return nil, nil, err
	}

	sp.requestHandler.SetEpoch(shardHdr.GetEpoch())

	return miniBlocks, processedMiniBlocksDestMeInfo, nil
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

// CommitBlock commits the block in the blockchain if everything was checked successfully
func (sp *shardProcessor) CommitBlock(
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {
	var err error
	sp.processStatusHandler.SetBusy("shardProcessor.CommitBlock")
	defer func() {
		if err != nil {
			sp.RevertCurrentBlock()
		}
		sp.processStatusHandler.SetIdle()
	}()

	err = checkForNils(headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	sp.store.SetEpochForPutOperation(headerHandler.GetEpoch())

	log.Debug("started committing block",
		"epoch", headerHandler.GetEpoch(),
		"shard", headerHandler.GetShardID(),
		"round", headerHandler.GetRound(),
		"nonce", headerHandler.GetNonce(),
	)

	err = sp.checkBlockValidity(headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	header, ok := headerHandler.(data.ShardHeaderHandler)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	err = sp.checkEpochCorrectnessCrossChain()
	if err != nil {
		return err
	}

	if header.IsStartOfEpochBlock() {
		sp.epochStartTrigger.SetProcessed(header, bodyHandler)
	}

	marshalizedHeader, err := sp.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	headerHash := sp.hasher.Compute(string(marshalizedHeader))

	sp.saveShardHeader(header, headerHash, marshalizedHeader)

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	sp.saveBody(body, header, headerHash)

	processedMetaHdrs, err := sp.getOrderedProcessedMetaBlocksFromHeader(header)
	if err != nil {
		return err
	}

	err = sp.addProcessedCrossMiniBlocksFromHeader(header)
	if err != nil {
		return err
	}

	selfNotarizedHeaders, selfNotarizedHeadersHashes, err := sp.getHighestHdrForOwnShardFromMetachain(processedMetaHdrs)
	if err != nil {
		return err
	}

	err = sp.saveLastNotarizedHeader(core.MetachainShardId, processedMetaHdrs)
	if err != nil {
		return err
	}

	err = sp.commitAll(headerHandler)
	if err != nil {
		return err
	}

	log.Info("shard block has been committed successfully",
		"epoch", header.GetEpoch(),
		"shard", header.GetShardID(),
		"round", header.GetRound(),
		"nonce", header.GetNonce(),
		"hash", headerHash,
	)
	sp.setNonceOfFirstCommittedBlock(headerHandler.GetNonce())

	sp.updateLastCommittedInDebugger(headerHandler.GetRound())

	err = sp.computeOwnShardStuckIfNeeded(headerHandler)
	if err != nil {
		return err
	}

	errNotCritical := sp.checkSentSignaturesAtCommitTime(headerHandler)
	if errNotCritical != nil {
		log.Debug("checkSentSignaturesBeforeCommitting", "error", errNotCritical.Error())
	}

	errNotCritical = sp.updateCrossShardInfo(processedMetaHdrs)
	if errNotCritical != nil {
		log.Debug("updateCrossShardInfo", "error", errNotCritical.Error())
	}

	errNotCritical = sp.forkDetector.AddHeader(header, headerHash, process.BHProcessed, selfNotarizedHeaders, selfNotarizedHeadersHashes)
	if errNotCritical != nil {
		log.Debug("forkDetector.AddHeader", "error", errNotCritical.Error())
	}

	currentHeader, currentHeaderHash := getLastSelfNotarizedHeaderByItself(sp.blockChain)
	sp.blockTracker.AddSelfNotarizedHeader(sp.shardCoordinator.SelfId(), currentHeader, currentHeaderHash)

	lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash := sp.getLastSelfNotarizedHeaderByMetachain()
	sp.blockTracker.AddSelfNotarizedHeader(core.MetachainShardId, lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash)

	sp.notifyFinalMetaHdrs(processedMetaHdrs)

	if sp.lastRestartNonce == 0 {
		sp.lastRestartNonce = header.GetNonce()
	}

	finalHeaderHash := headerHash
	if !common.IsFlagEnabledAfterEpochsStartBlock(header, sp.enableEpochsHandler, common.AndromedaFlag) {
		finalHeaderHash = currentHeaderHash
	}

	sp.updateState(selfNotarizedHeaders, header, finalHeaderHash)

	highestFinalBlockNonce := sp.forkDetector.GetHighestFinalBlockNonce()
	log.Debug("highest final shard block",
		"shard", sp.shardCoordinator.SelfId(),
		"nonce", highestFinalBlockNonce,
	)

	lastBlockHeader := sp.blockChain.GetCurrentBlockHeader()

	committedRootHash, err := sp.accountsDB[state.UserAccountsState].RootHash()
	if err != nil {
		return err
	}

	err = sp.blockChain.SetCurrentBlockHeaderAndRootHash(header, committedRootHash)
	if err != nil {
		return err
	}

	sp.blockChain.SetCurrentBlockHeaderHash(headerHash)
	sp.indexBlockIfNeeded(bodyHandler, headerHash, headerHandler, lastBlockHeader)
	sp.recordBlockInHistory(headerHash, headerHandler, bodyHandler)

	lastCrossNotarizedHeader, _, err := sp.blockTracker.GetLastCrossNotarizedHeader(core.MetachainShardId)
	if err != nil {
		return err
	}

	saveMetricsForCommittedShardBlock(
		sp.nodesCoordinator,
		sp.appStatusHandler,
		logger.DisplayByteSlice(headerHash),
		highestFinalBlockNonce,
		lastCrossNotarizedHeader,
		header,
		sp.managedPeersHolder,
	)

	headerInfo := bootstrapStorage.BootstrapHeaderInfo{
		ShardId: header.GetShardID(),
		Epoch:   header.GetEpoch(),
		Nonce:   header.GetNonce(),
		Hash:    headerHash,
	}

	nodesCoordinatorKey := sp.nodesCoordinator.GetSavedStateKey()
	epochStartKey := sp.epochStartTrigger.GetSavedStateKey()

	args := bootStorerDataArgs{
		headerInfo:                 headerInfo,
		round:                      header.GetRound(),
		lastSelfNotarizedHeaders:   sp.getBootstrapHeadersInfo(selfNotarizedHeaders, selfNotarizedHeadersHashes),
		highestFinalBlockNonce:     sp.forkDetector.GetHighestFinalBlockNonce(),
		processedMiniBlocks:        sp.processedMiniBlocksTracker.ConvertProcessedMiniBlocksMapToSlice(),
		nodesCoordinatorConfigKey:  nodesCoordinatorKey,
		epochStartTriggerConfigKey: epochStartKey,
	}

	sp.prepareDataForBootStorer(args)

	// write data to log
	go func() {
		sp.txCounter.headerExecuted(header)
		sp.txCounter.displayLogInfo(
			header,
			body,
			headerHash,
			sp.shardCoordinator.NumberOfShards(),
			sp.shardCoordinator.SelfId(),
			sp.dataPool,
			sp.blockTracker,
		)
	}()

	sp.blockSizeThrottler.Succeed(header.GetRound())

	sp.displayPoolsInfo()

	errNotCritical = sp.removeTxsFromPools(header, body)
	if errNotCritical != nil {
		log.Debug("removeTxsFromPools", "error", errNotCritical.Error())
	}

	sp.cleanupPools(headerHandler)

	sp.blockProcessingCutoffHandler.HandlePauseCutoff(header)

	return nil
}

func (sp *shardProcessor) notifyFinalMetaHdrs(processedMetaHeaders []data.HeaderHandler) {
	metaHeaders := make([]data.HeaderHandler, 0)
	metaHeadersHashes := make([][]byte, 0)

	for _, metaHeader := range processedMetaHeaders {
		metaHeaderHash, err := core.CalculateHash(sp.marshalizer, sp.hasher, metaHeader)
		if err != nil {
			log.Debug("shardProcessor.notifyFinalMetaHdrs", "error", err.Error())
			continue
		}

		metaHeaders = append(metaHeaders, metaHeader)
		metaHeadersHashes = append(metaHeadersHashes, metaHeaderHash)
	}

	if len(metaHeaders) > 0 {
		go sp.historyRepo.OnNotarizedBlocks(core.MetachainShardId, metaHeaders, metaHeadersHashes)
	}
}

func (sp *shardProcessor) displayPoolsInfo() {
	headersPool := sp.dataPool.Headers()
	miniBlocksPool := sp.dataPool.MiniBlocks()

	log.Trace("pools info",
		"shard", sp.shardCoordinator.SelfId(),
		"num headers", headersPool.GetNumHeaders(sp.shardCoordinator.SelfId()))

	log.Trace("pools info",
		"shard", core.MetachainShardId,
		"num headers", headersPool.GetNumHeaders(core.MetachainShardId))

	// numShardsToKeepHeaders represents the total number of shards for which shard node would keep tracking headers
	// (in this case this number is equal with: self shard + metachain)
	numShardsToKeepHeaders := 2
	capacity := headersPool.MaxSize() * numShardsToKeepHeaders
	log.Debug("pools info",
		"total headers", headersPool.Len(),
		"headers pool capacity", capacity,
		"total miniblocks", miniBlocksPool.Len(),
		"miniblocks pool capacity", miniBlocksPool.MaxSize(),
	)

	sp.displayMiniBlocksPool()
}

func (sp *shardProcessor) updateState(headers []data.HeaderHandler, currentHeader data.ShardHeaderHandler, currentHeaderHash []byte) {
	sp.snapShotEpochStartFromMeta(currentHeader)

	for _, header := range headers {
		if sp.forkDetector.GetHighestFinalBlockNonce() < header.GetNonce() {
			break
		}

		prevHeaderHash := header.GetPrevHash()
		prevHeader, errNotCritical := process.GetShardHeader(
			prevHeaderHash,
			sp.dataPool.Headers(),
			sp.marshalizer,
			sp.store,
		)
		if errNotCritical != nil {
			log.Debug("could not get shard header from storage")
			return
		}
		if header.IsStartOfEpochBlock() {
			sp.nodesCoordinator.ShuffleOutForEpoch(header.GetEpoch())
		}

		headerHash, err := core.CalculateHash(sp.marshalizer, sp.hasher, header)
		if err != nil {
			log.Debug("updateState.CalculateHash", "error", err.Error())
			return
		}

		headerRootHashForPruning := header.GetRootHash()
		prevHeaderRootHashForPruning := prevHeader.GetRootHash()

		scheduledHeaderRootHash, _ := sp.scheduledTxsExecutionHandler.GetScheduledRootHashForHeader(headerHash)
		headerAdditionalData := header.GetAdditionalData()
		if headerAdditionalData != nil && headerAdditionalData.GetScheduledRootHash() != nil {
			headerRootHashForPruning = headerAdditionalData.GetScheduledRootHash()
		}

		prevHeaderAdditionalData := prevHeader.GetAdditionalData()
		scheduledPrevHeaderRootHash, _ := sp.scheduledTxsExecutionHandler.GetScheduledRootHashForHeader(header.GetPrevHash())
		if prevHeaderAdditionalData != nil && prevHeaderAdditionalData.GetScheduledRootHash() != nil {
			prevHeaderRootHashForPruning = prevHeaderAdditionalData.GetScheduledRootHash()
		}

		log.Trace("updateState: prevHeader",
			"shard", prevHeader.GetShardID(),
			"epoch", prevHeader.GetEpoch(),
			"round", prevHeader.GetRound(),
			"nonce", prevHeader.GetNonce(),
			"root hash", prevHeader.GetRootHash(),
			"scheduled root hash for pruning", prevHeaderRootHashForPruning,
			"scheduled root hash after processing", scheduledPrevHeaderRootHash,
		)

		log.Trace("updateState: currHeader",
			"shard", header.GetShardID(),
			"epoch", header.GetEpoch(),
			"round", header.GetRound(),
			"nonce", header.GetNonce(),
			"root hash", header.GetRootHash(),
			"scheduled root hash for pruning", headerRootHashForPruning,
			"scheduled root hash after processing", scheduledHeaderRootHash,
		)

		sp.updateStateStorage(
			header,
			headerRootHashForPruning,
			prevHeaderRootHashForPruning,
			sp.accountsDB[state.UserAccountsState],
		)

		if sp.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, header.GetEpoch()) {
			continue
		}

		sp.setFinalizedHeaderHashInIndexer(header.GetPrevHash())

		sp.setFinalBlockInfo(header, headerHash, scheduledHeaderRootHash)
	}

	if !sp.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, currentHeader.GetEpoch()) {
		return
	}

	sp.setFinalizedHeaderHashInIndexer(currentHeaderHash)

	scheduledHeaderRootHash, _ := sp.scheduledTxsExecutionHandler.GetScheduledRootHashForHeader(currentHeaderHash)
	sp.setFinalBlockInfo(currentHeader, currentHeaderHash, scheduledHeaderRootHash)
}

func (sp *shardProcessor) setFinalBlockInfo(
	header data.HeaderHandler,
	headerHash []byte,
	scheduledHeaderRootHash []byte,
) {
	finalRootHash := scheduledHeaderRootHash
	if len(finalRootHash) == 0 {
		finalRootHash = header.GetRootHash()
	}

	sp.blockChain.SetFinalBlockInfo(header.GetNonce(), headerHash, finalRootHash)
}

func (sp *shardProcessor) snapShotEpochStartFromMeta(header data.ShardHeaderHandler) {
	accounts := sp.accountsDB[state.UserAccountsState]

	for _, metaHash := range header.GetMetaBlockHashes() {
		metaHdrInfo, ok := sp.hdrsForCurrBlock.GetHeaderInfo(string(metaHash))
		if !ok {
			continue
		}
		metaHdr, ok := metaHdrInfo.GetHeader().(*block.MetaBlock)
		if !ok {
			continue
		}
		if !metaHdr.IsStartOfEpochBlock() {
			continue
		}

		for _, epochStartShData := range metaHdr.EpochStart.LastFinalizedHeaders {
			if epochStartShData.ShardID != header.GetShardID() {
				continue
			}

			rootHash := epochStartShData.RootHash
			schRootHash := epochStartShData.GetScheduledRootHash()
			if schRootHash != nil {
				log.Debug("using scheduled root hash for snapshotting", "schRootHash", schRootHash)
				rootHash = schRootHash
			}
			epoch := sp.epochStartTrigger.MetaEpoch()
			log.Debug("shard trie snapshot from epoch start shard data", "rootHash", rootHash, "epoch", epoch)
			accounts.SnapshotState(rootHash, epoch)
			sp.markSnapshotDoneInPeerAccounts()
			saveEpochStartEconomicsMetrics(sp.appStatusHandler, metaHdr)
			go func() {
				err := sp.commitTrieEpochRootHashIfNeeded(metaHdr, rootHash)
				if err != nil {
					log.Warn("couldn't commit trie checkpoint", "epoch", header.GetEpoch(), "error", err)
				}
			}()
		}
	}
}

func (sp *shardProcessor) markSnapshotDoneInPeerAccounts() {
	peerAccounts := sp.accountsDB[state.PeerAccountsState]
	if check.IfNil(peerAccounts) {
		log.Warn("programming error: peerAccounts is nil while trying to take a snapshot on a shard node: this can cause OOM exceptions")
		return
	}

	peerAccountsHandler, ok := peerAccounts.(peerAccountsDBHandler)
	if !ok {
		log.Warn("programming error: peerAccounts is not of type peerAccountsDBHandler: this can cause OOM exceptions")
		return
	}

	peerAccountsHandler.MarkSnapshotDone()
	log.Debug("shardProcessor.markSnapshotDoneInPeerAccounts completed")
}

func (sp *shardProcessor) checkEpochCorrectnessCrossChain() error {
	currentHeader := sp.blockChain.GetCurrentBlockHeader()
	if check.IfNil(currentHeader) {
		return nil
	}
	if sp.epochStartTrigger.EpochStartRound() >= sp.epochStartTrigger.EpochFinalityAttestingRound() {
		return nil
	}

	lastSelfNotarizedHeader, _ := sp.getLastSelfNotarizedHeaderByMetachain()
	lastFinalizedRound := uint64(0)
	if !check.IfNil(lastSelfNotarizedHeader) {
		lastFinalizedRound = lastSelfNotarizedHeader.GetRound()
	}

	shouldRevertChain := false
	nonce := currentHeader.GetNonce()
	gracePeriodForEpoch, err := sp.epochChangeGracePeriodHandler.GetGracePeriodForEpoch(sp.epochStartTrigger.MetaEpoch())
	if err != nil {
		log.Debug("checkEpochCorrectnessCrossChain.GetGracePeriodForEpoch", "error", err.Error())
		return err
	}
	shouldEnterNewEpochRound := sp.epochStartTrigger.EpochFinalityAttestingRound() + uint64(gracePeriodForEpoch)

	for round := currentHeader.GetRound(); round > shouldEnterNewEpochRound && currentHeader.GetEpoch() < sp.epochStartTrigger.MetaEpoch(); round = currentHeader.GetRound() {
		if round <= lastFinalizedRound {
			break
		}

		shouldRevertChain = true
		prevHeader, err := process.GetShardHeader(
			currentHeader.GetPrevHash(),
			sp.dataPool.Headers(),
			sp.marshalizer,
			sp.store,
		)
		if err != nil {
			return err
		}

		nonce = currentHeader.GetNonce()
		currentHeader = prevHeader
	}

	if shouldRevertChain {
		log.Debug("blockchain is wrongly constructed",
			"reverted to nonce", nonce)

		sp.forkDetector.SetRollBackNonce(nonce)
		return process.ErrEpochDoesNotMatch
	}

	return nil
}

func (sp *shardProcessor) getLastSelfNotarizedHeaderByMetachain() (data.HeaderHandler, []byte) {
	if sp.forkDetector.GetHighestFinalBlockNonce() == sp.genesisNonce {
		return sp.blockChain.GetGenesisHeader(), sp.blockChain.GetGenesisHeaderHash()
	}

	hash := sp.forkDetector.GetHighestFinalBlockHash()
	header, err := process.GetShardHeader(hash, sp.dataPool.Headers(), sp.marshalizer, sp.store)
	if err != nil {
		log.Warn("getLastSelfNotarizedHeaderByMetachain.GetShardHeader", "error", err.Error(), "hash", hash, "nonce", sp.forkDetector.GetHighestFinalBlockNonce())
		return nil, nil
	}

	return header, hash
}

func (sp *shardProcessor) saveLastNotarizedHeader(shardId uint32, processedHdrs []data.HeaderHandler) error {
	lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash, err := sp.blockTracker.GetLastCrossNotarizedHeader(shardId)
	if err != nil {
		return err
	}

	lenProcessedHdrs := len(processedHdrs)
	if lenProcessedHdrs > 0 {
		if lastCrossNotarizedHeader.GetNonce() < processedHdrs[lenProcessedHdrs-1].GetNonce() {
			lastCrossNotarizedHeader = processedHdrs[lenProcessedHdrs-1]
			lastCrossNotarizedHeaderHash, err = core.CalculateHash(sp.marshalizer, sp.hasher, lastCrossNotarizedHeader)
			if err != nil {
				return err
			}
		}
	}

	sp.blockTracker.AddCrossNotarizedHeader(shardId, lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash)
	DisplayLastNotarized(sp.marshalizer, sp.hasher, lastCrossNotarizedHeader, shardId)

	return nil
}

// CreateNewHeaderProposal creates a new header proposal
func (sp *shardProcessor) CreateNewHeaderProposal(round uint64, nonce uint64) (data.HeaderHandler, error) {
	epoch := sp.epochStartTrigger.MetaEpoch()
	header := sp.versionedHeaderFactory.Create(epoch)

	shardHeader, ok := header.(data.ShardHeaderHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	err := shardHeader.SetRound(round)
	if err != nil {
		return nil, err
	}

	err = shardHeader.SetNonce(nonce)
	if err != nil {
		return nil, err
	}

	err = sp.addPendingExecutionResults(shardHeader)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func (sp *shardProcessor) addPendingExecutionResults(shardHeader data.ShardHeaderHandler) error {
	if check.IfNil(shardHeader) {
		return process.ErrNilHeaderHandler
	}

	// todo: add pending execution results to the header

	return nil
}

// CreateNewHeader creates a new header
func (sp *shardProcessor) CreateNewHeader(round uint64, nonce uint64) (data.HeaderHandler, error) {
	epoch := sp.epochStartTrigger.MetaEpoch()
	header := sp.versionedHeaderFactory.Create(epoch)

	shardHeader, ok := header.(data.ShardHeaderHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	err := shardHeader.SetRound(round)
	if err != nil {
		return nil, err
	}

	sp.roundNotifier.CheckRound(header)
	sp.epochNotifier.CheckEpoch(header)

	err = shardHeader.SetNonce(nonce)
	if err != nil {
		return nil, err
	}

	err = shardHeader.SetAccumulatedFees(big.NewInt(0))
	if err != nil {
		return nil, err
	}

	err = shardHeader.SetDeveloperFees(big.NewInt(0))
	if err != nil {
		return nil, err
	}

	err = sp.setHeaderVersionData(shardHeader)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func (sp *shardProcessor) setHeaderVersionData(shardHeader data.ShardHeaderHandler) error {
	if check.IfNil(shardHeader) {
		return process.ErrNilHeaderHandler
	}

	rootHash, err := sp.accountsDB[state.UserAccountsState].RootHash()
	if err != nil {
		return err
	}

	scheduledGasAndFees := sp.scheduledTxsExecutionHandler.GetScheduledGasAndFees()
	additionalVersionData := &headerVersionData.AdditionalData{
		ScheduledRootHash:        rootHash,
		ScheduledAccumulatedFees: scheduledGasAndFees.AccumulatedFees,
		ScheduledDeveloperFees:   scheduledGasAndFees.DeveloperFees,
		ScheduledGasProvided:     scheduledGasAndFees.GasProvided,
		ScheduledGasPenalized:    scheduledGasAndFees.GasPenalized,
		ScheduledGasRefunded:     scheduledGasAndFees.GasRefunded,
	}

	return shardHeader.SetAdditionalData(additionalVersionData)
}

// getHighestHdrForOwnShardFromMetachain calculates the highest shard header notarized by metachain
func (sp *shardProcessor) getHighestHdrForOwnShardFromMetachain(
	processedHdrs []data.HeaderHandler,
) ([]data.HeaderHandler, [][]byte, error) {

	ownShIdHdrs := make([]data.HeaderHandler, 0, len(processedHdrs))

	for i := 0; i < len(processedHdrs); i++ {
		hdr, ok := processedHdrs[i].(*block.MetaBlock)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}

		hdrs := sp.getHighestHdrForShardFromMetachain(sp.shardCoordinator.SelfId(), hdr)
		ownShIdHdrs = append(ownShIdHdrs, hdrs...)
	}

	process.SortHeadersByNonce(ownShIdHdrs)

	ownShIdHdrsHashes := make([][]byte, len(ownShIdHdrs))
	for i := 0; i < len(ownShIdHdrs); i++ {
		hash, _ := core.CalculateHash(sp.marshalizer, sp.hasher, ownShIdHdrs[i])
		ownShIdHdrsHashes[i] = hash
	}

	return ownShIdHdrs, ownShIdHdrsHashes, nil
}

func (sp *shardProcessor) getHighestHdrForShardFromMetachain(shardId uint32, hdr *block.MetaBlock) []data.HeaderHandler {
	ownShIdHdr := make([]data.HeaderHandler, 0, len(hdr.ShardInfo))

	for _, shardInfo := range hdr.ShardInfo {
		if shardInfo.ShardID != shardId {
			continue
		}

		ownHdr, err := process.GetShardHeader(shardInfo.HeaderHash, sp.dataPool.Headers(), sp.marshalizer, sp.store)
		if err != nil {
			go sp.requestHandler.RequestShardHeader(shardInfo.ShardID, shardInfo.HeaderHash)

			log.Debug("requested missing shard header",
				"hash", shardInfo.HeaderHash,
				"shard", shardInfo.ShardID,
			)
			continue
		}

		ownShIdHdr = append(ownShIdHdr, ownHdr)
	}

	return data.TrimHeaderHandlerSlice(ownShIdHdr)
}

// getOrderedProcessedMetaBlocksFromHeader returns all the meta blocks fully processed
func (sp *shardProcessor) getOrderedProcessedMetaBlocksFromHeader(header data.HeaderHandler) ([]data.HeaderHandler, error) {
	if check.IfNil(header) {
		return nil, process.ErrNilBlockHeader
	}

	miniBlockHeaders := header.GetMiniBlockHeaderHandlers()
	miniBlockHashes := make(map[int][]byte, len(miniBlockHeaders))
	for i := 0; i < len(miniBlockHeaders); i++ {
		miniBlockHashes[i] = miniBlockHeaders[i].GetHash()
	}

	log.Trace("cross mini blocks in body",
		"num miniblocks", len(miniBlockHashes),
	)

	processedMetaBlocks, err := sp.getOrderedProcessedMetaBlocksFromMiniBlockHashes(miniBlockHeaders, miniBlockHashes)
	if err != nil {
		return nil, err
	}

	return processedMetaBlocks, nil
}

func (sp *shardProcessor) addProcessedCrossMiniBlocksFromHeader(headerHandler data.HeaderHandler) error {
	if check.IfNil(headerHandler) {
		return process.ErrNilBlockHeader
	}

	shardHeader, ok := headerHandler.(data.ShardHeaderHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}
	miniBlockHashes := make(map[int][]byte, len(headerHandler.GetMiniBlockHeaderHandlers()))
	for i := 0; i < len(headerHandler.GetMiniBlockHeaderHandlers()); i++ {
		miniBlockHashes[i] = headerHandler.GetMiniBlockHeaderHandlers()[i].GetHash()
	}

	for _, metaBlockHash := range shardHeader.GetMetaBlockHashes() {
		headerInfo, found := sp.hdrsForCurrBlock.GetHeaderInfo(string(metaBlockHash))
		if !found {
			return fmt.Errorf("%w : addProcessedCrossMiniBlocksFromHeader metaBlockHash = %s",
				process.ErrMissingHeader, logger.DisplayByteSlice(metaBlockHash))
		}

		metaBlock, isMetaBlock := headerInfo.GetHeader().(*block.MetaBlock)
		if !isMetaBlock {
			return process.ErrWrongTypeAssertion
		}

		crossMiniBlockHashes := metaBlock.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
		for key, miniBlockHash := range miniBlockHashes {
			_, ok = crossMiniBlockHashes[string(miniBlockHash)]
			if !ok {
				continue
			}

			miniBlockHeader := process.GetMiniBlockHeaderWithHash(headerHandler, miniBlockHash)
			if miniBlockHeader == nil {
				log.Warn("shardProcessor.addProcessedCrossMiniBlocksFromHeader: GetMiniBlockHeaderWithHash", "mb hash", miniBlockHash, "error", process.ErrMissingMiniBlockHeader)
				continue
			}

			sp.processedMiniBlocksTracker.SetProcessedMiniBlockInfo(metaBlockHash, miniBlockHash, &processedMb.ProcessedMiniBlockInfo{
				FullyProcessed:         miniBlockHeader.IsFinal(),
				IndexOfLastTxProcessed: miniBlockHeader.GetIndexOfLastTxProcessed(),
			})

			delete(miniBlockHashes, key)
		}
	}

	return nil
}

func (sp *shardProcessor) getOrderedProcessedMetaBlocksFromMiniBlockHashes(
	miniBlockHeaders []data.MiniBlockHeaderHandler,
	miniBlockHashes map[int][]byte,
) ([]data.HeaderHandler, error) {
	headersInfo := sp.hdrsForCurrBlock.GetHeadersInfoMap()
	processedMetaHdrs := make([]data.HeaderHandler, 0, len(headersInfo))
	processedCrossMiniBlocksHashes := make(map[string]bool, len(headersInfo))

	for metaBlockHash, headerInfo := range headersInfo {
		if !headerInfo.UsedInBlock() {
			continue
		}

		metaBlock, ok := headerInfo.GetHeader().(*block.MetaBlock)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		log.Trace("meta header",
			"nonce", metaBlock.Nonce,
		)

		crossMiniBlockHashes := metaBlock.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
		for hash := range crossMiniBlockHashes {
			processedCrossMiniBlocksHashes[hash] = sp.processedMiniBlocksTracker.IsMiniBlockFullyProcessed([]byte(metaBlockHash), []byte(hash))
		}

		for key, miniBlockHash := range miniBlockHashes {
			_, ok = crossMiniBlockHashes[string(miniBlockHash)]
			if !ok {
				continue
			}

			processedCrossMiniBlocksHashes[string(miniBlockHash)] = miniBlockHeaders[key].IsFinal()

			delete(miniBlockHashes, key)
		}

		log.Trace("cross mini blocks in meta header",
			"num miniblocks", len(crossMiniBlockHashes),
		)

		processedAll := true
		for hash := range crossMiniBlockHashes {
			if !processedCrossMiniBlocksHashes[hash] {
				processedAll = false
				break
			}
		}

		if processedAll {
			processedMetaHdrs = append(processedMetaHdrs, metaBlock)
		}
	}

	process.SortHeadersByNonce(processedMetaHdrs)

	return processedMetaHdrs, nil
}

func (sp *shardProcessor) updateCrossShardInfo(processedMetaHdrs []data.HeaderHandler) error {
	lastCrossNotarizedHeader, _, err := sp.blockTracker.GetLastCrossNotarizedHeader(core.MetachainShardId)
	if err != nil {
		return err
	}

	// processedMetaHdrs is also sorted
	for i := 0; i < len(processedMetaHdrs); i++ {
		hdr := processedMetaHdrs[i]

		// remove process finished
		if hdr.GetNonce() > lastCrossNotarizedHeader.GetNonce() {
			continue
		}

		// metablock was processed and finalized
		marshalizedHeader, errMarshal := sp.marshalizer.Marshal(hdr)
		if errMarshal != nil {
			log.Debug("updateCrossShardInfo.Marshal", "error", errMarshal.Error())
			continue
		}

		headerHash := sp.hasher.Compute(string(marshalizedHeader))

		sp.saveMetaHeader(hdr, headerHash, marshalizedHeader)

		sp.processedMiniBlocksTracker.RemoveMetaBlockHash(headerHash)
	}

	return nil
}

func (sp *shardProcessor) verifyCrossShardMiniBlockDstMe(header data.ShardHeaderHandler) error {
	miniBlockMetaHashes, err := sp.getAllMiniBlockDstMeFromMeta(header)
	if err != nil {
		return err
	}

	crossMiniBlockHashes := header.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
	for hash := range crossMiniBlockHashes {
		if _, ok := miniBlockMetaHashes[hash]; !ok {
			return process.ErrCrossShardMBWithoutConfirmationFromMeta
		}
	}

	return nil
}

func (sp *shardProcessor) getAllMiniBlockDstMeFromMeta(header data.ShardHeaderHandler) (map[string][]byte, error) {
	lastCrossNotarizedHeader, _, err := sp.blockTracker.GetLastCrossNotarizedHeader(core.MetachainShardId)
	if err != nil {
		return nil, err
	}

	miniBlockMetaHashes := make(map[string][]byte)

	for _, metaBlockHash := range header.GetMetaBlockHashes() {
		headerInfo, ok := sp.hdrsForCurrBlock.GetHeaderInfo(string(metaBlockHash))
		if !ok {
			continue
		}
		metaBlock, ok := headerInfo.GetHeader().(*block.MetaBlock)
		if !ok {
			continue
		}
		if metaBlock.GetRound() > header.GetRound() {
			continue
		}
		if metaBlock.GetRound() <= lastCrossNotarizedHeader.GetRound() {
			continue
		}
		if metaBlock.GetNonce() <= lastCrossNotarizedHeader.GetNonce() {
			continue
		}

		crossMiniBlockHashes := metaBlock.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
		for hash := range crossMiniBlockHashes {
			miniBlockMetaHashes[hash] = metaBlockHash
		}
	}

	return miniBlockMetaHashes, nil
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

// full verification through metachain header
func (sp *shardProcessor) createAndProcessMiniBlocksDstMe(haveTime func() bool) (*createAndProcessMiniBlocksDestMeInfo, error) {
	log.Debug("createAndProcessMiniBlocksDstMe has been started")

	sw := core.NewStopWatch()
	sw.Start("ComputeLongestMetaChainFromLastNotarized")
	orderedMetaBlocks, orderedMetaBlocksHashes, err := sp.blockTracker.ComputeLongestMetaChainFromLastNotarized()
	sw.Stop("ComputeLongestMetaChainFromLastNotarized")
	log.Debug("measurements", sw.GetMeasurements()...)
	if err != nil {
		return nil, err
	}

	log.Debug("metablocks ordered",
		"num metablocks", len(orderedMetaBlocks),
	)

	lastMetaHdr, _, err := sp.blockTracker.GetLastCrossNotarizedHeader(core.MetachainShardId)
	if err != nil {
		return nil, err
	}

	haveAdditionalTimeFalse := func() bool {
		return false
	}

	createAndProcessInfo := &createAndProcessMiniBlocksDestMeInfo{
		haveTime:                   haveTime,
		haveAdditionalTime:         haveAdditionalTimeFalse,
		miniBlocks:                 make(block.MiniBlockSlice, 0),
		allProcessedMiniBlocksInfo: make(map[string]*processedMb.ProcessedMiniBlockInfo),
		numTxsAdded:                uint32(0),
		numHdrsAdded:               uint32(0),
		scheduledMode:              false,
	}

	// do processing in order
	for i := 0; i < len(orderedMetaBlocks); i++ {
		if !createAndProcessInfo.haveTime() && !createAndProcessInfo.haveAdditionalTime() {
			log.Debug("time is up after putting cross txs with destination to current shard",
				"scheduled mode", createAndProcessInfo.scheduledMode,
				"num txs added", createAndProcessInfo.numTxsAdded,
			)
			break
		}

		if createAndProcessInfo.numHdrsAdded >= process.MaxMetaHeadersAllowedInOneShardBlock {
			log.Debug("maximum meta headers allowed to be included in one shard block has been reached",
				"scheduled mode", createAndProcessInfo.scheduledMode,
				"meta headers added", createAndProcessInfo.numHdrsAdded,
			)
			break
		}

		createAndProcessInfo.currMetaHdr = orderedMetaBlocks[i]
		if createAndProcessInfo.currMetaHdr.GetNonce() > lastMetaHdr.GetNonce()+1 {
			log.Debug("skip searching",
				"scheduled mode", createAndProcessInfo.scheduledMode,
				"last meta hdr nonce", lastMetaHdr.GetNonce(),
				"curr meta hdr nonce", createAndProcessInfo.currMetaHdr.GetNonce())
			break
		}

		hasProofForHdr := sp.proofsPool.HasProof(core.MetachainShardId, orderedMetaBlocksHashes[i])
		shouldConsiderProofsForNotarization := sp.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, orderedMetaBlocks[i].GetEpoch())
		if !hasProofForHdr && shouldConsiderProofsForNotarization {
			log.Trace("no proof for meta header",
				"hash", logger.DisplayByteSlice(orderedMetaBlocksHashes[i]),
			)
			break
		}

		createAndProcessInfo.currMetaHdrHash = orderedMetaBlocksHashes[i]
		if len(createAndProcessInfo.currMetaHdr.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())) == 0 {
			sp.hdrsForCurrBlock.AddHeaderUsedInBlock(string(createAndProcessInfo.currMetaHdrHash), createAndProcessInfo.currMetaHdr)
			createAndProcessInfo.numHdrsAdded++
			lastMetaHdr = createAndProcessInfo.currMetaHdr
			continue
		}

		createAndProcessInfo.currProcessedMiniBlocksInfo = sp.processedMiniBlocksTracker.GetProcessedMiniBlocksInfo(createAndProcessInfo.currMetaHdrHash)
		createAndProcessInfo.hdrAdded = false

		shouldContinue, errCreated := sp.createMbsAndProcessCrossShardTransactionsDstMe(createAndProcessInfo)
		if errCreated != nil {
			return nil, errCreated
		}
		if !shouldContinue {
			break
		}

		lastMetaHdr = createAndProcessInfo.currMetaHdr
	}

	go sp.requestMetaHeadersIfNeeded(createAndProcessInfo.numHdrsAdded, lastMetaHdr)

	for _, miniBlock := range createAndProcessInfo.miniBlocks {
		log.Debug("mini block info",
			"type", miniBlock.Type,
			"sender shard", miniBlock.SenderShardID,
			"receiver shard", miniBlock.ReceiverShardID,
			"txs added", len(miniBlock.TxHashes))
	}

	log.Debug("createAndProcessMiniBlocksDstMe has been finished",
		"num txs added", createAndProcessInfo.numTxsAdded,
		"num hdrs added", createAndProcessInfo.numHdrsAdded)

	return createAndProcessInfo, nil
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

func (sp *shardProcessor) createMbsAndProcessCrossShardTransactionsDstMe(
	createAndProcessInfo *createAndProcessMiniBlocksDestMeInfo,
) (bool, error) {
	currMiniBlocksAdded, currNumTxsAdded, hdrProcessFinished, errCreated := sp.txCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe(
		createAndProcessInfo.currMetaHdr,
		createAndProcessInfo.currProcessedMiniBlocksInfo,
		createAndProcessInfo.haveTime,
		createAndProcessInfo.haveAdditionalTime,
		createAndProcessInfo.scheduledMode)
	if errCreated != nil {
		return false, errCreated
	}

	for miniBlockHash, processedMiniBlockInfo := range createAndProcessInfo.currProcessedMiniBlocksInfo {
		createAndProcessInfo.allProcessedMiniBlocksInfo[miniBlockHash] = &processedMb.ProcessedMiniBlockInfo{
			FullyProcessed:         processedMiniBlockInfo.FullyProcessed,
			IndexOfLastTxProcessed: processedMiniBlockInfo.IndexOfLastTxProcessed,
		}
	}

	// all txs processed, add to processed miniblocks
	createAndProcessInfo.miniBlocks = append(createAndProcessInfo.miniBlocks, currMiniBlocksAdded...)
	createAndProcessInfo.numTxsAdded += currNumTxsAdded

	if !createAndProcessInfo.hdrAdded && currNumTxsAdded > 0 {
		sp.hdrsForCurrBlock.AddHeaderUsedInBlock(string(createAndProcessInfo.currMetaHdrHash), createAndProcessInfo.currMetaHdr)
		createAndProcessInfo.numHdrsAdded++
		createAndProcessInfo.hdrAdded = true
	}

	if !hdrProcessFinished {
		log.Debug("meta block cannot be fully processed",
			"scheduled mode", createAndProcessInfo.scheduledMode,
			"round", createAndProcessInfo.currMetaHdr.GetRound(),
			"nonce", createAndProcessInfo.currMetaHdr.GetNonce(),
			"hash", createAndProcessInfo.currMetaHdrHash,
			"num mbs added", len(currMiniBlocksAdded),
			"num txs added", currNumTxsAdded)

		if sp.enableEpochsHandler.IsFlagEnabled(common.ScheduledMiniBlocksFlag) && !createAndProcessInfo.scheduledMode {
			createAndProcessInfo.scheduledMode = true
			createAndProcessInfo.haveAdditionalTime = process.HaveAdditionalTime()
			return sp.createMbsAndProcessCrossShardTransactionsDstMe(createAndProcessInfo)
		}

		return false, nil
	}

	return true, nil
}

func (sp *shardProcessor) requestMetaHeadersIfNeeded(hdrsAdded uint32, lastMetaHdr data.HeaderHandler) {
	log.Debug("meta headers added",
		"num", hdrsAdded,
		"highest nonce", lastMetaHdr.GetNonce(),
	)

	roundTooOld := sp.roundHandler.Index() > int64(lastMetaHdr.GetRound()+process.MaxRoundsWithoutNewBlockReceived)
	shouldRequestCrossHeaders := hdrsAdded == 0 && roundTooOld
	if shouldRequestCrossHeaders {
		fromNonce := lastMetaHdr.GetNonce() + 1
		toNonce := fromNonce + uint64(sp.metaBlockFinality)
		for nonce := fromNonce; nonce <= toNonce; nonce++ {
			sp.addHeaderIntoTrackerPool(nonce, core.MetachainShardId)
			sp.requestHandler.RequestMetaHeaderByNonce(nonce)
		}
	}
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

func (sp *shardProcessor) createMiniBlocks(haveTime func() bool, randomness []byte) (*block.Body, map[string]*processedMb.ProcessedMiniBlockInfo, error) {
	var miniBlocks block.MiniBlockSlice
	processedMiniBlocksDestMeInfo := make(map[string]*processedMb.ProcessedMiniBlockInfo)

	if sp.enableEpochsHandler.IsFlagEnabled(common.ScheduledMiniBlocksFlag) {
		miniBlocks = sp.scheduledTxsExecutionHandler.GetScheduledMiniBlocks()
		sp.txCoordinator.AddTxsFromMiniBlocks(miniBlocks)

		scheduledIntermediateTxs := sp.scheduledTxsExecutionHandler.GetScheduledIntermediateTxs()
		sp.txCoordinator.AddTransactions(scheduledIntermediateTxs[block.InvalidBlock], block.TxBlock)
	}

	// placeholder for shardProcessor.createMiniBlocks script

	if sp.accountsDB[state.UserAccountsState].JournalLen() != 0 {
		log.Error("shardProcessor.createMiniBlocks",
			"error", process.ErrAccountStateDirty,
			"stack", string(sp.accountsDB[state.UserAccountsState].GetStackDebugFirstEntry()))

		interMBs := sp.txCoordinator.CreatePostProcessMiniBlocks()
		if len(interMBs) > 0 {
			miniBlocks = append(miniBlocks, interMBs...)
		}

		log.Debug("creating mini blocks has been finished", "num miniblocks", len(miniBlocks))
		return &block.Body{MiniBlocks: miniBlocks}, processedMiniBlocksDestMeInfo, nil
	}

	if !haveTime() {
		log.Debug("shardProcessor.createMiniBlocks", "error", process.ErrTimeIsOut)

		interMBs := sp.txCoordinator.CreatePostProcessMiniBlocks()
		if len(interMBs) > 0 {
			miniBlocks = append(miniBlocks, interMBs...)
		}

		log.Debug("creating mini blocks has been finished", "num miniblocks", len(miniBlocks))
		return &block.Body{MiniBlocks: miniBlocks}, processedMiniBlocksDestMeInfo, nil
	}

	startTime := time.Now()
	createAndProcessMBsDestMeInfo, err := sp.createAndProcessMiniBlocksDstMe(haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to create mbs to me", "time", elapsedTime)
	if err != nil {
		log.Debug("createAndProcessCrossMiniBlocksDstMe", "error", err.Error())
	}
	if createAndProcessMBsDestMeInfo != nil {
		processedMiniBlocksDestMeInfo = createAndProcessMBsDestMeInfo.allProcessedMiniBlocksInfo
		if len(createAndProcessMBsDestMeInfo.miniBlocks) > 0 {
			miniBlocks = append(miniBlocks, createAndProcessMBsDestMeInfo.miniBlocks...)

			log.Debug("processed miniblocks and txs with destination in self shard",
				"num miniblocks", len(createAndProcessMBsDestMeInfo.miniBlocks),
				"num txs", createAndProcessMBsDestMeInfo.numTxsAdded,
				"num meta headers", createAndProcessMBsDestMeInfo.numHdrsAdded)
		}
	}

	if sp.blockTracker.IsShardStuck(core.MetachainShardId) {
		log.Warn("shardProcessor.createMiniBlocks",
			"error", process.ErrShardIsStuck,
			"shard", core.MetachainShardId)

		interMBs := sp.txCoordinator.CreatePostProcessMiniBlocks()
		if len(interMBs) > 0 {
			miniBlocks = append(miniBlocks, interMBs...)
		}

		log.Debug("creating mini blocks has been finished", "num miniblocks", len(miniBlocks))
		return &block.Body{MiniBlocks: miniBlocks}, processedMiniBlocksDestMeInfo, nil
	}

	startTime = time.Now()
	mbsFromMe := sp.txCoordinator.CreateMbsAndProcessTransactionsFromMe(haveTime, randomness)
	elapsedTime = time.Since(startTime)
	log.Debug("elapsed time to create mbs from me", "time", elapsedTime)

	if len(mbsFromMe) > 0 {
		miniBlocks = append(miniBlocks, mbsFromMe...)

		numTxs := 0
		for _, mb := range mbsFromMe {
			numTxs += len(mb.TxHashes)
		}

		log.Debug("processed miniblocks and txs from self shard",
			"num miniblocks", len(mbsFromMe),
			"num txs", numTxs)
	}

	log.Debug("creating mini blocks has been finished", "num miniblocks", len(miniBlocks))
	return &block.Body{MiniBlocks: miniBlocks}, processedMiniBlocksDestMeInfo, nil
}

// applyBodyToHeader creates a miniblock header list given a block body
func (sp *shardProcessor) applyBodyToHeader(
	shardHeader data.ShardHeaderHandler,
	body *block.Body,
	processedMiniBlocksDestMeInfo map[string]*processedMb.ProcessedMiniBlockInfo,
) (*block.Body, error) {
	sw := core.NewStopWatch()
	sw.Start("applyBodyToHeader")
	defer func() {
		sw.Stop("applyBodyToHeader")
		log.Debug("measurements", sw.GetMeasurements()...)
	}()

	var err error
	err = shardHeader.SetMiniBlockHeaderHandlers(nil)
	if err != nil {
		return nil, err
	}

	err = shardHeader.SetRootHash(sp.getRootHash())
	if err != nil {
		return nil, err
	}

	defer func() {
		go sp.checkAndRequestIfMetaHeadersMissing()
	}()

	if check.IfNil(body) {
		return nil, process.ErrNilBlockBody
	}

	var receiptsHash []byte
	sw.Start("CreateReceiptsHash")
	receiptsHash, err = sp.txCoordinator.CreateReceiptsHash()
	sw.Stop("CreateReceiptsHash")
	if err != nil {
		return nil, err
	}

	err = shardHeader.SetReceiptsHash(receiptsHash)
	if err != nil {
		return nil, err
	}

	newBody := deleteSelfReceiptsMiniBlocks(body)

	sw.Start("createMiniBlockHeaders")
	totalTxCount, miniBlockHeaderHandlers, err := sp.createMiniBlockHeaderHandlers(newBody, processedMiniBlocksDestMeInfo)
	sw.Stop("createMiniBlockHeaders")
	if err != nil {
		return nil, err
	}

	err = shardHeader.SetMiniBlockHeaderHandlers(miniBlockHeaderHandlers)
	if err != nil {
		return nil, err
	}

	err = shardHeader.SetTxCount(uint32(totalTxCount))
	if err != nil {
		return nil, err
	}

	err = shardHeader.SetAccumulatedFees(sp.feeHandler.GetAccumulatedFees())
	if err != nil {
		return nil, err
	}

	err = shardHeader.SetDeveloperFees(sp.feeHandler.GetDeveloperFees())
	if err != nil {
		return nil, err
	}

	sw.Start("sortHeaderHashesForCurrentBlockByNonce")
	metaBlockHashes, err := sp.sortHeaderHashesForCurrentBlockByNonce(true)
	sw.Stop("sortHeaderHashesForCurrentBlockByNonce")
	if err != nil {
		return nil, err
	}

	err = shardHeader.SetMetaBlockHashes(metaBlockHashes[core.MetachainShardId])
	if err != nil {
		return nil, err
	}

	err = sp.txCoordinator.VerifyCreatedMiniBlocks(shardHeader, newBody)
	if err != nil {
		return nil, err
	}

	sp.appStatusHandler.SetUInt64Value(common.MetricNumTxInBlock, uint64(totalTxCount))
	sp.appStatusHandler.SetUInt64Value(common.MetricNumMiniBlocks, uint64(len(body.MiniBlocks)))

	marshalizedBody, err := sp.marshalizer.Marshal(newBody)
	if err != nil {
		return nil, err
	}
	sp.blockSizeThrottler.Add(shardHeader.GetRound(), uint32(len(marshalizedBody)))

	return newBody, nil
}

// MarshalizedDataToBroadcast prepares underlying data into a marshalized object according to destination
func (sp *shardProcessor) MarshalizedDataToBroadcast(
	header data.HeaderHandler,
	bodyHandler data.BodyHandler,
) (map[uint32][]byte, map[string][][]byte, error) {

	if check.IfNil(bodyHandler) {
		return nil, nil, process.ErrNilMiniBlocks
	}

	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	// Remove mini blocks which are not final from "body" to avoid sending them cross shard
	newBodyToBroadcast, err := sp.getFinalMiniBlocks(header, body)
	if err != nil {
		return nil, nil, err
	}

	mrsTxs := sp.txCoordinator.CreateMarshalizedData(newBodyToBroadcast)

	bodies := make(map[uint32]block.MiniBlockSlice)
	for _, miniBlock := range newBodyToBroadcast.MiniBlocks {
		if miniBlock.SenderShardID != sp.shardCoordinator.SelfId() ||
			miniBlock.ReceiverShardID == sp.shardCoordinator.SelfId() {
			continue
		}
		bodies[miniBlock.ReceiverShardID] = append(bodies[miniBlock.ReceiverShardID], miniBlock)
	}

	mrsData := make(map[uint32][]byte, len(bodies))
	for shardId, subsetBlockBody := range bodies {
		bodyForShard := block.Body{MiniBlocks: subsetBlockBody}
		buff, errMarshal := sp.marshalizer.Marshal(&bodyForShard)
		if errMarshal != nil {
			log.Error("shardProcessor.MarshalizedDataToBroadcast.Marshal", "error", errMarshal.Error())
			continue
		}
		mrsData[shardId] = buff
	}

	return mrsData, mrsTxs, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sp *shardProcessor) IsInterfaceNil() bool {
	return sp == nil
}

// GetBlockBodyFromPool returns block body from pool for a given header
func (sp *shardProcessor) GetBlockBodyFromPool(headerHandler data.HeaderHandler) (data.BodyHandler, error) {
	header, ok := headerHandler.(data.ShardHeaderHandler)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	miniBlocksPool := sp.dataPool.MiniBlocks()
	var miniBlocks block.MiniBlockSlice

	for _, mbHeader := range header.GetMiniBlockHeaderHandlers() {
		obj, hashInPool := miniBlocksPool.Get(mbHeader.GetHash())
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

func (sp *shardProcessor) getBootstrapHeadersInfo(
	selfNotarizedHeaders []data.HeaderHandler,
	selfNotarizedHeadersHashes [][]byte,
) []bootstrapStorage.BootstrapHeaderInfo {

	numSelfNotarizedHeaders := len(selfNotarizedHeaders)

	highestNonceInSelfNotarizedHeaders := uint64(0)
	if numSelfNotarizedHeaders > 0 {
		highestNonceInSelfNotarizedHeaders = selfNotarizedHeaders[numSelfNotarizedHeaders-1].GetNonce()
	}

	isFinalNonceHigherThanSelfNotarized := sp.forkDetector.GetHighestFinalBlockNonce() > highestNonceInSelfNotarizedHeaders
	if isFinalNonceHigherThanSelfNotarized {
		numSelfNotarizedHeaders++
	}

	if numSelfNotarizedHeaders == 0 {
		return nil
	}

	lastSelfNotarizedHeaders := make([]bootstrapStorage.BootstrapHeaderInfo, 0, numSelfNotarizedHeaders)

	for index := range selfNotarizedHeaders {
		headerInfo := bootstrapStorage.BootstrapHeaderInfo{
			ShardId: selfNotarizedHeaders[index].GetShardID(),
			Nonce:   selfNotarizedHeaders[index].GetNonce(),
			Hash:    selfNotarizedHeadersHashes[index],
		}

		lastSelfNotarizedHeaders = append(lastSelfNotarizedHeaders, headerInfo)
	}

	if isFinalNonceHigherThanSelfNotarized {
		headerInfo := bootstrapStorage.BootstrapHeaderInfo{
			ShardId: sp.shardCoordinator.SelfId(),
			Nonce:   sp.forkDetector.GetHighestFinalBlockNonce(),
			Hash:    sp.forkDetector.GetHighestFinalBlockHash(),
		}

		lastSelfNotarizedHeaders = append(lastSelfNotarizedHeaders, headerInfo)
	}

	return lastSelfNotarizedHeaders
}

func (sp *shardProcessor) removeStartOfEpochBlockDataFromPools(
	_ data.HeaderHandler,
	_ data.BodyHandler,
) error {
	return nil
}

// Close - closes all underlying components
func (sp *shardProcessor) Close() error {
	return sp.baseProcessor.Close()
}

// DecodeBlockHeader method decodes block header from a given byte array
func (sp *shardProcessor) DecodeBlockHeader(dta []byte) data.HeaderHandler {
	if dta == nil {
		return nil
	}

	header, err := process.UnmarshalShardHeader(sp.marshalizer, dta)
	if err != nil {
		log.Debug("DecodeBlockHeader.UnmarshalShardHeader", "error", err.Error())
		return nil
	}

	return header
}
